package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Setup test environment variables
	os.Setenv("SERVER_ADDRESS", ":8085")
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")

	// Clean up after the test
	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
	}()

	// Load the config
	cfg, err := Load()

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.ServerAddress != ":8085" {
		t.Errorf("Expected ServerAddress to be ':8085', got '%s'", cfg.ServerAddress)
	}

	if cfg.DB.Host != "testhost" {
		t.Errorf("Expected DB.Host to be 'testhost', got '%s'", cfg.DB.Host)
	}

	if cfg.DB.Port != 5433 {
		t.Errorf("Expected DB.Port to be 5433, got %d", cfg.DB.Port)
	}

	if cfg.DB.User != "testuser" {
		t.Errorf("Expected DB.User to be 'testuser', got '%s'", cfg.DB.User)
	}

	if cfg.DB.Password != "testpass" {
		t.Errorf("Expected DB.Password to be 'testpass', got '%s'", cfg.DB.Password)
	}

	if cfg.DB.DBName != "testdb" {
		t.Errorf("Expected DB.DBName to be 'testdb', got '%s'", cfg.DB.DBName)
	}
}

func TestDBConfig_DSN(t *testing.T) {
	dbConfig := DBConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "myapp",
		SSLMode:  "disable",
	}

	expected := "host=localhost port=5432 user=postgres password=password dbname=myapp sslmode=disable"
	actual := dbConfig.DSN()

	if actual != expected {
		t.Errorf("Expected DSN '%s', got '%s'", expected, actual)
	}
}
