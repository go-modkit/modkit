package config

import (
	"time"

	mkconfig "github.com/go-modkit/modkit/modkit/config"
	"github.com/go-modkit/modkit/modkit/module"
)

const (
	TokenHTTPAddr           module.Token = "config.http_addr"
	TokenMySQLDSN           module.Token = "config.mysql_dsn"
	TokenJWTSecret          module.Token = "config.jwt_secret"
	TokenJWTIssuer          module.Token = "config.jwt_issuer"
	TokenJWTTTL             module.Token = "config.jwt_ttl"
	TokenAuthUsername       module.Token = "config.auth_username"
	TokenAuthPassword       module.Token = "config.auth_password"
	TokenCORSAllowedOrigins module.Token = "config.cors_allowed_origins"
	TokenCORSAllowedMethods module.Token = "config.cors_allowed_methods"
	TokenCORSAllowedHeaders module.Token = "config.cors_allowed_headers"
	TokenRateLimitPerSecond module.Token = "config.rate_limit_per_second"
	TokenRateLimitBurst     module.Token = "config.rate_limit_burst"
)

var exportedTokens = []module.Token{
	TokenHTTPAddr,
	TokenMySQLDSN,
	TokenJWTSecret,
	TokenJWTIssuer,
	TokenJWTTTL,
	TokenAuthUsername,
	TokenAuthPassword,
	TokenCORSAllowedOrigins,
	TokenCORSAllowedMethods,
	TokenCORSAllowedHeaders,
	TokenRateLimitPerSecond,
	TokenRateLimitBurst,
}

type Options struct {
	Source mkconfig.Source
}

type Module struct {
	opts Options
}

func NewModule(opts Options) module.Module {
	return &Module{opts: opts}
}

func (m *Module) Definition() module.ModuleDef {
	httpAddrDefault := ":8080"
	mySQLDSNDefault := "root:password@tcp(localhost:3306)/app?parseTime=true&multiStatements=true"
	jwtSecretDefault := "dev-secret-change-me"
	jwtIssuerDefault := "hello-mysql"
	jwtTTLDefault := 1 * time.Hour
	authUsernameDefault := "demo"
	authPasswordDefault := "demo"
	corsOriginsDefault := []string{"http://localhost:3000"}
	corsMethodsDefault := []string{"GET", "POST", "PUT", "DELETE"}
	corsHeadersDefault := []string{"Content-Type", "Authorization"}
	rateLimitPerSecondDefault := 5.0
	rateLimitBurstDefault := 10

	configOptions := []mkconfig.Option{
		mkconfig.WithModuleName("hello-mysql.config.values"),
		mkconfig.WithTyped(TokenHTTPAddr, mkconfig.ValueSpec[string]{
			Key:     "HTTP_ADDR",
			Default: &httpAddrDefault,
			Parse:   mkconfig.ParseString,
		}, true),
		mkconfig.WithTyped(TokenMySQLDSN, mkconfig.ValueSpec[string]{
			Key:     "MYSQL_DSN",
			Default: &mySQLDSNDefault,
			Parse:   mkconfig.ParseString,
		}, true),
		mkconfig.WithTyped(TokenJWTSecret, mkconfig.ValueSpec[string]{
			Key:       "JWT_SECRET",
			Default:   &jwtSecretDefault,
			Sensitive: true,
			Parse:     mkconfig.ParseString,
		}, true),
		mkconfig.WithTyped(TokenJWTIssuer, mkconfig.ValueSpec[string]{
			Key:     "JWT_ISSUER",
			Default: &jwtIssuerDefault,
			Parse:   mkconfig.ParseString,
		}, true),
		mkconfig.WithTyped(TokenJWTTTL, mkconfig.ValueSpec[time.Duration]{
			Key:     "JWT_TTL",
			Default: &jwtTTLDefault,
			Parse:   mkconfig.ParseDuration,
		}, true),
		mkconfig.WithTyped(TokenAuthUsername, mkconfig.ValueSpec[string]{
			Key:     "AUTH_USERNAME",
			Default: &authUsernameDefault,
			Parse:   mkconfig.ParseString,
		}, true),
		mkconfig.WithTyped(TokenAuthPassword, mkconfig.ValueSpec[string]{
			Key:       "AUTH_PASSWORD",
			Default:   &authPasswordDefault,
			Sensitive: true,
			Parse:     mkconfig.ParseString,
		}, true),
		mkconfig.WithTyped(TokenCORSAllowedOrigins, mkconfig.ValueSpec[[]string]{
			Key:     "CORS_ALLOWED_ORIGINS",
			Default: &corsOriginsDefault,
			Parse:   mkconfig.ParseCSV,
		}, true),
		mkconfig.WithTyped(TokenCORSAllowedMethods, mkconfig.ValueSpec[[]string]{
			Key:     "CORS_ALLOWED_METHODS",
			Default: &corsMethodsDefault,
			Parse:   mkconfig.ParseCSV,
		}, true),
		mkconfig.WithTyped(TokenCORSAllowedHeaders, mkconfig.ValueSpec[[]string]{
			Key:     "CORS_ALLOWED_HEADERS",
			Default: &corsHeadersDefault,
			Parse:   mkconfig.ParseCSV,
		}, true),
		mkconfig.WithTyped(TokenRateLimitPerSecond, mkconfig.ValueSpec[float64]{
			Key:     "RATE_LIMIT_PER_SECOND",
			Default: &rateLimitPerSecondDefault,
			Parse:   mkconfig.ParseFloat64,
		}, true),
		mkconfig.WithTyped(TokenRateLimitBurst, mkconfig.ValueSpec[int]{
			Key:     "RATE_LIMIT_BURST",
			Default: &rateLimitBurstDefault,
			Parse:   mkconfig.ParseInt,
		}, true),
	}

	if m.opts.Source != nil {
		configOptions = append(configOptions, mkconfig.WithSource(m.opts.Source))
	}

	configModule := mkconfig.NewModule(configOptions...)

	return module.ModuleDef{
		Name:    "config",
		Imports: []module.Module{configModule},
		Exports: exportedTokens,
	}
}
