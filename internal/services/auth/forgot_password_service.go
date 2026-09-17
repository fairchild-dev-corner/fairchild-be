package services

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	cc "fairchild_be/internal/constants"
	models "fairchild_be/internal/models/auth"
	"fairchild_be/internal/services/mail"
	"fairchild_be/internal/services/sms"
	"fairchild_be/internal/utils"

	"github.com/gin-gonic/gin"
)

// forgotPasswordOTPExpiry is short on purpose - a leaked reset code is more
// dangerous than a leaked login code, since on its own it ends with a new
// temporary password known only to whoever completes the challenge.
const forgotPasswordOTPExpiry = 2 * time.Minute

const otpPurposePasswordReset = "password_reset"

// ForgotPasswordService validates that Email belongs to an account, then
// creates and sends a password-reset OTP challenge via SMS and email. The
// client must follow up with VerifyForgotPasswordOTPService using
// ReferenceID to actually reset the password.
func (s *AuthService) ForgotPasswordService(ctx *gin.Context, req *models.ForgotPasswordRequest) (*models.ForgotPasswordOTPChallengeResponse, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, cc.ErrEmailNotAssociated
	}
	if user.MobileNumber == nil {
		return nil, cc.ErrFailedToSendOTP
	}

	// Stored mobile numbers aren't guaranteed to already be E.164 (e.g. rows
	// created before this normalization existed) - normalize at send time so
	// Movider, which requires E.164, doesn't reject an otherwise-valid number.
	mobileNumber, err := NormalizeMobileNumber(*user.MobileNumber)
	if err != nil {
		return nil, cc.ErrFailedToSendOTP
	}

	code, err := GenerateOTPCode()
	if err != nil {
		return nil, err
	}

	referenceID, err := GenerateOpaqueToken()
	if err != nil {
		return nil, err
	}

	var memberID string
	if user.MemberID != nil {
		memberID = *user.MemberID
	}

	challenge := &models.OTPVerification{
		ReferenceID:  referenceID,
		UserID:       &user.ID,
		MemberID:     memberID,
		OTPHash:      code,
		MobileNumber: mobileNumber,
		Purpose:      otpPurposePasswordReset,
		MaxAttempts:  otpMaxAttempts,
	}

	id, err := s.repo.CreateOTPVerification(ctx, challenge, int(forgotPasswordOTPExpiry.Minutes()))
	if err != nil {
		return nil, err
	}

	message := sms.ForgotPasswordOTPMessage(code, int(forgotPasswordOTPExpiry.Minutes()))
	if sendErr := s.smsSender.Send(ctx, mobileNumber, message); sendErr != nil {
		// Leave the row in place as 'failed' rather than deleting it - it's
		// unguessable (only the hash is stored) and useful for support to
		// see that delivery, not the user's input, is what failed.
		_ = s.repo.MarkOTPSendFailed(ctx, id, sendErr.Error())
		slog.Error("failed to send forgot-password OTP SMS", "user_id", user.ID, "otp_verification_id", id, "error", sendErr)
		return nil, fmt.Errorf("%w: %v", cc.ErrFailedToSendOTP, sendErr)
	}
	_ = s.repo.MarkOTPSent(ctx, id)

	// Email is a secondary delivery channel alongside SMS - SMS above is what
	// gates the challenge as usable (status='sent'), so any failure here
	// (bad address, render error, send error) is logged and swallowed rather
	// than failing the request; the user can still complete it with the code
	// from SMS.
	if err := utils.Validate.Var(user.Email, "required,email"); err != nil {
		slog.Warn("skipping forgot-password OTP email: invalid or missing email on file", "user_id", user.ID)

	} else if emailBody, renderErr := mail.RenderOTPVerificationEmail(code, int(forgotPasswordOTPExpiry.Minutes())); renderErr != nil {
		slog.Warn("failed to render forgot-password OTP email", "user_id", user.ID, "error", renderErr)

	} else {
		// Backgrounded for the same reason as the login OTP email - see
		// createAndSendLoginOTP in auth_service.go.
		bgCtx := ctx.Copy()
		go func() {
			if sendErr := s.mailSender.Send(bgCtx, user.Email, mail.OTPVerificationSubject, emailBody); sendErr != nil {
				slog.Warn("failed to send forgot-password OTP email", "user_id", user.ID, "error", sendErr)
			}
		}()
	}

	return &models.ForgotPasswordOTPChallengeResponse{
		ReferenceID:  referenceID,
		ExpiresIn:    int(forgotPasswordOTPExpiry.Seconds()),
		MaskedMobile: maskMobileNumber(*user.MobileNumber),
	}, nil
}

// VerifyForgotPasswordOTPService confirms the code for a challenge created
// by ForgotPasswordService. It never distinguishes "wrong code once" from
// "attempts exhausted" from "unknown/expired reference" in its returned
// error - all collapse to ErrInvalidOrExpiredOTP, matching
// VerifyLoginOTPService's behavior. On success it only marks the challenge
// verified (it is NOT consumed here) - the caller must still follow up with
// ResetForgotPasswordService, using the same ReferenceID, to actually set a
// new password.
func (s *AuthService) VerifyForgotPasswordOTPService(ctx *gin.Context, req *models.VerifyOTPRequest) error {
	challenge, err := s.repo.FindActiveOTPVerificationByReference(ctx, req.ReferenceID)
	if err != nil {
		return err
	}
	if challenge == nil || challenge.Purpose != otpPurposePasswordReset {
		return cc.ErrInvalidOrExpiredOTP
	}

	if challenge.Attempts >= challenge.MaxAttempts || req.OTPCode != challenge.OTPHash {
		// IncrementOTPAttempts also auto-consumes once max_attempts is hit,
		// atomically against the row's live count - see its doc comment.
		_ = s.repo.IncrementOTPAttempts(ctx, challenge.ID)
		return cc.ErrInvalidOrExpiredOTP
	}

	return s.repo.MarkOTPVerified(ctx, challenge.ID)
}

// ResetForgotPasswordService is the final step of the forgot-password flow.
// ReferenceID must belong to a challenge already confirmed via
// VerifyForgotPasswordOTPService - it consumes that challenge (so it can
// never be reused) and sets NewPassword as the account's password.
func (s *AuthService) ResetForgotPasswordService(ctx *gin.Context, req *models.ResetPasswordRequest) error {
	challenge, err := s.repo.FindVerifiedOTPVerificationByReference(ctx, req.ReferenceID)
	if err != nil {
		return err
	}
	if challenge == nil || challenge.Purpose != otpPurposePasswordReset || challenge.UserID == nil {
		return cc.ErrInvalidOrExpiredOTP
	}

	// consumed=false means another concurrent request (same reference_id)
	// already redeemed this challenge first - it must not also reset the
	// password a second time.
	consumed, err := s.repo.ConsumeOTPVerification(ctx, challenge.ID)
	if err != nil {
		return err
	}
	if !consumed {
		return cc.ErrInvalidOrExpiredOTP
	}

	user, err := s.repo.FindUserByID(ctx, *challenge.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(cc.USER_NOT_FOUND)
	}

	hash, err := HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.repo.UpdateUserPassword(ctx, user.ID, hash)
}
