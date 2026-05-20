package usecase

import (
	"context"
	"strings"

	"habit-tracker/internal/domain"
	"habit-tracker/pkg/hash"
	"habit-tracker/pkg/jwtutil"
)

type AuthService struct {
	users  domain.UserRepository
	jwt    *jwtutil.Manager
}

func NewAuthService(users domain.UserRepository, jwt *jwtutil.Manager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	Token string       `json:"token"`
	User  domain.User  `json:"user"`
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
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.jwt.Generate(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: *user}, nil
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

	token, err := s.jwt.Generate(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: *user}, nil
}
