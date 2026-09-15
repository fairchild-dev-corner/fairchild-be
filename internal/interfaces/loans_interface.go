package interfaces

import (
	models "fairchild_be/internal/models/loans"

	"github.com/gin-gonic/gin"
)

type LoansRepositoryInterface interface {
	FindActiveLoans(ctx *gin.Context, clientID, branchID string) ([]models.Loan, error)
}
