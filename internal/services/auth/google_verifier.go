package services

import (
	"context"
	cc "fairchild_be/internal/constants"
	models "fairchild_be/internal/models/auth"
	"fmt"

	"google.golang.org/api/idtoken"
)

// GoogleVerifier validates the Google ID token the client obtained from
// Google Sign-In against Google's public keys and the app's OAuth client id.
type GoogleVerifier struct {
	ClientID string
}

func NewGoogleVerifier(clientID string) *GoogleVerifier {
	return &GoogleVerifier{ClientID: clientID}
}

func (g *GoogleVerifier) Verify(ctx context.Context, rawIDToken string) (*models.SocialProfile, error) {
	payload, err := idtoken.Validate(ctx, rawIDToken, g.ClientID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cc.INVALID_SOCIAL_TOKEN, err)
	}

	sub, _ := payload.Claims["sub"].(string)
	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	if sub == "" || email == "" {
		return nil, fmt.Errorf("%s: missing sub/email claim", cc.INVALID_SOCIAL_TOKEN)
	}

	return &models.SocialProfile{
		Provider:       cc.PROVIDER_GOOGLE,
		ProviderUserID: sub,
		Email:          email,
		Name:           name,
		AvatarUrl:      picture,
		AccessToken:    rawIDToken,
	}, nil
}
