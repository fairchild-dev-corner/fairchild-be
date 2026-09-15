package interfaces

import (
	models "fairchild_be/internal/models/soa"

	"github.com/gin-gonic/gin"
)

type SOARepositoryInterface interface {
	FindSOAHistory(ctx *gin.Context, clientID, branchID string) ([]models.SOAHistoryEntry, error)
	FindSOAStatement(ctx *gin.Context, clientID, branchID string, ctrlNo int64) (*models.SOAStatement, error)
}
