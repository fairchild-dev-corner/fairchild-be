package services

import (
	"context"
	"encoding/json"
	cc "fairchild_be/internal/constants"
	models "fairchild_be/internal/models/auth"
	"fmt"
	"net/http"
	"time"
)

// YahooVerifier validates the client-supplied access token by asking Yahoo's
// OpenID Connect userinfo endpoint for the profile it belongs to - an
// invalid/expired token simply fails to resolve a profile.
type YahooVerifier struct {
	HTTPClient *http.Client
}

func NewYahooVerifier() *YahooVerifier {
	return &YahooVerifier{HTTPClient: &http.Client{Timeout: 5 * time.Second}}
}

type yahooProfile struct {
	Sub     string `json:"sub"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func (y *YahooVerifier) Verify(ctx context.Context, accessToken string) (*models.SocialProfile, error) {
	endpoint := "https://api.login.yahoo.com/openid/v1/userinfo"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := y.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cc.INVALID_SOCIAL_TOKEN, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: yahoo responded with status %d", cc.INVALID_SOCIAL_TOKEN, resp.StatusCode)
	}

	var profile yahooProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("%s: %w", cc.INVALID_SOCIAL_TOKEN, err)
	}

	if profile.Sub == "" || profile.Email == "" {
		return nil, fmt.Errorf("%s: missing sub/email in yahoo response", cc.INVALID_SOCIAL_TOKEN)
	}

	return &models.SocialProfile{
		Provider:       cc.PROVIDER_YAHOO,
		ProviderUserID: profile.Sub,
		Email:          profile.Email,
		Name:           profile.Name,
		AvatarUrl:      profile.Picture,
		AccessToken:    accessToken,
	}, nil
}
