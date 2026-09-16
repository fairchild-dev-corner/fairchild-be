package services

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	cc "fairchild_be/internal/constants"
	"fairchild_be/internal/interfaces"
	models "fairchild_be/internal/models/auth"
	"fairchild_be/internal/services/mail"
	"fairchild_be/internal/services/sms"
	"fairchild_be/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthService struct {
	repo       interfaces.AuthRepositoryInterface
	tokens     *TokenService
	verifiers  map[string]SocialVerifier
	smsSender  sms.Sender
	mailSender mail.Sender
}

func NewAuthService(
	repo interfaces.AuthRepositoryInterface,
	tokens *TokenService,
	verifiers map[string]SocialVerifier,
	smsSender sms.Sender,
	mailSender mail.Sender,
) *AuthService {
	return &AuthService{repo: repo, tokens: tokens, verifiers: verifiers, smsSender: smsSender, mailSender: mailSender}
}

// RegisterService is the final step of regular-member registration.
// ReferenceID must belong to a "register"-purpose OTP challenge already
// confirmed via VerifyRegisterOTPService (see register_service.go), and
// MobileNumber must match the number that challenge was sent to - otherwise
// a verified OTP for one mobile number could be reused to register an
// account under a different one.
func (s *AuthService) RegisterService(ctx *gin.Context, req *models.RegisterRequest) (*models.AuthTokenResponse, error) {
	challenge, err := s.repo.FindVerifiedOTPVerificationByReference(ctx, req.ReferenceID)
	if err != nil {
		return nil, err
	}
	if challenge == nil || challenge.Purpose != otpPurposeRegister {
		return nil, cc.ErrInvalidOrExpiredOTP
	}

	mobileNumber, err := NormalizeMobileNumber(req.MobileNumber)
	if err != nil {
		return nil, err
	}
	if mobileNumber != challenge.MobileNumber {
		return nil, cc.ErrInvalidOrExpiredOTP
	}

	existing, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, cc.ErrEmailAlreadyRegistered
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

	user := &models.User{
		Email:        req.Email,
		PasswordHash: &hash,
		Status:       models.UserStatusActive,
		MemberType:   models.MemberTypeRegular,
		FirstName:    nullIfEmptyStr(req.FirstName),
		MiddleName:   nullIfEmptyStr(req.MiddleName),
		LastName:     nullIfEmptyStr(req.LastName),
		Suffix:       nullIfEmptyStr(req.Suffix),
		MemberID:     nullIfEmptyStr(req.MemberID),
		DateOfBirth:  nullIfEmptyStr(req.DateOfBirth),
		Gender:       nullIfEmptyStr(req.Gender),
		MobileNumber: nullIfEmptyStr(mobileNumber),
		Location:     nullIfEmptyStr(req.Location),
	}

	id, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	user.ID = id

	// Only consume the challenge once the account actually exists - if
	// CreateUser fails (e.g. a raced duplicate), the client can still retry
	// registration with the same reference_id instead of having to redo the
	// OTP step.
	consumed, err := s.repo.ConsumeOTPVerification(ctx, challenge.ID)
	if err != nil {
		return nil, err
	}
	if !consumed {
		return nil, cc.ErrInvalidOrExpiredOTP
	}

	return s.issueTokens(ctx, user)
}

const (
	loginLockoutThreshold = 5
	loginLockoutWindow    = 15 * time.Minute // documents the policy; the actual boundary is computed in SQL
)

