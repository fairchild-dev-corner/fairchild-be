package sms

import "context"

// Sender delivers a text message to a phone number. Never trust OTP delivery
// to have succeeded without checking the returned error - see
// MoviderClient.Send for provider-specific failure modes (e.g. a 200
// response that still failed for a specific number).
type Sender interface {
	Send(ctx context.Context, to, message string) error
}
