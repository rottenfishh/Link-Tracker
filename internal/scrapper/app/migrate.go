package app

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn string) error {
	wd, _ := os.Getwd()
	fmt.Println("working dir:", wd)
	m, err := migrate.New(
		"file://database",
		dsn,
	)
	if err != nil {
		return fmt.Errorf("migration failed to initialize: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}

	slog.Info("Migrations applied successfully!")
	return nil
}