// LoginService is the first factor of login: member_id + password. On
// success it does not issue tokens - it creates an OTP challenge, sends the
// code via SMS, and the caller must complete VerifyLoginOTPService to
// actually get tokens. No "success" audit row is written here; that only
// happens once the OTP itself is verified.
func (s *AuthService) LoginService(ctx *gin.Context, req *models.LoginRequest) (*models.LoginOTPChallengeResponse, error) {
	ip := ctx.ClientIP()

	// Not transactional with the audit-log insert below - a burst of
	// concurrent requests can briefly overshoot the threshold before the
	// next request sees the true count. Bounded and self-correcting on the
	// next request, same eventual-consistency character as IPRateLimiter.
	identifierFailures, ipFailures, err := s.repo.CountRecentFailedLogins(ctx, req.MemberID, ip)
	if err != nil {
		return nil, err
	}

	if err := s.enforceLoginLockout(ctx, req.MemberID, ip, identifierFailures, ipFailures); err != nil {
		return nil, err
	}

	user, err := s.repo.FindUserByMemberID(ctx, req.MemberID)
	if err != nil {
		return nil, err
	}

	if user == nil || user.PasswordHash == nil {
		s.recordLoginAttempt(ctx, nil, req.MemberID, ip, models.LoginAttemptFailed, cc.INVALID_CREDENTIALS)
		return nil, errors.New(cc.INVALID_CREDENTIALS)
	}

	if err := CheckPassword(*user.PasswordHash, req.Password); err != nil {
		s.recordLoginAttempt(ctx, &user.ID, req.MemberID, ip, models.LoginAttemptFailed, cc.INVALID_CREDENTIALS)
		return nil, errors.New(cc.INVALID_CREDENTIALS)
	}

	// Checked after the password so an unauthenticated guesser can't use
	// this to probe account status - only a caller who already knows the
	// password reaches this branch. UserStatusPendingReview (e.g. a Young
	// Saver account not yet reviewed) is deliberately NOT blocked here -
	// it can log in like any other account for now; restricting what a
	// pending account can actually do is a dashboard-side concern for a
	// later change, not a login-time one.
	if user.Status == models.UserStatusSuspended {
		s.recordLoginAttempt(ctx, &user.ID, req.MemberID, ip, models.LoginAttemptFailed, "account suspended")
		return nil, cc.ErrAccountSuspended
	}

	return s.createAndSendLoginOTP(ctx, user)
}

