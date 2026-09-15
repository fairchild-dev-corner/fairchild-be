package core

import (
	"database/sql"
	"fmt"
	"log"

	activity_handler "fairchild_be/internal/handlers/activity"
	amortization_handler "fairchild_be/internal/handlers/amortization"
	ar_handler "fairchild_be/internal/handlers/ar"
	auth_handler "fairchild_be/internal/handlers/auth"
	balance_handler "fairchild_be/internal/handlers/balance"
	health_handler "fairchild_be/internal/handlers/health"
	loans_handler "fairchild_be/internal/handlers/loans"
	profile_handler "fairchild_be/internal/handlers/profile"
	soa_handler "fairchild_be/internal/handlers/soa"
	transactions_handler "fairchild_be/internal/handlers/transactions"
	middlewares "fairchild_be/internal/middlewares"
	models "fairchild_be/internal/models/config"
	account_repo "fairchild_be/internal/repositories/account"
	activity_repo "fairchild_be/internal/repositories/activity"
	amortization_repo "fairchild_be/internal/repositories/amortization"
	ar_repo "fairchild_be/internal/repositories/ar"
	auth_repo "fairchild_be/internal/repositories/auth"
	balance_repo "fairchild_be/internal/repositories/balance"
	health_repo "fairchild_be/internal/repositories/health"
	loans_repo "fairchild_be/internal/repositories/loans"
	profile_repo "fairchild_be/internal/repositories/profile"
	soa_repo "fairchild_be/internal/repositories/soa"
	transactions_repo "fairchild_be/internal/repositories/transactions"
	router "fairchild_be/internal/routes"
	activity_service "fairchild_be/internal/services/activity"
	amortization_service "fairchild_be/internal/services/amortization"
	ar_service "fairchild_be/internal/services/ar"
	auth_service "fairchild_be/internal/services/auth"
	balance_service "fairchild_be/internal/services/balance"
	health_service "fairchild_be/internal/services/health"
	loans_service "fairchild_be/internal/services/loans"
	"fairchild_be/internal/services/mail"
	profile_service "fairchild_be/internal/services/profile"
	"fairchild_be/internal/services/sms"
	soa_service "fairchild_be/internal/services/soa"
	transactions_service "fairchild_be/internal/services/transactions"

	cc "fairchild_be/internal/constants"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
)

type APIServer struct {
	server         *gin.Engine
	hostConf       *models.HostConfig
	mysqlDB        *sql.DB
	jwtConf        *models.JWTConfig
	googleClientID string
	moviderConf    *models.MoviderConfig
	smtpConf       *models.SMTPConfig
	ssoConf        *models.SSOConfig
}

func NewAPIServer(
	s *gin.Engine,
	h *models.HostConfig,
	mysqlDB *sql.DB,
	jwtConf *models.JWTConfig,
	googleClientID string,
	moviderConf *models.MoviderConfig,
	smtpConf *models.SMTPConfig,
	ssoConf *models.SSOConfig,
) *APIServer {
	return &APIServer{
		server:         s,
		hostConf:       h,
		mysqlDB:        mysqlDB,
		jwtConf:        jwtConf,
		googleClientID: googleClientID,
		moviderConf:    moviderConf,
		smtpConf:       smtpConf,
		ssoConf:        ssoConf,
	}
}

