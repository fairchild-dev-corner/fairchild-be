package models

import "time"

// Loan is one of the caller's active loans (loan_balance > 0), as returned
// by LoansRepository.FindActiveLoans. Ported from the legacy sync API's
// fetchLoans_test query (fccmpc-sync-api handleMemberSync/handleLoansTest):
// principal_due/interest_due sum every scheduled amortsched installment due
// on/before today and net against actual sldtl payments, so they reflect
// what's actually outstanding (including any overdue backlog), not merely
// the next scheduled installment. PenaltyDue is nil once the loan matured
// over a month ago (IsMaturedOverdue true) - the portal falls back to
// "Confirm at branch" for those rather than showing a possibly-stale figure.
type Loan struct {
	BadgeNo             string     `json:"badge_no"`
	RefNo               string     `json:"ref_no"`
	LoanType            *string    `json:"loan_type"`
	ReleaseDate         *time.Time `json:"release_date"`
	LoanAmount          *float64   `json:"loan_amount"`
	Term                *string    `json:"term"`
	InterestRate        float64    `json:"interest_rate"`
	MatureDate          *time.Time `json:"mature_date"`
	LastTransactionDate *time.Time `json:"last_transaction_date"`
	LoanBalance         float64    `json:"loan_balance"`
	PrincipalDue        float64    `json:"principal_due"`
	InterestDue         float64    `json:"interest_due"`
	PenaltyDue          *float64   `json:"penalty_due"`
	IsMaturedOverdue    bool       `json:"is_matured_overdue"`
}
