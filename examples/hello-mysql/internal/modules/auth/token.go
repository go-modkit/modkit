package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func IssueToken(cfg Config, user User) (string, error) {
	if cfg.Secret == "" {
		return "", fmt.Errorf("auth: missing jwt secret")
	}
	if cfg.TTL <= 0 {
		return "", fmt.Errorf("auth: invalid jwt ttl")
	}

	subject := user.ID
	if subject == "" {
		subject = user.Email
	}

	claims := jwt.MapClaims{
		"iss": cfg.Issuer,
		"exp": time.Now().Add(cfg.TTL).Unix(),
	}
	if subject != "" {
		claims["sub"] = subject
	}
	if user.Email != "" {
		claims["email"] = user.Email
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}
