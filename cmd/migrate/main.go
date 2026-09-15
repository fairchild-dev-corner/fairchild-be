package main

import (
	"errors"
	"log"
	"log/slog"
	"os"

	conf "fairchild_be/config"
	cc "fairchild_be/internal/constants"
	"fairchild_be/pkg/database"

	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var BuildEnv string

const migrationsPath = "file://cmd/migrate/migrations"

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down>")
	}

	direction := os.Args[1]
	if direction != "up" && direction != "down" {
		log.Fatalf("unknown direction %q: expected \"up\" or \"down\"", direction)
	}

	mysqlConf, mErr := conf.GetMySQLConfig(BuildEnv)
	if mErr != nil {
		slog.Error("Failed to initialize MySQL config", "error", mErr)
		panic(cc.UNINITIALIZED_PANIC_STATE)
	}

	db, err := database.MySQLInitConnector(mysqlConf, true)
	if err != nil {
		log.Fatalf("failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		log.Fatalf("failed to init migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		log.Fatalf("failed to init migrate instance: %v", err)
	}

	if err := runMigration(m, direction); err != nil {
		log.Fatalf("migration %s failed: %v", direction, err)
	}

	log.Printf("migration %s complete", direction)
}

// runMigration applies every pending migration (up) or rolls back every
// applied migration (down). migrate.ErrNoChange just means the schema was
// already at the target state - not a real failure.
func runMigration(m *migrate.Migrate, direction string) error {
	var err error

	if direction == "up" {
		err = m.Up()
	} else {
		err = m.Down()
	}

	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	return err
}
