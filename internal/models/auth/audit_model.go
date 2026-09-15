package models

import "time"

type LoginAttemptStatus string

const (
	LoginAttemptSuccess LoginAttemptStatus = "success"
	LoginAttemptFailed  LoginAttemptStatus = "failed"
)

// LoginAuditLog mirrors `login_audit_logs`. ProviderID is always nil today -
// this table is only populated by member_id+password login (see
// AuthService.LoginService/VerifyLoginOTPService); wiring social login into
// it is a future extension, not implemented here. IdentifierAttempted holds
// the member_id that was attempted (renamed from the original
// email-based login flow via migration 000005).
type LoginAuditLog struct {
	ID                  int64              `json:"id"`
	UserID              *int64             `json:"user_id,omitempty"`
	ProviderID          *int8              `json:"provider_id,omitempty"`
	IdentifierAttempted *string            `json:"identifier_attempted,omitempty"`
	IPAddress           *string            `json:"ip_address,omitempty"`
	UserAgent           *string            `json:"user_agent,omitempty"`
	Status              LoginAttemptStatus `json:"status"`
	Reason              *string            `json:"reason,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
}
