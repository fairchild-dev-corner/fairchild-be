package models

import "time"

// ActivityMonth is one calendar month's SAVINGS + TIME DEPOSIT movement for
// a client - see ActivityRepository.FindMonthlyActivity. DepositAmt and
// WithdrawalAmt are both positive magnitudes; NetAmt is their difference.
type ActivityMonth struct {
	MonthStart    time.Time `json:"month_start"`
	DepositAmt    float64   `json:"deposit_amt"`
	WithdrawalAmt float64   `json:"withdrawal_amt"`
	NetAmt        float64   `json:"net_amt"`
}
