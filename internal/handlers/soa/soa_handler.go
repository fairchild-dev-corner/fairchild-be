package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	cc "fairchild_be/internal/constants"
	middlewares "fairchild_be/internal/middlewares"
	common "fairchild_be/internal/models/common"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handleSOAHistory returns the authenticated caller's own SOA history (last
// 3 months, or their single most recent statement) - clientID/branchID are
// resolved once per request by middlewares.RequireMemberAuth, not
// re-resolved here.
func (h *SOAHandler) handleSOAHistory(ctx *gin.Context) {
	clientID, branchID, ok := middlewares.ClientAccountFromContext(ctx)
	if !ok {
		loggers.GetCommonError(ctx, cc.USER_NOT_FOUND, http.StatusUnauthorized)
		return
	}

	history, err := h.soaService.SOAHistoryService(ctx, clientID, branchID)
	if err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_SOA_FETCHED,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       history,
	})
}

// handleSOAStatement returns the full detail of one of the authenticated
// caller's own statements, identified by the ":ctrl_no" path param (the
// literal "SOA No." printed on the legacy document) - clientID/branchID are
// resolved once per request by middlewares.RequireMemberAuth; see
// SOAService.SOAStatementService for why a foreign or nonexistent ctrl_no
// collapses to the same 404 rather than distinguishing the two cases.
func (h *SOAHandler) handleSOAStatement(ctx *gin.Context) {
	clientID, branchID, ok := middlewares.ClientAccountFromContext(ctx)
	if !ok {
		loggers.GetCommonError(ctx, cc.USER_NOT_FOUND, http.StatusUnauthorized)
		return
	}

	ctrlNo, err := strconv.ParseInt(ctx.Param("ctrl_no"), 10, 64)
	if err != nil {
		loggers.GetCommonError(ctx, "invalid ctrl_no", http.StatusBadRequest)
		return
	}

	statement, err := h.soaService.SOAStatementService(ctx, clientID, branchID, ctrlNo)
	if err != nil {
		if errors.Is(err, cc.ErrSOAStatementNotFound) {
			loggers.GetCommonError(ctx, err.Error(), http.StatusNotFound)
			return
		}
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_SOA_FETCHED,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       statement,
	})
}