// createAndSendLoginOTP generates a code, stores its hash as a new
// otp_verifications row, and sends it via SMS. member_id and
// mobile_number are denormalized onto the challenge row so verification can
// feed the same lockout counters as LoginService without an extra lookup.
func (s *AuthService) createAndSendLoginOTP(ctx *gin.Context, user *models.User) (*models.LoginOTPChallengeResponse, error) {
	if user.MobileNumber == nil || user.MemberID == nil {
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

	challenge := &models.OTPVerification{
		ReferenceID:  referenceID,
		UserID:       &user.ID,
		MemberID:     *user.MemberID,
		OTPHash:      code,
		MobileNumber: mobileNumber,
		Purpose:      otpPurposeLogin,
		MaxAttempts:  otpMaxAttempts,
	}

	id, err := s.repo.CreateOTPVerification(ctx, challenge, int(otpExpiry.Minutes()))
	if err != nil {
		return nil, err
	}

	message := sms.LoginOTPMessage(code, int(otpExpiry.Minutes()))
	if sendErr := s.smsSender.Send(ctx, mobileNumber, message); sendErr != nil {
		// Leave the row in place as 'failed' rather than deleting it - it's
		// unguessable (only the hash is stored) and useful for support to
		// see that delivery, not the user's input, is what failed.
		_ = s.repo.MarkOTPSendFailed(ctx, id, sendErr.Error())
		slog.Error("failed to send login OTP SMS", "user_id", user.ID, "otp_verification_id", id, "error", sendErr)
		return nil, fmt.Errorf("%w: %v", cc.ErrFailedToSendOTP, sendErr)
	}
	_ = s.repo.MarkOTPSent(ctx, id)

	// Email is a secondary delivery channel alongside SMS - SMS above is what
	// gates the challenge as usable (status='sent'), so any failure here
	// (bad address, render error, send error) is logged and swallowed rather
	// than failing the login attempt; the user can still complete it with
	// the code from SMS.
	if err := utils.Validate.Var(user.Email, "required,email"); err != nil {
		slog.Warn("skipping login OTP email: invalid or missing email on file", "user_id", user.ID)

	} else if emailBody, renderErr := mail.RenderOTPVerificationEmail(code, int(otpExpiry.Minutes())); renderErr != nil {
		slog.Warn("failed to render login OTP email", "user_id", user.ID, "error", renderErr)

	} else if sendErr := s.mailSender.Send(ctx, user.Email, mail.OTPVerificationSubject, emailBody); sendErr != nil {
		slog.Warn("failed to send login OTP email", "user_id", user.ID, "error", sendErr)
	}

	return &models.LoginOTPChallengeResponse{
		ReferenceID:  referenceID,
		ExpiresIn:    int(otpExpiry.Seconds()),
		MaskedMobile: maskMobileNumber(*user.MobileNumber),
	}, nil
}

// VerifyLoginOTPService is the second factor of login. It never distinguishes
// "wrong code once" from "attempts exhausted" from "unknown/expired
// reference" in its returned error - all collapse to ErrInvalidOrExpiredOTP,
// so a client can't use the response to narrow down a guess.
func (s *AuthService) VerifyLoginOTPService(ctx *gin.Context, req *models.VerifyOTPRequest) (*models.AuthTokenResponse, error) {
	ip := ctx.ClientIP()

	challenge, err := s.repo.FindActiveOTPVerificationByReference(ctx, req.ReferenceID)
	if err != nil {
		return nil, err
	}
	if challenge == nil || challenge.Purpose != otpPurposeLogin || challenge.UserID == nil {
		return nil, cc.ErrInvalidOrExpiredOTP
	}

	// Re-check the same lockout gate LoginService uses - a mix of bad
	// passwords and bad OTP guesses (possibly against different challenges)
	// should still lock the badge/IP out, not just this one challenge.
	identifierFailures, ipFailures, err := s.repo.CountRecentFailedLogins(ctx, challenge.MemberID, ip)
	if err != nil {
		return nil, err
	}
	if err := s.enforceLoginLockout(ctx, challenge.MemberID, ip, identifierFailures, ipFailures); err != nil {
		return nil, err
	}

	if challenge.Attempts >= challenge.MaxAttempts || req.OTPCode != challenge.OTPHash {
		// IncrementOTPAttempts also auto-consumes once max_attempts is hit,
		// atomically against the row's live count - see its doc comment.
		_ = s.repo.IncrementOTPAttempts(ctx, challenge.ID)
		s.recordLoginAttempt(ctx, challenge.UserID, challenge.MemberID, ip, models.LoginAttemptFailed, "invalid otp code")
		return nil, cc.ErrInvalidOrExpiredOTP
	}

	// consumed=false means another concurrent request (same reference_id,
	// same correct code) already redeemed this challenge first - it must
	// not also get treated as a valid login, or one OTP could mint two
	// token pairs.
	consumed, err := s.repo.ConsumeOTPVerification(ctx, challenge.ID)
	if err != nil {
		return nil, err
	}
	if !consumed {
		return nil, cc.ErrInvalidOrExpiredOTP
	}

	user, err := s.repo.FindUserByID(ctx, *challenge.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(cc.USER_NOT_FOUND)
	}

	s.recordLoginAttempt(ctx, &user.ID, challenge.MemberID, ip, models.LoginAttemptSuccess, "")
	_ = s.repo.TouchLastLogin(ctx, user.ID)

	return s.issueTokens(ctx, user)
}

// recordLoginAttempt is fire-and-forget, matching this file's existing
// tolerance for TouchLastLogin failures - an audit-log write error should
// never fail the login flow itself.
func (s *AuthService) recordLoginAttempt(ctx *gin.Context, userID *int64, identifier, ip string, status models.LoginAttemptStatus, reason string) {
	_ = s.repo.CreateLoginAuditLog(ctx, &models.LoginAuditLog{
		UserID:              userID,
		IdentifierAttempted: nullIfEmptyStr(identifier),
		IPAddress:           nullIfEmptyStr(ip),
		UserAgent:           nullIfEmptyStr(ctx.Request.UserAgent()),
		Status:              status,
		Reason:              nullIfEmptyStr(reason),
	})
}

// memberLockDurationDays is how long a locked_member_accounts row - and the
// copy in its notification email - says the lock lasts for. It does not
// itself gate login: CountRecentFailedLogins' 15-minute rolling window is
// what actually decides whether a request is currently locked out; this is
// only the persisted record/notification of that decision.
const memberLockDurationDays = 10

// enforceLoginLockout returns cc.ErrAccountTemporarilyLocked once either
// counter has crossed loginLockoutThreshold, and best-effort persists the
// event plus emails the affected member via lockMemberAccount. Shared by
// LoginService and VerifyLoginOTPService, which both gate on the same
// counters from CountRecentFailedLogins.
func (s *AuthService) enforceLoginLockout(ctx *gin.Context, identifier, ip string, identifierFailures, ipFailures int) error {
	if identifierFailures < loginLockoutThreshold && ipFailures < loginLockoutThreshold {
		return nil
	}

	attempts := identifierFailures
	reason := "too many failed login attempts for this member ID"
	if ipFailures > identifierFailures {
		attempts = ipFailures
		reason = "too many failed login attempts from this IP address"
	}

	s.lockMemberAccount(ctx, identifier, ip, attempts, reason)
	return cc.ErrAccountTemporarilyLocked
}

// lockMemberAccount is fire-and-forget, matching this file's tolerance for
// audit/notification failures elsewhere (see recordLoginAttempt) - a
// persistence or email problem must never surface as anything other than
// cc.ErrAccountTemporarilyLocked to the caller.
//
// While an unexpired, still-locked locked_member_accounts row already
// exists for identifier, this only refreshes its attempt_count - it does
// not insert a duplicate row or send a second email for the same lockout
// episode, so a member being repeatedly retried during the lockout window
// isn't emailed on every single request.
func (s *AuthService) lockMemberAccount(ctx *gin.Context, identifier, ip string, attempts int, reason string) {
	active, err := s.repo.FindActiveLockedMemberAccount(ctx, identifier)
	if err != nil {
		slog.Error("failed to check existing account lock", "member_id", identifier, "error", err)
		return
	}
	if active != nil {
		if attempts > active.AttemptCount {
			_ = s.repo.UpdateLockedMemberAccountAttempts(ctx, active.ID, attempts)
		}
		return
	}

	user, err := s.repo.FindUserByMemberID(ctx, identifier)
	if err != nil {
		slog.Error("failed to look up user for account lock", "member_id", identifier, "error", err)
	}

	lock := &models.LockedMemberAccount{
		MemberID:     identifier,
		IPAddress:    nullIfEmptyStr(ip),
		AttemptCount: attempts,
		Reason:       reason,
	}
	if user != nil {
		lock.UserID = &user.ID
	}

	id, err := s.repo.CreateLockedMemberAccount(ctx, lock)
	if err != nil {
		slog.Error("failed to record account lock", "member_id", identifier, "error", err)
		return
	}

	if user == nil {
		return
	}
	if err := utils.Validate.Var(user.Email, "required,email"); err != nil {
		slog.Warn("skipping account-lock email: invalid or missing email on file", "user_id", user.ID)
		return
	}

	emailBody, renderErr := mail.RenderAccountLockedEmail(memberLockDurationDays)
	if renderErr != nil {
		slog.Warn("failed to render account-lock email", "user_id", user.ID, "error", renderErr)
		return
	}
	if sendErr := s.mailSender.Send(ctx, user.Email, mail.AccountLockedSubject, emailBody); sendErr != nil {
		slog.Warn("failed to send account-lock email", "user_id", user.ID, "error", sendErr)
		return
	}
	_ = s.repo.MarkLockedMemberAccountNotified(ctx, id)
}

func (s *AuthService) SocialLoginService(ctx *gin.Context, req *models.SocialLoginRequest) (*models.AuthTokenResponse, error) {
	verifier, ok := s.verifiers[req.Provider]
	if !ok {
		return nil, errors.New(cc.INVALID_SOCIAL_PROVIDER)
	}

	profile, err := verifier.Verify(ctx, req.AccessToken)
	if err != nil {
		return nil, err
	}

	provider, err := s.repo.FindProviderByCode(ctx, req.Provider)
	if err != nil {
		return nil, err
	}
	if provider == nil || !provider.IsEnabled {
		return nil, errors.New(cc.INVALID_SOCIAL_PROVIDER)
	}

	user, err := s.repo.FindLinkedMemberFromSocialProfile(ctx, provider, profile)
	if err != nil {
		return nil, err
	}
	if user.Status == models.UserStatusSuspended {
		return nil, cc.ErrAccountSuspended
	}

	_ = s.repo.TouchLastLogin(ctx, user.ID)

	return s.issueTokens(ctx, user)
}

// SSOLoginService completes login for an identity goth has already verified
// via the provider's OAuth2 redirect flow - unlike SocialLoginService, there
// is no token to re-verify here, since trust was established by the
// redirect round-trip itself (see internal/handlers/auth/sso_handler.go).
func (s *AuthService) SSOLoginService(ctx *gin.Context, profile *models.SocialProfile) (*models.AuthTokenResponse, error) {
	provider, err := s.repo.FindProviderByCode(ctx, profile.Provider)
	if err != nil {
		return nil, err
	}
	if provider == nil || !provider.IsEnabled {
		return nil, errors.New(cc.INVALID_SOCIAL_PROVIDER)
	}
	if profile.Email == "" {
		return nil, errors.New(cc.INVALID_SOCIAL_TOKEN)
	}

	user, err := s.repo.FindLinkedMemberFromSocialProfile(ctx, provider, profile)
	if err != nil {
		return nil, err
	}
	if user.Status == models.UserStatusSuspended {
		return nil, cc.ErrAccountSuspended
	}

	_ = s.repo.TouchLastLogin(ctx, user.ID)

	return s.issueTokens(ctx, user)
}

func (s *AuthService) RefreshTokenService(ctx *gin.Context, refreshToken string) (*models.AuthTokenResponse, error) {
	hash := utils.MustStringSHA256(refreshToken)

	session, err := s.repo.FindActiveSessionByTokenHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New(cc.INVALID_REFRESH_TOKEN)
	}

	user, err := s.repo.FindUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(cc.USER_NOT_FOUND)
	}

	// Rotate: the presented refresh token is single-use - revoke it and
	// issue a brand new access/refresh pair.
	if err := s.repo.RevokeSession(ctx, hash); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user)
}

