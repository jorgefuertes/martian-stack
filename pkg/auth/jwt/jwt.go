package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when token validation fails
	ErrInvalidToken = errors.New("invalid token")

	// ErrExpiredToken is returned when token has expired
	ErrExpiredToken = errors.New("token has expired")

	// ErrInvalidClaims is returned when token claims are invalid
	ErrInvalidClaims = errors.New("invalid token claims")
)

// Claims represents JWT claims for authentication
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Config represents JWT configuration
type Config struct {
	// Secret key for signing tokens
	SecretKey []byte

	// Issuer of the token
	Issuer string

	// Access token expiration duration
	AccessTokenExpiry time.Duration

	// Refresh token expiration duration
	RefreshTokenExpiry time.Duration
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig(secretKey string) *Config {
	return &Config{
		SecretKey:          []byte(secretKey),
		Issuer:             "martian-stack",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour, // 7 days
	}
}

// Service provides JWT token operations
type Service struct {
	config *Config
}

// NewService creates a new JWT service
func NewService(cfg *Config) *Service {
	return &Service{
		config: cfg,
	}
}

// GenerateAccessToken generates a new access token
func (s *Service) GenerateAccessToken(userID, username, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessTokenExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.config.SecretKey)
}

// GenerateRefreshToken generates a new refresh token
func (s *Service) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshTokenExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.config.SecretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.config.SecretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (s *Service) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Refresh tokens should only contain UserID
	if claims.Username != "" || claims.Email != "" {
		return "", ErrInvalidToken
	}

	// Note: In production, you'd fetch user details from database here
	// For now, we return a token with only UserID
	return s.GenerateAccessToken(claims.UserID, "", "", "")
}

// GetExpiryTime returns the expiry time from a token
func (s *Service) GetExpiryTime(tokenString string) (time.Time, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, ErrInvalidClaims
	}

	return claims.ExpiresAt.Time, nil
}

// IsExpired checks if a token is expired
func (s *Service) IsExpired(tokenString string) bool {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return true
	}

	if claims.ExpiresAt == nil {
		return true
	}

	return claims.ExpiresAt.Time.Before(time.Now())
}
