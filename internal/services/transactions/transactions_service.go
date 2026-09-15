package services

import (
	interfaces "fairchild_be/internal/interfaces"
	models "fairchild_be/internal/models/transactions"

	"github.com/gin-gonic/gin"
)

type TransactionsService struct {
	repo interfaces.TransactionsRepositoryInterface
}

func NewTransactionsService(repo interfaces.TransactionsRepositoryInterface) *TransactionsService {
	return &TransactionsService{repo: repo}
}

// LastTransactionsService returns the authenticated caller's own last 10
// transactions. clientID/branchID are already resolved by
// middlewares.RequireMemberAuth - a caller with no linked ledger account
// (clientID == "") gets an empty list rather than an error - there's simply
// nothing to show yet.
func (s *TransactionsService) LastTransactionsService(ctx *gin.Context, clientID, branchID string) ([]models.Transaction, error) {
	if clientID == "" {
		return []models.Transaction{}, nil
	}

	return s.repo.FindLastTransactions(ctx, clientID, branchID)
}

// TransactionsByRefService returns every posting for one specific loan or
// AR account the authenticated caller owns (identified by its
// SLC_CODE-SLT_CODE-REF_NO triple, as shown in LoanDto.ref_no /
// AccountReceivableDto.ref_no) - full history, not LastTransactionsService's
// last-10/3-month cap. clientID/branchID are already resolved by
// middlewares.RequireMemberAuth - a caller with no linked ledger account
// (clientID == "") gets an empty list rather than an error.
func (s *TransactionsService) TransactionsByRefService(ctx *gin.Context, clientID, branchID, slcCode, sltCode, refNo string) ([]models.Transaction, error) {
	if clientID == "" {
		return []models.Transaction{}, nil
	}

	return s.repo.FindTransactionsByRef(ctx, clientID, branchID, slcCode, sltCode, refNo)
}
