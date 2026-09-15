package services

import (
	interfaces "fairchild_be/internal/interfaces"
	models "fairchild_be/internal/models/activity"

	"github.com/gin-gonic/gin"
)

type ActivityService struct {
	repo interfaces.ActivityRepositoryInterface
}

func NewActivityService(repo interfaces.ActivityRepositoryInterface) *ActivityService {
	return &ActivityService{repo: repo}
}

// MonthlyActivityService returns the authenticated caller's own trailing
// 12-month savings/time-deposit deposit vs withdrawal activity. clientID/
// branchID are already resolved by middlewares.RequireMemberAuth - a caller
// with no linked ledger account (clientID == "") gets an empty list rather
// than an error - there's simply nothing to show yet.
func (s *ActivityService) MonthlyActivityService(ctx *gin.Context, clientID, branchID string) ([]models.ActivityMonth, error) {
	if clientID == "" {
		return []models.ActivityMonth{}, nil
	}

	return s.repo.FindMonthlyActivity(ctx, clientID, branchID)
}
