package models

// SendRegisterOTPRequest starts pre-registration mobile verification for
// both regular-member and Young Saver registration - see
// AuthService.SendRegisterOTPService. For a regular member MobileNumber is
// the applicant's own number; for a Young Saver it's the parent/guardian's
// number (verifying the code doubles as guardian consent). The client must
// follow up with VerifyOTPRequest, then either RegisterRequest or
// RegisterYoungSaverRequest using the same ReferenceID.
type SendRegisterOTPRequest struct {
	MobileNumber string `json:"mobile_number" validate:"required"`
	Email        string `json:"email" validate:"omitempty,email"`
}

// UpdateProfileRequest - PATCH /auth/profile. Edits the caller's own name/
// email/mobile fields on the `users` table only - the legacy client-master
// data surfaced by GET /profile (address, employment, TIN, etc.) has no
// write path yet. MobileNumber accepts PH local formats as well as E.164,
// same as RegisterRequest - AuthService.UpdateProfileService normalizes it.
type UpdateProfileRequest struct {
	FirstName    string `json:"first_name" validate:"required,max=100"`
	MiddleName   string `json:"middle_name" validate:"omitempty,max=100"`
	LastName     string `json:"last_name" validate:"required,max=100"`
	Email        string `json:"email" validate:"required,email"`
	MobileNumber string `json:"mobile_number" validate:"required"`
}

// ChangePasswordRequest - PATCH /auth/password. Requires the caller's
// current password (not just a valid access token) before overwriting it,
// same defense-in-depth as most "change password while logged in" flows -
// AuthService.ChangePasswordService rejects accounts with no password set
// (social/SSO-only) via ErrNoPasswordSet.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// UpdateSettingsRequest - PATCH /auth/settings. Address is the caller's own
// member-editable mailing address on the `users` table, distinct from the
// read-only legacy client-master address GET /profile returns (no write
// path there - see UpdateProfileRequest). NotificationsEnabled is a single
// on/off preference, not per-channel.
type UpdateSettingsRequest struct {
	Address              string `json:"address" validate:"omitempty,max=255"`
	NotificationsEnabled bool   `json:"notifications_enabled"`
}

// RegisterRequest is for regular-member registration.
// ReferenceID must belong to a "register"-purpose OTP challenge already
// confirmed via VerifyOTPRequest (AuthService.VerifyRegisterOTPService), and
// MobileNumber must match the number that challenge was sent to.
// MobileNumber accepts PH local formats (e.g. 09171234567) as well as
// E.164 - AuthService.RegisterService normalizes it to E.164 before storage
// via NormalizeMobileNumber, so format is validated there.
type RegisterRequest struct {
	ReferenceID  string `json:"reference_id" validate:"required,len=64,hexadecimal"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	FirstName    string `json:"first_name" validate:"required,max=100"`
	MiddleName   string `json:"middle_name" validate:"omitempty,max=100"`
	LastName     string `json:"last_name" validate:"required,max=100"`
	Suffix       string `json:"suffix" validate:"omitempty,max=20"`
	MemberID     string `json:"member_id" validate:"required,max=50"`
	DateOfBirth  string `json:"date_of_birth" validate:"required,datetime=2006-01-02"`
	Gender       string `json:"gender" validate:"required,oneof=male female prefer-not-to-say"`
	MobileNumber string `json:"mobile_number" validate:"required"`
	Location     string `json:"location" validate:"omitempty,max=255"`
}

// RegisterYoungSaverRequest enrolls a dependent (under 18) as a Young Saver.
// Unlike RegisterRequest, Email and MobileNumber (the dependent's own) are
// optional - the account's verified contact channel is the guardian's
// mobile number, which ReferenceID's OTP challenge must have been sent to
// (see AuthService.RegisterYoungSaverService). GuardianConsent must be true;
// verifying the OTP is what actually proves the guardian controls that
// number, this flag just records that they affirmatively checked the box.
// The resulting account starts as UserStatusPendingReview, not active.
type RegisterYoungSaverRequest struct {
	ReferenceID          string `json:"reference_id" validate:"required,len=64,hexadecimal"`
	Password             string `json:"password" validate:"required,min=8"`
	FirstName            string `json:"first_name" validate:"required,max=100"`
	MiddleName           string `json:"middle_name" validate:"omitempty,max=100"`
	LastName             string `json:"last_name" validate:"required,max=100"`
	Suffix               string `json:"suffix" validate:"omitempty,max=20"`
	MemberID             string `json:"member_id" validate:"required,max=50"`
	DateOfBirth          string `json:"date_of_birth" validate:"required,datetime=2006-01-02"`
	Gender               string `json:"gender" validate:"required,oneof=male female prefer-not-to-say"`
	Email                string `json:"email" validate:"omitempty,email"`
	MobileNumber         string `json:"mobile_number" validate:"omitempty"`
	GuardianFullName     string `json:"guardian_full_name" validate:"required,max=150"`
	GuardianMobileNumber string `json:"guardian_mobile_number" validate:"required"`
	GuardianConsent      bool   `json:"guardian_consent" validate:"required"`
}

// LoginRequest is the first factor of login: member_id + password.
// On success, an OTP is sent to the account's mobile_number - see
// VerifyOTPRequest for the second factor that completes the login.
type LoginRequest struct {
	MemberID string `json:"member_id" validate:"required,max=50"`
	Password string `json:"password" validate:"required"`
}

// VerifyOTPRequest completes an OTP challenge created by either LoginRequest
// or ForgotPasswordRequest. ReferenceID identifies the challenge (a
// hex-encoded opaque token, not a UUID - see
// AuthService.createAndSendLoginOTP), and which flow it belongs to is
// resolved server-side from the challenge's stored purpose.
type VerifyOTPRequest struct {
	ReferenceID string `json:"reference_id" validate:"required,len=64,hexadecimal"`
	OTPCode     string `json:"otp_code" validate:"required,len=6,numeric"`
}

// ForgotPasswordRequest starts the password-reset flow. If Email matches an
// account, an OTP is sent via SMS and email - see VerifyOTPRequest for the
// step that confirms the code, and ResetPasswordRequest for the step that
// actually sets the new password.
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest is the final step of the forgot-password flow.
// ReferenceID must belong to a challenge already confirmed via
// VerifyOTPRequest (AuthService.VerifyForgotPasswordOTPService) - the code
// itself is not re-submitted here.
type ResetPasswordRequest struct {
	ReferenceID string `json:"reference_id" validate:"required,len=64,hexadecimal"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// SocialLoginRequest carries the token the client already obtained from the
// provider's native SDK (Google/Facebook/Apple/GitHub/Yahoo). The backend
// re-verifies it server-side before trusting any identity claim.
type SocialLoginRequest struct {
	Provider    string `json:"provider" validate:"required,oneof=google facebook apple github yahoo"`
	AccessToken string `json:"access_token" validate:"required"`
}

// UnlockAccountRequest is a development-only convenience for clearing a
// login lockout without waiting out the real 15-minute window - see
// AuthHandler.RegisterDevRoutes. IPAddress defaults to the caller's own
// request IP when omitted.
type UnlockAccountRequest struct {
	MemberID  string `json:"member_id" validate:"required,max=50"`
	IPAddress string `json:"ip_address" validate:"omitempty,max=45"`
}
