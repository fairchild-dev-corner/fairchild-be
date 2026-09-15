package models

// DepositBalance is one deposit sub-product's current balance for a client
// - see BalanceRepository.FindDepositBalances. A client typically has one
// entry per distinct SLC_CODE/SLT_CODE combination they've actually posted
// against (e.g. a separate row for "SD - Regular" vs "SD - Retirement"),
// spanning SAVINGS (SLC_CODE 22), TIME DEPOSIT (24), and SHARE CAPITAL (31).
type DepositBalance struct {
	MemberID      *string `json:"member_id"`
	SLCCode       int     `json:"slc_code"`
	SLTCode       int     `json:"slt_code"`
	DepositName   *string `json:"deposit_name"`
	DepositAmount float64 `json:"deposit_amount"`
}
