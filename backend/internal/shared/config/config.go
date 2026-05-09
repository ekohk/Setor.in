// Package config loads runtime configuration from environment variables.
//
// Order of precedence (highest first):
//   1. Real environment variables
//   2. .env file in working directory (loaded via godotenv)
//   3. Defaults defined here
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all runtime settings.
type Config struct {
	App      AppConfig
	DB       DBConfig
	Keycloak KeycloakConfig
	JWT      JWTConfig
	CORS     CORSConfig
	SMTP     SMTPConfig
	Rate     RateConfig
}

type AppConfig struct {
	Name     string
	Env      string // development | staging | production
	Port     int
	LogLevel string // debug | info | warn | error
}

type DBConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	DSN             string // computed
}

type KeycloakConfig struct {
	BaseURL             string
	Realm               string
	Issuer              string
	JWKSURL             string
	BackendClientID     string
	BackendClientSecret string
	WebClientID         string
}

type JWTConfig struct {
	Audience          []string
	ClockSkewSeconds  int
	JWKSCacheTTL      time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

type RateConfig struct {
	AuthSyncPerMin int
}

// IsProduction returns true if APP_ENV=production.
func (c *Config) IsProduction() bool { return c.App.Env == "production" }

// IsDevelopment returns true if APP_ENV=development.
func (c *Config) IsDevelopment() bool { return c.App.Env == "development" }

// Load reads config from .env (if present) and environment variables.
// It returns an error if any required field is missing or malformed.
func Load() (*Config, error) {
	// Best-effort load .env; absence is not fatal (production uses real env vars).
	_ = godotenv.Load()

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefaults(v)

	c := &Config{
		App: AppConfig{
			Name:     v.GetString("APP_NAME"),
			Env:      v.GetString("APP_ENV"),
			Port:     v.GetInt("APP_PORT"),
			LogLevel: v.GetString("APP_LOG_LEVEL"),
		},
		DB: DBConfig{
			Host:            v.GetString("DB_HOST"),
			Port:            v.GetInt("DB_PORT"),
			User:            v.GetString("DB_USER"),
			Password:        v.GetString("DB_PASSWORD"),
			Name:            v.GetString("DB_NAME"),
			SSLMode:         v.GetString("DB_SSLMODE"),
			MaxConns:        int32(v.GetInt("DB_MAX_CONNS")),
			MinConns:        int32(v.GetInt("DB_MIN_CONNS")),
			MaxConnLifetime: v.GetDuration("DB_MAX_CONN_LIFETIME"),
		},
		Keycloak: KeycloakConfig{
			BaseURL:             v.GetString("KEYCLOAK_BASE_URL"),
			Realm:               v.GetString("KEYCLOAK_REALM"),
			Issuer:              v.GetString("KEYCLOAK_ISSUER"),
			JWKSURL:             v.GetString("KEYCLOAK_JWKS_URL"),
			BackendClientID:     v.GetString("KEYCLOAK_BACKEND_CLIENT_ID"),
			BackendClientSecret: v.GetString("KEYCLOAK_BACKEND_CLIENT_SECRET"),
			WebClientID:         v.GetString("KEYCLOAK_WEB_CLIENT_ID"),
		},
		JWT: JWTConfig{
			Audience:         splitCSV(v.GetString("JWT_AUDIENCE")),
			ClockSkewSeconds: v.GetInt("JWT_CLOCK_SKEW_SECONDS"),
			JWKSCacheTTL:     v.GetDuration("JWKS_CACHE_TTL"),
		},
		CORS: CORSConfig{
			AllowedOrigins: splitCSV(v.GetString("CORS_ALLOWED_ORIGINS")),
		},
		SMTP: SMTPConfig{
			Host:     v.GetString("SMTP_HOST"),
			Port:     v.GetInt("SMTP_PORT"),
			Username: v.GetString("SMTP_USERNAME"),
			Password: v.GetString("SMTP_PASSWORD"),
			From:     v.GetString("SMTP_FROM"),
			FromName: v.GetString("SMTP_FROM_NAME"),
		},
		Rate: RateConfig{
			AuthSyncPerMin: v.GetInt("RATE_LIMIT_AUTH_SYNC_PER_MIN"),
		},
	}

	c.DB.DSN = buildDSN(c.DB)

	if err := validate(c); err != nil {
		return nil, err
	}
	return c, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_NAME", "setorin-api")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_PORT", 8000)
	v.SetDefault("APP_LOG_LEVEL", "info")

	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_MAX_CONNS", 10)
	v.SetDefault("DB_MIN_CONNS", 2)
	v.SetDefault("DB_MAX_CONN_LIFETIME", "30m")

	v.SetDefault("JWT_CLOCK_SKEW_SECONDS", 60)
	v.SetDefault("JWKS_CACHE_TTL", "1h")

	v.SetDefault("RATE_LIMIT_AUTH_SYNC_PER_MIN", 10)
}

// validate ensures required fields are present.
func validate(c *Config) error {
	var missing []string
	check := func(name, val string) {
		if strings.TrimSpace(val) == "" {
			missing = append(missing, name)
		}
	}

	check("DB_HOST", c.DB.Host)
	check("DB_USER", c.DB.User)
	check("DB_PASSWORD", c.DB.Password)
	check("DB_NAME", c.DB.Name)

	check("KEYCLOAK_BASE_URL", c.Keycloak.BaseURL)
	check("KEYCLOAK_REALM", c.Keycloak.Realm)
	check("KEYCLOAK_ISSUER", c.Keycloak.Issuer)
	check("KEYCLOAK_JWKS_URL", c.Keycloak.JWKSURL)
	check("KEYCLOAK_BACKEND_CLIENT_ID", c.Keycloak.BackendClientID)
	check("KEYCLOAK_BACKEND_CLIENT_SECRET", c.Keycloak.BackendClientSecret)

	if c.Keycloak.BackendClientSecret == "CHANGE_ME_AFTER_KEYCLOAK_SETUP" {
		return errors.New("KEYCLOAK_BACKEND_CLIENT_SECRET still set to placeholder; update .env with real client secret from Keycloak")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	if c.DB.Port == 0 {
		return errors.New("DB_PORT must be set")
	}
	return nil
}

func buildDSN(d DBConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
