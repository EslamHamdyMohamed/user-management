package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	App      AppConfig      `yaml:"app"`
}

type ServerConfig struct {
	Port            string        `yaml:"port" env:"PORT" env-default:"8080"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-default:"10s"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" env-default:"60s"`
	CORS            CORSConfig    `yaml:"cors"`
}

type RedisConfig struct {
	URL string `yaml:"url" env:"REDIS_URL" env-default:"redis://localhost:6379"`
}

type CORSConfig struct {
	AllowOrigins []string `yaml:"allow_origins" env:"CORS_ALLOW_ORIGINS" env-default:"*"`
	AllowMethods []string `yaml:"allow_methods" env:"CORS_ALLOW_METHODS" env-default:"GET,POST,PUT,DELETE,OPTIONS"`
	AllowHeaders []string `yaml:"allow_headers" env:"CORS_ALLOW_HEADERS" env-default:"Origin,Content-Type,Accept,Authorization"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port            string        `yaml:"port" env:"DB_PORT" env-default:"5432"`
	Name            string        `yaml:"name" env:"DB_NAME" env-default:"userdb"`
	User            string        `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password        string        `yaml:"password" env:"DB_PASS" env-default:"password"`
	SSLMode         string        `yaml:"ssl_mode" env:"DB_SSL_MODE" env-default:"disable"`
	MaxOpenConns    int           `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS" env-default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" env-default:"10"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME" env-default:"5m"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env:"DB_CONN_MAX_IDLE_TIME" env-default:"1m"`
}

type JWTConfig struct {
	Secret            string        `yaml:"secret" env:"JWT_SECRET" env-default:"your-super-secret-jwt-key-change-in-production"`
	AccessExpiration  time.Duration `yaml:"expiration" env:"JWT_EXPIRY" env-default:"24h"`
	RefreshExpiration time.Duration `yaml:"refresh_expiration" env:"JWT_REFRESH_EXPIRY" env-default:"168h"`
	Issuer            string        `yaml:"issuer" env:"JWT_ISSUER" env-default:"user-management"`
}

type AppConfig struct {
	Environment string `yaml:"environment" env:"APP_ENV" env-default:"development"`
	LogLevel    string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
	Version     string `yaml:"version" env:"APP_VERSION" env-default:"1.0.0"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Detect environment
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if env == "" {
		env = "development"
	}

	// Load .env only in non-production environments
	if env != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println(" .env file not found, using environment variables")
		}
	}

	// Load from environment variables
	if err := loadFromEnv(cfg); err != nil {
		return nil, err
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	log.Printf("✅ Config loaded successfully [%s]\n", env)

	return cfg, nil
}

func (c *Config) Validate() error {
	// --- Server ---
	if c.Server.Port == "" {
		return errors.New("server.port is required")
	}

	// --- Database ---
	db := c.Database
	if db.Host == "" || db.Port == "" || db.Name == "" || db.User == "" {
		return errors.New("database configuration incomplete")
	}

	if c.App.Environment == "production" && db.Password == "" {
		return errors.New("DB_PASS must be set in production")
	}

	// --- JWT ---
	jwt := c.JWT
	if jwt.Secret == "" {
		return errors.New("JWT_SECRET is required")
	}

	if c.App.Environment == "production" &&
		strings.Contains(jwt.Secret, "change-in-production") {
		return errors.New("insecure JWT secret used in production")
	}

	// --- App ---
	switch c.App.Environment {
	case "development", "staging", "production":
	default:
		return fmt.Errorf("invalid APP_ENV: %s", c.App.Environment)
	}

	switch c.App.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid LOG_LEVEL: %s", c.App.LogLevel)
	}

	return nil
}

func loadFromEnv(cfg *Config) error {
	cfg.Server.Port = getEnv("PORT", "8082")
	cfg.Server.ShutdownTimeout, _ = time.ParseDuration(getEnv("SHUTDOWN_TIMEOUT", "10s"))
	cfg.Server.ReadTimeout, _ = time.ParseDuration(getEnv("READ_TIMEOUT", "5s"))
	cfg.Server.WriteTimeout, _ = time.ParseDuration(getEnv("WRITE_TIMEOUT", "10s"))
	cfg.Server.IdleTimeout, _ = time.ParseDuration(getEnv("IDLE_TIMEOUT", "60s"))

	// CORS
	cfg.Server.CORS.AllowOrigins = getEnvSlice("CORS_ALLOW_ORIGINS", []string{"*"})
	cfg.Server.CORS.AllowMethods = getEnvSlice("CORS_ALLOW_METHODS", []string{
		"GET", "POST", "PUT", "DELETE", "OPTIONS",
	})
	cfg.Server.CORS.AllowHeaders = getEnvSlice("CORS_ALLOW_HEADERS", []string{
		"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With",
	})

	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.Name = getEnv("DB_NAME", "userdb")
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASS", "password")
	cfg.Database.SSLMode = getEnv("DB_SSL_MODE", "disable")

	// Parse database timeouts
	cfg.Database.ConnMaxLifetime, _ = time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m"))
	cfg.Database.ConnMaxIdleTime, _ = time.ParseDuration(getEnv("DB_CONN_MAX_IDLE_TIME", "1m"))
	cfg.Database.MaxOpenConns = 25 // Simplified
	cfg.Database.MaxIdleConns = 10

	cfg.Redis.URL = getEnv("REDIS_URL", "redis://localhost:6379")

	cfg.JWT.Secret = getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production")
	cfg.JWT.AccessExpiration, _ = time.ParseDuration(getEnv("JWT_ACCESS_EXPIRY", "24h"))
	cfg.JWT.RefreshExpiration, _ = time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"))
	cfg.JWT.Issuer = getEnv("JWT_ISSUER", "user-management")

	cfg.App.Environment = getEnv("APP_ENV", "development")
	cfg.App.LogLevel = getEnv("LOG_LEVEL", "info")
	cfg.App.Version = getEnv("APP_VERSION", "1.0.0")

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvSlice(key string, defaultVal []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultVal
}
