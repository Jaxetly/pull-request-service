package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type ServerConfig struct {
	Port int
}

type ValidationError struct {
	MissingVars []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("missing required environment variables: %s",
		strings.Join(e.MissingVars, ", "))
}

func Load() (Config, error) {
	missing := &ValidationError{MissingVars: []string{}}

	dbHost := getEnvRequired("POSTGRES_HOST", missing)
	dbPort := getEnvRequiredAsInt("POSTGRES_PORT", missing)
	dbUser := getEnvRequired("POSTGRES_USER", missing)
	dbPassword := getEnvRequired("POSTGRES_PASSWORD", missing)
	dbName := getEnvRequired("POSTGRES_DB", missing)

	appPort := getEnvOrDefaultAsInt("APP_PORT", 8080)
	maxConns := int32(getEnvOrDefaultAsInt("DB_MAX_CONNS", 10))
	minConns := int32(getEnvOrDefaultAsInt("DB_MIN_CONNS", 2))

	if len(missing.MissingVars) > 0 {
		return Config{}, missing
	}

	cfg := Config{
		Database: DatabaseConfig{
			Host:            dbHost,
			Port:            dbPort,
			User:            dbUser,
			Password:        dbPassword,
			DBName:          dbName,
			MaxConns:        maxConns,
			MinConns:        minConns,
			MaxConnLifetime: time.Hour,
			MaxConnIdleTime: 30 * time.Minute,
		},
		Server: ServerConfig{
			Port: appPort,
		},
	}

	return cfg, nil
}

func (c *DatabaseConfig) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
	)
}

func getEnvRequired(key string, missing *ValidationError) string {
	value := os.Getenv(key)
	if value == "" {
		missing.MissingVars = append(missing.MissingVars, key)
		return ""
	}
	return value
}

func getEnvRequiredAsInt(key string, missing *ValidationError) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		missing.MissingVars = append(missing.MissingVars, key)
		return 0
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		missing.MissingVars = append(missing.MissingVars,
			fmt.Sprintf("%s (invalid integer: %s)", key, valueStr))
		return 0
	}

	return value
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	// TODO: warning
	return defaultValue
}

func getEnvOrDefaultAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	// TODO: warning
	return defaultValue
}
