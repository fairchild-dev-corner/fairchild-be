package database

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	models "fairchild_be/internal/models/config"

	"github.com/go-sql-driver/mysql"
)

/**
* MySQLInitConnector opens a pooled *sql.DB and verifies connectivity with
* a bounded retry/backoff loop. multiStatements enables sending several
* ;-separated statements in a single Exec - required by cmd/migrate (whose
* .sql files batch multiple DDL statements) but left off for the app's
* runtime connection, which only ever runs single parameterized queries.
**/
func MySQLInitConnector(dbConfig *models.MySQLConfig, multiStatements bool) (*sql.DB, error) {

	cfg := toDriverConfig(dbConfig, multiStatements)

	db, err := mySQLConnectWithRetry(cfg, 5, 2*time.Second)

	if err != nil {
		log.Fatalf("Failed to establish MySQL connection: %v", err)
	}

	// Connection pool tuning - best practice for a request-scoped web service.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	return db, nil
}

// toDriverConfig maps our app-level MySQLConfig onto the driver's own
// mysql.Config{} - so credentials/host/db name are carried as typed fields
// instead of a hand-built DSN string.
func toDriverConfig(c *models.MySQLConfig, multiStatements bool) *mysql.Config {
	return &mysql.Config{
		User:            c.User,
		Passwd:          c.Password,
		Net:             "tcp",
		Addr:            fmt.Sprintf("%s:%s", c.Host, c.Port),
		DBName:          c.DBName,
		ParseTime:       true,
		Loc:             time.UTC,
		Collation:       "utf8mb4_general_ci",
		MultiStatements: multiStatements,
	}
}

func mySQLConnectWithRetry(cfg *mysql.Config, maxRetries int, baseDelay time.Duration) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < maxRetries; i++ {

		log.Printf("Trying to connect to MySQL (attempt %d/%d)...", i+1, maxRetries)

		connector, connErr := mysql.NewConnector(cfg)
		if connErr != nil {
			// Malformed config (bad TLS name, invalid auth plugin, etc.) -
			// no amount of retrying fixes this, fail immediately.
			return nil, fmt.Errorf("invalid MySQL config: %w", connErr)
		}

		db = sql.OpenDB(connector)

		if pingErr := db.Ping(); pingErr != nil {
			slog.Error("Failed to ping MySQL", "error", pingErr)
			db.Close()
			err = pingErr

			sleep := baseDelay * time.Duration(1<<i)
			log.Printf("Retrying in %v...", sleep)
			time.Sleep(sleep)
			continue
		}

		slog.Info("Successfully connected DB")
		return db, nil
	}

	return nil, fmt.Errorf("could not connect to MySQL after %d attempts: %v", maxRetries, err)
}
