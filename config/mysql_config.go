package config

import (
	"errors"
	"strconv"

	models "fairchild_be/internal/models/config"
)

/**
* Custom MySQL (Auth DB) Configuration & Enviroment Setup
* @Description: Define Setup Configuration {Host, Port, User, Password, DBName}
**/
func GetMySQLConfig(env string) (*models.MySQLConfig, error) {
	switch env {
	case "dev":
		return models.NewMySQLConfig(
			Envs.MySQLLocalHost,
			Envs.MySQLLocalPort,
			Envs.MySQLLocalUser,
			Envs.MySQLLocalPassword,
			Envs.MySQLLocalDBName,
			env,
		), nil

	case "stage":
		return models.NewMySQLConfig(
			Envs.MySQLStagHost,
			Envs.MySQLStagPort,
			Envs.MySQLStagUser,
			Envs.MySQLStagPassword,
			Envs.MySQLStagDBName,
			env,
		), nil

	default:
		return nil, errors.New("unknown environment: " + env)
	}
}

// GetLegacyMySQLConfig connects to the old portal's MySQL server - read-only
// source for the one-off cmd/migrate_legacy_* backfill tools. There's only
// ever one legacy target, so unlike GetMySQLConfig there's no dev/stage switch.
func GetLegacyMySQLConfig() (*models.MySQLConfig, error) {
	if Envs.LegacyMySQLHost == "" || Envs.LegacyMySQLDBName == "" {
		return nil, errors.New("missing LEGACY_MYSQL_HOST or LEGACY_MYSQL_DB_NAME")
	}

	return models.NewMySQLConfig(
		Envs.LegacyMySQLHost,
		Envs.LegacyMySQLPort,
		Envs.LegacyMySQLUser,
		Envs.LegacyMySQLPassword,
		Envs.LegacyMySQLDBName,
		"legacy",
	), nil
}

// GetJWTConfig loads signing secrets and TTLs, falling back to safe defaults
// (15m access / 30d refresh) when the TTL env vars are absent or invalid.
func GetJWTConfig() (*models.JWTConfig, error) {
	if Envs.JWTAccessSecret == "" || Envs.JWTRefreshSecret == "" {
		return nil, errors.New("missing JWT_ACCESS_SECRET or JWT_REFRESH_SECRET")
	}

	accessTTL, err := strconv.Atoi(Envs.JWTAccessTTLMinutes)
	if err != nil || accessTTL <= 0 {
		accessTTL = 15
	}

	refreshTTL, err := strconv.Atoi(Envs.JWTRefreshTTLDays)
	if err != nil || refreshTTL <= 0 {
		refreshTTL = 30
	}

	return models.NewJWTConfig(Envs.JWTAccessSecret, Envs.JWTRefreshSecret, accessTTL, refreshTTL), nil
}
