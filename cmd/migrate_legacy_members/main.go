// Command migrate_legacy_members is a one-off backfill: it reads the old
// portal's member_record table (on a separate legacy MySQL server) and
// creates the matching `users` row (plus a paired, still-passwordless
// `user_account_cred` row - see AuthRepository.CreateUser, which keeps the
// same invariant for every user created through the app) for each legacy
// member that doesn't already exist.
//
// It must run BEFORE cmd/migrate_legacy_credentials, which resolves
// user_id by member_id and only updates rows that already exist here.
//
// Safe to re-run: any member_id already present in `users` is skipped.
package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	conf "fairchild_be/config"
	authModels "fairchild_be/internal/models/auth"
	"fairchild_be/pkg/database"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var BuildEnv string

const skippedReportPath = "unmigrated_legacy_members.csv"

type legacyMember struct {
	MemberID       string
	ClientID       sql.NullString
	Type           sql.NullString
	FirstName      sql.NullString
	MiddleName     sql.NullString
	LastName       sql.NullString
	MobileNumber   sql.NullString
	Email          sql.NullString
	Gender         sql.NullString
	Nationality    sql.NullString
	CurrentAddr    sql.NullString
	MembershipDate sql.NullString
	Status         sql.NullString
}

type skipRecord struct {
	memberID string
	reason   string
}

// stats tallies data-quality issues that get silently defaulted rather than
// aborting the run - printed as a summary, and worth a manual look afterward.
type stats struct {
	truncatedMobile      int
	truncatedLocation    int
	badMembershipDate    int
	duplicateEmailNulled int
	unmappedStatus       map[string]int
	unmappedType         map[string]int
}

func main() {
	legacyConf, err := conf.GetLegacyMySQLConfig()
	if err != nil {
		log.Fatalf("legacy MySQL config: %v", err)
	}
	appConf, err := conf.GetMySQLConfig(BuildEnv)
	if err != nil {
		log.Fatalf("app MySQL config: %v", err)
	}

	legacyDB, err := database.MySQLInitConnector(legacyConf, false)
	if err != nil {
		log.Fatalf("connect legacy DB: %v", err)
	}
	defer legacyDB.Close()

	appDB, err := database.MySQLInitConnector(appConf, false)
	if err != nil {
		log.Fatalf("connect app DB: %v", err)
	}
	defer appDB.Close()

	rows, err := legacyDB.Query(`
		SELECT m_id, m_client_id, m_type, f_name, m_name, l_name, m_no, m_email, m_gender,
		       m_nationality, m_currentAddr, m_membershp_date, m_status
		FROM member_record`)
	if err != nil {
		log.Fatalf("query member_record: %v", err)
	}
	defer rows.Close()

	st := &stats{unmappedStatus: map[string]int{}, unmappedType: map[string]int{}}
	var total, inserted, alreadyMigrated int
	var skipped []skipRecord

	for rows.Next() {
		var m legacyMember
		if err := rows.Scan(&m.MemberID, &m.ClientID, &m.Type, &m.FirstName, &m.MiddleName, &m.LastName,
			&m.MobileNumber, &m.Email, &m.Gender, &m.Nationality, &m.CurrentAddr, &m.MembershipDate, &m.Status); err != nil {
			log.Fatalf("scan member_record row: %v", err)
		}
		total++

		memberID := strings.TrimSpace(m.MemberID)
		if memberID == "" {
			skipped = append(skipped, skipRecord{m.MemberID, "empty m_id"})
			continue
		}

		var existingID int64
		lookupErr := appDB.QueryRow(`SELECT id FROM users WHERE member_id = ?`, memberID).Scan(&existingID)
		if lookupErr == nil {
			alreadyMigrated++
			continue
		}
		if lookupErr != sql.ErrNoRows {
			log.Fatalf("lookup existing user for member_id %s: %v", memberID, lookupErr)
		}

		if _, insertErr := insertUser(appDB, memberID, m, st); insertErr != nil {
			skipped = append(skipped, skipRecord{memberID, insertErr.Error()})
			continue
		}
		inserted++
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("iterate member_record rows: %v", err)
	}

	writeSkippedReport(skipped)

	log.Printf("member_record backfill complete: total=%d inserted=%d already_migrated=%d skipped=%d",
		total, inserted, alreadyMigrated, len(skipped))
	log.Printf("data-quality: mobile_number truncated=%d location truncated=%d unparsed membership_date=%d duplicate email nulled=%d",
		st.truncatedMobile, st.truncatedLocation, st.badMembershipDate, st.duplicateEmailNulled)
	for v, n := range st.unmappedStatus {
		log.Printf("unmapped m_status %q defaulted to active: %d row(s)", v, n)
	}
	for v, n := range st.unmappedType {
		log.Printf("unmapped m_type %q defaulted to regular: %d row(s)", v, n)
	}
	if len(skipped) > 0 {
		log.Printf("skipped rows written to %s", skippedReportPath)
	}
}

