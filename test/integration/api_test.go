package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dyrober/AgencyCRM/internal/config"
	"github.com/dyrober/AgencyCRM/internal/domain"
	"github.com/dyrober/AgencyCRM/internal/repository"
	"github.com/dyrober/AgencyCRM/internal/server"
	"github.com/dyrober/AgencyCRM/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Setup for integration tests using a test database
var (
	testDB     *sql.DB
	testServer *httptest.Server
	baseURL    string
)

func TestMain(m *testing.M) {
	// Set up integration test environment
	if err := setupIntegrationTest(); err != nil {
		fmt.Printf("Failed to set up integration test: %v\n", err)
		os.Exit(1)
	}

	// Run the tests
	code := m.Run()

	// Clean up
	testServer.Close()
	testDB.Close()

	os.Exit(code)
}

func setupIntegrationTest() error {
	// Set up test database connection
	// NOTE: In a real environment, you might want to use Docker to spin up a test database
	// For this example, we'll assume you have a test database already set up
	dbConfig := config.DBConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("TEST_DB_NAME", "myapp_test"),
		SSLMode:  "disable",
	}

	// Connect to test database
	var err error
	testDB, err = sql.Open("pgx", dbConfig.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Create tables for testing
	if err := setupTestDatabase(testDB); err != nil {
		return fmt.Errorf("failed to set up test database: %w", err)
	}

	// Create repository, service and server
	repo := repository.NewRepository(testDB)
	svc := service.NewService(repo)

	// Create server config
	serverConfig := &config.Config{
		ServerAddress:      ":0", // Let the OS choose a port
		ServerReadTimeout:  10 * time.Second,
		ServerWriteTimeout: 10 * time.Second,
	}

	// Create HTTP server
	srv := server.NewServer(serverConfig, svc)

	// Start test server
	testServer = httptest.NewServer(srv.Server.Handler)
	baseURL = testServer.URL

	return nil
}

func setupTestDatabase(db *sql.DB) error {
	// Clear any existing data and set up tables
	_, err := db.Exec(`
        DROP TABLE IF EXISTS users;
        
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) NOT NULL UNIQUE,
            created_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL
        );
    `)
	return err
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Integration tests
func TestUserAPIIntegration(t *testing.T) {
	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test; set INTEGRATION_TESTS=true to run")
	}

	// Test creating a user
	t.Run("Create user", func(t *testing.T) {
		// Create a unique email for testing
		email := fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
		createReq := domain.CreateUserRequest{
			Name:  "Integration Test User",
			Email: email,
		}

		// Create the user
		userID := createUserForTest(t, createReq)

		// Get the user and verify
		getUserAndVerify(t, userID, createReq.Name, createReq.Email)
	})
}

func createUserForTest(t *testing.T, req domain.CreateUserRequest) int {
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	resp, err := http.Post(baseURL+"/api/v1/users", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var createResp map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	userID, exists := createResp["id"]
	if !exists {
		t.Fatalf("Response does not contain user ID")
	}

	return userID
}

func getUserAndVerify(t *testing.T, id int, expectedName, expectedEmail string) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/users/%d", baseURL, id))
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var user domain.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.Name != expectedName {
		t.Errorf("Expected user name %s, got %s", expectedName, user.Name)
	}

	if user.Email != expectedEmail {
		t.Errorf("Expected user email %s, got %s", expectedEmail, user.Email)
	}
}
