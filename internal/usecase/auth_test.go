package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"habit-tracker/internal/domain"
	"habit-tracker/pkg/hash"
	"habit-tracker/pkg/jwtutil"
)

func TestAuthService_Register_Success(t *testing.T) {
	users := newMockUserRepo()
	svc := NewAuthService(users, jwtutil.New("test-secret", time.Hour))

	result, err := svc.Register(context.Background(), RegisterInput{
		Email:    "User@Example.com",
		Password: "secret12",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if result.User.Email != "user@example.com" {
		t.Errorf("email = %q, want user@example.com", result.User.Email)
	}
	if result.User.XP != 0 || result.User.Level != 1 {
		t.Errorf("gamification = xp:%d level:%d, want xp:0 level:1", result.User.XP, result.User.Level)
	}
	if !hash.Compare("secret12", result.User.PasswordHash) {
		t.Error("password hash mismatch")
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	users := newMockUserRepo()
	svc := NewAuthService(users, jwtutil.New("test-secret", time.Hour))
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterInput{Email: "dup@test.com", Password: "secret12"})
	if err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	_, err = svc.Register(ctx, RegisterInput{Email: "dup@test.com", Password: "other123"})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("second Register() error = %v, want ErrAlreadyExists", err)
	}
}

func TestAuthService_Register_InvalidInput(t *testing.T) {
	users := newMockUserRepo()
	svc := NewAuthService(users, jwtutil.New("test-secret", time.Hour))

	_, err := svc.Register(context.Background(), RegisterInput{Email: "", Password: "secret12"})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("empty email: got %v, want ErrInvalidInput", err)
	}

	_, err = svc.Register(context.Background(), RegisterInput{Email: "a@b.c", Password: "123"})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("short password: got %v, want ErrInvalidInput", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	users := newMockUserRepo()
	svc := NewAuthService(users, jwtutil.New("test-secret", time.Hour))
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterInput{Email: "login@test.com", Password: "secret12"})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	result, err := svc.Login(ctx, LoginInput{Email: "login@test.com", Password: "secret12"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	users := newMockUserRepo()
	svc := NewAuthService(users, jwtutil.New("test-secret", time.Hour))
	ctx := context.Background()

	_, _ = svc.Register(ctx, RegisterInput{Email: "x@test.com", Password: "secret12"})

	_, err := svc.Login(ctx, LoginInput{Email: "x@test.com", Password: "wrongpass"})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("Login() error = %v, want ErrUnauthorized", err)
	}
}

func TestAuthService_GetMe(t *testing.T) {
	users := newMockUserRepo()
	svc := NewAuthService(users, jwtutil.New("test-secret", time.Hour))
	ctx := context.Background()

	reg, _ := svc.Register(ctx, RegisterInput{Email: "me@test.com", Password: "secret12"})
	user, err := svc.GetMe(ctx, reg.User.ID)
	if err != nil {
		t.Fatalf("GetMe() error = %v", err)
	}
	if user.Email != "me@test.com" {
		t.Errorf("email = %q", user.Email)
	}
}
