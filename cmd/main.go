package main

import (
	"log/slog"

	conf "fairchild_be/config"
	cc "fairchild_be/internal/constants"
	"fairchild_be/pkg/database"
)

var BuildEnv string

func main() {

	hostConf, hErr := conf.GetHostConfig(BuildEnv)
	if hErr != nil {
		slog.Error("Failed to initialize host config", "error", hErr)
		panic(cc.UNINITIALIZED_PANIC_STATE)
	}

	mysqlConf, mErr := conf.GetMySQLConfig(BuildEnv)
	if mErr != nil {
		slog.Error("Failed to initialize MySQL config", "error", mErr)
		panic(cc.UNINITIALIZED_PANIC_STATE)
	}

	jwtConf, jErr := conf.GetJWTConfig()
	if jErr != nil {
		slog.Error("Failed to initialize JWT config", "error", jErr)
		panic(cc.UNINITIALIZED_PANIC_STATE)
	}

	moviderConf, mvErr := conf.GetMoviderConfig()
	if mvErr != nil {
		slog.Error("Failed to initialize Movider config", "error", mvErr)
		panic(cc.UNINITIALIZED_PANIC_STATE)
	}

	smtpConf, smtpErr := conf.GetSMTPConfig()
	if smtpErr != nil {
		slog.Error("Failed to initialize SMTP config", "error", smtpErr)
		panic(cc.UNINITIALIZED_PANIC_STATE)
	}

	/*** Initialize MySQL Connector (Auth DB) ***/
	mysqlDB, mysqlErr := database.MySQLInitConnector(mysqlConf, false)
	if mysqlErr != nil {
		slog.Error("Error connecting to MySQL", "error", mysqlErr)
		panic(cc.UNINITIALIZED_PANIC_STATE)
	}

	defer mysqlDB.Close()
	// Schema is managed via versioned migrations - see cmd/migrate.
	// Run `make migrate-up` (or `go run cmd/migrate/main.go up`) before starting the app.

	// Connect Server
	conf.ServerConfig(hostConf, mysqlDB, jwtConf, conf.Envs.GoogleClientID, moviderConf, smtpConf, conf.GetSSOConfig())
}
