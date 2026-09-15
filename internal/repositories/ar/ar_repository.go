package repository

import (
	"database/sql"

	models "fairchild_be/internal/models/ar"

	"github.com/gin-gonic/gin"
)

/**
* ARRepository
* @Description: Wraps the pooled MySQL connection for reads against the
* accounts-receivable ledger (ar, sldtl SLC_CODE = 13).
**/
type ARRepository struct {
	DB *sql.DB
}

func NewARRepository(db *sql.DB) *ARRepository {
	return &ARRepository{DB: db}
}

// activeARQuery returns every one of the caller's accounts-receivable
// entries with an outstanding balance. Ported from the legacy sync API's
// fetchAR query (fccmpc-sync-api public/index.php:fetchAR), restructured
// to drive FROM the ar table directly (one row per REF_NO, like
// loans_repository.go activeLoansQuery drives FROM loan) instead of
// GROUP BY over sldtl - the original's GROUP BY shape hits MySQL's
// ONLY_FULL_GROUP_BY here (the driving fetchAR connection runs with
// sql_mode=” and never sees it), since several correlated subqueries in
// the SELECT list reference the left-joined ar row's non-aggregated
// columns. Verified to produce identical figures to the original grouped
// form. Same scheduled-due-vs-actually-paid approach as activeLoansQuery:
// principal_due/interest_due sum every amortsched installment due
// on/before today, net against the original AR amount plus actual sldtl
// payments; penalty_due replays the per-installment penalty formula since
// the last penalty settlement, only when there's an actual unpaid overdue
// balance.
const activeARQuery = `
SELECT
    badge_no, ref_no, ar_name, release_date, principal_amount, mature_date,
    principal_balance, principal_due, interest_due, penalty_due
FROM (
    SELECT
        c.OldClientID AS badge_no,
        CONCAT(a.ARSLC_CODE, '-', a.ARSLT_CODE, '-', a.ARREF_NO) AS ref_no,
        st.SLTypeM_DESC AS ar_name,
        a.ARTR_DATE AS release_date,
        a.ARPAMT AS principal_amount,
        a.ARMAT_DATE AS mature_date,

        (SELECT COALESCE(SUM(s.AMT), 0) FROM sldtl s
         WHERE s.SL_BRCODE = a.ARBR_CODE AND s.SL_CLIENTID = a.ClientIDAR
           AND s.REF_NO = a.ARREF_NO AND s.SLC_CODE = a.ARSLC_CODE AND s.SLT_CODE = a.ARSLT_CODE
           AND s.SLE_CODE = 11) AS principal_balance,

        GREATEST(0, (SELECT COALESCE(SUM(am.PRINDUE_PP), 0) - a.ARPAMT + COALESCE(
            (SELECT SUM(s.AMT) FROM sldtl s
             WHERE s.SL_BRCODE = a.ARBR_CODE AND s.SL_CLIENTID = a.ClientIDAR
               AND s.REF_NO = a.ARREF_NO AND s.SLC_CODE = a.ARSLC_CODE AND s.SLT_CODE = a.ARSLT_CODE
               AND s.SLE_CODE = 11), 0)
         FROM amortsched am
         WHERE am.BR_CODE = a.ARBR_CODE AND am.ClientID = a.ClientIDAR
           AND am.REF_NO = a.ARREF_NO AND am.SLC_CODE = a.ARSLC_CODE AND am.SLT_CODE = a.ARSLT_CODE
           AND am.DUEDATE <= CURDATE())) AS principal_due,

        GREATEST(0, (SELECT COALESCE(SUM(am.INTDUE_PP), 0) + COALESCE(
            (SELECT SUM(s.AMT) FROM sldtl s
             WHERE s.SL_BRCODE = a.ARBR_CODE AND s.SL_CLIENTID = a.ClientIDAR
               AND s.REF_NO = a.ARREF_NO AND s.SLC_CODE = a.ARSLC_CODE AND s.SLT_CODE = a.ARSLT_CODE
               AND s.SLE_CODE = 23), 0)
         FROM amortsched am
         WHERE am.BR_CODE = a.ARBR_CODE AND am.ClientID = a.ClientIDAR
           AND am.REF_NO = a.ARREF_NO AND am.SLC_CODE = a.ARSLC_CODE AND am.SLT_CODE = a.ARSLT_CODE
           AND am.DUEDATE <= CURDATE())) AS interest_due,

        CASE WHEN (
            (SELECT COALESCE(SUM(am2.PRINDUE_PP), 0) FROM amortsched am2
             WHERE am2.BR_CODE = a.ARBR_CODE AND am2.ClientID = a.ClientIDAR
               AND am2.REF_NO = a.ARREF_NO AND am2.SLC_CODE = a.ARSLC_CODE AND am2.SLT_CODE = a.ARSLT_CODE
               AND am2.PENFLAG = 'Y'
               AND am2.DUEDATE < (CURDATE() - INTERVAL 15 DAY))
            - a.ARPAMT
            + COALESCE(
                (SELECT SUM(s3.AMT) FROM sldtl s3
                 WHERE s3.SL_BRCODE = a.ARBR_CODE AND s3.SL_CLIENTID = a.ClientIDAR
                   AND s3.REF_NO = a.ARREF_NO AND s3.SLC_CODE = a.ARSLC_CODE AND s3.SLT_CODE = a.ARSLT_CODE
                   AND s3.SLE_CODE = 11), 0)
        ) > 0
        THEN COALESCE(
            (SELECT am.PRINDUE_PP * (COALESCE(a.ARPEN_RATE, 0) / 100 / 12) /
                CASE COALESCE(a.ARIPMT_MODE, 3)
                    WHEN 1 THEN 30
                    WHEN 2 THEN 4
                    WHEN 3 THEN 2
                    WHEN 4 THEN 1
                    WHEN 5 THEN 0.333
                    WHEN 6 THEN 0.167
                    WHEN 7 THEN 0.083
                    ELSE 2
                END
            FROM amortsched am
            WHERE am.BR_CODE = a.ARBR_CODE AND am.ClientID = a.ClientIDAR
                AND am.REF_NO = a.ARREF_NO AND am.SLC_CODE = a.ARSLC_CODE AND am.SLT_CODE = a.ARSLT_CODE
                AND am.PENFLAG = 'Y'
                AND am.DUEDATE < (CURDATE() - INTERVAL 15 DAY)
                AND am.PRINDUE_PP > 0
            ORDER BY am.DUEDATE DESC
            LIMIT 1
            ), 0)
        ELSE 0
        END AS penalty_due
    FROM ar a
    INNER JOIN client c ON c.ClientIDBrCode = a.ARBR_CODE AND c.ClientID = a.ClientIDAR
    LEFT JOIN sltype st ON st.SLTypeBR_CODE = a.ARBR_CODE AND st.SLTypeSLC_CODE = a.ARSLC_CODE AND st.SLTypeSLT_CODE = a.ARSLT_CODE
    WHERE a.ClientIDAR = ?
      AND a.ARBR_CODE = ?
      AND a.ARSLC_CODE = 13
) base
WHERE principal_balance > 0
ORDER BY release_date DESC`

// FindActiveAR returns every one of the caller's accounts-receivable
// entries that still has an outstanding balance.
func (r *ARRepository) FindActiveAR(ctx *gin.Context, clientID, branchID string) ([]models.AccountReceivable, error) {
	rows, err := r.DB.QueryContext(ctx, activeARQuery, clientID, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]models.AccountReceivable, 0)
	for rows.Next() {
		var e models.AccountReceivable
		var arName sql.NullString
		var releaseDate, matureDate sql.NullTime
		var principalAmount sql.NullFloat64

		if err := rows.Scan(
			&e.BadgeNo,
			&e.RefNo,
			&arName,
			&releaseDate,
			&principalAmount,
			&matureDate,
			&e.PrincipalBalance,
			&e.PrincipalDue,
			&e.InterestDue,
			&e.PenaltyDue,
		); err != nil {
			return nil, err
		}

		if arName.Valid {
			e.ARName = &arName.String
		}
		if releaseDate.Valid {
			e.ReleaseDate = &releaseDate.Time
		}
		if matureDate.Valid {
			e.MatureDate = &matureDate.Time
		}
		if principalAmount.Valid {
			e.PrincipalAmount = &principalAmount.Float64
		}

		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
