package repository

import (
	"database/sql"

	models "fairchild_be/internal/models/loans"

	"github.com/gin-gonic/gin"
)

/**
* LoansRepository
* @Description: Wraps the pooled MySQL connection for reads against the
* loan master (loan, amortsched, sldtl, sltype, term).
**/
type LoansRepository struct {
	DB *sql.DB
}

func NewLoansRepository(db *sql.DB) *LoansRepository {
	return &LoansRepository{DB: db}
}

// activeLoansQuery returns every one of the caller's loans (LoanSLC_CODE =
// 12) with an outstanding balance, newest release first. Ported verbatim
// (positional params swapped in for the PDO named ones) from the legacy
// sync API's fetchLoans_test query - see fccmpc-sync-api
// public/index.php:handleMemberSync/handleLoansTest, verified against a
// live member's statement of account (exact match on principal balance,
// principal due, interest due, and penalty due).
//
// principal_due/interest_due sum every amortsched installment due on/before
// today (PRINDUE_PP/INTDUE_PP), net against the original loan amount, plus
// actual sldtl payments (SLE_CODE 11/23) - i.e. what SHOULD have been paid
// by now minus what actually has, which correctly surfaces overdue
// backlog rather than just the next scheduled installment.
//
// penalty_due replays the legacy per-installment penalty formula
// (PRINDUE_PP x LoanPEN_RATE/100/12 x mode-factor) for each overdue,
// still-outstanding installment since the last penalty settlement
// (SLE_CODE 41/42), but is forced to NULL once the loan matured over a
// month ago (is_matured_overdue) so the portal shows "Confirm at branch"
// instead of a potentially-stale figure - same fallback the legacy portal
// uses.
const activeLoansQuery = `
SELECT
    badge_no, ref_no, loan_type, release_date, loan_amount, term,
    interest_rate, mature_date, last_transaction_date, loan_balance,
    principal_due, interest_due,
    CASE WHEN is_matured_overdue = 1 THEN NULL ELSE COALESCE(pendue_current, 0) END AS pendue,
    is_matured_overdue
FROM (
    SELECT
        c.OldClientID AS badge_no,
        CONCAT(l.LoanSLC_CODE, '-', l.LoanSLT_CODE, '-', l.LoanREF_NO) AS ref_no,
        st.SLTypeM_DESC AS loan_type,
        l.LoanTR_DATE AS release_date,
        l.LoanPAMT AS loan_amount,
        CONCAT(l.LoanTERMS, ' ', t.TermDesc) AS term,
        l.LoanINT_RATE AS interest_rate,
        l.LoanMAT_DATE AS mature_date,

        CASE WHEN l.LoanMAT_DATE < (CURDATE() - INTERVAL 1 MONTH) THEN 1 ELSE 0 END AS is_matured_overdue,

        (SELECT MAX(s.TR_DATE) FROM sldtl s
         WHERE s.SL_BRCODE = l.LoanBR_CODE AND s.SL_CLIENTID = l.ClientIDLoan
           AND s.REF_NO = l.LoanREF_NO AND s.SLC_CODE = l.LoanSLC_CODE AND s.SLT_CODE = l.LoanSLT_CODE
           AND s.SLE_CODE = 11) AS last_transaction_date,

        (SELECT COALESCE(SUM(s.AMT), 0) FROM sldtl s
         WHERE s.SL_BRCODE = l.LoanBR_CODE AND s.SL_CLIENTID = l.ClientIDLoan
           AND s.REF_NO = l.LoanREF_NO AND s.SLC_CODE = l.LoanSLC_CODE AND s.SLT_CODE = l.LoanSLT_CODE
           AND s.SLE_CODE = 11) AS loan_balance,

        GREATEST(0, (SELECT COALESCE(SUM(a.PRINDUE_PP), 0) - l.LoanPAMT + COALESCE(
            (SELECT SUM(s.AMT) FROM sldtl s
             WHERE s.SL_BRCODE = l.LoanBR_CODE AND s.SL_CLIENTID = l.ClientIDLoan
               AND s.REF_NO = l.LoanREF_NO AND s.SLC_CODE = l.LoanSLC_CODE AND s.SLT_CODE = l.LoanSLT_CODE
               AND s.SLE_CODE = 11), 0)
         FROM amortsched a
         WHERE a.BR_CODE = l.LoanBR_CODE AND a.ClientID = l.ClientIDLoan
           AND a.REF_NO = l.LoanREF_NO AND a.SLC_CODE = l.LoanSLC_CODE AND a.SLT_CODE = l.LoanSLT_CODE
           AND a.DUEDATE <= CURDATE())) AS principal_due,

        GREATEST(0, (SELECT COALESCE(SUM(a.INTDUE_PP), 0) + COALESCE(
            (SELECT SUM(s.AMT) FROM sldtl s
             WHERE s.SL_BRCODE = l.LoanBR_CODE AND s.SL_CLIENTID = l.ClientIDLoan
               AND s.REF_NO = l.LoanREF_NO AND s.SLC_CODE = l.LoanSLC_CODE AND s.SLT_CODE = l.LoanSLT_CODE
               AND s.SLE_CODE = 23), 0)
         FROM amortsched a
         WHERE a.BR_CODE = l.LoanBR_CODE AND a.ClientID = l.ClientIDLoan
           AND a.REF_NO = l.LoanREF_NO AND a.SLC_CODE = l.LoanSLC_CODE AND a.SLT_CODE = l.LoanSLT_CODE
           AND a.DUEDATE <= CURDATE())) AS interest_due,

        ROUND((
            SELECT SUM(a.PRINDUE_PP * l.LoanPEN_RATE / 100 / 12 *
                CASE l.LoanPEN_MODE
                    WHEN 1 THEN 1/30 WHEN 2 THEN 1/4 WHEN 3 THEN 1/2 WHEN 4 THEN 1
                    WHEN 5 THEN 3 WHEN 6 THEN 6 WHEN 7 THEN 12 ELSE 0.5 END)
            FROM amortsched a
            WHERE a.BR_CODE = l.LoanBR_CODE AND a.ClientID = l.ClientIDLoan
              AND a.REF_NO = l.LoanREF_NO AND a.SLC_CODE = l.LoanSLC_CODE AND a.SLT_CODE = l.LoanSLT_CODE
              AND a.PENFLAG = 'Y'
              AND a.DUEDATE >= DATE_SUB(COALESCE(
                  (SELECT MAX(s.TR_DATE) FROM sldtl s
                   WHERE s.SL_BRCODE = l.LoanBR_CODE AND s.SL_CLIENTID = l.ClientIDLoan
                     AND s.REF_NO = l.LoanREF_NO AND s.SLC_CODE = l.LoanSLC_CODE AND s.SLT_CODE = l.LoanSLT_CODE
                     AND s.SLE_CODE IN (41, 42) AND s.TR_DATE < CURDATE()), '1900-01-01'), INTERVAL 1 DAY)
              AND DATEDIFF(CURDATE(), a.DUEDATE) > l.LoanPENGP
              AND a.PRINBAL < (SELECT SUM(x.AMT) FROM sldtl x
                  WHERE x.SL_BRCODE = l.LoanBR_CODE AND x.SL_CLIENTID = l.ClientIDLoan
                    AND x.REF_NO = l.LoanREF_NO AND x.SLC_CODE = l.LoanSLC_CODE AND x.SLT_CODE = l.LoanSLT_CODE
                    AND x.SLE_CODE = 11 AND x.TR_DATE < CURDATE())
        ), 2) AS pendue_current

    FROM loan l
    INNER JOIN client c ON c.ClientIDBrCode = l.LoanBR_CODE AND c.ClientID = l.ClientIDLoan
    LEFT JOIN sltype st ON st.SLTypeBR_CODE = l.LoanBR_CODE AND st.SLTypeSLC_CODE = l.LoanSLC_CODE AND st.SLTypeSLT_CODE = l.LoanSLT_CODE
    LEFT JOIN term t ON t.TermID = l.LoanTERM_PERD
    WHERE l.ClientIDLoan = ?
      AND l.LoanBR_CODE = ?
      AND l.LoanSLC_CODE = 12
    HAVING loan_balance > 0
) base
ORDER BY release_date DESC`

