package interfaces

import (
	models "fairchild_be/internal/models/profile"

	"github.com/gin-gonic/gin"
)

type ProfileRepositoryInterface interface {
	FindMemberIDByUserID(ctx *gin.Context, userID int64) (memberID string, err error)
	FindProfileByMemberID(ctx *gin.Context, userID int64, memberID string) (*models.Profile, error)
}
