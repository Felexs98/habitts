package usecase

import (
	"context"
	"strings"
	"time"

	"habit-tracker/internal/domain"
	"habit-tracker/pkg/hash"
	"habit-tracker/pkg/jwtutil"

	"github.com/google/uuid"
)

type AuthService struct {
	users   domain.UserRepository
	refresh domain.RefreshTokenRepository
	jwt     *jwtutil.Manager
}

func NewAuthService(
	users domain.UserRepository,
	refresh domain.RefreshTokenRepository,
	jwt *jwtutil.Manager,
) *AuthService {
	return &AuthService{users: users, refresh: refresh, jwt: jwt}
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type RefreshInput struct {
	RefreshToken string
}

type UpdateProfileInput struct {
	Email           *string
	DisplayName     *string
	Password        *string
	CurrentPassword *string
}

type AuthResult struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	User         domain.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	if email == "" || len(in.Password) < 6 {
		return nil, domain.ErrInvalidInput
	}

	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, domain.ErrAlreadyExists
	} else if err != domain.ErrNotFound {
		return nil, err
	}

	passwordHash, err := hash.Hash(in.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: passwordHash,
		XP:           0,
		Level:        1,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if !hash.Compare(in.Password, user.PasswordHash) {
		return nil, domain.ErrUnauthorized
	}

	return s.issueTokens(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, in RefreshInput) (*AuthResult, error) {
	token := strings.TrimSpace(in.RefreshToken)
	if token == "" {
		return nil, domain.ErrInvalidInput
	}

	claims, err := s.jwt.ParseRefresh(token)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	exists, err := s.refresh.Exists(ctx, claims.JTI)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrUnauthorized
	}

	if err := s.refresh.Delete(ctx, claims.JTI); err != nil {
		return nil, err
	}

	user, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.issueTokens(ctx, user)
}

// Logout отзывает refresh-токен (идемпотентно).
func (s *AuthService) Logout(ctx context.Context, in RefreshInput) error {
	token := strings.TrimSpace(in.RefreshToken)
	if token == "" {
		return domain.ErrInvalidInput
	}
	claims, err := s.jwt.ParseRefresh(token)
	if err != nil {
		return nil
	}
	_ = s.refresh.Delete(ctx, claims.JTI)
	return nil
}

func (s *AuthService) GetMe(ctx context.Context, userID uint) (*domain.User, error) {
	return s.users.FindByID(ctx, userID)
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uint, in UpdateProfileInput) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if in.Password != nil {
		if in.CurrentPassword == nil || *in.CurrentPassword == "" {
			return nil, domain.ErrInvalidInput
		}
		if !hash.Compare(*in.CurrentPassword, user.PasswordHash) {
			return nil, domain.ErrUnauthorized
		}
		if len(*in.Password) < 6 {
			return nil, domain.ErrInvalidInput
		}
		hashVal, err := hash.Hash(*in.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hashVal
	}

	if in.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*in.Email))
		if email == "" {
			return nil, domain.ErrInvalidInput
		}
		if email != user.Email {
			if _, err := s.users.FindByEmail(ctx, email); err == nil {
				return nil, domain.ErrAlreadyExists
			} else if err != domain.ErrNotFound {
				return nil, err
			}
			user.Email = email
		}
	}

	if in.DisplayName != nil {
		name := strings.TrimSpace(*in.DisplayName)
		if name == "" {
			user.DisplayName = nil
		} else {
			user.DisplayName = &name
		}
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) issueTokens(ctx context.Context, user *domain.User) (*AuthResult, error) {
	jti := uuid.New().String()
	access, expiresIn, err := s.jwt.GenerateAccess(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	refresh, err := s.jwt.GenerateRefresh(user.ID, jti)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().UTC().Add(s.jwt.RefreshTTL())
	if err := s.refresh.Store(ctx, user.ID, jti, expiresAt); err != nil {
		return nil, err
	}
	return &AuthResult{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    expiresIn,
		User:         *user,
	}, nil
}
