package models

import "time"

// AuthSession mirrors `auth_sessions`. Only the SHA-256 hash of the refresh
// token is ever persisted - the raw token is returned to the client once
// and never stored.
type AuthSession struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	RefreshTokenHash string     `json:"-"`
	UserAgent        *string    `json:"user_agent,omitempty"`
	IPAddress        *string    `json:"ip_address,omitempty"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// SessionInfo is the response shape for GET /auth/sessions - a trimmed view
// of AuthSession for the caller's own device list. No geolocation: IP-based
// location lookups aren't reliable enough for the coop to present as fact,
// so callers only ever see the device, the raw IP, and the sign-in time.
type SessionInfo struct {
	ID        int64     `json:"id"`
	UserAgent *string   `json:"user_agent,omitempty"`
	IPAddress *string   `json:"ip_address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Current   bool      `json:"current"`
}
