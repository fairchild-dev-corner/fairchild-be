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

// handleForgotPassword is the first step of the forgot-password flow: given
// an email, it creates and sends an OTP challenge (SMS + email) that must be
// completed via handleVerifyForgotPasswordOTP.
func (h *AuthHandler) handleForgotPassword(ctx *gin.Context) {
	var req models.ForgotPasswordRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	res, err := h.authService.ForgotPasswordService(ctx, &req)
	if err != nil {
		if errors.Is(err, cc.ErrFailedToSendOTP) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
			return
		}
		if errors.Is(err, cc.ErrEmailNotAssociated) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusNotFound)
			return
		}
		loggers.GetCommonError(ctx, err.Error(), http.StatusBadRequest)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_FORGOT_PASSWORD_OTP_SENT,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       res,
	})
}

// handleVerifyForgotPasswordOTP is the second step of the forgot-password
// flow - it confirms the code for the challenge created by
// handleForgotPassword. It does not reset the password: the caller must
// follow up with handleResetForgotPassword using the same reference_id.
func (h *AuthHandler) handleVerifyForgotPasswordOTP(ctx *gin.Context) {
	var req models.VerifyOTPRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	if err := h.authService.VerifyForgotPasswordOTPService(ctx, &req); err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
		return
	}

	loggers.SuccessResponse(ctx, cc.SUCCESS_FORGOT_PASSWORD_OTP_VERIFIED)
}

// handleResetForgotPassword is the final step of the forgot-password flow -
// it completes a challenge already confirmed via handleVerifyForgotPasswordOTP
// and sets the submitted new_password as the account's password.
func (h *AuthHandler) handleResetForgotPassword(ctx *gin.Context) {
	var req models.ResetPasswordRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	if err := h.authService.ResetForgotPasswordService(ctx, &req); err != nil {
		if err.Error() == cc.USER_NOT_FOUND {
			loggers.GetCommonError(ctx, err.Error(), http.StatusNotFound)
			return
		}
		loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
		return
	}

	loggers.SuccessResponse(ctx, cc.SUCCESS_PASSWORD_RESET)
}
