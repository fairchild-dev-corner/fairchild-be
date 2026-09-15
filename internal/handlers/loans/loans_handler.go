package handlers

import (
	"net/http"
	"strings"
	"time"

	cc "fairchild_be/internal/constants"
	middlewares "fairchild_be/internal/middlewares"
	common "fairchild_be/internal/models/common"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handleActiveLoans returns the authenticated caller's own active loans -
// clientID/branchID are resolved once per request by
// middlewares.RequireMemberAuth, not re-resolved here.
func (h *LoansHandler) handleActiveLoans(ctx *gin.Context) {
	clientID, branchID, ok := middlewares.ClientAccountFromContext(ctx)
	if !ok {
		loggers.GetCommonError(ctx, cc.USER_NOT_FOUND, http.StatusUnauthorized)
		return
	}

	loans, err := h.loansService.ActiveLoansService(ctx, clientID, branchID)
	if err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_LOANS_FETCHED,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       loans,
	})
}

// handleLoanTransactions returns the full posting history for one of the
// authenticated caller's own loans or AR accounts, identified by the
// "SLC-SLT-REF" ref_no path param (e.g. "12-2-29836") already exposed on
// GET /loans and GET /ar rows - clientID/branchID are resolved once per
// request by middlewares.RequireMemberAuth, not re-resolved here.
func (h *LoansHandler) handleLoanTransactions(ctx *gin.Context) {
	clientID, branchID, ok := middlewares.ClientAccountFromContext(ctx)
	if !ok {
		loggers.GetCommonError(ctx, cc.USER_NOT_FOUND, http.StatusUnauthorized)
		return
	}

	parts := strings.Split(ctx.Param("ref_no"), "-")
	if len(parts) != 3 {
		loggers.GetCommonError(ctx, "invalid ref_no", http.StatusBadRequest)
		return
	}
	slcCode, sltCode, refNo := parts[0], parts[1], parts[2]

	txns, err := h.transactionsService.TransactionsByRefService(ctx, clientID, branchID, slcCode, sltCode, refNo)
	if err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS_LOAN_TRANSACTIONS_FETCHED,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       txns,
	})
}
