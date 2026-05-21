package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email,omitempty"`
	TokenType string `json:"typ"`
	JTI       string `json:"jti,omitempty"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *Manager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

func (m *Manager) GenerateAccess(userID uint, email string) (token string, expiresIn int64, err error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}
	return signed, int64(m.accessTTL.Seconds()), nil
}

func (m *Manager) GenerateRefresh(userID uint, jti string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		TokenType: TokenTypeRefresh,
		JTI:       jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) ParseAccess(tokenString string) (*Claims, error) {
	claims, err := m.parse(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (m *Manager) ParseRefresh(tokenString string) (*Claims, error) {
	claims, err := m.parse(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeRefresh || claims.JTI == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (m *Manager) parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// Generate сохраняет совместимость со старым однотокенным API (access).
func (m *Manager) Generate(userID uint, email string) (string, error) {
	token, _, err := m.GenerateAccess(userID, email)
	return token, err
}

// Parse — access-токен (для middleware).
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	return m.ParseAccess(tokenString)
}
