package config

import "os"

type Config struct {
	HTTPAddr string
	MySQLDSN string
}

func Load() Config {
	return Config{
		HTTPAddr: envOrDefault("HTTP_ADDR", ":8080"),
		MySQLDSN: envOrDefault("MYSQL_DSN", "root:password@tcp(localhost:3306)/app?parseTime=true&multiStatements=true"),
	}
}

func envOrDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}
