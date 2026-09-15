package interfaces

import (
	models "fairchild_be/internal/models/transactions"

	"github.com/gin-gonic/gin"
)

type TransactionsRepositoryInterface interface {
	FindLastTransactions(ctx *gin.Context, clientID, branchID string) ([]models.Transaction, error)
	FindTransactionsByRef(ctx *gin.Context, clientID, branchID, slcCode, sltCode, refNo string) ([]models.Transaction, error)
}
