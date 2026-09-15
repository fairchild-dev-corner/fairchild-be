package models

/**
* MySQL Database Configuration
* @Description: Holds connection parameters, mapped onto the driver's
* mysql.Config{} by pkg/database.MySQLInitConnector.
**/
type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	BuildEnv string
}

func NewMySQLConfig(host, port, user, password, dbName, buildEnv string) *MySQLConfig {
	return &MySQLConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		BuildEnv: buildEnv,
	}
}

/**
* JWT Configuration
* @Description: Holds signing secrets and token lifetimes for issuing/validating
* access and refresh tokens.
**/
type JWTConfig struct {
	AccessSecret     string
	RefreshSecret    string
	AccessTTLMinutes int
	RefreshTTLDays   int
}

func NewJWTConfig(accessSecret, refreshSecret string, accessTTLMinutes, refreshTTLDays int) *JWTConfig {
	return &JWTConfig{
		AccessSecret:     accessSecret,
		RefreshSecret:    refreshSecret,
		AccessTTLMinutes: accessTTLMinutes,
		RefreshTTLDays:   refreshTTLDays,
	}
}

/**
* Movider (SMS OTP) Configuration
* @Description: Holds the credentials used to send login OTPs via the
* Movider API. See internal/services/auth/movider_client.go.
**/
type MoviderConfig struct {
	APIKey     string
	APISecret  string
	SenderName string
}

func NewMoviderConfig(apiKey, apiSecret, senderName string) *MoviderConfig {
	return &MoviderConfig{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		SenderName: senderName,
	}
}

/**
* SMTP Configuration
* @Description: Holds the credentials used to send emails (e.g. OTP
* verification) via net/smtp. See internal/services/mail/smtp_client.go.
**/
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func NewSMTPConfig(host, port, username, password, from string) *SMTPConfig {
	return &SMTPConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

/**
* SSO Configuration
* @Description: Holds the credentials and endpoints for the goth-based
* OAuth2 redirect flow (see internal/handlers/auth/sso_handler.go). Distinct
* from GoogleClientID, which is only used to verify a token the client
* already obtained from a native SDK.
**/
type SSOConfig struct {
	SessionSecret       string
	CallbackBaseURL     string
	FrontendRedirectURL string

	GoogleClientID     string
	GoogleClientSecret string

	FacebookClientID     string
	FacebookClientSecret string

	YahooClientID     string
	YahooClientSecret string
}

func NewSSOConfig(
	sessionSecret, callbackBaseURL, frontendRedirectURL string,
	googleClientID, googleClientSecret string,
	facebookClientID, facebookClientSecret string,
	yahooClientID, yahooClientSecret string,
) *SSOConfig {
	return &SSOConfig{
		SessionSecret:        sessionSecret,
		CallbackBaseURL:      callbackBaseURL,
		FrontendRedirectURL:  frontendRedirectURL,
		GoogleClientID:       googleClientID,
		GoogleClientSecret:   googleClientSecret,
		FacebookClientID:     facebookClientID,
		FacebookClientSecret: facebookClientSecret,
		YahooClientID:        yahooClientID,
		YahooClientSecret:    yahooClientSecret,
	}
}
