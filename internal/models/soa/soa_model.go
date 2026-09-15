package models

import "time"

// SOAHistoryEntry is one summarized row of a member's SOA history list (see
// SOARepository.FindSOAHistory) - one row per CTRLNO, the literal "SOA No."
// printed on the legacy document. CutoffDate (not StatementDate) is the
// true "as of" date the TotalDue figure reflects - see the repository's
// query comments. Only statements with CutoffDate <= today are ever
// returned; future-dated/projected statements are excluded at the query
// level, never filtered client-side.
type SOAHistoryEntry struct {
	CtrlNo        int64     `json:"ctrl_no"`
	StatementDate time.Time `json:"statement_date"`
	CutoffDate    time.Time `json:"cutoff_date"`
	TotalDue      float64   `json:"total_due"`
}

// SOALineItem is one wkfcsoa row (one SEQNO) of a single statement - one
// deposit/AR/loan account line. AccountGroup is already TRIM()med by the
// query (the source data has a stray leading space on " DEPOSITS").
type SOALineItem struct {
	SeqNo             int64    `json:"seq_no"`
	AccountGroup      string   `json:"account_group"` // "DEPOSITS" | "ACCOUNTS RECEIVABLE" | "LOANS RECEIVABLE"
	SLDescr           *string  `json:"sl_descr"`
	RefNo             *string  `json:"ref_no"`
	PrincipalAmount   *float64 `json:"principal_amount"`
	PrincipalBalance  *float64 `json:"principal_balance"`
	AvailableBalance  *float64 `json:"available_balance"`
	PrincipalDue      float64  `json:"principal_due"`
	InterestDue       float64  `json:"interest_due"`
	PenaltyDue        float64  `json:"penalty_due"`
	InsuranceDue      float64  `json:"insurance_due"`
	PrincipalOverpaid float64  `json:"principal_overpaid"` // PRINOVR - the "ahead of schedule" amount floored out of PrincipalDue
	InterestOverpaid  float64  `json:"interest_overpaid"`
	PenaltyOverpaid   float64  `json:"penalty_overpaid"`
	AccountStatus     *string  `json:"account_status"`
	MaturityDate      *string  `json:"maturity_date"`
	Term              *string  `json:"term"`
}

// SOAStatement is the full detail of one pre-generated statement (see
// SOARepository.FindSOAStatement). Header fields come straight off the
// first SEQNO row - wkfcsoa duplicates them per line rather than
// normalizing, so no join back to client/profile is needed to render a
// complete document.
type SOAStatement struct {
	CtrlNo        int64     `json:"ctrl_no"`
	StatementDate time.Time `json:"statement_date"`
	CutoffDate    time.Time `json:"cutoff_date"`
	ClientID      string    `json:"client_id"`
	ClientName    string    `json:"client_name"`
	HomeAddress   *string   `json:"home_address"`
	MsMr          *string   `json:"ms_mr"`
	BranchManager *string   `json:"branch_manager"`
	// IDNo is EMPLOYEEID off wkfcsoa - despite the column name, the legacy
	// generator stores a ready-to-print label+value string here (e.g.
	// "ID No.: 010116"), not a raw employee id. Rendered as-is, not
	// relabeled, to avoid duplicating or mismatching its baked-in label.
	IDNo           *string       `json:"id_no"`
	MembershipDate *string       `json:"membership_date"`
	OrgAcronym     *string       `json:"org_acronym"`
	Lines          []SOALineItem `json:"lines"`
}
