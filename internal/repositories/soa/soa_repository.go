package repository

import (
	"database/sql"
	"time"

	models "fairchild_be/internal/models/soa"

	"github.com/gin-gonic/gin"
)

/**
* SOARepository
* @Description: Wraps the pooled MySQL connection for reads against
* wkfcsoa - the legacy system's cache of pre-generated Statement of Account
* documents. Each row is one line item (one SEQNO) of one generated
* statement (one CTRLNO, the literal "SOA No." printed on the legacy
* document).
**/
type SOARepository struct {
	DB *sql.DB
}

func NewSOARepository(db *sql.DB) *SOARepository {
	return &SOARepository{DB: db}
}

// soaHistoryWindowQuery groups the caller's own wkfcsoa rows by CTRLNO -
// never by STATEMENTDATE, which is not unique per client (two distinct
// CTRLNOs can share a STATEMENTDATE with different CUTOFFDATEs and
// different line items - grouping by STATEMENTDATE would silently merge
// two separate generated documents into one row). CUTOFFDATE <= CURDATE()
// excludes forward-projected statements unconditionally: CUTOFFDATE, not
// STATEMENTDATE, is the true "as of" date the DUE figures reflect (verified
// against a real row where STATEMENTDATE and CUTOFFDATE differ and the
// figures match the CUTOFFDATE), and a meaningful fraction of wkfcsoa rows
// carry a future CUTOFFDATE - collections "what if nothing more is paid"
// projections, never something a member should see as "their SOA". CTRLNO
// is not time-monotonic either (a verified low CTRLNO carries the single
// furthest-future CUTOFFDATE in that client's history), so ordering is by
// cutoff_date, with CTRLNO only as a tiebreaker. MIN() on
// STATEMENTDATE/CUTOFFDATE is a deterministic single-value pick, not a real
// aggregation - every SEQNO row sharing a CTRLNO carries identical values
// for both.
const soaHistoryWindowQuery = `
SELECT
    CTRLNO,
    MIN(STATEMENTDATE) AS statement_date,
    MIN(CUTOFFDATE) AS cutoff_date,
    SUM(COALESCE(PRINDUE,0) + COALESCE(INTDUE,0) + COALESCE(PENDUE,0) + COALESCE(INSDUE,0)) AS total_due
FROM wkfcsoa
WHERE CLIENTIDx = ? AND BR_CODE = ?
  AND CUTOFFDATE <= CURDATE()
  AND CUTOFFDATE >= DATE_SUB(CURDATE(), INTERVAL 3 MONTH)
GROUP BY CTRLNO
ORDER BY cutoff_date DESC, CTRLNO DESC`

// soaHistoryFallbackQuery runs only when the 3-month window is empty - it
// picks the single most recent pre-today CTRLNO regardless of age, so a
// member whose last statement predates the window still sees their most
// recent one instead of a blank page.
const soaHistoryFallbackQuery = `
SELECT
    CTRLNO,
    MIN(STATEMENTDATE) AS statement_date,
    MIN(CUTOFFDATE) AS cutoff_date,
    SUM(COALESCE(PRINDUE,0) + COALESCE(INTDUE,0) + COALESCE(PENDUE,0) + COALESCE(INSDUE,0)) AS total_due
FROM wkfcsoa
WHERE CLIENTIDx = ? AND BR_CODE = ?
  AND CTRLNO = (
      SELECT CTRLNO FROM wkfcsoa
       WHERE CLIENTIDx = ? AND BR_CODE = ? AND CUTOFFDATE <= CURDATE()
       ORDER BY CUTOFFDATE DESC, CTRLNO DESC
       LIMIT 1
  )
GROUP BY CTRLNO`

// soaStatementDetailQuery returns every SEQNO line of one specific
// statement, scoped to the caller's own CLIENTIDx/BR_CODE - a foreign or
// nonexistent CTRLNO simply yields zero rows (the service maps that to a
// 404), never another member's data. TRIM() on ACCOUNTGROUP matches the
// history query - see wkfcsoa's leading-space " DEPOSITS" value.
// MATURITYDATE/MEMBERSHIPDATE are free-text varchar columns in wkfcsoa
// (not real DATE columns), so they're carried through as strings rather
// than parsed. CUTOFFDATE <= CURDATE() is enforced here too, not just in
// the history query - a member should never be able to view a
// forward-projected statement (see soaHistoryWindowQuery's comment) even
// by requesting its ctrl_no directly, e.g. via a stale link or by
// incrementing a known real one.
const soaStatementDetailQuery = `
SELECT
    SEQNO, TRIM(ACCOUNTGROUP), SLDESCR, REF_NO,
    PRINAMT, PRINBAL, AVBLBAL, PRINDUE, INTDUE, PENDUE, INSDUE,
    PRINOVR, INTOVR, PENOVR, ACCOUNTSTATUS, MATURITYDATE, TERM,
    CLIENTID, CLIENTNAME, HOMEADDRESS, MS_MR, STATEMENTDATE, CUTOFFDATE,
    BRANCHMANAGER, EMPLOYEEID, MEMBERSHIPDATE, ORGACRONYM
FROM wkfcsoa
WHERE CLIENTIDx = ? AND BR_CODE = ? AND CTRLNO = ? AND CUTOFFDATE <= CURDATE()
ORDER BY SEQNO ASC`

