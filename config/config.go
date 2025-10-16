package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type (
	Config struct {
		App       AppConfig
		Log       LogConfig
		Database  DatabaseConfig
		HTTP      HTTPConfig
		CORS      CORSConfig
		RateLimit RateLimitConfig
		Auth      AuthConfig
	}

	AppConfig struct {
		Name string
		Env  string // dev|staging|prod
	}

	LogConfig struct {
		Level  string // debug|info|warn|error
		Format string // json|text
		File   string // path ke log file (kosong = stderr saja)
		AddSrc bool   // true untuk AddSource
	}

	DatabaseConfig struct {
		URL             string
		Host            string
		Port            int
		User            string
		Pass            string
		Name            string
		SSLMode         string
		MaxConns        int32
		MinConns        int32
		MaxConnLifetime time.Duration
		MaxConnIdleTime time.Duration
		HealthTimeout   time.Duration
	}

	HTTPConfig struct {
		Host           string
		Port           int
		Prefork        bool
		ReadTimeout    time.Duration
		WriteTimeout   time.Duration
		IdleTimeout    time.Duration
		BodyLimitBytes int
		EnableETag     bool
		BaseURL        string
	}

	CORSConfig struct {
		AllowOrigins  string
		AllowMethods  string
		AllowHeaders  string
		ExposeHeaders string
		Credentials   bool
	}

	RateLimitConfig struct {
		Enabled   bool
		Max       int
		Window    time.Duration
		KeyHeader string
	}

	AuthConfig struct {
		GuestEnabled       bool
		GuestRatePerMinute int
		JWTSecret          string        // minimal 32 chars
		JWTAccessTTL       time.Duration // ex: 15m
		JWTRefreshTTL      time.Duration // ex: 720h (30d)
	}
)

func atoiDef(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func Parse() *Config {
	app := AppConfig{
		Name: os.Getenv("APP_NAME"),
		Env:  os.Getenv("APP_ENV"),
	}

	log := LogConfig{
		Level:  os.Getenv("LOG_LEVEL"),
		Format: os.Getenv("LOG_FORMAT"),
		File:   os.Getenv("LOG_FILE"),
		AddSrc: os.Getenv("LOG_ADD_SOURCE") == "true",
	}

	database := DatabaseConfig{
		URL:             os.Getenv("DATABASE_URL"),
		Host:            os.Getenv("DB_HOST"),
		Port:            atoiDef(os.Getenv("DB_PORT"), 5432),
		User:            os.Getenv("DB_USER"),
		Pass:            os.Getenv("DB_PASSWORD"),
		Name:            os.Getenv("DB_NAME"),
		SSLMode:         os.Getenv("DB_SSLMODE"),
		MaxConns:        int32(atoiDef(os.Getenv("DB_MAX_CONNS"), 15)),
		MinConns:        int32(atoiDef(os.Getenv("DB_MIN_CONNS"), 2)),
		MaxConnLifetime: time.Duration(atoiDef(os.Getenv("DB_MAX_CONN_LIFETIME_SEC"), 3600)) * time.Second,
		MaxConnIdleTime: time.Duration(atoiDef(os.Getenv("DB_MAX_CONN_IDLE_SEC"), 300)) * time.Second,
		HealthTimeout:   time.Duration(atoiDef(os.Getenv("DB_HEALTH_TIMEOUT_MS"), 1500)) * time.Millisecond,
	}
	if database.URL == "" {
		database.URL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			database.User, database.Pass, database.Host, database.Port, database.Name, database.SSLMode,
		)
	}

	http := HTTPConfig{
		Host:           os.Getenv("HTTP_HOST"),
		Port:           atoiDef(os.Getenv("HTTP_PORT"), 8080),
		Prefork:        os.Getenv("HTTP_PREFORK") == "true",
		ReadTimeout:    time.Duration(atoiDef(os.Getenv("HTTP_READ_TIMEOUT_MS"), 10000)) * time.Millisecond,
		WriteTimeout:   time.Duration(atoiDef(os.Getenv("HTTP_WRITE_TIMEOUT_MS"), 10000)) * time.Millisecond,
		IdleTimeout:    time.Duration(atoiDef(os.Getenv("HTTP_IDLE_TIMEOUT_MS"), 60000)) * time.Millisecond,
		BodyLimitBytes: atoiDef(os.Getenv("HTTP_BODY_LIMIT_BYTES"), 10<<20), // 10MB
		EnableETag:     os.Getenv("HTTP_ETAG") == "true",
		BaseURL:        os.Getenv("HTTP_BASE_URL"),
	}

	cors := CORSConfig{
		AllowOrigins:  os.Getenv("CORS_ALLOW_ORIGINS"),
		AllowMethods:  os.Getenv("CORS_ALLOW_METHODS"),
		AllowHeaders:  os.Getenv("CORS_ALLOW_HEADERS"),
		ExposeHeaders: os.Getenv("CORS_EXPOSE_HEADERS"),
		Credentials:   os.Getenv("CORS_CREDENTIALS") == "true",
	}

	rateLimit := RateLimitConfig{
		Enabled:   os.Getenv("RATE_LIMIT_ENABLED") == "true",
		Max:       atoiDef(os.Getenv("RATE_LIMIT_MAX"), 120),
		Window:    time.Duration(atoiDef(os.Getenv("RATE_LIMIT_WINDOW_SEC"), 60)) * time.Second,
		KeyHeader: os.Getenv("RATE_LIMIT_KEY_HEADER"),
	}

	auth := AuthConfig{
		GuestEnabled:       os.Getenv("GUEST_ENABLED") == "true",
		GuestRatePerMinute: atoiDef(os.Getenv("GUEST_SIGNIN_RATE_PER_MIN"), 10),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		JWTAccessTTL:       time.Duration(atoiDef(os.Getenv("JWT_ACCESS_TTL_MIN"), 15)) * time.Minute,
		JWTRefreshTTL:      time.Duration(atoiDef(os.Getenv("JWT_REFRESH_TTL_HOURS"), 720)) * time.Hour,
	}

	cfg := &Config{
		App:       app,
		Log:       log,
		Database:  database,
		HTTP:      http,
		CORS:      cors,
		RateLimit: rateLimit,
		Auth:      auth,
	}

	return cfg
}
