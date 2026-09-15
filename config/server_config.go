package config

import (
	"database/sql"
	core "fairchild_be/cmd/api"
	models "fairchild_be/internal/models/config"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func ServerConfig(
	hostConf *models.HostConfig,
	mysqlDB *sql.DB,
	jwtConf *models.JWTConfig,
	googleClientID string,
	moviderConf *models.MoviderConfig,
	smtpConf *models.SMTPConfig,
	ssoConf *models.SSOConfig,
) {

	gin.ForceConsoleColor()

	ginServer := gin.Default()

	// Middlewares
	ShowMascot()
	slog.Info("App Running", "addr", hostConf.Host)
	slog.Info("Running in --" + hostConf.BuildEnv + " mode & setup config")
	slog.Info("Starting Server", "port", hostConf.Port)

	// Run Server
	core.NewAPIServer(ginServer, hostConf, mysqlDB, jwtConf, googleClientID, moviderConf, smtpConf, ssoConf).Run()

}
