package database

import (
	"crypto/tls"
	"crypto/x509"
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

	cfg, err := toDriverConfig(dbConfig, multiStatements)
	if err != nil {
		return nil, err
	}

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
func toDriverConfig(c *models.MySQLConfig, multiStatements bool) (*mysql.Config, error) {
	cfg := &mysql.Config{
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

	if c.CACert != "" {
		tlsConfigName, err := registerTLSConfig(c.BuildEnv, c.CACert)
		if err != nil {
			return nil, err
		}
		cfg.TLSConfig = tlsConfigName
	}

	return cfg, nil
}

// registerTLSConfig parses a PEM-encoded CA certificate and registers a
// named TLS config with the driver (e.g. for DigitalOcean managed MySQL,
// which requires TLS and provides its own CA cert rather than a
// publicly-trusted one). name scopes the registration per environment so
// dev/stage/legacy don't clobber each other's TLS config.
func registerTLSConfig(name, caCertPEM string) (string, error) {
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM([]byte(caCertPEM)); !ok {
		return "", fmt.Errorf("failed to parse CA certificate for MySQL TLS config %q", name)
	}

	if err := mysql.RegisterTLSConfig(name, &tls.Config{RootCAs: pool}); err != nil {
		return "", fmt.Errorf("failed to register MySQL TLS config %q: %w", name, err)
	}

	return name, nil
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