func (s *APIServer) Run() {

	// Define Routes Config
	routes := router.NewRouterConfig(s.server, s.hostConf.AllowedOrigins)
	routes.InitConfig()

	// auth: repo -> service -> handler
	authRepo := auth_repo.NewAuthRepository(s.mysqlDB)
	tokenService := auth_service.NewTokenService(s.jwtConf)

	socialVerifiers := map[string]auth_service.SocialVerifier{
		cc.PROVIDER_GOOGLE:   auth_service.NewGoogleVerifier(s.googleClientID),
		cc.PROVIDER_FACEBOOK: auth_service.NewFacebookVerifier(),
		cc.PROVIDER_YAHOO:    auth_service.NewYahooVerifier(),
	}

	smsSender := sms.NewMoviderClient(s.moviderConf.APIKey, s.moviderConf.APISecret, s.moviderConf.SenderName)
	mailSender := mail.NewSMTPClient(s.smtpConf.Host, s.smtpConf.Port, s.smtpConf.Username, s.smtpConf.Password, s.smtpConf.From)
	authService := auth_service.NewAuthService(authRepo, tokenService, socialVerifiers, smsSender, mailSender)
	authHandler := auth_handler.NewHandler(authService, s.hostConf.BuildEnv)
	authHandler.RegisterRoutes(routes.RoutesGroup())

	// Dev-only convenience routes (e.g. clearing a login lockout while
	// testing) - never registered in a stage/prod binary, see
	// AuthHandler.RegisterDevRoutes.
	if s.hostConf.BuildEnv == "dev" {
		authHandler.RegisterDevRoutes(routes.RoutesGroup())
	}

	// sso: goth-based OAuth2 redirect flow, mounted under /api/v1/auth/sso
	gothic.Store = sessions.NewCookieStore([]byte(s.ssoConf.SessionSecret))
	auth_service.RegisterGothProviders(s.ssoConf)
	ssoHandler := auth_handler.NewSSOHandler(authService, s.ssoConf.FrontendRedirectURL, s.hostConf.BuildEnv)
	ssoHandler.RegisterRoutes(routes.RoutesGroup())

	// requireMemberAuth: shared by every per-member feature below (loans,
	// ar, amortization, soa, transactions, health, activity, balance) -
	// validates the access token AND resolves the caller's own ClientID/
	// branch exactly once per request, instead of each feature
	// independently re-querying the same account lookup. See
	// middlewares.RequireMemberAuth.
	accountRepo := account_repo.NewAccountRepository(s.mysqlDB)
	requireMemberAuth := middlewares.RequireMemberAuth(tokenService, accountRepo)

	// transactions: repo -> service -> handler
	transactionsRepo := transactions_repo.NewTransactionsRepository(s.mysqlDB)
	transactionsService := transactions_service.NewTransactionsService(transactionsRepo)
	transactionsHandler := transactions_handler.NewHandler(transactionsService)
	transactionsHandler.RegisterRoutes(routes.RoutesGroup(), requireMemberAuth)

	// health: repo -> service -> handler
	healthRepo := health_repo.NewHealthRepository(s.mysqlDB)
	healthService := health_service.NewHealthService(healthRepo)
	healthHandler := health_handler.NewHandler(healthService)
	healthHandler.RegisterRoutes(routes.RoutesGroup(), requireMemberAuth)

	// profile: repo -> service -> handler
	profileRepo := profile_repo.NewProfileRepository(s.mysqlDB)
	profileService := profile_service.NewProfileService(profileRepo)
	profileHandler := profile_handler.NewHandler(profileService, tokenService)
	profileHandler.RegisterRoutes(routes.RoutesGroup())

	// amortization: repo -> service -> handler
	amortizationRepo := amortization_repo.NewAmortizationRepository(s.mysqlDB)
	amortizationService := amortization_service.NewAmortizationService(amortizationRepo)
	amortizationHandler := amortization_handler.NewHandler(amortizationService)
	amortizationHandler.RegisterRoutes(routes.RoutesGroup(), requireMemberAuth)

	// loans: repo -> service -> handler
	loansRepo := loans_repo.NewLoansRepository(s.mysqlDB)
	loansService := loans_service.NewLoansService(loansRepo)
	loansHandler := loans_handler.NewHandler(loansService, transactionsService)
	loansHandler.RegisterRoutes(routes.RoutesGroup(), requireMemberAuth)

	// ar: repo -> service -> handler
	arRepo := ar_repo.NewARRepository(s.mysqlDB)
	arService := ar_service.NewARService(arRepo)
	arHandler := ar_handler.NewHandler(arService)
	arHandler.RegisterRoutes(routes.RoutesGroup(), requireMemberAuth)

	// activity: repo -> service -> handler
	activityRepo := activity_repo.NewActivityRepository(s.mysqlDB)
	activityService := activity_service.NewActivityService(activityRepo)
	activityHandler := activity_handler.NewHandler(activityService)
	activityHandler.RegisterRoutes(routes.RoutesGroup(), requireMemberAuth)

	// balance: repo -> service -> handler
	balanceRepo := balance_repo.NewBalanceRepository(s.mysqlDB)
	balanceService := balance_service.NewBalanceService(balanceRepo)
	balanceHandler := balance_handler.NewHandler(balanceService)
	balanceHandler.RegisterRoutes(routes.RoutesGroup(), requireMemberAuth)

	// soa: repo -> service -> handler
	soaRepo := soa_repo.NewSOARepository(s.mysqlDB)
	soaService := soa_service.NewSOAService(soaRepo)
	soaHandler := soa_handler.NewHandler(soaService)
	soaHandler.RegisterRoutes(routes.RoutesGroup(), requireMemberAuth)

	addr := fmt.Sprintf(":%s", s.hostConf.Port)
	if err := s.server.Run(addr); err != nil {
		log.Fatalf("unable to serve: %v", err)
	}
}
