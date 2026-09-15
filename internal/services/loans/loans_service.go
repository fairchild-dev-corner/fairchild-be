package services

import (
	interfaces "fairchild_be/internal/interfaces"
	models "fairchild_be/internal/models/loans"

	"github.com/gin-gonic/gin"
)

type LoansService struct {
	repo interfaces.LoansRepositoryInterface
}

func NewLoansService(repo interfaces.LoansRepositoryInterface) *LoansService {
	return &LoansService{repo: repo}
}

// ActiveLoansService returns the authenticated caller's own active loans
// (outstanding balance > 0), newest release first. clientID/branchID are
// already resolved by middlewares.RequireMemberAuth - a caller with no
// linked ledger account (clientID == "") gets an empty slice rather than an
// error - there's simply nothing to show.
func (s *LoansService) ActiveLoansService(ctx *gin.Context, clientID, branchID string) ([]models.Loan, error) {
	if clientID == "" {
		return []models.Loan{}, nil
	}

	return s.repo.FindActiveLoans(ctx, clientID, branchID)
}
