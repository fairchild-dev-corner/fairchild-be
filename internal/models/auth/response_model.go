package models

// RefreshToken is deliberately excluded from JSON (json:"-") - it is never
// returned to the client in a response body or URL fragment. Handlers set it
// as an HttpOnly refresh_token cookie instead (see
// AuthHandler.setRefreshTokenCookie) so it isn't reachable from JS.
type AuthTokenResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"-"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int           `json:"expires_in"`
	User         *UserResponse `json:"user"`
}

// LoginOTPChallengeResponse is returned by the first factor of login
// (member_id + password) once credentials are valid - the client must
// follow up with VerifyOTPRequest using ReferenceID to complete login.
type LoginOTPChallengeResponse struct {
	ReferenceID  string `json:"reference_id"`
	ExpiresIn    int    `json:"expires_in"`
	MaskedMobile string `json:"masked_mobile"`
}

// ForgotPasswordOTPChallengeResponse is returned once a password-reset OTP
// challenge has been created and sent - the client must follow up with
// VerifyOTPRequest using ReferenceID to complete the reset and receive a new
// temporary password by email.
type ForgotPasswordOTPChallengeResponse struct {
	ReferenceID  string `json:"reference_id"`
	ExpiresIn    int    `json:"expires_in"`
	MaskedMobile string `json:"masked_mobile"`
}

// SettingsResponse - GET/PATCH /auth/settings.
type SettingsResponse struct {
	Address              *string `json:"address"`
	NotificationsEnabled bool    `json:"notifications_enabled"`
}

// RegisterOTPChallengeResponse is returned by SendRegisterOTPRequest - the
// client must follow up with VerifyOTPRequest using ReferenceID, then
// complete registration (RegisterRequest or RegisterYoungSaverRequest) using
// the same ReferenceID.
type RegisterOTPChallengeResponse struct {
	ReferenceID  string `json:"reference_id"`
	ExpiresIn    int    `json:"expires_in"`
	MaskedMobile string `json:"masked_mobile"`
}
