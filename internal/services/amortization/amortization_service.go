package services

import (
	interfaces "fairchild_be/internal/interfaces"
	models "fairchild_be/internal/models/amortization"

	"github.com/gin-gonic/gin"
)

type AmortizationService struct {
	repo interfaces.AmortizationRepositoryInterface
}

func NewAmortizationService(repo interfaces.AmortizationRepositoryInterface) *AmortizationService {
	return &AmortizationService{repo: repo}
}

// AmortizationScheduleService returns the authenticated caller's own next
// due installment per active loan. clientID/branchID are already resolved
// by middlewares.RequireMemberAuth - a caller with no linked ledger account
// (clientID == "") gets an empty list rather than an error - there's simply
// nothing to show yet.
func (s *AmortizationService) AmortizationScheduleService(ctx *gin.Context, clientID, branchID string) ([]models.AmortizationScheduleEntry, error) {
	if clientID == "" {
		return []models.AmortizationScheduleEntry{}, nil
	}

	return s.repo.FindAmortizationSchedule(ctx, clientID, branchID)
}

// LoanScheduleService returns the full projected installment schedule for
// one specific loan the authenticated caller owns (identified by its
// SLC_CODE/SLT_CODE/REF_NO). clientID/branchID are already resolved by
// middlewares.RequireMemberAuth - a caller with no linked ledger account
// gets an empty list rather than an error.
func (s *AmortizationService) LoanScheduleService(ctx *gin.Context, clientID, branchID, slcCode, sltCode, refNo string) ([]models.AmortizationScheduleEntry, error) {
	if clientID == "" {
		return []models.AmortizationScheduleEntry{}, nil
	}

	return s.repo.FindLoanSchedule(ctx, clientID, branchID, slcCode, sltCode, refNo)
}
