package handlers

import (
	"errors"
	"net/http"
	"time"

	cc "fairchild_be/internal/constants"
	models "fairchild_be/internal/models/auth"
	common "fairchild_be/internal/models/common"
	"fairchild_be/internal/utils"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *AuthHandler) handleRegister(ctx *gin.Context) {
	var req models.RegisterRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	res, err := h.authService.RegisterService(ctx, &req)

	if err != nil {
		if errors.Is(err, cc.ErrEmailAlreadyRegistered) || errors.Is(err, cc.ErrMemberIDAlreadyRegistered) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, cc.ErrInvalidOrExpiredOTP) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
			return
		}
		loggers.GetCommonError(ctx, err.Error(), http.StatusBadRequest)
		return
	}

	h.respondWithTokens(ctx, cc.SUCCESS_REGISTER, res)
}

// handleSendRegisterOTP starts pre-registration mobile verification, shared
// by regular-member and Young Saver registration - see
// AuthService.SendRegisterOTPService.
func (h *AuthHandler) handleSendRegisterOTP(ctx *gin.Context) {
	var req models.SendRegisterOTPRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	res, err := h.authService.SendRegisterOTPService(ctx, &req)
	if err != nil {
		if errors.Is(err, cc.ErrFailedToSendOTP) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
			return
		}
		loggers.GetCommonError(ctx, err.Error(), http.StatusBadRequest)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_REGISTER_OTP_SENT,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       res,
	})
}

// handleVerifyRegisterOTP completes the challenge created by
// handleSendRegisterOTP - it only confirms the code (see
// AuthService.VerifyRegisterOTPService), it does not create an account.
func (h *AuthHandler) handleVerifyRegisterOTP(ctx *gin.Context) {
	var req models.VerifyOTPRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	if err := h.authService.VerifyRegisterOTPService(ctx, &req); err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
		return
	}

	loggers.SuccessResponse(ctx, cc.SUCCESS_REGISTER_OTP_VERIFIED)
}

// handleRegisterYoungSaver enrolls a dependent under 18 - unlike
// handleRegister, it does not issue tokens (the account starts pending
// review, see AuthService.RegisterYoungSaverService).
func (h *AuthHandler) handleRegisterYoungSaver(ctx *gin.Context) {
	var req models.RegisterYoungSaverRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	res, err := h.authService.RegisterYoungSaverService(ctx, &req)
	if err != nil {
		if errors.Is(err, cc.ErrEmailAlreadyRegistered) || errors.Is(err, cc.ErrMemberIDAlreadyRegistered) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, cc.ErrInvalidOrExpiredOTP) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
			return
		}
		loggers.GetCommonError(ctx, err.Error(), http.StatusBadRequest)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_YOUNG_SAVER_REGISTER,
		HttpCode:   http.StatusCreated,
		ResponseAt: time.Now(),
		Body:       res,
	})
}

// handleLogin is the first factor of login: member_id + password. On
// success it does not return tokens - it returns an OTP challenge that must
// be completed via handleVerifyLoginOTP.
func (h *AuthHandler) handleLogin(ctx *gin.Context) {
	var req models.LoginRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	res, err := h.authService.LoginService(ctx, &req)

	if err != nil {
		if errors.Is(err, cc.ErrAccountTemporarilyLocked) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusLocked)
			return
		}
		if errors.Is(err, cc.ErrFailedToSendOTP) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
			return
		}
		if errors.Is(err, cc.ErrAccountSuspended) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusForbidden)
			return
		}
		loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_OTP_SENT,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       res,
	})
}

// handleVerifyLoginOTP is the second factor of login - it completes the
// challenge created by handleLogin and issues tokens.
func (h *AuthHandler) handleVerifyLoginOTP(ctx *gin.Context) {
	var req models.VerifyOTPRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	res, err := h.authService.VerifyLoginOTPService(ctx, &req)
	if err != nil {
		if errors.Is(err, cc.ErrAccountTemporarilyLocked) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusLocked)
			return
		}

		loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
		return
	}

	h.respondWithTokens(ctx, cc.SUCCESS_LOGIN, res)
}

