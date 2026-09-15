package repository

import (
	"database/sql"

	models "fairchild_be/internal/models/activity"

	"github.com/gin-gonic/gin"
)

/**
* ActivityRepository
* @Description: Wraps the pooled MySQL connection for reads against the
* client ledger (sldtl) to compute a client's trailing-12-month savings /
* time deposit activity (deposits vs withdrawals per month).
**/
type ActivityRepository struct {
	DB *sql.DB
}

func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{DB: db}
}

// monthlyActivityQuery returns, for the trailing 12 calendar months
// (including the current one), the SAVINGS (SLC_CODE 22) + TIME DEPOSIT
// (SLC_CODE 24) deposit/withdrawal split for one client/branch. In the raw
// ledger a negative AMT is a deposit/credit posting and a positive AMT is a
// withdrawal/debit posting - deposit_amt and withdrawal_amt are both
// reported as positive magnitudes, with net_amt as their difference.
// Placeholders are positional (?); args must be passed in the order they
// appear below: branchID, clientID.
const monthlyActivityQuery = `
WITH RECURSIVE months AS (
  SELECT DATE_FORMAT(CURDATE(), '%Y-%m-01') AS month_start, 0 AS n
  UNION ALL
  SELECT DATE_SUB(month_start, INTERVAL 1 MONTH), n + 1
  FROM months WHERE n < 5
),
movement_by_month AS (
  SELECT
      DATE_FORMAT(s.TR_DATE, '%Y-%m-01') AS month_start,
      SUM(CASE WHEN s.AMT < 0 THEN -s.AMT ELSE 0 END) AS deposit_amt,
      SUM(CASE WHEN s.AMT > 0 THEN  s.AMT ELSE 0 END) AS withdrawal_amt
  FROM sldtl s
  WHERE s.SL_BRCODE = ?
    AND s.SL_CLIENTID = ?
    AND s.SLC_CODE IN (22, 24)
    AND s.TR_DATE >= (SELECT MIN(month_start) FROM months)
  GROUP BY DATE_FORMAT(s.TR_DATE, '%Y-%m-01')
)
SELECT
    CAST(m.month_start AS DATE)                                  AS month_start,
    COALESCE(mb.deposit_amt, 0)                                  AS deposit_amt,
    COALESCE(mb.withdrawal_amt, 0)                                AS withdrawal_amt,
    COALESCE(mb.deposit_amt, 0) - COALESCE(mb.withdrawal_amt, 0)  AS net_amt
FROM months m
LEFT JOIN movement_by_month mb ON mb.month_start = m.month_start
ORDER BY m.month_start ASC`

func (r *ActivityRepository) FindMonthlyActivity(ctx *gin.Context, clientID, branchID string) ([]models.ActivityMonth, error) {
	rows, err := r.DB.QueryContext(ctx, monthlyActivityQuery, branchID, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	months := make([]models.ActivityMonth, 0, 12)
	for rows.Next() {
		var m models.ActivityMonth
		if err := rows.Scan(&m.MonthStart, &m.DepositAmt, &m.WithdrawalAmt, &m.NetAmt); err != nil {
			return nil, err
		}
		months = append(months, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return months, nil
}