// FindActiveLoans returns every one of the caller's loans that still has an
// outstanding balance, newest release first.
func (r *LoansRepository) FindActiveLoans(ctx *gin.Context, clientID, branchID string) ([]models.Loan, error) {
	rows, err := r.DB.QueryContext(ctx, activeLoansQuery, clientID, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	loans := make([]models.Loan, 0)
	for rows.Next() {
		var l models.Loan
		var loanType, term sql.NullString
		var releaseDate, matureDate, lastTxnDate sql.NullTime
		var loanAmount sql.NullFloat64
		var penaltyDue sql.NullFloat64
		var isMaturedOverdue int64

		if err := rows.Scan(
			&l.BadgeNo,
			&l.RefNo,
			&loanType,
			&releaseDate,
			&loanAmount,
			&term,
			&l.InterestRate,
			&matureDate,
			&lastTxnDate,
			&l.LoanBalance,
			&l.PrincipalDue,
			&l.InterestDue,
			&penaltyDue,
			&isMaturedOverdue,
		); err != nil {
			return nil, err
		}
		l.IsMaturedOverdue = isMaturedOverdue != 0

		if loanType.Valid {
			l.LoanType = &loanType.String
		}
		if term.Valid {
			l.Term = &term.String
		}
		if releaseDate.Valid {
			l.ReleaseDate = &releaseDate.Time
		}
		if matureDate.Valid {
			l.MatureDate = &matureDate.Time
		}
		if lastTxnDate.Valid {
			l.LastTransactionDate = &lastTxnDate.Time
		}
		if loanAmount.Valid {
			l.LoanAmount = &loanAmount.Float64
		}
		if penaltyDue.Valid {
			l.PenaltyDue = &penaltyDue.Float64
		}

		loans = append(loans, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return loans, nil
}
