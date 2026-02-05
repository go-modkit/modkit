package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func IssueToken(cfg Config, user User) (string, error) {
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