// insertUser mirrors AuthRepository.CreateUser's users + user_account_cred
// pairing so every downstream INNER JOIN between the two tables keeps working
// for these backfilled accounts too. password_hash stays NULL here -
// cmd/migrate_legacy_credentials fills it in from tb_account afterward.
func insertUser(db *sql.DB, memberID string, m legacyMember, st *stats) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op once committed

	email := nullIfJunkEmail(m.Email.String)
	firstName := nullIfEmpty(m.FirstName.String)
	middleName := nullIfEmpty(m.MiddleName.String)
	lastName := nullIfEmpty(m.LastName.String)
	gender := nullIfEmpty(m.Gender.String)
	nationality := nullIfEmpty(m.Nationality.String)
	clientID := nullIfEmpty(m.ClientID.String)

	mobile := truncate(nullIfEmpty(m.MobileNumber.String), 20, &st.truncatedMobile)
	location := truncate(nullIfEmpty(m.CurrentAddr.String), 255, &st.truncatedLocation)

	status := mapStatus(m.Status.String, st)
	memberType := mapMemberType(m.Type.String, st)
	membershipDate := parseMembershipDate(m.MembershipDate.String, st)

	userUUID := uuid.NewString()
	res, err := tx.Exec(`
		INSERT INTO users (
			uuid, email, first_name, middle_name, last_name, member_id, gender, mobile_number, location,
			status, member_type, nationality, membership_date, user_client_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userUUID, email, firstName, middleName, lastName, memberID, gender, mobile, location,
		status, memberType, nationality, membershipDate, clientID,
	)
	if dupErr := translateDuplicateError(err); dupErr != nil && strings.Contains(dupErr.Error(), "uq_users_email") {
		// Legacy data has no email uniqueness guarantee - the same person
		// often has multiple member_ids (e.g. a parent's account and their
		// child's Kiddie Depositor account). Keep the row (and the profile
		// data it carries) by dropping the email on this later duplicate
		// rather than losing the member entirely.
		st.duplicateEmailNulled++
		email = sql.NullString{}
		res, err = tx.Exec(`
			INSERT INTO users (
				uuid, email, first_name, middle_name, last_name, member_id, gender, mobile_number, location,
				status, member_type, nationality, membership_date, user_client_id
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			userUUID, email, firstName, middleName, lastName, memberID, gender, mobile, location,
			status, memberType, nationality, membershipDate, clientID,
		)
	}
	if err != nil {
		return 0, translateDuplicateError(err)
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if _, err := tx.Exec(`INSERT INTO user_account_cred (user_id, member_id) VALUES (?, ?)`, userID, memberID); err != nil {
		return 0, err
	}

	return userID, tx.Commit()
}

// translateDuplicateError flags a raw ER_DUP_ENTRY (most commonly
// uq_users_email, since the legacy data has no email uniqueness guarantee)
// as-is - the driver's message already names the offending key, and the skip
// report just needs a reason string.
func translateDuplicateError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return mysqlErr
	}
	return err
}

func mapStatus(raw string, st *stats) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "active", "":
		return string(authModels.UserStatusActive)
	case "inactive":
		return string(authModels.UserStatusInactive)
	default:
		st.unmappedStatus[raw]++
		return string(authModels.UserStatusActive)
	}
}

func mapMemberType(raw string, st *stats) string {
	switch strings.TrimSpace(raw) {
	case "Regular Member", "Associate Member", "":
		return string(authModels.MemberTypeRegular)
	case "Kiddie Depositor", "Aflatoun":
		return string(authModels.MemberTypeYoungSaver)
	default:
		st.unmappedType[raw]++
		return string(authModels.MemberTypeRegular)
	}
}

// parseMembershipDate accepts either "YYYY-MM-DD" or "YYYY-MM-DD HH:MM:SS"
// (both appear in the legacy dump) and drops the time component - the new
// membership_date column is DATE, not DATETIME.
func parseMembershipDate(raw string, st *stats) sql.NullString {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return sql.NullString{}
	}
	datePart := raw
	if len(raw) >= 10 {
		datePart = raw[:10]
	}
	if _, err := time.Parse("2006-01-02", datePart); err != nil {
		st.badMembershipDate++
		return sql.NullString{}
	}
	return sql.NullString{String: datePart, Valid: true}
}

func nullIfEmpty(s string) sql.NullString {
	s = strings.TrimSpace(s)
	return sql.NullString{String: s, Valid: s != ""}
}

// junkEmailValues are placeholder strings front-desk staff typed into
// m_email instead of leaving it blank - not real addresses, so they should
// be treated the same as an empty field rather than fought over by
// nullIfEmpty's duplicate-key retry.
var junkEmailValues = map[string]bool{
	"n": true, "na": true, "n/a": true, "none": true, "-": true, "n.a.": true,
}

func nullIfJunkEmail(s string) sql.NullString {
	trimmed := strings.ToLower(strings.TrimSpace(s))
	if junkEmailValues[trimmed] {
		return sql.NullString{}
	}
	return nullIfEmpty(s)
}

func truncate(v sql.NullString, max int, counter *int) sql.NullString {
	if v.Valid && len(v.String) > max {
		v.String = v.String[:max]
		*counter++
	}
	return v
}

func writeSkippedReport(skipped []skipRecord) {
	if len(skipped) == 0 {
		return
	}

	f, err := os.Create(skippedReportPath)
	if err != nil {
		log.Printf("could not write %s: %v", skippedReportPath, err)
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{"member_id", "reason"})
	for _, s := range skipped {
		_ = w.Write([]string{s.memberID, s.reason})
	}
}
