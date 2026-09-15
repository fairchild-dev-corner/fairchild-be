package interfaces

import (
	models "fairchild_be/internal/models/activity"

	"github.com/gin-gonic/gin"
)

type ActivityRepositoryInterface interface {
	FindMonthlyActivity(ctx *gin.Context, clientID, branchID string) ([]models.ActivityMonth, error)
}
