package repository

import (
	"database/sql"

	models "fairchild_be/internal/models/balance"

	"github.com/gin-gonic/gin"
)

/**
* BalanceRepository
* @Description: Wraps the pooled MySQL connection for reads against the
* client ledger (sldtl) to compute a client's current deposit balances,
* broken out per sub-product (SAVINGS, TIME DEPOSIT, SHARE CAPITAL).
**/
type BalanceRepository struct {
	DB *sql.DB
}

func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{DB: db}
}

// depositBalancesQuery sums every all-time posting for one client/branch,
// grouped per sub-product actually posted against (SLC_CODE/SLT_CODE),
// across SAVINGS (SLC_CODE 22 - Regular, Aflatoun, Kiddies, Plus,
// Retirement, Youth, Equity are all distinct SLT_CODE rows here), TIME
// DEPOSIT (24), and SHARE CAPITAL (31). SLE_CODE = 11 is the "principal"
// ledger entry type (11=Principal/23=Interest/41=Penalty, same mapping as
// loan postings) - verified DB-wide that every SAVINGS/TIME DEPOSIT/SHARE
// CAPITAL posting uses SLE_CODE 11 exclusively, so this filter is
// currently a no-op but documents the intent defensively. In the raw
// ledger a negative AMT is a deposit/credit posting and a positive AMT is
// a withdrawal/debit posting, so AMT * -1 yields the running balance
// directly. The client-eligibility filter (regular members, or Kiddies/
// Aflatoun minors under 18) mirrors TransactionsRepository's
// lastTransactionsQuery. Placeholders are positional (?); args must be
// passed in the order they appear below: branchID, clientID.
const depositBalancesQuery = `
SELECT
    MAX(c.OldClientID) AS member_id,
    s.SLC_CODE,
    s.SLT_CODE,
    MAX(st.SLTypeM_DESC) AS deposit_name,
    SUM(s.AMT * -1) AS deposit_amount
FROM sldtl s
LEFT JOIN client c ON c.ClientID = s.SL_CLIENTID
                  AND c.ClientIDBrCode = s.SL_BRCODE
LEFT JOIN sltype st ON st.SLTypeBR_CODE  = s.SL_BRCODE
                   AND st.SLTypeSLC_CODE = s.SLC_CODE
                   AND st.SLTypeSLT_CODE = s.SLT_CODE
WHERE s.SL_BRCODE = ?
  AND c.ClientID = ?
  AND s.SLC_CODE IN (22, 31, 24)
  AND s.SLE_CODE = 11
  AND s.TR_DATE <= CURDATE()
  AND (
        c.ClientType IN (1, 2)
     OR (
            c.ClientType IN (3, 4)
        AND c.AccountType IN (SELECT AccTypeID FROM accttype WHERE TRIM(UPPER(AcctTypeDesc)) IN ('KIDDIES','AFLATOUN'))
        AND TIMESTAMPDIFF(YEAR, c.DateOfBirth, CURDATE()) < 18
        )
  )
GROUP BY s.SLC_CODE, s.SLT_CODE
ORDER BY FIELD(s.SLC_CODE, 22, 31, 24), deposit_name`

func (r *BalanceRepository) FindDepositBalances(ctx *gin.Context, clientID, branchID string) ([]models.DepositBalance, error) {
	rows, err := r.DB.QueryContext(ctx, depositBalancesQuery, branchID, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balances := make([]models.DepositBalance, 0)
	for rows.Next() {
		var b models.DepositBalance
		if err := rows.Scan(&b.MemberID, &b.SLCCode, &b.SLTCode, &b.DepositName, &b.DepositAmount); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return balances, nil
}
