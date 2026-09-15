package services

import (
	cc "fairchild_be/internal/constants"
	interfaces "fairchild_be/internal/interfaces"
	models "fairchild_be/internal/models/soa"

	"github.com/gin-gonic/gin"
)

type SOAService struct {
	repo interfaces.SOARepositoryInterface
}

func NewSOAService(repo interfaces.SOARepositoryInterface) *SOAService {
	return &SOAService{repo: repo}
}

// SOAHistoryService returns the authenticated caller's own SOA history
// (last 3 months, or their single most recent statement if none fall in
// that window). clientID/branchID are already resolved by
// middlewares.RequireMemberAuth - a caller with no linked ledger account
// (clientID == "") gets an empty slice rather than an error - there's
// simply nothing to show yet.
func (s *SOAService) SOAHistoryService(ctx *gin.Context, clientID, branchID string) ([]models.SOAHistoryEntry, error) {
	if clientID == "" {
		return []models.SOAHistoryEntry{}, nil
	}

	return s.repo.FindSOAHistory(ctx, clientID, branchID)
}

// SOAStatementService returns one full statement the caller owns.
// clientID/branchID are already resolved by middlewares.RequireMemberAuth -
// ctrlNo alone is never enough to fetch someone else's document. An
// unlinked caller, or a ctrlNo that isn't theirs (or doesn't exist), both
// surface as cc.ErrSOAStatementNotFound so neither case can be used to
// probe for other members' statement numbers.
func (s *SOAService) SOAStatementService(ctx *gin.Context, clientID, branchID string, ctrlNo int64) (*models.SOAStatement, error) {
	if clientID == "" {
		return nil, cc.ErrSOAStatementNotFound
	}

	stmt, err := s.repo.FindSOAStatement(ctx, clientID, branchID, ctrlNo)
	if err != nil {
		return nil, err
	}
	if stmt == nil {
		return nil, cc.ErrSOAStatementNotFound
	}

	return stmt, nil
}
