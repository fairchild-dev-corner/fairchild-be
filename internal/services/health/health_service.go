package services

import (
	interfaces "fairchild_be/internal/interfaces"
	models "fairchild_be/internal/models/health"

	"github.com/gin-gonic/gin"
)

type HealthService struct {
	repo interfaces.HealthRepositoryInterface
}

func NewHealthService(repo interfaces.HealthRepositoryInterface) *HealthService {
	return &HealthService{repo: repo}
}

// CoopHealthService returns the authenticated caller's own coop health
// score. clientID/branchID are already resolved by
// middlewares.RequireMemberAuth - a caller with no linked ledger account
// (clientID == "") gets nil rather than an error - there's simply no score
// to show yet.
func (s *HealthService) CoopHealthService(ctx *gin.Context, clientID, branchID string) (*models.CoopHealthScore, error) {
	if clientID == "" {
		return nil, nil
	}

	return s.repo.FindCoopHealthScore(ctx, clientID, branchID)
}
