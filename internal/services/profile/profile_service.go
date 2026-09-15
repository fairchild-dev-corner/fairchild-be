package services

import (
	interfaces "fairchild_be/internal/interfaces"
	models "fairchild_be/internal/models/profile"

	"github.com/gin-gonic/gin"
)

type ProfileService struct {
	repo interfaces.ProfileRepositoryInterface
}

func NewProfileService(repo interfaces.ProfileRepositoryInterface) *ProfileService {
	return &ProfileService{repo: repo}
}

// ProfileService returns the authenticated caller's own realtime profile. A
// caller with no linked member id, or whose member id has no matching
// active client record, gets nil rather than an error - there's simply no
// profile to show.
func (s *ProfileService) ProfileService(ctx *gin.Context, userID int64) (*models.Profile, error) {
	memberID, err := s.repo.FindMemberIDByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if memberID == "" {
		return nil, nil
	}

	return s.repo.FindProfileByMemberID(ctx, userID, memberID)
}
