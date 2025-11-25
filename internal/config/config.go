package config

import (
	"context"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
)

type Config struct {
	Environment   string
	HTTPAddress   string
	GracefulDelay time.Duration `env:"GRACEFUL_DELAY" envDefault:"5s"`

	// Shared Tenant Database (NEW)
	TenantDatabase DatabaseConfig

	// Service-specific Database
	Database DatabaseConfig

	Security       SecurityConfig
	Storage        StorageConfig
	InternalSecret string `env:"FREGO_INTERNAL_SECRET"`
}

type DatabaseConfig struct {
	URL               string        `env:"URL,required"`
	MaxOpenConns      int           `env:"MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns      int           `env:"MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime   time.Duration `env:"CONN_MAX_LIFETIME" envDefault:"5m"`
	PreferSimpleProto bool          `env:"PREFER_SIMPLE_PROTO" envDefault:"false"`
	User              string        `env:"USER" envDefault:"erp_user"`
}

type SecurityConfig struct {
	KeycloakIssuer      string    `env:"KEYCLOAK_ISSUER,required"`
	KeycloakAudience    []string  `env:"KEYCLOAK_AUDIENCE" envSeparator:","`
	KeycloakTenantClaim string    `env:"KEYCLOAK_TENANT_CLAIM" envDefault:"tenant_id"`
	DefaultTenant       uuid.UUID `env:"DEFAULT_TENANT"`
	AllowedOrigins      []string  `env:"ALLOWED_ORIGINS" envSeparator:","`
}

type StorageConfig struct {
	Bucket          string `env:"S3_BUCKET"`
	Region          string `env:"S3_REGION"`
	Endpoint        string `env:"S3_ENDPOINT"`
	AccessKeyID     string `env:"S3_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"S3_SECRET_ACCESS_KEY"`
	UsePathStyle    bool   `env:"S3_USE_PATH_STYLE" envDefault:"false"`
	KeyPrefix       string `env:"S3_KEY_PREFIX" envDefault:"finance/"`
	MaxUploadSize   int64  `env:"MAX_UPLOAD_SIZE" envDefault:"10485760"` // 10MB
}

func Load(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	// Load environment
	cfg.Environment = getEnvOrDefault("ENVIRONMENT", "development")
	cfg.HTTPAddress = getEnvOrDefault("HTTP_ADDRESS", ":8080")

	// Load tenant database config
	if err := env.ParseWithOptions(&cfg.TenantDatabase, env.Options{
		Prefix: "TENANT_DB_",
	}); err != nil {
		return nil, fmt.Errorf("parse tenant database config: %w", err)
	}

	// Load service database config
	if err := env.ParseWithOptions(&cfg.Database, env.Options{
		Prefix: "DB_",
	}); err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	// Load security config
	if err := env.Parse(&cfg.Security); err != nil {
		return nil, fmt.Errorf("parse security config: %w", err)
	}

	// Load storage config
	if err := env.Parse(&cfg.Storage); err != nil {
		return nil, fmt.Errorf("parse storage config: %w", err)
	}

	// Parse graceful delay
	if delayStr := getEnvOrDefault("GRACEFUL_DELAY", "5s"); delayStr != "" {
		if d, err := time.ParseDuration(delayStr); err == nil {
			cfg.GracefulDelay = d
		}
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := getEnv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnv(key string) string {
	// Implementation depends on your env package
	return ""
}
