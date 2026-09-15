package models

import "time"

// AccountReceivable is one of the caller's active accounts receivable
// entries (outstanding balance > 0) - a separate ledger type from regular
// loans (SLC_CODE = 13 vs 12), e.g. staff/other receivables. Ported from
// the legacy sync API's fetchAR query - see fccmpc-sync-api
// public/index.php:fetchAR. Shape mirrors models/loans.Loan so both can
// render in the same table.
type AccountReceivable struct {
	BadgeNo          string     `json:"badge_no"`
	RefNo            string     `json:"ref_no"`
	ARName           *string    `json:"ar_name"`
	ReleaseDate      *time.Time `json:"release_date"`
	PrincipalAmount  *float64   `json:"principal_amount"`
	MatureDate       *time.Time `json:"mature_date"`
	PrincipalBalance float64    `json:"principal_balance"`
	PrincipalDue     float64    `json:"principal_due"`
	InterestDue      float64    `json:"interest_due"`
	PenaltyDue       float64    `json:"penalty_due"`
}
