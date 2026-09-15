package config

import (
	"errors"

	models "fairchild_be/internal/models/config"
)

// GetMoviderConfig loads the Movider SMS credentials used to send login
// OTPs. SenderName is optional per Movider's API docs.
func GetMoviderConfig() (*models.MoviderConfig, error) {
	if Envs.MoviderAPIKey == "" || Envs.MoviderAPISecret == "" {
		return nil, errors.New("missing MOVIDER_API_KEY or MOVIDER_API_SECRET")
	}

	return models.NewMoviderConfig(Envs.MoviderAPIKey, Envs.MoviderAPISecret, Envs.MoviderSenderName), nil
}