// UnlockAccountService clears the failed-login counters and any active lock
// record for memberID/ip, bypassing the real 15-minute lockout window.
// Development-only: the only caller is AuthHandler.handleUnlockAccount, and
// that route is only ever registered when hostConf.BuildEnv == "dev" (see
// cmd/api/api.go) - it must never be reachable in stage/prod, since it would
// otherwise defeat the brute-force protection entirely.
func (s *AuthService) UnlockAccountService(ctx *gin.Context, memberID, ip string) error {
	if err := s.repo.DeleteRecentFailedLogins(ctx, memberID, ip); err != nil {
		return err
	}
	return s.repo.UnlockMemberAccount(ctx, memberID)
}

func (s *AuthService) LogoutService(ctx *gin.Context, refreshToken string) error {
	hash := utils.MustStringSHA256(refreshToken)
	return s.repo.RevokeSession(ctx, hash)
}

// ListSessionsService returns the caller's own active sessions, most recent
// first, trimmed to what the profile page's device list needs. Current
// flags whichever entry matches the refresh_token cookie on this request -
// an empty currentRefreshToken (no cookie, e.g. an access token that hasn't
// been refreshed yet) just means nothing gets flagged, not an error.
func (s *AuthService) ListSessionsService(ctx *gin.Context, userID int64, currentRefreshToken string) ([]models.SessionInfo, error) {
	sessions, err := s.repo.ListActiveSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var currentHash string
	if currentRefreshToken != "" {
		currentHash = utils.MustStringSHA256(currentRefreshToken)
	}

	infos := make([]models.SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		infos = append(infos, models.SessionInfo{
			ID:        sess.ID,
			UserAgent: sess.UserAgent,
			IPAddress: sess.IPAddress,
			CreatedAt: sess.CreatedAt,
			Current:   currentHash != "" && sess.RefreshTokenHash == currentHash,
		})
	}

	return infos, nil
}

