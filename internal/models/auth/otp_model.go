package models

import "time"

type OTPDeliveryStatus string

const (
	// OTPDeliveryPending is the transient state between the challenge row
	// being inserted and the SMS send attempt resolving - it should never
	// be observed at rest outside of a crash between those two writes.
	OTPDeliveryPending OTPDeliveryStatus = "pending"
	OTPDeliverySent    OTPDeliveryStatus = "sent"
	OTPDeliveryFailed  OTPDeliveryStatus = "failed"
)

// OTPVerification mirrors `otp_verifications`. MemberID and MobileNumber
// are denormalized copies of the user's values at challenge-creation time, so
// verification/lockout checks don't need an extra user lookup. UserID is nil
// for Purpose == "register" challenges, which are created before any user
// row exists (see AuthService.SendRegisterOTPService) - every other purpose
// always sets it.
type OTPVerification struct {
	ID           int64             `json:"id"`
	ReferenceID  string            `json:"-"`
	UserID       *int64            `json:"user_id,omitempty"`
	MemberID     string            `json:"-"`
	OTPHash      string            `json:"-"`
	Status       OTPDeliveryStatus `json:"status"`
	SendError    *string           `json:"-"`
	MobileNumber string            `json:"-"`
	Purpose      string            `json:"purpose"`
	Attempts     int               `json:"attempts"`
	MaxAttempts  int               `json:"max_attempts"`
	ExpiresAt    time.Time         `json:"expires_at"`
	// VerifiedAt is set once the correct code has been submitted, before the
	// challenge is consumed - see AuthService.VerifyForgotPasswordOTPService.
	// A challenge can be verified without being consumed (password not yet
	// reset), but never consumed without first being verified.
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	ConsumedAt *time.Time `json:"consumed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
