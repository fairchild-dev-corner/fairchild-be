package services

import (
	"context"
	"encoding/json"
	cc "fairchild_be/internal/constants"
	models "fairchild_be/internal/models/auth"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// FacebookVerifier validates the client-supplied access token by asking the
// Facebook Graph API for the profile it belongs to - an invalid/expired
// token simply fails to resolve a profile.
type FacebookVerifier struct {
	HTTPClient *http.Client
}

func NewFacebookVerifier() *FacebookVerifier {
	return &FacebookVerifier{HTTPClient: &http.Client{Timeout: 5 * time.Second}}
}

type facebookProfile struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

func (f *FacebookVerifier) Verify(ctx context.Context, accessToken string) (*models.SocialProfile, error) {
	endpoint := "https://graph.facebook.com/v19.0/me?fields=id,name,email,picture&access_token=" + url.QueryEscape(accessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cc.INVALID_SOCIAL_TOKEN, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: facebook responded with status %d", cc.INVALID_SOCIAL_TOKEN, resp.StatusCode)
	}

	var profile facebookProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("%s: %w", cc.INVALID_SOCIAL_TOKEN, err)
	}

	if profile.ID == "" || profile.Email == "" {
		return nil, fmt.Errorf("%s: missing id/email in facebook response", cc.INVALID_SOCIAL_TOKEN)
	}

	return &models.SocialProfile{
		Provider:       cc.PROVIDER_FACEBOOK,
		ProviderUserID: profile.ID,
		Email:          profile.Email,
		Name:           profile.Name,
		AvatarUrl:      profile.Picture.Data.URL,
		AccessToken:    accessToken,
	}, nil
}
