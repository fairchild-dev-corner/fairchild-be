package interfaces

import (
	models "fairchild_be/internal/models/ar"

	"github.com/gin-gonic/gin"
)

type ARRepositoryInterface interface {
	FindActiveAR(ctx *gin.Context, clientID, branchID string) ([]models.AccountReceivable, error)
}
