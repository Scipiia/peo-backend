package auth_ldap

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"vue-golang/internal/config"
)

type JWTService struct {
	secret          string
	expirationHours int
}

func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{
		secret:          cfg.Secret,
		expirationHours: cfg.ExpirationHours,
	}
}

func (j *JWTService) GenerateToken(user *User, permissions []string) (string, error) {
	claims := CustomClaims{
		UID:         user.UID,
		FullName:    user.FullName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "system-norm-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWTService) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
