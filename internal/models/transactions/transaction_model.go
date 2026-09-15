package models

import "time"

// Transaction mirrors one row of a client's ledger (sldtl), scoped to loans,
// savings, time deposits and share capital - see
// TransactionsRepository.FindLastTransactions.
type Transaction struct {
	RefID    string    `json:"ref_id"`
	MemberID string    `json:"member_id"`
	Category string    `json:"category"`
	Name     *string   `json:"name,omitempty"`
	Amount   float64   `json:"amount"`
	Date     time.Time `json:"date"`
}
