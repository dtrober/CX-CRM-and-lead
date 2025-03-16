package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dyrober/AgencyCRM/internal/config"
	"github.com/dyrober/AgencyCRM/internal/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	testRepo *Repository
	testDB   *sql.DB
)

// TestMain sets up and tears down the test database
func TestMain(m *testing.M) {
	// Skip repository tests if not explicitly enabled
	// This prevents them from running during regular unit test runs
	if os.Getenv("REPO_TESTS") != "true" {
		fmt.Println("Skipping repository tests; set REPO_TESTS=true to run")
		os.Exit(0)
	}

	// Set up test database
	var err error
	testDB, err = setupTestDB()
	if err != nil {
		fmt.Printf("Failed to set up test database: %v\n", err)
		os.Exit(1)
	}

	// Create repository with test database
	testRepo = NewRepository(testDB)

	// Run the tests
	code := m.Run()

	// Clean up
	if err := teardownTestDB(testDB); err != nil {
		fmt.Printf("Failed to tear down test database: %v\n", err)
	}

	os.Exit(code)
}

// setupTestDB creates a connection to the test database and sets up the schema
func setupTestDB() (*sql.DB, error) {
	// Create database configuration for tests
	dbConfig := config.DBConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvIntOrDefault("TEST_DB_PORT", 5432),
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("TEST_DB_NAME", "myapp_test"),
		SSLMode:  "disable",
	}

	// Connect to database
	db, err := sql.Open("pgx", dbConfig.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute)

	// Create test schema
	if err := createTestSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create test schema: %w", err)
	}

	return db, nil
}

// createTestSchema sets up the necessary tables for testing
func createTestSchema(db *sql.DB) error {
	// Clear any existing data and recreate tables
	_, err := db.Exec(`
		DROP TABLE IF EXISTS users;
		
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);
		
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`)
	return err
}

// teardownTestDB closes the database connection and performs cleanup
func teardownTestDB(db *sql.DB) error {
	// Clean up data (not dropping tables to avoid schema validation errors in other tests)
	_, err := db.Exec(`TRUNCATE TABLE users RESTART IDENTITY`)
	if err != nil {
		return err
	}
	return db.Close()
}

// Helper functions for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		var result int
		fmt.Sscanf(value, "%d", &result)
		return result
	}
	return defaultValue
}

// Test the GetUser function
func TestRepository_GetUser(t *testing.T) {
	// Create a test context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Insert a test user
	testUser := domain.User{
		Name:  "Test User " + testID,
		Email: fmt.Sprintf("testrepo_%s@example.com", testID),
	}

	userID, err := testRepo.CreateUser(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Test cases
	t.Run("existing user", func(t *testing.T) {
		// Get the user we just created
		user, err := testRepo.GetUser(ctx, userID)

		// Verify no error occurred
		if err != nil {
			t.Fatalf("Failed to get existing user: %v", err)
		}

		// Verify the user details
		if user.ID != userID {
			t.Errorf("Expected user ID %d, got %d", userID, user.ID)
		}

		if user.Name != testUser.Name {
			t.Errorf("Expected user name '%s', got '%s'", testUser.Name, user.Name)
		}

		if user.Email != testUser.Email {
			t.Errorf("Expected user email '%s', got '%s'", testUser.Email, user.Email)
		}

		// Verify timestamps are set
		if user.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}

		if user.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set")
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		// Try to get a user with an ID that doesn't exist
		_, err := testRepo.GetUser(ctx, 9999)

		// Verify we got an error
		if err == nil {
			t.Fatal("Expected error for non-existent user, got nil")
		}
	})
}

// Test the CreateUser function
func TestRepository_CreateUser(t *testing.T) {
	// Create a test context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("create new user", func(t *testing.T) {
		// Create a unique email for testing
		uniqueEmail := fmt.Sprintf("unique%d@example.com", time.Now().UnixNano())

		newUser := domain.User{
			Name:  "New Test User",
			Email: uniqueEmail,
		}

		// Create the user
		userID, err := testRepo.CreateUser(ctx, newUser)

		// Verify no error occurred
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Verify we got a valid ID
		if userID <= 0 {
			t.Errorf("Expected positive user ID, got %d", userID)
		}

		time.Sleep(100 * time.Millisecond)

		// Verify we can retrieve the user
		createdUser, err := testRepo.GetUser(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to get created user: %v", err)
		}

		if createdUser.Name != newUser.Name {
			t.Errorf("Expected user name '%s', got '%s'", newUser.Name, createdUser.Name)
		}

		if createdUser.Email != newUser.Email {
			t.Errorf("Expected user email '%s', got '%s'", newUser.Email, createdUser.Email)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		// Create a user
		user1 := domain.User{
			Name:  "Duplicate Test User 1",
			Email: "duplicate@example.com",
		}

		// Create the first user
		_, err := testRepo.CreateUser(ctx, user1)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Try to create another user with the same email
		user2 := domain.User{
			Name:  "Duplicate Test User 2",
			Email: "duplicate@example.com", // Same email as user1
		}

		// This should fail due to unique constraint on email
		_, err = testRepo.CreateUser(ctx, user2)

		// Verify we got an error
		if err == nil {
			t.Fatal("Expected error for duplicate email, got nil")
		}
	})
}
