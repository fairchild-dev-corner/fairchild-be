package services

import (
	models "fairchild_be/internal/models/config"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/yahoo"
)

// RegisterGothProviders tells goth which OAuth2 providers the server-initiated
// SSO redirect flow supports (see internal/handlers/auth/sso_handler.go).
// Adding another provider later - anything goth already implements - is just
// one more line here plus its client id/secret in config, no new handler code.
func RegisterGothProviders(conf *models.SSOConfig) {
	goth.UseProviders(
		google.New(conf.GoogleClientID, conf.GoogleClientSecret, conf.CallbackBaseURL+"/api/v1/auth/sso/google/callback", "email", "profile"),
		facebook.New(conf.FacebookClientID, conf.FacebookClientSecret, conf.CallbackBaseURL+"/api/v1/auth/sso/facebook/callback", "email"),
		yahoo.New(conf.YahooClientID, conf.YahooClientSecret, conf.CallbackBaseURL+"/api/v1/auth/sso/yahoo/callback", "openid", "profile", "email"),
	)
}
