package models

import "time"

// LockedMemberAccount mirrors `locked_member_accounts` - a persisted record
// of a login-lockout event (see AuthService.lockMemberAccount), separate
// from the rolling-window check in login_audit_logs that actually decides
// whether a request is currently locked out. Kept for 10 days from when it
// was first recorded (ExpiresAt); while an unexpired, still-locked row
// exists for a member_id, a repeat lockout trip updates it in place
// rather than creating a duplicate row or re-sending the notification email.
//
// IsLocked/UnlockedAt exist so a support action can clear a lock before its
// natural 10-day expiry without deleting the audit trail - nothing in this
// codebase writes UnlockedAt yet.
type LockedMemberAccount struct {
	ID           int64      `json:"id"`
	UserID       *int64     `json:"user_id,omitempty"`
	MemberID     string     `json:"member_id"`
	IPAddress    *string    `json:"ip_address,omitempty"`
	AttemptCount int        `json:"attempt_count"`
	Reason       string     `json:"reason"`
	IsLocked     bool       `json:"is_locked"`
	LockedAt     time.Time  `json:"locked_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	NotifiedAt   *time.Time `json:"notified_at,omitempty"`
	UnlockedAt   *time.Time `json:"unlocked_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