// UpdateProfileService edits the caller's own first/middle/last name, email,
// and mobile number - see UpdateProfileRequest for why this is limited to
// the users table (not the legacy client-master fields GET /profile also
// returns). A changed email is pre-checked against another account the same
// way registration is; UpdateUserProfile's own unique-constraint error is
// still the real enforcement against a concurrent race.
func (s *AuthService) UpdateProfileService(ctx *gin.Context, userID int64, req *models.UpdateProfileRequest) (*models.User, error) {
	existing, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != userID {
		return nil, cc.ErrEmailAlreadyRegistered
	}

	mobileNumber, err := NormalizeMobileNumber(req.MobileNumber)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateUserProfile(
		ctx, userID, req.FirstName, req.LastName, nullIfEmptyStr(req.MiddleName), req.Email, mobileNumber,
	); err != nil {
		return nil, err
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(cc.USER_NOT_FOUND)
	}
	return user, nil
}

// ChangePasswordService is the self-service "change my password while
// logged in" flow - distinct from the forgot-password OTP flow
// (forgot_password_service.go), this requires the caller to prove they
// already know the current password rather than a mobile OTP. A
// social/SSO-only account (no PasswordHash) has nothing to verify against,
// so it's rejected with ErrNoPasswordSet rather than silently setting one.
func (s *AuthService) ChangePasswordService(ctx *gin.Context, userID int64, req *models.ChangePasswordRequest) error {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(cc.USER_NOT_FOUND)
	}
	if user.PasswordHash == nil {
		return cc.ErrNoPasswordSet
	}
	if err := CheckPassword(*user.PasswordHash, req.CurrentPassword); err != nil {
		return cc.ErrIncorrectPassword
	}

	hash, err := HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdateUserPassword(ctx, userID, hash)
}

