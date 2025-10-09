package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents the JWT claims
type Claims struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	HouseholdID string `json:"household_id,omitempty"`
	jwt.RegisteredClaims
}

// TokenManager handles JWT operations
type TokenManager struct {
	secretKey       []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
}

// NewTokenManager creates a new token manager
func NewTokenManager(secretKey string, accessDuration, refreshDuration time.Duration) *TokenManager {
	return &TokenManager{
		secretKey:       []byte(secretKey),
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (tm *TokenManager) GenerateTokenPair(userID, email, householdID string) (string, string, error) {
	accessToken, err := tm.generateToken(userID, email, householdID, tm.accessDuration)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := tm.generateToken(userID, email, householdID, tm.refreshDuration)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// generateToken creates a new JWT token
func (tm *TokenManager) generateToken(userID, email, householdID string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:      userID,
		Email:       email,
		HouseholdID: householdID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "house-helper-api",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}

// VerifyToken validates and parses a JWT token
func (tm *TokenManager) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return tm.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken generates a new access token from a valid refresh token
func (tm *TokenManager) RefreshToken(refreshToken string) (string, error) {
	claims, err := tm.VerifyToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Generate new access token
	newToken, err := tm.generateToken(claims.UserID, claims.Email, claims.HouseholdID, tm.accessDuration)
	if err != nil {
		return "", err
	}

	return newToken, nil
}
