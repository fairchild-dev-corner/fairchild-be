package services

import (
	"context"
	models "fairchild_be/internal/models/auth"
)

// SocialVerifier re-validates a provider access/id token server-side and
// returns the normalized identity claimed by it. Never trust a client-supplied
// profile without going through one of these.
type SocialVerifier interface {
	Verify(ctx context.Context, token string) (*models.SocialProfile, error)
}
