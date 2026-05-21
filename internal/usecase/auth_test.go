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

type mockRefreshTokenRepo struct {
	byJTI map[string]domain.RefreshToken
}

func newMockRefreshTokenRepo() *mockRefreshTokenRepo {
	return &mockRefreshTokenRepo{byJTI: make(map[string]domain.RefreshToken)}
}

func (m *mockRefreshTokenRepo) Store(_ context.Context, userID uint, jti string, expiresAt time.Time) error {
	m.byJTI[jti] = domain.RefreshToken{UserID: userID, JTI: jti, ExpiresAt: expiresAt}
	return nil
}

func (m *mockRefreshTokenRepo) Exists(_ context.Context, jti string) (bool, error) {
	rt, ok := m.byJTI[jti]
	if !ok {
		return false, nil
	}
	return rt.ExpiresAt.After(time.Now().UTC()), nil
}

func (m *mockRefreshTokenRepo) Delete(_ context.Context, jti string) error {
	delete(m.byJTI, jti)
	return nil
}

func newTestAuthService() *AuthService {
	users := newMockUserRepo()
	refresh := newMockRefreshTokenRepo()
	jwt := jwtutil.New("test-secret", 15*time.Minute, 7*24*time.Hour)
	return NewAuthService(users, refresh, jwt)
}

func TestAuthService_Register_Success(t *testing.T) {
	svc := newTestAuthService()

	result, err := svc.Register(context.Background(), RegisterInput{
		Email:    "User@Example.com",
		Password: "secret12",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Fatal("expected access and refresh tokens")
	}
	if result.ExpiresIn <= 0 {
		t.Errorf("expires_in = %d", result.ExpiresIn)
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
	svc := newTestAuthService()
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
	svc := newTestAuthService()

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
	svc := newTestAuthService()
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterInput{Email: "login@test.com", Password: "secret12"})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	result, err := svc.Login(ctx, LoginInput{Email: "login@test.com", Password: "secret12"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	_, _ = svc.Register(ctx, RegisterInput{Email: "x@test.com", Password: "secret12"})

	_, err := svc.Login(ctx, LoginInput{Email: "x@test.com", Password: "wrongpass"})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("Login() error = %v, want ErrUnauthorized", err)
	}
}

func TestAuthService_Refresh_Success(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	reg, err := svc.Register(ctx, RegisterInput{Email: "refresh@test.com", Password: "secret12"})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	refreshed, err := svc.Refresh(ctx, RefreshInput{RefreshToken: reg.RefreshToken})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshed.AccessToken == "" {
		t.Error("expected new access token")
	}
	if refreshed.RefreshToken == reg.RefreshToken {
		t.Error("expected rotated refresh token")
	}
}

func TestAuthService_Refresh_Invalid(t *testing.T) {
	svc := newTestAuthService()
	_, err := svc.Refresh(context.Background(), RefreshInput{RefreshToken: "invalid"})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("Refresh() error = %v, want ErrUnauthorized", err)
	}
}

func TestAuthService_UpdateProfile(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	reg, _ := svc.Register(ctx, RegisterInput{Email: "patch@test.com", Password: "secret12"})
	name := "Tracker"
	newPass := "newsecret12"
	cur := "secret12"

	user, err := svc.UpdateProfile(ctx, reg.User.ID, UpdateProfileInput{
		DisplayName:     &name,
		Password:        &newPass,
		CurrentPassword: &cur,
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if user.DisplayName == nil || *user.DisplayName != "Tracker" {
		t.Errorf("display_name = %v", user.DisplayName)
	}
	if !hash.Compare("newsecret12", user.PasswordHash) {
		t.Error("password was not updated")
	}
}

func TestAuthService_GetMe(t *testing.T) {
	svc := newTestAuthService()
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
