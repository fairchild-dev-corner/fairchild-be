package interfaces

import (
	models "fairchild_be/internal/models/health"

	"github.com/gin-gonic/gin"
)

type HealthRepositoryInterface interface {
	FindCoopHealthScore(ctx *gin.Context, clientID, branchID string) (*models.CoopHealthScore, error)
}
