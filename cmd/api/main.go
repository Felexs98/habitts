package main

import (
	"log"
	"time"

	_ "habit-tracker/docs" // swagger docs (генерируется: swag init -g cmd/api/main.go)

	"habit-tracker/internal/config"
	delivery "habit-tracker/internal/delivery/http"
	"habit-tracker/internal/repository"
	"habit-tracker/internal/usecase"
	"habit-tracker/pkg/jwtutil"
)

// @title           Habit Tracker API
// @version         1.0
// @description     REST API трекера привычек (дипломный MVP)
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description     JWT токен. Формат: Bearer &lt;token&gt;
func main() {
	cfg := config.Load()

	db, err := repository.NewPostgres(cfg.DSN())
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	habitRepo := repository.NewHabitRepository(db)
	checkInRepo := repository.NewCheckInRepository(db)
	achievementRepo := repository.NewAchievementRepository(db)

	jwtManager := jwtutil.New(cfg.JWTSecret, 24*time.Hour)

	authUC := usecase.NewAuthService(userRepo, jwtManager)
	habitUC := usecase.NewHabitService(habitRepo, checkInRepo, userRepo, achievementRepo)
	achievementUC := usecase.NewAchievementService(achievementRepo)

	handlers := delivery.NewHandlers(authUC, habitUC, achievementUC)
	router := delivery.NewRouter(handlers, jwtManager, cfg)

	addr := ":" + cfg.AppPort
	log.Printf("server starting on %s (env=%s)", addr, cfg.AppEnv)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
