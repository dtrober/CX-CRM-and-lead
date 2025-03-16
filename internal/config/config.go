package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// this holds all the configs for the app
type Config struct {
	ServerAddress      string
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	DB                 DBConfig
	StaticDir          string
	TemplatesDir       string
}

// This holds the configs for the DB
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN returns the PostgresSQL connection string
func (c *DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	readTimeout, err := strconv.Atoi(getEnv("SERVER_READ_TIMEOUT", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := strconv.Atoi(getEnv("SERVER_WRITE_TIMEOUT", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_WRITE_TIMEOUT: %w", err)
	}

	return &Config{
		ServerAddress:      getEnv("SERVER_ADDRESS", ":8080"),
		ServerReadTimeout:  time.Duration(readTimeout) * time.Second,
		ServerWriteTimeout: time.Duration(writeTimeout) * time.Second,
		StaticDir:          getEnv("STATIC_DIR", "/app/web/static"),
		TemplatesDir:       getEnv("TEMPLATES_DIR", "/app/web/templates"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "myapp"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
