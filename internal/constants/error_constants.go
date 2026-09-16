package cc

import "errors"

const (
	UNINITIALIZED_PANIC_STATE = "[panic state] setup/un-initialized"

	// Auth Errors
	EMAIL_ALREADY_REGISTERED     = "Email is already registered"
	MEMBER_ID_ALREADY_REGISTERED = "Member ID is already registered"
	INVALID_CREDENTIALS          = "Invalid member ID or password"
	USER_NOT_FOUND               = "User not found"
	EMAIL_NOT_ASSOCIATED         = "This email is not associated with any account"
	INVALID_SOCIAL_PROVIDER      = "Unsupported or unknown social provider"
	INVALID_SOCIAL_TOKEN         = "Unable to verify social provider token"
	INVALID_REFRESH_TOKEN        = "Invalid or expired refresh token"
	SESSION_REVOKED              = "Session has been revoked"
	ACCOUNT_TEMPORARILY_LOCKED   = "Too many failed login attempts, please try again later"
	INVALID_OR_EXPIRED_OTP       = "Invalid or expired otp code"
	FAILED_TO_SEND_OTP           = "Unable to send otp, please try again later"
	INVALID_MOBILE_NUMBER        = "Mobile number is invalid"
	ACCOUNT_SUSPENDED            = "This account has been suspended. Please contact support"
	YOUNG_SAVER_MUST_BE_MINOR    = "Young Saver accounts are only for dependents under 18 years old"
	INCORRECT_PASSWORD           = "Current password is incorrect"
	NO_PASSWORD_SET              = "This account has no password set - sign in with social login instead"

	// Statement of Account
	SOA_STATEMENT_NOT_FOUND = "statement not found"
)

// ErrEmailAlreadyRegistered is the sentinel error returned when a
// registration is attempted for an email that already exists, so callers
// can distinguish it from other errors via errors.Is instead of matching
// on message text.
var ErrEmailAlreadyRegistered = errors.New(EMAIL_ALREADY_REGISTERED)

// ErrMemberIDAlreadyRegistered is the sentinel error returned when a
// registration is attempted for a member_id that already exists.
var ErrMemberIDAlreadyRegistered = errors.New(MEMBER_ID_ALREADY_REGISTERED)

// ErrEmailNotAssociated is the sentinel error returned when
// ForgotPasswordService is given an email that doesn't match any account.
var ErrEmailNotAssociated = errors.New(EMAIL_NOT_ASSOCIATED)

// ErrAccountTemporarilyLocked is the sentinel error returned when a login is
// rejected because the member_id or source IP has too many recent
// failures (bad password or bad OTP, see AuthService).
var ErrAccountTemporarilyLocked = errors.New(ACCOUNT_TEMPORARILY_LOCKED)

// ErrInvalidOrExpiredOTP is the sentinel error returned for any OTP-verify
// failure - unknown reference, expired, already consumed, wrong code, or
// attempts exhausted. Deliberately generic: never gives the client a way to
// distinguish "wrong code once" from "exhausted".
var ErrInvalidOrExpiredOTP = errors.New(INVALID_OR_EXPIRED_OTP)

// ErrFailedToSendOTP is the sentinel error returned when the credentials
// were valid but the SMS provider (Movider) failed to deliver the OTP.
var ErrFailedToSendOTP = errors.New(FAILED_TO_SEND_OTP)

// ErrInvalidMobileNumber is the sentinel error returned when a mobile
// number can't be normalized into E.164, the format Movider requires.
var ErrInvalidMobileNumber = errors.New(INVALID_MOBILE_NUMBER)

// ErrAccountSuspended is the sentinel error returned by LoginService when
// credentials are valid but the account's status is UserStatusSuspended.
// UserStatusPendingReview (e.g. a Young Saver not yet reviewed) is
// deliberately not gated here - see LoginService's status check.
var ErrAccountSuspended = errors.New(ACCOUNT_SUSPENDED)

// ErrYoungSaverMustBeMinor is the sentinel error returned when
// RegisterYoungSaverService is given a date_of_birth for someone 18 or
// older.
var ErrYoungSaverMustBeMinor = errors.New(YOUNG_SAVER_MUST_BE_MINOR)

// ErrSOAStatementNotFound is the sentinel error returned when a requested
// ctrl_no doesn't belong to the authenticated caller, or doesn't exist at
// all - both cases collapse to the same 404 so a foreign ctrl_no can't be
// probed to determine whether it's valid (IDOR-safe).
var ErrSOAStatementNotFound = errors.New(SOA_STATEMENT_NOT_FOUND)

// ErrIncorrectPassword is the sentinel error returned by
// AuthService.ChangePasswordService when CurrentPassword doesn't match the
// account's stored hash.
var ErrIncorrectPassword = errors.New(INCORRECT_PASSWORD)

// ErrNoPasswordSet is the sentinel error returned by
// AuthService.ChangePasswordService for a social/SSO-only account (no
// PasswordHash) - there is nothing to verify CurrentPassword against.
var ErrNoPasswordSet = errors.New(NO_PASSWORD_SET)
