// Command migrate_legacy_credentials is a one-off backfill: it reads the old
// portal's tb_account table (on a separate legacy MySQL server) and fills in
// password_hash/last_password_change_at/password_change_required/created_at
// on the matching user_account_cred row.
//
// It must run AFTER cmd/migrate_legacy_members, which is what creates the
// user_account_cred row (member_id resolves to it via a JOIN to users) - a
// tb_account row with no matching user yet is skipped and reported, safe to
// retry once that member has been migrated.
//
// Safe to re-run: every row is a plain UPDATE, not an INSERT.
package main

import (
	"database/sql"
	"encoding/csv"
	"log"
	"os"
	"strings"

	conf "fairchild_be/config"
	"fairchild_be/pkg/database"
)

var BuildEnv string

const unmatchedReportPath = "unmatched_legacy_credentials.csv"
const failedReportPath = "failed_legacy_credentials.csv"

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
		SELECT member_id, a_upass, last_password_change, password_change_required, created_at
		FROM tb_account`)
	if err != nil {
		log.Fatalf("query tb_account: %v", err)
	}
	defer rows.Close()

	var total, updated int
	var unmatched []string
	var failed [][2]string // {member_id, reason}

	for rows.Next() {
		var (
			memberID               string
			passwordHash           string
			lastPasswordChange     sql.NullTime
			passwordChangeRequired sql.NullBool
			createdAt              sql.NullTime
		)
		if err := rows.Scan(&memberID, &passwordHash, &lastPasswordChange, &passwordChangeRequired, &createdAt); err != nil {
			log.Fatalf("scan tb_account row: %v", err)
		}
		total++

		memberID = strings.TrimSpace(memberID)

		// Resolve the match with its own SELECT rather than reading the
		// UPDATE's RowsAffected - MySQL reports rows *changed*, not rows
		// *matched*, so re-running this tool against already-correct rows
		// would otherwise make every one of them look "unmatched".
		var userID int64
		lookupErr := appDB.QueryRow(`SELECT id FROM users WHERE member_id = ?`, memberID).Scan(&userID)
		if lookupErr == sql.ErrNoRows {
			unmatched = append(unmatched, memberID)
			continue
		}
		if lookupErr != nil {
			failed = append(failed, [2]string{memberID, lookupErr.Error()})
			continue
		}

		_, err = appDB.Exec(`
			UPDATE user_account_cred
			SET password_hash = ?,
			    last_password_change_at = ?,
			    password_change_required = ?,
			    created_at = COALESCE(?, created_at)
			WHERE user_id = ?`,
			passwordHash, lastPasswordChange, passwordChangeRequired, createdAt, userID,
		)
		if err != nil {
			// A single malformed legacy row shouldn't abort the whole batch -
			// record it and keep going.
			failed = append(failed, [2]string{memberID, err.Error()})
			continue
		}
		updated++
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("iterate tb_account rows: %v", err)
	}

	writeUnmatchedReport(unmatched)
	writeFailedReport(failed)

	log.Printf("tb_account backfill complete: total=%d updated=%d unmatched=%d failed=%d", total, updated, len(unmatched), len(failed))
	if len(unmatched) > 0 {
		log.Printf("unmatched member_ids written to %s - retry after they're migrated into users", unmatchedReportPath)
	}
	if len(failed) > 0 {
		log.Printf("failed rows written to %s", failedReportPath)
	}
}

func writeUnmatchedReport(unmatched []string) {
	if len(unmatched) == 0 {
		return
	}

	f, err := os.Create(unmatchedReportPath)
	if err != nil {
		log.Printf("could not write %s: %v", unmatchedReportPath, err)
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{"member_id"})
	for _, id := range unmatched {
		_ = w.Write([]string{id})
	}
}

func writeFailedReport(failed [][2]string) {
	if len(failed) == 0 {
		return
	}

	f, err := os.Create(failedReportPath)
	if err != nil {
		log.Printf("could not write %s: %v", failedReportPath, err)
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{"member_id", "reason"})
	for _, row := range failed {
		_ = w.Write(row[:])
	}
}
