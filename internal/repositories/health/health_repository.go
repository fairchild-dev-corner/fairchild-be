package repository

import (
	"database/sql"

	models "fairchild_be/internal/models/health"

	"github.com/gin-gonic/gin"
)

/**
* HealthRepository
* @Description: Wraps the pooled MySQL connection for reads against the
* client ledger (sldtl/loan) to compute a client's rolling coop health score.
**/
type HealthRepository struct {
	DB *sql.DB
}

func NewHealthRepository(db *sql.DB) *HealthRepository {
	return &HealthRepository{DB: db}
}

// coopHealthScoreQuery computes, over the trailing 12 calendar months for one
// client/branch:
//   - share_capital_buildup_pct / savings_consistency_pct: the percentage of
//     those 12 months with a net-positive posting (SLC_CODE 31 / 22).
//   - loan_repayment_pct: the percentage of paid loan amounts on loans not in
//     a written-off/closed-bad status (LoanSTATUS 12/13), defaulting to 100
//     when the client has no loans at all.
//   - coop_health_score: the average of the three.
//
// Placeholders are positional (?) in place of the original named
// (:client_id / :branch_id) params; args must be passed in the exact order
// they appear below: branchID, clientID, branchID, clientID, branchID,
// clientID, clientID, branchID.
const coopHealthScoreQuery = `
WITH RECURSIVE months AS (
  SELECT DATE_FORMAT(CURDATE(), '%Y-%m-01') AS month_start, 0 AS n
  UNION ALL
  SELECT DATE_SUB(month_start, INTERVAL 1 MONTH), n + 1
  FROM months WHERE n < 11
),
savings_by_month AS (
  SELECT DATE_FORMAT(s.TR_DATE, '%Y-%m-01') AS month_start, SUM(s.AMT * -1) AS net_amt
  FROM sldtl s
  WHERE s.SL_BRCODE = ?
    AND s.SL_CLIENTID = ?
    AND s.SLC_CODE = 22
    AND s.TR_DATE >= (SELECT MIN(month_start) FROM months)
  GROUP BY DATE_FORMAT(s.TR_DATE, '%Y-%m-01')
),
sharecap_by_month AS (
  SELECT DATE_FORMAT(s.TR_DATE, '%Y-%m-01') AS month_start, SUM(s.AMT * -1) AS net_amt
  FROM sldtl s
  WHERE s.SL_BRCODE = ?
    AND s.SL_CLIENTID = ?
    AND s.SLC_CODE = 31
    AND s.TR_DATE >= (SELECT MIN(month_start) FROM months)
  GROUP BY DATE_FORMAT(s.TR_DATE, '%Y-%m-01')
),
savings_score AS (
  SELECT ROUND(100.0 * SUM(CASE WHEN sb.net_amt > 0 THEN 1 ELSE 0 END) / COUNT(*), 0) AS pct
  FROM months m LEFT JOIN savings_by_month sb ON sb.month_start = m.month_start
),
sharecap_score AS (
  SELECT ROUND(100.0 * SUM(CASE WHEN sc.net_amt > 0 THEN 1 ELSE 0 END) / COUNT(*), 0) AS pct
  FROM months m LEFT JOIN sharecap_by_month sc ON sc.month_start = m.month_start
),
loan_repayment AS (
  SELECT
    COALESCE(ROUND(100.0 * SUM(CASE WHEN l.LoanSTATUS NOT IN (12,13) THEN l.LoanPAMT ELSE 0 END)
                   / NULLIF(SUM(l.LoanPAMT), 0), 0), 100) AS pct
  FROM loan l
  WHERE l.LoanBR_CODE = ?
    AND l.ClientIDLoan = ?
    AND l.LoanTR_DATE <= CURDATE()
    AND (l.LoanSTATUS <> 17 OR l.LoanMAT_DATE >= (SELECT MIN(month_start) FROM months))
)
SELECT
    ? AS client_id,
    ? AS branch_id,
    sc.pct AS share_capital_buildup_pct,
    sv.pct AS savings_consistency_pct,
    lr.pct AS loan_repayment_pct,
    ROUND((sc.pct + sv.pct + lr.pct) / 3, 0) AS coop_health_score
FROM sharecap_score sc, savings_score sv, loan_repayment lr`

func (r *HealthRepository) FindCoopHealthScore(ctx *gin.Context, clientID, branchID string) (*models.CoopHealthScore, error) {
	var score models.CoopHealthScore
	err := r.DB.QueryRowContext(ctx, coopHealthScoreQuery,
		branchID, clientID,
		branchID, clientID,
		branchID, clientID,
		clientID, branchID,
	).Scan(
		&score.ClientID,
		&score.BranchID,
		&score.ShareCapitalBuildupPct,
		&score.SavingsConsistencyPct,
		&score.LoanRepaymentPct,
		&score.CoopHealthScore,
	)
	if err != nil {
		return nil, err
	}

	return &score, nil
}
