package handlers

import (
	"net/http"

	cc "fairchild_be/internal/constants"
	models "fairchild_be/internal/models/auth"
	"fairchild_be/internal/utils"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
)

// handleClaimAccount lets a member with no email/mobile on file (so Forgot
// Password can't reach them - e.g. a legacy account migrated without portal
// credentials) submit a request to attach contact info to their account.
// Always returns the same generic success response regardless of whether
// member_id matches a real account - see AuthService.SubmitClaimRequestService.
func (h *AuthHandler) handleClaimAccount(ctx *gin.Context) {
	var req models.ClaimAccountRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	if err := h.authService.SubmitClaimRequestService(ctx, &req); err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.SuccessResponse(ctx, cc.SUCCESS_CLAIM_REQUEST_SUBMITTED)
}
