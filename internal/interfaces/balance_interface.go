package interfaces

import (
	models "fairchild_be/internal/models/balance"

	"github.com/gin-gonic/gin"
)

type BalanceRepositoryInterface interface {
	FindDepositBalances(ctx *gin.Context, clientID, branchID string) ([]models.DepositBalance, error)
}
