package handlers

import (
	"net/http"
	"time"

	cc "fairchild_be/internal/constants"
	middlewares "fairchild_be/internal/middlewares"
	common "fairchild_be/internal/models/common"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handleDepositBalances returns the authenticated caller's own current
// deposit balances (SAVINGS, TIME DEPOSIT, SHARE CAPITAL) - clientID/
// branchID are resolved once per request by middlewares.RequireMemberAuth,
// not re-resolved here.
func (h *BalanceHandler) handleDepositBalances(ctx *gin.Context) {
	clientID, branchID, ok := middlewares.ClientAccountFromContext(ctx)
	if !ok {
		loggers.GetCommonError(ctx, cc.USER_NOT_FOUND, http.StatusUnauthorized)
		return
	}

	balances, err := h.balanceService.DepositBalancesService(ctx, clientID, branchID)
	if err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_BALANCE_FETCHED,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       balances,
	})
}