// GetSettingsService returns the caller's own mailing address and
// notification preference - the pair PersonalInformationCard/OtherInformationCard
// don't already cover (both come from the legacy client master, read-only).
func (s *AuthService) GetSettingsService(ctx *gin.Context, userID int64) (*models.SettingsResponse, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(cc.USER_NOT_FOUND)
	}

	enabled, err := s.repo.GetNotificationPreference(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.SettingsResponse{Address: user.Address, NotificationsEnabled: enabled}, nil
}

// UpdateSettingsService overwrites the caller's own mailing address and
// notification preference together - see UpdateSettingsRequest for why
// these two (not identity fields like name/email/mobile, which stay on
// UpdateProfileService) are grouped here.
func (s *AuthService) UpdateSettingsService(ctx *gin.Context, userID int64, req *models.UpdateSettingsRequest) (*models.SettingsResponse, error) {
	address := nullIfEmptyStr(req.Address)

	if err := s.repo.UpdateUserAddress(ctx, userID, address); err != nil {
		return nil, err
	}
	if err := s.repo.UpsertNotificationPreference(ctx, userID, req.NotificationsEnabled); err != nil {
		return nil, err
	}

	return &models.SettingsResponse{Address: address, NotificationsEnabled: req.NotificationsEnabled}, nil
}

func (s *AuthService) MeService(ctx *gin.Context, userID int64) (*models.User, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(cc.USER_NOT_FOUND)
	}
	return user, nil
}

func (s *AuthService) ParseAccessToken(tokenStr string) (*Claims, error) {
	return s.tokens.ParseAccessToken(tokenStr)
}

// RefreshTokenTTL exposes TokenService's refresh TTL so handlers can set the
// refresh_token cookie's Max-Age without reaching into the token service
// directly.
func (s *AuthService) RefreshTokenTTL() time.Duration {
	return s.tokens.RefreshTokenTTL()
}

// issueTokens signs a new access token and persists a new refresh-token
// session bound to the requesting device (user agent + IP), then returns
// both tokens plus the safe user projection.
func (s *AuthService) issueTokens(ctx *gin.Context, user *models.User) (*models.AuthTokenResponse, error) {
	access, expiresAt, err := s.tokens.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshRaw, err := GenerateOpaqueToken()
	if err != nil {
		return nil, err
	}
	refreshHash := utils.MustStringSHA256(refreshRaw)

	session := &models.AuthSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        nullIfEmptyStr(ctx.Request.UserAgent()),
		IPAddress:        nullIfEmptyStr(ctx.ClientIP()),
		ExpiresAt:        time.Now().Add(s.tokens.RefreshTokenTTL()),
	}

	if _, err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return &models.AuthTokenResponse{
		AccessToken:  access,
		RefreshToken: refreshRaw,
		TokenType:    "Bearer",
		ExpiresIn:    int(time.Until(expiresAt).Seconds()),
		User:         user.ToResponse(),
	}, nil
}

func nullIfEmptyStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// maskMobileNumber hides everything but the last 4 digits behind a
// fixed-width mask - not length-preserving, since the original length is
// itself metadata worth not leaking, e.g. "+15551234567" -> "•••• 4567".
func maskMobileNumber(s string) string {
	if len(s) <= 4 {
		return "•••• " + s
	}
	return "•••• " + s[len(s)-4:]
}
