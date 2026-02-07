package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddr     string
	MySQLDSN     string
	JWTSecret    string
	JWTIssuer    string
	JWTTTL       string
	AuthUsername string
	AuthPassword string
}

func Load() Config {
	return Config{
		HTTPAddr:     envOrDefault("HTTP_ADDR", ":8080"),
		MySQLDSN:     envOrDefault("MYSQL_DSN", "root:password@tcp(localhost:3306)/app?parseTime=true&multiStatements=true"),
		JWTSecret:    envOrDefault("JWT_SECRET", "dev-secret-change-me"),
		JWTIssuer:    envOrDefault("JWT_ISSUER", "hello-mysql"),
		JWTTTL:       envOrDefault("JWT_TTL", "1h"),
		AuthUsername: envOrDefault("AUTH_USERNAME", "demo"),
		AuthPassword: envOrDefault("AUTH_PASSWORD", "demo"),
	}
}

func envOrDefault(key, def string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return def
	}

	val = strings.TrimSpace(val)
	if val == "" {
		return def
	}

	return val
}
