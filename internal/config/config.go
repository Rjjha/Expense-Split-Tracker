package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Security SecurityConfig
	Logging  LoggingConfig
	Features FeatureConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	DSN      string
}

type ServerConfig struct {
	Host string
	Port int
	Env  string
}

type SecurityConfig struct {
	JWTSecret string
}

type LoggingConfig struct {
	Level string
}

type FeatureConfig struct {
	IdempotencyTTL time.Duration
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load("config.env"); err != nil {
		// It's okay if the file doesn't exist, we'll use environment variables
		fmt.Println("Warning: config.env file not found, using environment variables")
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "3306"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}

	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %v", err)
	}

	idempotencyTTLHours, err := strconv.Atoi(getEnv("IDEMPOTENCY_TTL_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid IDEMPOTENCY_TTL_HOURS: %v", err)
	}

	dbConfig := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     dbPort,
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", "password"),
		Name:     getEnv("DB_NAME", "expense_split_tracker"),
	}

	// Create DSN
	dbConfig.DSN = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)

	config := &Config{
		Database: dbConfig,
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: serverPort,
			Env:  getEnv("ENV", "development"),
		},
		Security: SecurityConfig{
			JWTSecret: getEnv("JWT_SECRET", "default-jwt-secret-change-in-production"),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Features: FeatureConfig{
			IdempotencyTTL: time.Duration(idempotencyTTLHours) * time.Hour,
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
