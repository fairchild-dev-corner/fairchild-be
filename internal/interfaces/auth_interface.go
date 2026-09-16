package interfaces

import (
	models "fairchild_be/internal/models/auth"

	"github.com/gin-gonic/gin"
)

type AuthRepositoryInterface interface {
	// Users
	CreateUser(ctx *gin.Context, u *models.User) (int64, error)
	FindUserByEmail(ctx *gin.Context, email string) (*models.User, error)
	FindUserByMemberID(ctx *gin.Context, memberID string) (*models.User, error)
	FindUserByID(ctx *gin.Context, id int64) (*models.User, error)
	TouchLastLogin(ctx *gin.Context, userID int64) error
	UpdateUserPassword(ctx *gin.Context, userID int64, passwordHash string) error
	UpdateUserProfile(ctx *gin.Context, userID int64, firstName, lastName string, middleName *string, email, mobileNumber string) error
	UpdateUserAddress(ctx *gin.Context, userID int64, address *string) error

	// Notification Preferences
	GetNotificationPreference(ctx *gin.Context, userID int64) (bool, error)
	UpsertNotificationPreference(ctx *gin.Context, userID int64, enabled bool) error

	// Social Accounts / Providers
	FindProviderByCode(ctx *gin.Context, code string) (*models.AuthProvider, error)
	FindSocialAccount(ctx *gin.Context, providerID int8, providerUserID string) (*models.SocialAccount, error)
	FindLinkedMemberFromSocialProfile(ctx *gin.Context, provider *models.AuthProvider, profile *models.SocialProfile) (*models.User, error)

	// Sessions (refresh tokens)
	CreateSession(ctx *gin.Context, session *models.AuthSession) (int64, error)
	FindActiveSessionByTokenHash(ctx *gin.Context, tokenHash string) (*models.AuthSession, error)
	RevokeSession(ctx *gin.Context, tokenHash string) error
	ListActiveSessionsByUserID(ctx *gin.Context, userID int64) ([]models.AuthSession, error)

	// Login Audit
	CreateLoginAuditLog(ctx *gin.Context, log *models.LoginAuditLog) error
	CountRecentFailedLogins(ctx *gin.Context, identifier, ip string) (identifierFailures, ipFailures int, err error)
	DeleteRecentFailedLogins(ctx *gin.Context, identifier, ip string) error

	// OTP Verifications
	CreateOTPVerification(ctx *gin.Context, c *models.OTPVerification, expiresInMinutes int) (int64, error)
	FindActiveOTPVerificationByReference(ctx *gin.Context, referenceID string) (*models.OTPVerification, error)
	FindVerifiedOTPVerificationByReference(ctx *gin.Context, referenceID string) (*models.OTPVerification, error)
	IncrementOTPAttempts(ctx *gin.Context, id int64) error
	ConsumeOTPVerification(ctx *gin.Context, id int64) (consumed bool, err error)
	MarkOTPSent(ctx *gin.Context, id int64) error
	MarkOTPSendFailed(ctx *gin.Context, id int64, reason string) error
	MarkOTPVerified(ctx *gin.Context, id int64) error

	// Locked Member Accounts
	FindActiveLockedMemberAccount(ctx *gin.Context, memberID string) (*models.LockedMemberAccount, error)
	CreateLockedMemberAccount(ctx *gin.Context, l *models.LockedMemberAccount) (int64, error)
	UpdateLockedMemberAccountAttempts(ctx *gin.Context, id int64, attemptCount int) error
	MarkLockedMemberAccountNotified(ctx *gin.Context, id int64) error
	UnlockMemberAccount(ctx *gin.Context, memberID string) error

	// Account Claim Requests
	CreateClaimRequest(ctx *gin.Context, c *models.ClaimRequest) (int64, error)
}
