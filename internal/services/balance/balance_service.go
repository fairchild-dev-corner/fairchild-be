package services

import (
	interfaces "fairchild_be/internal/interfaces"
	models "fairchild_be/internal/models/balance"

	"github.com/gin-gonic/gin"
)

type BalanceService struct {
	repo interfaces.BalanceRepositoryInterface
}

func NewBalanceService(repo interfaces.BalanceRepositoryInterface) *BalanceService {
	return &BalanceService{repo: repo}
}

// DepositBalancesService returns the authenticated caller's own current
// deposit balances, one entry per SAVINGS/TIME DEPOSIT/SHARE CAPITAL
// sub-product they've posted against. clientID/branchID are already
// resolved by middlewares.RequireMemberAuth - a caller with no linked
// ledger account (clientID == "") gets an empty list rather than an error -
// there's simply nothing to show yet.
func (s *BalanceService) DepositBalancesService(ctx *gin.Context, clientID, branchID string) ([]models.DepositBalance, error) {
	if clientID == "" {
		return []models.DepositBalance{}, nil
	}

	return s.repo.FindDepositBalances(ctx, clientID, branchID)
}