// handleSocialLogin logs in an existing member using a token the client
// already obtained from the provider's native SDK (Google/Facebook/Yahoo).
// The token is re-verified server-side. It never creates a new account - an
// email with no matching member fails with ErrEmailNotAssociated.
func (h *AuthHandler) handleSocialLogin(ctx *gin.Context) {
	var req models.SocialLoginRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	res, err := h.authService.SocialLoginService(ctx, &req)
	if err != nil {
		if errors.Is(err, cc.ErrEmailNotAssociated) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusForbidden)
			return
		}
		if errors.Is(err, cc.ErrAccountSuspended) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusForbidden)
			return
		}
		loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
		return
	}

	h.respondWithTokens(ctx, cc.SUCCESS_SOCIAL_LOGIN, res)
}

// handleRefreshToken reads the refresh token from the HttpOnly cookie set by
// respondWithTokens - it is never accepted from the request body, since the
// client is never given the raw value.
func (h *AuthHandler) handleRefreshToken(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie(refreshTokenCookieName)
	if err != nil || refreshToken == "" {
		loggers.GetCommonError(ctx, cc.INVALID_REFRESH_TOKEN, http.StatusUnauthorized)
		return
	}

	res, err := h.authService.RefreshTokenService(ctx, refreshToken)
	if err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
		return
	}

	h.respondWithTokens(ctx, cc.SUCCESS_TOKEN_REFRESH, res)
}

// handleLogout reads the refresh token from the same HttpOnly cookie as
// handleRefreshToken. A missing cookie means the client is already logged
// out, not an error.
func (h *AuthHandler) handleLogout(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie(refreshTokenCookieName)
	if err == nil && refreshToken != "" {
		if err := h.authService.LogoutService(ctx, refreshToken); err != nil {
			loggers.GetCommonError(ctx, err.Error(), http.StatusBadRequest)
			return
		}
	}

	clearRefreshTokenCookie(ctx, h.buildEnv)
	loggers.SuccessResponse(ctx, cc.SUCCESS_LOGOUT)
}

// handleSessions returns the caller's own active login sessions - device,
// IP, and sign-in time, with the current one flagged. No location: the coop
// doesn't run IP geolocation, so nothing beyond the raw IP is shown as fact.
func (h *AuthHandler) handleSessions(ctx *gin.Context) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		loggers.GetCommonError(ctx, cc.USER_NOT_FOUND, http.StatusUnauthorized)
		return
	}

	refreshToken, _ := ctx.Cookie(refreshTokenCookieName)

	sessions, err := h.authService.ListSessionsService(ctx, userID, refreshToken)
	if err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_SESSIONS_FETCHED,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       sessions,
	})
}

// handleUpdateProfile edits the caller's own first/middle/last name, email,
// and mobile number - see UpdateProfileRequest for why this is limited to
// the users table.
func (h *AuthHandler) handleUpdateProfile(ctx *gin.Context) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		loggers.GetCommonError(ctx, cc.USER_NOT_FOUND, http.StatusUnauthorized)
		return
	}

	var req models.UpdateProfileRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	user, err := h.authService.UpdateProfileService(ctx, userID, &req)
	if err != nil {
		if errors.Is(err, cc.ErrEmailAlreadyRegistered) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, cc.ErrInvalidMobileNumber) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusBadRequest)
			return
		}
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_PROFILE_UPDATED,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       user.ToResponse(),
	})
}

func (h *AuthHandler) handleProfile(ctx *gin.Context) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		loggers.GetCommonError(ctx, cc.USER_NOT_FOUND, http.StatusUnauthorized)
		return
	}

	user, err := h.authService.MeService(ctx, userID)
	if err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusNotFound)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       user.ToResponse(),
	})

}

// respondWithTokens sets the refresh token as an HttpOnly cookie (it is
// excluded from AuthTokenResponse's JSON, see response_model.go) and returns
// the access token in the body as before.
func (h *AuthHandler) respondWithTokens(ctx *gin.Context, status string, res *models.AuthTokenResponse) {
	setRefreshTokenCookie(ctx, h.buildEnv, res.RefreshToken, h.authService.RefreshTokenTTL())

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     status,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       res,
	})
}
