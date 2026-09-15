package config

import (
	models "fairchild_be/internal/models/config"
)

// GetSSOConfig loads the settings for the goth-based OAuth2 redirect flow.
// Unlike Movider/SMTP, this does not error on missing values - an app can
// legitimately run with only some providers configured (e.g. Yahoo but not
// Facebook yet), same as the existing GoogleClientID handling.
func GetSSOConfig() *models.SSOConfig {
	return models.NewSSOConfig(
		Envs.SessionSecret,
		Envs.SSOCallbackBaseURL,
		Envs.SSOFrontendRedirectURL,
		Envs.GoogleClientID,
		Envs.GoogleClientSecret,
		Envs.FacebookClientID,
		Envs.FacebookClientSecret,
		Envs.YahooClientID,
		Envs.YahooClientSecret,
	)
}
