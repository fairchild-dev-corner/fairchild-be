package models

import "time"

// Profile mirrors one row of the client master (client + lookup tables),
// scoped to the caller's own active member record - see
// ProfileRepository.FindProfileByMemberID. Every field beyond the branch/
// client id pair comes from an outer join or a nullable client column, so
// they're all pointers.
type Profile struct {
	BranchID        uint16     `json:"branch_id"`
	ClientID        uint32     `json:"client_id"`
	OldClientID     *string    `json:"old_client_id"`
	ClientTypeDesc  *string    `json:"client_type_desc"`
	AccountName     *string    `json:"account_name"`
	LastName        *string    `json:"last_name"`
	FirstName       *string    `json:"first_name"`
	SuffixName      *string    `json:"suffix_name"`
	MiddleName      *string    `json:"middle_name"`
	ContactNumber   *string    `json:"contact_number"`
	EmailAddress    *string    `json:"email_address"`
	AvatarUrl       *string    `json:"avatar_url"`
	GenderDesc      *string    `json:"gender_desc"`
	DateOfBirth     *time.Time `json:"date_of_birth"`
	CivilStatDesc   *string    `json:"civil_stat_desc"`
	ProvAddStreet   *string    `json:"prov_add_street"`
	ProvAddBarangay *string    `json:"prov_add_barangay"`
	ProvAddCity     *string    `json:"prov_add_city"`
	ProvAddProvince *string    `json:"prov_add_province"`
	ProvAddZipCode  *string    `json:"prov_add_zip_code"`
	ResAddStreet    *string    `json:"res_add_street"`
	ResAddBarangay  *string    `json:"res_add_barangay"`
	ResAddCity      *string    `json:"res_add_city"`
	ResAddProvince  *string    `json:"res_add_province"`
	ResAddZipCode   *string    `json:"res_add_zip_code"`
	BusJobTitle     *string    `json:"bus_job_title"`
	TINNum          *string    `json:"tin_num"`
	AcctStatDesc    *string    `json:"acct_stat_desc"`
	DateOpened      *time.Time `json:"date_opened"`
	ClientType      *int       `json:"client_type"`
	DeptDesc        *string    `json:"dept_desc"`
	MemberCategory  *string    `json:"member_category"`
}
