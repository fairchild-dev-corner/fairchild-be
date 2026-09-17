package services

import (
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

const (
	otpPurposeLogin    = "login"
	otpPurposeRegister = "register"

	youngSaverMaxAge = 18
)

// SendRegisterOTPService creates and sends a "register"-purpose OTP
// challenge for mobile-number verification before any user row exists. It's
// shared by both regular-member registration (mobile is the applicant's
// own) and Young Saver registration (mobile is the parent/guardian's) - the
// caller must follow up with VerifyRegisterOTPService, then RegisterService
// or RegisterYoungSaverService, all using the same ReferenceID.
func (s *AuthService) SendRegisterOTPService(ctx *gin.Context, req *models.SendRegisterOTPRequest) (*models.RegisterOTPChallengeResponse, error) {
	mobileNumber, err := NormalizeMobileNumber(req.MobileNumber)
	if err != nil {
		return nil, err
	}

	code, err := GenerateOTPCode()
	if err != nil {
		return nil, err
	}

	referenceID, err := GenerateOpaqueToken()
	if err != nil {
		return nil, err
	}

	challenge := &models.OTPVerification{
		ReferenceID:  referenceID,
		UserID:       nil,
		OTPHash:      code,
		MobileNumber: mobileNumber,
		Purpose:      otpPurposeRegister,
		MaxAttempts:  otpMaxAttempts,
	}

	id, err := s.repo.CreateOTPVerification(ctx, challenge, int(otpExpiry.Minutes()))
	if err != nil {
		return nil, err
	}

	message := sms.RegistrationOTPMessage(code, int(otpExpiry.Minutes()))
	if sendErr := s.smsSender.Send(ctx, mobileNumber, message); sendErr != nil {
		// Leave the row in place as 'failed' rather than deleting it - it's
		// unguessable (only the hash is stored) and useful for support to
		// see that delivery, not the user's input, is what failed.
		_ = s.repo.MarkOTPSendFailed(ctx, id, sendErr.Error())
		slog.Error("failed to send registration OTP SMS", "otp_verification_id", id, "error", sendErr)
		return nil, fmt.Errorf("%w: %v", cc.ErrFailedToSendOTP, sendErr)
	}
	_ = s.repo.MarkOTPSent(ctx, id)

	// Email is optional here (Young Saver's guardian doesn't supply one, and
	// a regular member's own email is collected but not required to be
	// present at this step) - best-effort secondary channel, same tolerance
	// as the login/forgot-password OTP flows.
	if req.Email != "" {
		if err := utils.Validate.Var(req.Email, "required,email"); err != nil {
			slog.Warn("skipping registration OTP email: invalid email", "error", err)
		} else if emailBody, renderErr := mail.RenderOTPVerificationEmail(code, int(otpExpiry.Minutes())); renderErr != nil {
			slog.Warn("failed to render registration OTP email", "error", renderErr)
		} else {
			// Backgrounded for the same reason as the login OTP email - see
			// createAndSendLoginOTP in auth_service.go.
			bgCtx := ctx.Copy()
			go func() {
				if sendErr := s.mailSender.Send(bgCtx, req.Email, mail.OTPVerificationSubject, emailBody); sendErr != nil {
					slog.Warn("failed to send registration OTP email", "error", sendErr)
				}
			}()
		}
	}

	return &models.RegisterOTPChallengeResponse{
		ReferenceID:  referenceID,
		ExpiresIn:    int(otpExpiry.Seconds()),
		MaskedMobile: maskMobileNumber(mobileNumber),
	}, nil
}

// VerifyRegisterOTPService confirms the code for a challenge created by
// SendRegisterOTPService, mirroring VerifyForgotPasswordOTPService's
// verify-without-consume split: the challenge is only marked verified here,
// not consumed - the caller must still follow up with RegisterService or
// RegisterYoungSaverService, using the same ReferenceID, to actually create
// the account.
func (s *AuthService) VerifyRegisterOTPService(ctx *gin.Context, req *models.VerifyOTPRequest) error {
	challenge, err := s.repo.FindActiveOTPVerificationByReference(ctx, req.ReferenceID)
	if err != nil {
		return err
	}
	if challenge == nil || challenge.Purpose != otpPurposeRegister {
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

// RegisterYoungSaverService enrolls a dependent under 18 as a Young Saver.
// ReferenceID must belong to a "register"-purpose OTP challenge already
// confirmed via VerifyRegisterOTPService, and that challenge must have been
// sent to req.GuardianMobileNumber - verifying it is what actually proves
// guardian control of that number, req.GuardianConsent is just the
// affirmative checkbox on top of that. The resulting account starts as
// UserStatusPendingReview until the membership team reviews it - it can
// still log in normally (see UserStatusPendingReview's doc comment), but
// this endpoint itself doesn't issue a session: registration and login are
// kept as separate steps here, same as RegisterService.
func (s *AuthService) RegisterYoungSaverService(ctx *gin.Context, req *models.RegisterYoungSaverRequest) (*models.UserResponse, error) {
	minor, err := isDateOfBirthUnder(req.DateOfBirth, youngSaverMaxAge)
	if err != nil {
		return nil, err
	}
	if !minor {
		return nil, cc.ErrYoungSaverMustBeMinor
	}

	challenge, err := s.repo.FindVerifiedOTPVerificationByReference(ctx, req.ReferenceID)
	if err != nil {
		return nil, err
	}
	if challenge == nil || challenge.Purpose != otpPurposeRegister {
		return nil, cc.ErrInvalidOrExpiredOTP
	}

	guardianMobile, err := NormalizeMobileNumber(req.GuardianMobileNumber)
	if err != nil {
		return nil, err
	}
	if guardianMobile != challenge.MobileNumber {
		return nil, cc.ErrInvalidOrExpiredOTP
	}

	if req.Email != "" {
		existing, err := s.repo.FindUserByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, cc.ErrEmailAlreadyRegistered
		}
	}

	existingMember, err := s.repo.FindUserByMemberID(ctx, req.MemberID)
	if err != nil {
		return nil, err
	}
	if existingMember != nil {
		return nil, cc.ErrMemberIDAlreadyRegistered
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var mobileNumber string
	if req.MobileNumber != "" {
		mobileNumber, err = NormalizeMobileNumber(req.MobileNumber)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now()
	user := &models.User{
		Email:                req.Email,
		PasswordHash:         &hash,
		Status:               models.UserStatusPendingReview,
		MemberType:           models.MemberTypeYoungSaver,
		FirstName:            nullIfEmptyStr(req.FirstName),
		MiddleName:           nullIfEmptyStr(req.MiddleName),
		LastName:             nullIfEmptyStr(req.LastName),
		Suffix:               nullIfEmptyStr(req.Suffix),
		MemberID:             nullIfEmptyStr(req.MemberID),
		DateOfBirth:          nullIfEmptyStr(req.DateOfBirth),
		Gender:               nullIfEmptyStr(req.Gender),
		MobileNumber:         nullIfEmptyStr(mobileNumber),
		GuardianFullName:     nullIfEmptyStr(req.GuardianFullName),
		GuardianMobileNumber: nullIfEmptyStr(guardianMobile),
		GuardianConsentAt:    &now,
	}

	id, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	user.ID = id

	consumed, err := s.repo.ConsumeOTPVerification(ctx, challenge.ID)
	if err != nil {
		return nil, err
	}
	if !consumed {
		return nil, cc.ErrInvalidOrExpiredOTP
	}

	return user.ToResponse(), nil
}

// isDateOfBirthUnder reports whether dateOfBirth (format "2006-01-02")
// describes someone younger than maxAge as of now.
func isDateOfBirthUnder(dateOfBirth string, maxAge int) (bool, error) {
	dob, err := time.Parse("2006-01-02", dateOfBirth)
	if err != nil {
		return false, err
	}

	now := time.Now()
	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}

	return age < maxAge, nil
}
