package app

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // register postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/database"
)

func RunMigrations(dsn string) error {
	driver, err := iofs.New(database.BotMigrations, "bot")
	if err != nil {
		return fmt.Errorf("error creating driver for embedded migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		driver,
		dsn,
	)

	if err != nil {
		return fmt.Errorf("migration failed to initialize: %w", err)
	}

	if upErr := m.Up(); upErr != nil {
		if errors.Is(upErr, migrate.ErrNoChange) {
			slog.Info("no migrations applied")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", upErr)
	}
	slog.Info("running migrations", "dsn", dsn)
	slog.Info("Migrations applied successfully!")
	return nil
}
