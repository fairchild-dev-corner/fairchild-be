package services

import (
	interfaces "fairchild_be/internal/interfaces"
	models "fairchild_be/internal/models/ar"

	"github.com/gin-gonic/gin"
)

type ARService struct {
	repo interfaces.ARRepositoryInterface
}

func NewARService(repo interfaces.ARRepositoryInterface) *ARService {
	return &ARService{repo: repo}
}

// ActiveARService returns the authenticated caller's own active
// accounts-receivable entries. clientID/branchID are already resolved by
// middlewares.RequireMemberAuth - a caller with no linked ledger account
// (clientID == "") gets an empty slice rather than an error - most members
// simply have none.
func (s *ARService) ActiveARService(ctx *gin.Context, clientID, branchID string) ([]models.AccountReceivable, error) {
	if clientID == "" {
		return []models.AccountReceivable{}, nil
	}

	return s.repo.FindActiveAR(ctx, clientID, branchID)
}
