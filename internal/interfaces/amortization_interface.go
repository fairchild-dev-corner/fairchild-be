package interfaces

import (
	models "fairchild_be/internal/models/amortization"

	"github.com/gin-gonic/gin"
)

type AmortizationRepositoryInterface interface {
	FindAmortizationSchedule(ctx *gin.Context, clientID, branchID string) ([]models.AmortizationScheduleEntry, error)
	FindLoanSchedule(ctx *gin.Context, clientID, branchID, slcCode, sltCode, refNo string) ([]models.AmortizationScheduleEntry, error)
}
