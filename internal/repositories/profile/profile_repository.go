package repository

import (
	"database/sql"
	"errors"

	models "fairchild_be/internal/models/profile"

	"github.com/gin-gonic/gin"
)

/**
* ProfileRepository
* @Description: Wraps the pooled MySQL connection for reads against the
* client master (client + clienttype/gender/acctstat/civilstat/dept).
**/
type ProfileRepository struct {
	DB *sql.DB
}

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{DB: db}
}

// profileQuery resolves one active member's realtime profile by their
// legacy member id (client.OldClientID), restricted the same way the
// transactions/health queries are - regular members (ClientType 1,2) or
// Kiddies/Aflatoun minors (ClientType 3,4 with a matching account type and
// age under 18) - and to an active account status (AcctStatID = 1).
//
// Identity fields (name, email, mobile number, avatar) are deliberately
// sourced from `users` - the app's own record the member can update -
// rather than the read-only legacy `client` table, which we don't want to
// touch. The `users` join is scoped by id (the authenticated caller), not
// by member_id, so args are (userID, memberID) in that order.
const profileQuery = `
SELECT DISTINCT
    c.ClientIDBrCode AS branchID,
    c.ClientID AS newID,
    c.OldClientID,
    ct.ClientTypeDesc,
    c.AccountName,
    u.last_name AS Lname,
    u.first_name AS Fname,
    c.SName,
    u.middle_name AS MName,
    u.mobile_number AS ContactNumber,
    u.email AS EmailAddress,
    u.avatar_url AS AvatarUrl,
    g.GenderDesc,
    c.DateOfBirth,
    cs.CivilStatDesc,
    c.ProvAddStreet,
    c.ProvAddBarangay,
    c.ProvAddCity,
    c.ProvAddProvince,
    c.ProvAddZipCode,
    c.ResAddStreet,
    c.ResAddBarangay,
    c.ResAddCity,
    c.ResAddProvince,
    c.ResAddZipCode,
    c.BusJobTitle,
    c.TINNum,
    a.AcctStatDesc,
    c.DateOpened,
    c.ClientType,
    d.DeptDesc,
    CASE
        WHEN c.ClientType IN (1, 2) THEN 'adult'
        WHEN c.ClientType IN (3, 4) THEN 'kiddie'
    END AS MemberCategory
FROM client c
LEFT JOIN users      u  ON u.id = ?
LEFT JOIN clienttype ct ON ct.ClientTypeID = c.ClientType
LEFT JOIN gender     g  ON g.GenderID      = c.Gender
LEFT JOIN acctstat   a  ON a.AcctStatID    = c.AccountStatus
LEFT JOIN civilstat  cs ON cs.CivilStatID  = c.CivilStatus
LEFT JOIN dept       d  ON d.DeptID        = c.ClientDEPT
WHERE c.OldClientID = ?
  AND a.AcctStatID = 1
  AND (
        c.ClientType IN (1, 2)
     OR (
            c.ClientType IN (3, 4)
        AND c.AccountType IN (SELECT AccTypeID FROM accttype WHERE TRIM(UPPER(AcctTypeDesc)) IN ('KIDDIES','AFLATOUN'))
        AND TIMESTAMPDIFF(YEAR, c.DateOfBirth, CURDATE()) < 18
        )
  )
LIMIT 1`

// FindMemberIDByUserID resolves the caller's own legacy member id
// (users.member_id) - callers are never trusted to supply their member_id
// directly (see FindProfileByMemberID), since that would let any
// authenticated user read another member's profile. A user with no linked
// member id returns "", nil - not an error - so the caller can treat it as
// "no profile" rather than a failure.
func (r *ProfileRepository) FindMemberIDByUserID(ctx *gin.Context, userID int64) (memberID string, err error) {
	err = r.DB.QueryRowContext(ctx, `
		SELECT member_id FROM users WHERE id = ? AND member_id IS NOT NULL`,
		userID,
	).Scan(&memberID)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return memberID, nil
}

// FindProfileByMemberID returns nil, nil (not an error) when the member id
// has no matching active client record - e.g. a suspended/closed account -
// so the caller can treat it as "no profile" rather than a failure.
func (r *ProfileRepository) FindProfileByMemberID(ctx *gin.Context, userID int64, memberID string) (*models.Profile, error) {
	var p models.Profile
	err := r.DB.QueryRowContext(ctx, profileQuery, userID, memberID).Scan(
		&p.BranchID,
		&p.ClientID,
		&p.OldClientID,
		&p.ClientTypeDesc,
		&p.AccountName,
		&p.LastName,
		&p.FirstName,
		&p.SuffixName,
		&p.MiddleName,
		&p.ContactNumber,
		&p.EmailAddress,
		&p.AvatarUrl,
		&p.GenderDesc,
		&p.DateOfBirth,
		&p.CivilStatDesc,
		&p.ProvAddStreet,
		&p.ProvAddBarangay,
		&p.ProvAddCity,
		&p.ProvAddProvince,
		&p.ProvAddZipCode,
		&p.ResAddStreet,
		&p.ResAddBarangay,
		&p.ResAddCity,
		&p.ResAddProvince,
		&p.ResAddZipCode,
		&p.BusJobTitle,
		&p.TINNum,
		&p.AcctStatDesc,
		&p.DateOpened,
		&p.ClientType,
		&p.DeptDesc,
		&p.MemberCategory,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}
