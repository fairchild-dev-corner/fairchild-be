package config

import (
	"errors"

	models "fairchild_be/internal/models/config"
)

// GetSMTPConfig loads the SMTP credentials used to send emails (e.g. OTP
// verification) via net/smtp.
func GetSMTPConfig() (*models.SMTPConfig, error) {
	if Envs.SMTPHost == "" || Envs.SMTPPort == "" || Envs.SMTPUsername == "" || Envs.SMTPPassword == "" || Envs.SMTPFrom == "" {
		return nil, errors.New("missing SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD, or SMTP_FROM")
	}

	return models.NewSMTPConfig(Envs.SMTPHost, Envs.SMTPPort, Envs.SMTPUsername, Envs.SMTPPassword, Envs.SMTPFrom), nil
}
