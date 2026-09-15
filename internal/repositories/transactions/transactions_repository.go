package repository

import (
	"database/sql"

	models "fairchild_be/internal/models/transactions"

	"github.com/gin-gonic/gin"
)

/**
* TransactionsRepository
* @Description: Wraps the pooled MySQL connection for reads against the
* client ledger (sldtl/client/sltype/accttype).
**/
type TransactionsRepository struct {
	DB *sql.DB
}

func NewTransactionsRepository(db *sql.DB) *TransactionsRepository {
	return &TransactionsRepository{DB: db}
}

// lastTransactionsQuery returns a client's most recent loan/savings/time
// deposit/share capital postings from the last 3 months, restricted to
// regular members (ClientType 1,2) or Kiddies/Aflatoun minors (ClientType
// 3,4 with a matching account type and age under 18).
const lastTransactionsQuery = `
SELECT
    CONCAT(c.OldClientID, '_', s.SL_BRCODE, '_', s.SLC_CODE, '_', s.SLT_CODE, '_', s.SLE_CODE, '_', s.REF_NO, '_', s.SEQNO, '_', s.TR_DATE) AS dt_refID,
    c.OldClientID AS m_id,
    CASE s.SLC_CODE
        WHEN 12 THEN 'LOAN'
        WHEN 22 THEN 'SAVINGS'
        WHEN 24 THEN 'TIME DEPOSIT'
        WHEN 31 THEN 'SHARE CAPITAL'
    END AS txn_category,
    st.SLTypeM_DESC AS txn_name,
    s.AMT * -1 AS txn_amount,
    s.TR_DATE AS txn_date
FROM sldtl s
INNER JOIN client c ON c.ClientID = s.SL_CLIENTID
                   AND c.ClientIDBrCode = s.SL_BRCODE
LEFT JOIN sltype st ON st.SLTypeBR_CODE  = s.SL_BRCODE
                   AND st.SLTypeSLC_CODE = s.SLC_CODE
                   AND st.SLTypeSLT_CODE = s.SLT_CODE
WHERE s.SL_BRCODE = ?
  AND c.ClientID = ?
  AND s.SLC_CODE IN (12, 22, 24, 31)
  AND s.TR_DATE >= CURDATE() - INTERVAL 3 MONTH
  AND (
        c.ClientType IN (1, 2)
     OR (
            c.ClientType IN (3, 4)
        AND c.AccountType IN (SELECT AccTypeID FROM accttype WHERE TRIM(UPPER(AcctTypeDesc)) IN ('KIDDIES','AFLATOUN'))
        AND TIMESTAMPDIFF(YEAR, c.DateOfBirth, CURDATE()) < 18
        )
  )
ORDER BY s.TR_DATE DESC LIMIT 10`

func (r *TransactionsRepository) FindLastTransactions(ctx *gin.Context, clientID, branchID string) ([]models.Transaction, error) {
	rows, err := r.DB.QueryContext(ctx, lastTransactionsQuery, branchID, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTransactionRows(rows)
}

// transactionsByRefQuery returns every posting against one specific loan or
// AR account (matched by SLC_CODE/SLT_CODE/REF_NO, the same triple exposed
// as LoanDto.ref_no / AccountReceivableDto.ref_no), full history rather
// than lastTransactionsQuery's last-10/3-month cap. SLC_CODE 13 (AR) is
// included here even though lastTransactionsQuery excludes it, since a
// caller-selected AR account should show its own postings regardless of
// which SLC types the general "recent activity" feed surfaces.
const transactionsByRefQuery = `
SELECT
    CONCAT(c.OldClientID, '_', s.SL_BRCODE, '_', s.SLC_CODE, '_', s.SLT_CODE, '_', s.SLE_CODE, '_', s.REF_NO, '_', s.SEQNO, '_', s.TR_DATE) AS dt_refID,
    c.OldClientID AS m_id,
    CASE s.SLC_CODE
        WHEN 12 THEN 'LOAN'
        WHEN 13 THEN 'A/R'
        WHEN 22 THEN 'SAVINGS'
        WHEN 24 THEN 'TIME DEPOSIT'
        WHEN 31 THEN 'SHARE CAPITAL'
        ELSE 'OTHER'
    END AS txn_category,
    st.SLTypeM_DESC AS txn_name,
    s.AMT * -1 AS txn_amount,
    s.TR_DATE AS txn_date
FROM sldtl s
INNER JOIN client c ON c.ClientID = s.SL_CLIENTID
                   AND c.ClientIDBrCode = s.SL_BRCODE
LEFT JOIN sltype st ON st.SLTypeBR_CODE  = s.SL_BRCODE
                   AND st.SLTypeSLC_CODE = s.SLC_CODE
                   AND st.SLTypeSLT_CODE = s.SLT_CODE
WHERE s.SL_BRCODE = ?
  AND c.ClientID = ?
  AND s.SLC_CODE = ?
  AND s.SLT_CODE = ?
  AND s.REF_NO = ?
ORDER BY s.TR_DATE DESC`

// FindTransactionsByRef returns every posting for one specific loan/AR
// account the caller owns - callers must resolve clientID/branchID via
// FindClientAccountByUserID first (see FindLastTransactions), never
// trusting a caller-supplied client/branch directly.
func (r *TransactionsRepository) FindTransactionsByRef(ctx *gin.Context, clientID, branchID, slcCode, sltCode, refNo string) ([]models.Transaction, error) {
	rows, err := r.DB.QueryContext(ctx, transactionsByRefQuery, branchID, clientID, slcCode, sltCode, refNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTransactionRows(rows)
}

func scanTransactionRows(rows *sql.Rows) ([]models.Transaction, error) {
	txns := make([]models.Transaction, 0)
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.RefID, &t.MemberID, &t.Category, &t.Name, &t.Amount, &t.Date); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return txns, nil
}
