package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type (
	JWT        = string
	JWTManager struct {
		secretKey     string
		tokenDuration time.Duration
		issuer        string
	}

	Claims struct {
		UserID string `json:"user_id"`
		jwt.RegisteredClaims
	}
)

func NewJWTManager(secretKey string, issuer string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
		issuer:        issuer,
	}
}

func (m *JWTManager) GenerateToken(userId string) (JWT, error) {
	claims := Claims{
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    m.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

func (m *JWTManager) ExtractClaimsFrom(token JWT) (*Claims, error) {
	parsedJwt, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected sign: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %v", err)
	}

	if !parsedJwt.Valid {
		return nil, errors.New("invalid JWT")
	}

	claims, ok := parsedJwt.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid JWT claims")
	}

	return claims, nil
}
