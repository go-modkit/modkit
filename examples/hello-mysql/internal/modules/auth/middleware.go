package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewJWTMiddleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := bearerToken(r.Header.Get("Authorization"))
			if tokenStr == "" {
				w.Header().Set("WWW-Authenticate", "Bearer")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			token, err := parseToken(tokenStr, cfg, time.Now())
			if err != nil {
				w.Header().Set("WWW-Authenticate", "Bearer")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				w.Header().Set("WWW-Authenticate", "Bearer")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			user, ok := userFromClaims(claims)
			if !ok {
				w.Header().Set("WWW-Authenticate", "Bearer")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}

	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return ""
	}

	return token
}

func parseToken(tokenStr string, cfg Config, now time.Time) (*jwt.Token, error) {
	parser := jwt.NewParser(
		jwt.WithIssuer(cfg.Issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)

	return parser.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(cfg.Secret), nil
	})
}

func userFromClaims(claims jwt.MapClaims) (User, bool) {
	user := User{}
	if subject, ok := claims["sub"].(string); ok {
		user.ID = subject
	}
	if email, ok := claims["email"].(string); ok {
		user.Email = email
	}
	if user.ID == "" && user.Email == "" {
		return User{}, false
	}
	return user, true
}
