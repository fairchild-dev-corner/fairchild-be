package repository

import (
	"database/sql"

	models "fairchild_be/internal/models/amortization"

	"github.com/gin-gonic/gin"
)

/**
* AmortizationRepository
* @Description: Wraps the pooled MySQL connection for reads against the
* loan amortization schedule (amortsched).
**/
type AmortizationRepository struct {
	DB *sql.DB
}

func NewAmortizationRepository(db *sql.DB) *AmortizationRepository {
	return &AmortizationRepository{DB: db}
}

// amortizationScheduleQuery returns, per loan/AR reference that still has an
// outstanding sldtl balance, the earliest amortsched installment that
// hasn't been paid off yet (PRINBAL < the account's current balance) -
// i.e. what's genuinely still owed next, whether or not its calendar
// DUEDATE has already passed.
//
// This replaces an earlier version ported from the legacy PHP
// handleAmortSchedule, which instead picked the nearest row with
// DUEDATE >= CURDATE() and had no balance filter at all. That was wrong in
// two ways, both reproducible against the dev DB: (1) amortsched is a
// static snapshot here - unlike production, nothing periodically re-syncs
// it from IACCS - so closed/fully-paid loan and AR references can still
// carry future-dated schedule rows; DUEDATE >= CURDATE() happily included
// all of them, inflating the dashboard's "Next due" total to over 4x the
// real figure. (2) even for a loan that's still genuinely active, once the
// borrower falls behind, the calendar-future row is the WRONG row - the
// borrower's real next payment is the earlier, already-overdue
// installment that was skipped, which is exactly the row this member's
// SOA lists under "Amount Collectible" and what the /loans, /ar
// principal_due/interest_due formula (see loans_repository.go
// activeLoansQuery) already computes correctly.
//
// The balances CTE mirrors the loan_balance/principal_balance subqueries
// in loans_repository.go/ar_repository.go (sum of sldtl SLE_CODE=11 per
// reference) rather than joining sldtl directly here, so a reference with
// a zero balance is excluded up front instead of merely producing a
// bogus row.
const amortizationScheduleQuery = `
WITH balances AS (
    SELECT l.LoanSLC_CODE AS slc_code, l.LoanSLT_CODE AS slt_code, l.LoanREF_NO AS ref_no,
           COALESCE((SELECT SUM(s.AMT) FROM sldtl s
                      WHERE s.SL_BRCODE = l.LoanBR_CODE AND s.SL_CLIENTID = l.ClientIDLoan
                        AND s.REF_NO = l.LoanREF_NO AND s.SLC_CODE = l.LoanSLC_CODE AND s.SLT_CODE = l.LoanSLT_CODE
                        AND s.SLE_CODE = 11), 0) AS balance
      FROM loan l
     WHERE l.ClientIDLoan = ? AND l.LoanBR_CODE = ? AND l.LoanSLC_CODE = 12
    UNION ALL
    SELECT a.ARSLC_CODE, a.ARSLT_CODE, a.ARREF_NO,
           COALESCE((SELECT SUM(s.AMT) FROM sldtl s
                      WHERE s.SL_BRCODE = a.ARBR_CODE AND s.SL_CLIENTID = a.ClientIDAR
                        AND s.REF_NO = a.ARREF_NO AND s.SLC_CODE = a.ARSLC_CODE AND s.SLT_CODE = a.ARSLT_CODE
                        AND s.SLE_CODE = 11), 0)
      FROM ar a
     WHERE a.ClientIDAR = ? AND a.ARBR_CODE = ? AND a.ARSLC_CODE = 13
),
ranked AS (
    SELECT am.BR_CODE, am.CLIENTID, am.SLC_CODE, am.SLT_CODE, am.REF_NO, am.DUEDATE,
           am.PRINDUE, am.INTDUE, am.PRINBAL, am.PENFLAG,
           ROW_NUMBER() OVER (PARTITION BY am.SLC_CODE, am.SLT_CODE, am.REF_NO ORDER BY am.DUEDATE ASC) AS rn
      FROM amortsched am
      INNER JOIN balances b ON b.slc_code = am.SLC_CODE AND b.slt_code = am.SLT_CODE AND b.ref_no = am.REF_NO
     WHERE am.CLIENTID = ? AND am.BR_CODE = ?
       AND b.balance > 0
       AND am.PRINBAL < b.balance
)
SELECT BR_CODE, CLIENTID, SLC_CODE, SLT_CODE, REF_NO, DUEDATE, PRINDUE, INTDUE, PRINBAL, PENFLAG
  FROM ranked
 WHERE rn = 1
 ORDER BY REF_NO`

// loanScheduleQuery returns every amortsched row for one specific loan
// (matched by SLC_CODE/SLT_CODE/REF_NO), oldest due date first - the full
// projected installment schedule, not just the next due row.
const loanScheduleQuery = `
SELECT a.BR_CODE, a.CLIENTID, a.SLC_CODE, a.SLT_CODE, a.REF_NO, a.DUEDATE,
       a.PRINDUE, a.INTDUE, a.PRINBAL, a.PENFLAG
  FROM amortsched a
 WHERE a.BR_CODE = ? AND a.CLIENTID = ?
   AND a.SLC_CODE = ? AND a.SLT_CODE = ? AND a.REF_NO = ?
ORDER BY a.DUEDATE ASC`

func (r *AmortizationRepository) FindAmortizationSchedule(ctx *gin.Context, clientID, branchID string) ([]models.AmortizationScheduleEntry, error) {
	rows, err := r.DB.QueryContext(ctx, amortizationScheduleQuery,
		clientID, branchID, // balances CTE: loan side
		clientID, branchID, // balances CTE: ar side
		clientID, branchID, // ranked: am.CLIENTID / am.BR_CODE
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanScheduleRows(rows)
}

// FindLoanSchedule returns the full projected installment schedule for one
// specific loan the caller owns - callers must resolve clientID/branchID
// via FindClientAccountByUserID first (see FindAmortizationSchedule),
// never trusting a caller-supplied client/branch directly.
func (r *AmortizationRepository) FindLoanSchedule(ctx *gin.Context, clientID, branchID, slcCode, sltCode, refNo string) ([]models.AmortizationScheduleEntry, error) {
	rows, err := r.DB.QueryContext(ctx, loanScheduleQuery, branchID, clientID, slcCode, sltCode, refNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanScheduleRows(rows)
}

func scanScheduleRows(rows *sql.Rows) ([]models.AmortizationScheduleEntry, error) {
	schedule := make([]models.AmortizationScheduleEntry, 0)
	for rows.Next() {
		var e models.AmortizationScheduleEntry
		var prinDue, intDue, prinBal sql.NullFloat64

		if err := rows.Scan(
			&e.BranchID,
			&e.ClientID,
			&e.SLCCode,
			&e.SLTCode,
			&e.RefNo,
			&e.DueDate,
			&prinDue,
			&intDue,
			&prinBal,
			&e.PenaltyFlag,
		); err != nil {
			return nil, err
		}

		e.PrincipalDue = prinDue.Float64
		e.InterestDue = intDue.Float64
		e.PrincipalBalance = prinBal.Float64
		schedule = append(schedule, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedule, nil
}