// FindSOAHistory returns the last 3 months of the caller's own generated
// statements, newest first, one row per CTRLNO. Falls back to the single
// latest pre-window statement if the window is empty - see
// soaHistoryFallbackQuery.
func (r *SOARepository) FindSOAHistory(ctx *gin.Context, clientID, branchID string) ([]models.SOAHistoryEntry, error) {
	entries, err := scanHistory(r.DB.QueryContext(ctx, soaHistoryWindowQuery, clientID, branchID))
	if err != nil {
		return nil, err
	}
	if len(entries) > 0 {
		return entries, nil
	}

	return scanHistory(r.DB.QueryContext(ctx, soaHistoryFallbackQuery, clientID, branchID, clientID, branchID))
}

func scanHistory(rows *sql.Rows, err error) ([]models.SOAHistoryEntry, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]models.SOAHistoryEntry, 0)
	for rows.Next() {
		var e models.SOAHistoryEntry
		if err := rows.Scan(&e.CtrlNo, &e.StatementDate, &e.CutoffDate, &e.TotalDue); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// FindSOAStatement returns the full line-item detail of one statement the
// caller owns. Returns (nil, nil) - not an error - when CTRLNO doesn't
// belong to this clientID/branchID (or doesn't exist at all); the service
// maps that to a 404 without distinguishing the two cases, so a foreign
// CTRLNO can't be probed to determine whether it exists.
func (r *SOARepository) FindSOAStatement(ctx *gin.Context, clientID, branchID string, ctrlNo int64) (*models.SOAStatement, error) {
	rows, err := r.DB.QueryContext(ctx, soaStatementDetailQuery, clientID, branchID, ctrlNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stmt *models.SOAStatement
	for rows.Next() {
		var line models.SOALineItem
		var sldescr, refNo, accStatus, maturityDate, term, homeAddr, msMr, branchMgr, idNo, membershipDate, org sql.NullString
		var prinAmt, prinBal, avblBal sql.NullFloat64
		var lineClientID, lineClientName string
		var statementDate, cutoffDate time.Time

		if err := rows.Scan(
			&line.SeqNo, &line.AccountGroup, &sldescr, &refNo,
			&prinAmt, &prinBal, &avblBal, &line.PrincipalDue, &line.InterestDue, &line.PenaltyDue, &line.InsuranceDue,
			&line.PrincipalOverpaid, &line.InterestOverpaid, &line.PenaltyOverpaid, &accStatus, &maturityDate, &term,
			&lineClientID, &lineClientName, &homeAddr, &msMr, &statementDate, &cutoffDate,
			&branchMgr, &idNo, &membershipDate, &org,
		); err != nil {
			return nil, err
		}

		if sldescr.Valid {
			line.SLDescr = &sldescr.String
		}
		if refNo.Valid {
			line.RefNo = &refNo.String
		}
		if prinAmt.Valid {
			line.PrincipalAmount = &prinAmt.Float64
		}
		if prinBal.Valid {
			line.PrincipalBalance = &prinBal.Float64
		}
		if avblBal.Valid {
			line.AvailableBalance = &avblBal.Float64
		}
		if accStatus.Valid {
			line.AccountStatus = &accStatus.String
		}
		if maturityDate.Valid {
			line.MaturityDate = &maturityDate.String
		}
		if term.Valid {
			line.Term = &term.String
		}

		if stmt == nil {
			stmt = &models.SOAStatement{
				CtrlNo:        ctrlNo,
				StatementDate: statementDate,
				CutoffDate:    cutoffDate,
				ClientID:      lineClientID,
				ClientName:    lineClientName,
				Lines:         make([]models.SOALineItem, 0),
			}
			if homeAddr.Valid {
				stmt.HomeAddress = &homeAddr.String
			}
			if msMr.Valid {
				stmt.MsMr = &msMr.String
			}
			if branchMgr.Valid {
				stmt.BranchManager = &branchMgr.String
			}
			if idNo.Valid {
				stmt.IDNo = &idNo.String
			}
			if membershipDate.Valid {
				stmt.MembershipDate = &membershipDate.String
			}
			if org.Valid {
				stmt.OrgAcronym = &org.String
			}
		}

		stmt.Lines = append(stmt.Lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stmt, nil
}
