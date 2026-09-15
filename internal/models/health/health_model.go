package models

// CoopHealthScore is a client's rolling 12-month "coop health score" -
// share capital build-up consistency, savings consistency, and loan
// repayment standing, each as a 0-100 percentage, plus their average -
// see HealthRepository.FindCoopHealthScore.
type CoopHealthScore struct {
	ClientID               string   `json:"client_id"`
	BranchID               string   `json:"branch_id"`
	ShareCapitalBuildupPct *float64 `json:"share_capital_buildup_pct"`
	SavingsConsistencyPct  *float64 `json:"savings_consistency_pct"`
	LoanRepaymentPct       *float64 `json:"loan_repayment_pct"`
	CoopHealthScore        *float64 `json:"coop_health_score"`
}
