package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PublicHost string
	Port       string

	Dev_Allowed_Origins  string
	Stag_Allowed_Origins string

	// MySQL (Auth DB)
	MySQLLocalHost     string
	MySQLLocalPort     string
	MySQLLocalUser     string
	MySQLLocalPassword string
	MySQLLocalDBName   string
	MySQLStagHost      string
	MySQLStagPort      string
	MySQLStagUser      string
	MySQLStagPassword  string
	MySQLStagDBName    string
	// MySQLStagCACert is the PEM-encoded CA certificate for verifying the
	// staging MySQL server's TLS cert (e.g. DigitalOcean managed MySQL).
	MySQLStagCACert string

	// Legacy MySQL (read-only source for one-off cmd/migrate_legacy_* tools)
	LegacyMySQLHost     string
	LegacyMySQLPort     string
	LegacyMySQLUser     string
	LegacyMySQLPassword string
	LegacyMySQLDBName   string

	// Auth / JWT
	JWTAccessSecret     string
	JWTRefreshSecret    string
	JWTAccessTTLMinutes string
	JWTRefreshTTLDays   string

	// Social Auth Providers (native SDK token verification)
	GoogleClientID string

	// SSO (goth-based OAuth2 redirect flow - see internal/handlers/auth/sso_handler.go)
	SessionSecret          string
	SSOCallbackBaseURL     string
	SSOFrontendRedirectURL string
	GoogleClientSecret     string
	FacebookClientID       string
	FacebookClientSecret   string
	YahooClientID          string
	YahooClientSecret      string

	// Movider (SMS OTP)
	MoviderAPIKey     string
	MoviderAPISecret  string
	MoviderSenderName string

	// SMTP (Email OTP)
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

/*Create Singleton*/
var Envs = initConfig()

func initConfig() Config {

	//reload enviroment var
	godotenv.Load()

	return Config{
		PublicHost: getEnv("PUBLIC_HOST", os.Getenv("PUBLIC_HOST")),
		Port:       getEnv("PORT", os.Getenv("PORT")),

		Dev_Allowed_Origins:  getEnv("DEV_ALLOWED_ORIGINS", os.Getenv("DEV_ALLOWED_ORIGINS")),
		Stag_Allowed_Origins: getEnv("STAG_ALLOWED_ORIGINS", os.Getenv("STAG_ALLOWED_ORIGINS")),

		MySQLLocalHost:     getEnv("MYSQL_LOCAL_HOST", os.Getenv("MYSQL_LOCAL_HOST")),
		MySQLLocalPort:     getEnv("MYSQL_LOCAL_PORT", os.Getenv("MYSQL_LOCAL_PORT")),
		MySQLLocalUser:     getEnv("MYSQL_LOCAL_USER", os.Getenv("MYSQL_LOCAL_USER")),
		MySQLLocalPassword: getEnv("MYSQL_LOCAL_PASSWORD", os.Getenv("MYSQL_LOCAL_PASSWORD")),
		MySQLLocalDBName:   getEnv("MYSQL_LOCAL_DB_NAME", os.Getenv("MYSQL_LOCAL_DB_NAME")),
		MySQLStagHost:      getEnv("MYSQL_STAG_HOST", os.Getenv("MYSQL_STAG_HOST")),
		MySQLStagPort:      getEnv("MYSQL_STAG_PORT", os.Getenv("MYSQL_STAG_PORT")),
		MySQLStagUser:      getEnv("MYSQL_STAG_USER", os.Getenv("MYSQL_STAG_USER")),
		MySQLStagPassword:  getEnv("MYSQL_STAG_PASSWORD", os.Getenv("MYSQL_STAG_PASSWORD")),
		MySQLStagDBName:    getEnv("MYSQL_STAG_DB_NAME", os.Getenv("MYSQL_STAG_DB_NAME")),
		MySQLStagCACert:    getEnv("MYSQL_STAG_CA_CERT", os.Getenv("MYSQL_STAG_CA_CERT")),

		LegacyMySQLHost:     getEnv("LEGACY_MYSQL_HOST", os.Getenv("LEGACY_MYSQL_HOST")),
		LegacyMySQLPort:     getEnv("LEGACY_MYSQL_PORT", os.Getenv("LEGACY_MYSQL_PORT")),
		LegacyMySQLUser:     getEnv("LEGACY_MYSQL_USER", os.Getenv("LEGACY_MYSQL_USER")),
		LegacyMySQLPassword: getEnv("LEGACY_MYSQL_PASSWORD", os.Getenv("LEGACY_MYSQL_PASSWORD")),
		LegacyMySQLDBName:   getEnv("LEGACY_MYSQL_DB_NAME", os.Getenv("LEGACY_MYSQL_DB_NAME")),

		JWTAccessSecret:     getEnv("JWT_ACCESS_SECRET", os.Getenv("JWT_ACCESS_SECRET")),
		JWTRefreshSecret:    getEnv("JWT_REFRESH_SECRET", os.Getenv("JWT_REFRESH_SECRET")),
		JWTAccessTTLMinutes: getEnv("JWT_ACCESS_TTL_MINUTES", os.Getenv("JWT_ACCESS_TTL_MINUTES")),
		JWTRefreshTTLDays:   getEnv("JWT_REFRESH_TTL_DAYS", os.Getenv("JWT_REFRESH_TTL_DAYS")),

		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", os.Getenv("GOOGLE_CLIENT_ID")),

		SessionSecret:          getEnv("SESSION_SECRET", os.Getenv("SESSION_SECRET")),
		SSOCallbackBaseURL:     getEnv("SSO_CALLBACK_BASE_URL", os.Getenv("SSO_CALLBACK_BASE_URL")),
		SSOFrontendRedirectURL: getEnv("SSO_FRONTEND_REDIRECT_URL", os.Getenv("SSO_FRONTEND_REDIRECT_URL")),
		GoogleClientSecret:     getEnv("GOOGLE_CLIENT_SECRET", os.Getenv("GOOGLE_CLIENT_SECRET")),
		FacebookClientID:       getEnv("FACEBOOK_CLIENT_ID", os.Getenv("FACEBOOK_CLIENT_ID")),
		FacebookClientSecret:   getEnv("FACEBOOK_CLIENT_SECRET", os.Getenv("FACEBOOK_CLIENT_SECRET")),
		YahooClientID:          getEnv("YAHOO_CLIENT_ID", os.Getenv("YAHOO_CLIENT_ID")),
		YahooClientSecret:      getEnv("YAHOO_CLIENT_SECRET", os.Getenv("YAHOO_CLIENT_SECRET")),

		MoviderAPIKey:     getEnv("MOVIDER_API_KEY", os.Getenv("MOVIDER_API_KEY")),
		MoviderAPISecret:  getEnv("MOVIDER_API_SECRET", os.Getenv("MOVIDER_API_SECRET")),
		MoviderSenderName: getEnv("MOVIDER_SENDER_NAME", os.Getenv("MOVIDER_SENDER_NAME")),

		SMTPHost:     getEnv("SMTP_HOST", os.Getenv("SMTP_HOST")),
		SMTPPort:     getEnv("SMTP_PORT", os.Getenv("SMTP_PORT")),
		SMTPUsername: getEnv("SMTP_USERNAME", os.Getenv("SMTP_USERNAME")),
		SMTPPassword: getEnv("SMTP_PASSWORD", os.Getenv("SMTP_PASSWORD")),
		SMTPFrom:     getEnv("SMTP_FROM", os.Getenv("SMTP_FROM")),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
