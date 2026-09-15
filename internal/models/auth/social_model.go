package models

import "time"

// AuthProvider mirrors the `auth_providers` reference table.
type AuthProvider struct {
	ID        int8   `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsEnabled bool   `json:"is_enabled"`
}

// SocialAccount mirrors `user_social_accounts` - the join between a user and
// one identity issued by a social provider (many social accounts : one user).
type SocialAccount struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	ProviderID     int8       `json:"provider_id"`
	ProviderUserID string     `json:"provider_user_id"`
	ProviderEmail  *string    `json:"provider_email,omitempty"`
	AccessToken    *string    `json:"-"`
	RefreshToken   *string    `json:"-"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// SocialProfile is the normalized identity a provider verifier returns once
// it has validated the client-supplied token against the provider's API.
type SocialProfile struct {
	Provider       string
	ProviderUserID string
	Email          string
	Name           string
	AvatarUrl      string
	AccessToken    string
	RefreshToken   string
	ExpiresAt      *time.Time
}
