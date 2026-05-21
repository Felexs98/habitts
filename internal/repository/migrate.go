package repository

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(migrateURL, migrationsPath string) (err error) {
	source := fmt.Sprintf("file://%s", migrationsPath)
	m, err := migrate.New(source, migrateURL)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			err = errors.Join(err, fmt.Errorf("migrate close source: %w", sourceErr))
		}
		if dbErr != nil {
			err = errors.Join(err, fmt.Errorf("migrate close db: %w", dbErr))
		}
	}()

	if upErr := m.Up(); upErr != nil && !errors.Is(upErr, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", upErr)
	}
	return nil
}
