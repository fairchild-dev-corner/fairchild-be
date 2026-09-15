package models

import "time"

// AmortizationScheduleEntry is one loan's next upcoming due installment -
// see AmortizationRepository.FindAmortizationSchedule. A member typically
// has one entry per active loan (grouped by SLC_CODE/SLT_CODE/REF_NO).
type AmortizationScheduleEntry struct {
	BranchID         int       `json:"branch_id"`
	ClientID         int       `json:"client_id"`
	SLCCode          int       `json:"slc_code"`
	SLTCode          int       `json:"slt_code"`
	RefNo            string    `json:"ref_no"`
	DueDate          time.Time `json:"due_date"`
	PrincipalDue     float64   `json:"principal_due"`
	InterestDue      float64   `json:"interest_due"`
	PrincipalBalance float64   `json:"principal_balance"`
	PenaltyFlag      *string   `json:"penalty_flag"`
}
