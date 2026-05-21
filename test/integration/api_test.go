//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"habit-tracker/internal/config"
	delivery "habit-tracker/internal/delivery/http"
	"habit-tracker/internal/repository"
	"habit-tracker/internal/usecase"
	"habit-tracker/pkg/jwtutil"

	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type apiClient struct {
	router *gin.Engine
}

func findMigrationsPath() string {
	if p := os.Getenv("MIGRATIONS_PATH"); p != "" {
		return p
	}
	candidates := []string{"migrations", "../../migrations", "../../../migrations"}
	for _, p := range candidates {
		if _, err := os.Stat(filepath.Join(p, "000001_init.up.sql")); err == nil {
			return p
		}
		if _, err := os.Stat(filepath.Join(p, "000002_auth_profile.up.sql")); err == nil {
			return p
		}
	}
	return "migrations"
}

func setupTestAPI(t *testing.T) (*apiClient, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("habit_test"),
		postgres.WithUsername("habit"),
		postgres.WithPassword("habit_secret"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("mapped port: %v", err)
	}

	cfg := &config.Config{
		JWTSecret:        "integration-test-secret",
		JWTAccessTTL:     15 * time.Minute,
		JWTRefreshTTL:    24 * time.Hour,
		MigrationsPath:   findMigrationsPath(),
		DBHost:           host,
		DBPort:           port.Port(),
		DBUser:           "habit",
		DBPassword:       "habit_secret",
		DBName:           "habit_test",
		DBSSLMode:        "disable",
		RateLimitEnabled: false,
	}

	if err := repository.RunMigrations(cfg.MigrateURL(), cfg.MigrationsPath); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	db, err := gorm.Open(pgdriver.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	if err := repository.SeedAchievements(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	habitRepo := repository.NewHabitRepository(db)
	checkInRepo := repository.NewCheckInRepository(db)
	achievementRepo := repository.NewAchievementRepository(db)

	jwtManager := jwtutil.New(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	authUC := usecase.NewAuthService(userRepo, refreshRepo, jwtManager)
	habitUC := usecase.NewHabitService(habitRepo, checkInRepo, userRepo, achievementRepo)
	achievementUC := usecase.NewAchievementService(achievementRepo)

	handlers := delivery.NewHandlers(authUC, habitUC, achievementUC)
	router := delivery.NewRouter(handlers, jwtManager, cfg)

	cleanup := func() {
		_ = container.Terminate(ctx)
	}
	return &apiClient{router: router}, cleanup
}

func (c *apiClient) do(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	c.router.ServeHTTP(w, req)
	return w
}

func TestIntegration_AuthFlow(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	reg := api.do(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "int@test.com",
		"password": "secret12",
	}, "")
	if reg.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", reg.Code, reg.Body.String())
	}

	var authResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(reg.Body.Bytes(), &authResp); err != nil {
		t.Fatalf("decode register: %v", err)
	}
	if authResp.AccessToken == "" || authResp.RefreshToken == "" {
		t.Fatal("missing tokens")
	}

	me := api.do(t, http.MethodGet, "/api/v1/me", nil, authResp.AccessToken)
	if me.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", me.Code, me.Body.String())
	}

	refresh := api.do(t, http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": authResp.RefreshToken,
	}, "")
	if refresh.Code != http.StatusOK {
		t.Fatalf("refresh status=%d body=%s", refresh.Code, refresh.Body.String())
	}

	var refreshed struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(refresh.Body.Bytes(), &refreshed); err != nil {
		t.Fatalf("decode refresh: %v", err)
	}
	if refreshed.AccessToken == "" {
		t.Fatal("empty access after refresh")
	}
}

func TestIntegration_PatchMe(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	reg := api.do(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "patch-int@test.com",
		"password": "secret12",
	}, "")
	var authResp struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.Unmarshal(reg.Body.Bytes(), &authResp)

	patch := api.do(t, http.MethodPatch, "/api/v1/me", map[string]string{
		"display_name": "Integration User",
	}, authResp.AccessToken)
	if patch.Code != http.StatusOK {
		t.Fatalf("patch status=%d body=%s", patch.Code, patch.Body.String())
	}

	var user struct {
		DisplayName *string `json:"display_name"`
	}
	if err := json.Unmarshal(patch.Body.Bytes(), &user); err != nil {
		t.Fatalf("decode patch: %v", err)
	}
	if user.DisplayName == nil || *user.DisplayName != "Integration User" {
		t.Fatalf("display_name = %v", user.DisplayName)
	}
}

func TestIntegration_HabitCheckIn(t *testing.T) {
	api, cleanup := setupTestAPI(t)
	defer cleanup()

	reg := api.do(t, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    fmt.Sprintf("habit-%d@test.com", time.Now().UnixNano()),
		"password": "secret12",
	}, "")
	var authResp struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.Unmarshal(reg.Body.Bytes(), &authResp)

	create := api.do(t, http.MethodPost, "/api/v1/habits", map[string]string{
		"title":     "Read",
		"frequency": "daily",
	}, authResp.AccessToken)
	if create.Code != http.StatusCreated {
		t.Fatalf("create habit status=%d body=%s", create.Code, create.Body.String())
	}

	var habit struct {
		ID uint `json:"id"`
	}
	_ = json.Unmarshal(create.Body.Bytes(), &habit)

	checkIn := api.do(t, http.MethodPost, fmt.Sprintf("/api/v1/habits/%d/check-ins", habit.ID), map[string]string{}, authResp.AccessToken)
	if checkIn.Code != http.StatusCreated {
		t.Fatalf("check-in status=%d body=%s", checkIn.Code, checkIn.Body.String())
	}
}
