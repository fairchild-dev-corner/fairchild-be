package handlers

import (
	"net/http"
	"time"

	cc "fairchild_be/internal/constants"
	common "fairchild_be/internal/models/common"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handleProfile returns the authenticated caller's own realtime profile -
// see ProfileService.ProfileService for how the caller's own member id is
// resolved server-side from their user id.
func (h *ProfileHandler) handleProfile(ctx *gin.Context) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		loggers.GetCommonError(ctx, cc.USER_NOT_FOUND, http.StatusUnauthorized)
		return
	}

	profile, err := h.profileService.ProfileService(ctx, userID)
	if err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_PROFILE_FETCHED,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       profile,
	})
}
