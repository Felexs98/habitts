package repository

import (
	"fmt"
	"log"

	"habit-tracker/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgres(cfg *config.Config) (*gorm.DB, error) {
	if err := RunMigrations(cfg.MigrateURL(), cfg.MigrationsPath); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if err := SeedAchievements(db); err != nil {
		return nil, fmt.Errorf("seed achievements: %w", err)
	}

	log.Println("database ready (migrations applied)")
	return db, nil
}
