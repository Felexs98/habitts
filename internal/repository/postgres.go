package repository

import (
	"fmt"
	"log"

	"habit-tracker/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgres(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Habit{},
		&domain.CheckIn{},
		&domain.Achievement{},
		&domain.UserAchievement{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	if err := SeedAchievements(db); err != nil {
		return nil, fmt.Errorf("seed achievements: %w", err)
	}

	log.Println("database migrated successfully")
	return db, nil
}
