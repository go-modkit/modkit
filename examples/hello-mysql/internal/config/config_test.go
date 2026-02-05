package config

import (
	"os"
	"reflect"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear all relevant env vars
	envVars := []string{
		"HTTP_ADDR",
		"MYSQL_DSN",
		"CORS_ALLOWED_ORIGINS",
		"CORS_ALLOWED_METHODS",
		"CORS_ALLOWED_HEADERS",
		"RATE_LIMIT_PER_SECOND",
		"RATE_LIMIT_BURST",
	}
	for _, key := range envVars {
		os.Unsetenv(key)
	}

	cfg := Load()

	if cfg.HTTPAddr != ":8080" {
		t.Errorf("expected HTTPAddr :8080, got %q", cfg.HTTPAddr)
	}
	if cfg.MySQLDSN != "root:password@tcp(localhost:3306)/app?parseTime=true&multiStatements=true" {
		t.Errorf("unexpected MySQLDSN: %q", cfg.MySQLDSN)
	}
	if !reflect.DeepEqual(cfg.CORSAllowedOrigins, []string{"http://localhost:3000"}) {
		t.Errorf("expected CORSAllowedOrigins [http://localhost:3000], got %v", cfg.CORSAllowedOrigins)
	}
	if !reflect.DeepEqual(cfg.CORSAllowedMethods, []string{"GET", "POST", "PUT", "DELETE"}) {
		t.Errorf("expected CORSAllowedMethods [GET POST PUT DELETE], got %v", cfg.CORSAllowedMethods)
	}
	if !reflect.DeepEqual(cfg.CORSAllowedHeaders, []string{"Content-Type", "Authorization"}) {
		t.Errorf("expected CORSAllowedHeaders [Content-Type Authorization], got %v", cfg.CORSAllowedHeaders)
	}
	if cfg.RateLimitPerSecond != 5 {
		t.Errorf("expected RateLimitPerSecond 5, got %f", cfg.RateLimitPerSecond)
	}
	if cfg.RateLimitBurst != 10 {
		t.Errorf("expected RateLimitBurst 10, got %d", cfg.RateLimitBurst)
	}
}

func TestLoad_WithEnvVars(t *testing.T) {
	os.Setenv("HTTP_ADDR", ":9090")
	os.Setenv("MYSQL_DSN", "user:pass@tcp(host:3307)/db")
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://example.com,https://example.org")
	os.Setenv("CORS_ALLOWED_METHODS", "GET,POST")
	os.Setenv("CORS_ALLOWED_HEADERS", "X-Custom-Header")
	os.Setenv("RATE_LIMIT_PER_SECOND", "10.5")
	os.Setenv("RATE_LIMIT_BURST", "20")
	defer func() {
		os.Unsetenv("HTTP_ADDR")
		os.Unsetenv("MYSQL_DSN")
		os.Unsetenv("CORS_ALLOWED_ORIGINS")
		os.Unsetenv("CORS_ALLOWED_METHODS")
		os.Unsetenv("CORS_ALLOWED_HEADERS")
		os.Unsetenv("RATE_LIMIT_PER_SECOND")
		os.Unsetenv("RATE_LIMIT_BURST")
	}()

	cfg := Load()

	if cfg.HTTPAddr != ":9090" {
		t.Errorf("expected HTTPAddr :9090, got %q", cfg.HTTPAddr)
	}
	if cfg.MySQLDSN != "user:pass@tcp(host:3307)/db" {
		t.Errorf("unexpected MySQLDSN: %q", cfg.MySQLDSN)
	}
	if !reflect.DeepEqual(cfg.CORSAllowedOrigins, []string{"http://example.com", "https://example.org"}) {
		t.Errorf("expected CORSAllowedOrigins [http://example.com https://example.org], got %v", cfg.CORSAllowedOrigins)
	}
	if !reflect.DeepEqual(cfg.CORSAllowedMethods, []string{"GET", "POST"}) {
		t.Errorf("expected CORSAllowedMethods [GET POST], got %v", cfg.CORSAllowedMethods)
	}
	if !reflect.DeepEqual(cfg.CORSAllowedHeaders, []string{"X-Custom-Header"}) {
		t.Errorf("expected CORSAllowedHeaders [X-Custom-Header], got %v", cfg.CORSAllowedHeaders)
	}
	if cfg.RateLimitPerSecond != 10.5 {
		t.Errorf("expected RateLimitPerSecond 10.5, got %f", cfg.RateLimitPerSecond)
	}
	if cfg.RateLimitBurst != 20 {
		t.Errorf("expected RateLimitBurst 20, got %d", cfg.RateLimitBurst)
	}
}

func TestEnvOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setValue string
		default_ string
		want     string
	}{
		{
			name:     "env var set",
			key:      "TEST_VAR",
			setValue: "custom-value",
			default_: "default-value",
			want:     "custom-value",
		},
		{
			name:     "env var empty",
			key:      "TEST_VAR_EMPTY",
			setValue: "",
			default_: "default-value",
			want:     "default-value",
		},
		{
			name:     "env var not set",
			key:      "TEST_VAR_NOT_SET",
			setValue: "",
			default_: "default-value",
			want:     "default-value",
		},
		{
			name:     "env var with spaces",
			key:      "TEST_VAR_SPACES",
			setValue: "  trimmed  ",
			default_: "default-value",
			want:     "trimmed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			} else {
				os.Unsetenv(tt.key)
			}
			defer os.Unsetenv(tt.key)

			got := envOrDefault(tt.key, tt.default_)
			if got != tt.want {
				t.Errorf("envOrDefault(%q, %q) = %q, want %q", tt.key, tt.default_, got, tt.want)
			}
		})
	}
}

func TestSplitEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setValue string
		default_ string
		want     []string
	}{
		{
			name:     "comma-separated values",
			key:      "TEST_SPLIT",
			setValue: "a,b,c",
			default_: "default",
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "values with spaces",
			key:      "TEST_SPLIT_SPACES",
			setValue: " a , b , c ",
			default_: "default",
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "empty string in list",
			key:      "TEST_SPLIT_EMPTY",
			setValue: "a,,c",
			default_: "default",
			want:     []string{"a", "c"},
		},
		{
			name:     "only empty strings",
			key:      "TEST_SPLIT_ALL_EMPTY",
			setValue: ",,",
			default_: "default",
			want:     []string{},
		},
		{
			name:     "env var not set",
			key:      "TEST_SPLIT_NOT_SET",
			setValue: "",
			default_: "x,y,z",
			want:     []string{"x", "y", "z"},
		},
		{
			name:     "single value",
			key:      "TEST_SPLIT_SINGLE",
			setValue: "single",
			default_: "default",
			want:     []string{"single"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			} else {
				os.Unsetenv(tt.key)
			}
			defer os.Unsetenv(tt.key)

			got := splitEnv(tt.key, tt.default_)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitEnv(%q, %q) = %v, want %v", tt.key, tt.default_, got, tt.want)
			}
		})
	}
}

func TestEnvFloat(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setValue string
		default_ float64
		want     float64
	}{
		{
			name:     "valid float",
			key:      "TEST_FLOAT",
			setValue: "3.14",
			default_: 1.0,
			want:     3.14,
		},
		{
			name:     "valid integer as float",
			key:      "TEST_FLOAT_INT",
			setValue: "42",
			default_: 1.0,
			want:     42.0,
		},
		{
			name:     "invalid float",
			key:      "TEST_FLOAT_INVALID",
			setValue: "not-a-number",
			default_: 1.0,
			want:     1.0,
		},
		{
			name:     "empty env var",
			key:      "TEST_FLOAT_EMPTY",
			setValue: "",
			default_: 1.0,
			want:     1.0,
		},
		{
			name:     "env var not set",
			key:      "TEST_FLOAT_NOT_SET",
			setValue: "",
			default_: 2.5,
			want:     2.5,
		},
		{
			name:     "negative float",
			key:      "TEST_FLOAT_NEGATIVE",
			setValue: "-5.5",
			default_: 1.0,
			want:     -5.5,
		},
		{
			name:     "float with spaces",
			key:      "TEST_FLOAT_SPACES",
			setValue: "  7.5  ",
			default_: 1.0,
			want:     7.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			} else {
				os.Unsetenv(tt.key)
			}
			defer os.Unsetenv(tt.key)

			got := envFloat(tt.key, tt.default_)
			if got != tt.want {
				t.Errorf("envFloat(%q, %f) = %f, want %f", tt.key, tt.default_, got, tt.want)
			}
		})
	}
}

func TestEnvInt(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setValue string
		default_ int
		want     int
	}{
		{
			name:     "valid int",
			key:      "TEST_INT",
			setValue: "42",
			default_: 1,
			want:     42,
		},
		{
			name:     "zero",
			key:      "TEST_INT_ZERO",
			setValue: "0",
			default_: 1,
			want:     0,
		},
		{
			name:     "negative int",
			key:      "TEST_INT_NEGATIVE",
			setValue: "-10",
			default_: 1,
			want:     -10,
		},
		{
			name:     "invalid int",
			key:      "TEST_INT_INVALID",
			setValue: "not-a-number",
			default_: 1,
			want:     1,
		},
		{
			name:     "float string",
			key:      "TEST_INT_FLOAT",
			setValue: "3.14",
			default_: 1,
			want:     1,
		},
		{
			name:     "empty env var",
			key:      "TEST_INT_EMPTY",
			setValue: "",
			default_: 1,
			want:     1,
		},
		{
			name:     "env var not set",
			key:      "TEST_INT_NOT_SET",
			setValue: "",
			default_: 5,
			want:     5,
		},
		{
			name:     "int with spaces",
			key:      "TEST_INT_SPACES",
			setValue: "  99  ",
			default_: 1,
			want:     99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
			} else {
				os.Unsetenv(tt.key)
			}
			defer os.Unsetenv(tt.key)

			got := envInt(tt.key, tt.default_)
			if got != tt.want {
				t.Errorf("envInt(%q, %d) = %d, want %d", tt.key, tt.default_, got, tt.want)
			}
		})
	}
}
