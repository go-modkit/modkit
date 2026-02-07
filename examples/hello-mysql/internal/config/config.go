package config

import (
	"os"
	"strings"

	mkconfig "github.com/go-modkit/modkit/modkit/config"
)

type Config struct {
	HTTPAddr           string
	MySQLDSN           string
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string
	RateLimitPerSecond float64
	RateLimitBurst     int
}

func Load() Config {
	return Config{
		HTTPAddr:           envOrDefault("HTTP_ADDR", ":8080"),
		MySQLDSN:           envOrDefault("MYSQL_DSN", "root:password@tcp(localhost:3306)/app?parseTime=true&multiStatements=true"),
		CORSAllowedOrigins: splitEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		CORSAllowedMethods: splitEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE"),
		CORSAllowedHeaders: splitEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization"),
		RateLimitPerSecond: envFloat("RATE_LIMIT_PER_SECOND", 5),
		RateLimitBurst:     envInt("RATE_LIMIT_BURST", 10),
	}
}

func envOrDefault(key, def string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	return val
}

func splitEnv(key, def string) []string {
	raw := envOrDefault(key, def)
	out, err := mkconfig.ParseCSV(raw)
	if err != nil {
		return []string{}
	}
	return out
}

func envFloat(key string, def float64) float64 {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	parsed, err := mkconfig.ParseFloat64(val)
	if err != nil {
		return def
	}
	return parsed
}

func envInt(key string, def int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	parsed, err := mkconfig.ParseInt(val)
	if err != nil {
		return def
	}
	return parsed
}
