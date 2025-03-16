package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dyrober/AgencyCRM/internal/domain"
	"github.com/stretchr/testify/assert"
)

// TestE2E runs end-to-end tests against a deployed API server
func TestE2E(t *testing.T) {
	// Skip if not in e2e test environment
	if os.Getenv("E2E_TESTS") != "true" {
		t.Skip("Skipping e2e test; set E2E_TESTS=true to run")
	}

	// Get API base URL from environment or use default
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Test the health check endpoint
	t.Run("Health check", testHealthCheck(baseURL))

	// Test user creation and retrieval
	t.Run("User CRUD operations", testUserOperations(baseURL))

	// Test error handling
	t.Run("Error handling", testErrorHandling(baseURL))
}

// Test health check endpoint
func testHealthCheck(baseURL string) func(t *testing.T) {
	return func(t *testing.T) {
		// Send request to health check endpoint
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer resp.Body.Close()

		// Check status code
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Health check should return 200 OK")

		// Decode response
		var response map[string]string
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err, "Response should be valid JSON")

		// Validate response content
		assert.Equal(t, "ok", response["status"], "Health check should return status 'ok'")
		assert.NotEmpty(t, response["time"], "Health check should include server time")
	}
}

// Test user creation and retrieval
func testUserOperations(baseURL string) func(t *testing.T) {
	return func(t *testing.T) {
		// Create unique user data for testing
		timestamp := time.Now().UnixNano()
		userName := fmt.Sprintf("E2E Test User %d", timestamp)
		email := fmt.Sprintf("e2e_test_%d@example.com", timestamp)

		// 1. Create a user
		userID := createUser(t, baseURL, userName, email)
		t.Logf("Created user with ID: %d", userID)

		// 2. Get the user and verify details
		getUserAndVerify(t, baseURL, userID, userName, email)

		// 3. Try to create a user with the same email (should fail)
		testDuplicateEmail(t, baseURL, userName+" Duplicate", email)
	}
}

// Test error handling
func testErrorHandling(baseURL string) func(t *testing.T) {
	return func(t *testing.T) {
		// Test invalid user ID
		t.Run("Invalid user ID", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/api/v1/users/invalid", baseURL))
			assert.NoError(t, err, "Request should not fail")
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request for invalid ID")

			var errorResp domain.ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&errorResp)
			assert.NoError(t, err, "Response should be valid JSON")
			assert.Contains(t, errorResp.Error, "Invalid user ID", "Error message should be descriptive")
		})

		// Test non-existent user ID
		t.Run("Non-existent user ID", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/api/v1/users/99999", baseURL))
			assert.NoError(t, err, "Request should not fail")
			defer resp.Body.Close()

			// Your API might return 404 Not Found or 500 Internal Server Error depending on implementation
			assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusInternalServerError,
				"Should return appropriate error status for non-existent user")
		})

		// Test missing required fields
		t.Run("Missing required fields", func(t *testing.T) {
			// Create an incomplete user request
			incompleteReq := domain.CreateUserRequest{
				Name:  "", // Missing name
				Email: "incomplete@example.com",
			}

			body, err := json.Marshal(incompleteReq)
			assert.NoError(t, err, "Should marshal request")

			resp, err := http.Post(baseURL+"/api/v1/users", "application/json", bytes.NewBuffer(body))
			assert.NoError(t, err, "Request should not fail")
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request for missing fields")

			var errorResp domain.ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&errorResp)
			assert.NoError(t, err, "Response should be valid JSON")
			assert.Contains(t, errorResp.Error, "required", "Error message should indicate missing fields")
		})
	}
}

// Helper function to test duplicate email handling
func testDuplicateEmail(t *testing.T, baseURL, name, email string) {
	createReq := domain.CreateUserRequest{
		Name:  name,
		Email: email,
	}

	body, err := json.Marshal(createReq)
	assert.NoError(t, err, "Should marshal request")

	resp, err := http.Post(baseURL+"/api/v1/users", "application/json", bytes.NewBuffer(body))
	assert.NoError(t, err, "Request should not fail")
	defer resp.Body.Close()

	// Should return an error status code (likely 400 or 500 depending on implementation)
	assert.NotEqual(t, http.StatusCreated, resp.StatusCode, "Should not create user with duplicate email")
}

// Helper function to create a user
func createUser(t *testing.T, baseURL, name, email string) int {
	createReq := domain.CreateUserRequest{
		Name:  name,
		Email: email,
	}

	body, err := json.Marshal(createReq)
	assert.NoError(t, err, "Should marshal request")

	resp, err := http.Post(baseURL+"/api/v1/users", "application/json", bytes.NewBuffer(body))
	assert.NoError(t, err, "Request to create user should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Should return 201 Created status")

	var createResp map[string]int
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	assert.NoError(t, err, "Response should be valid JSON")

	userID, exists := createResp["id"]
	assert.True(t, exists, "Response should contain user ID")
	assert.Greater(t, userID, 0, "User ID should be positive")

	return userID
}

// Helper function to get a user and verify their details
func getUserAndVerify(t *testing.T, baseURL string, id int, expectedName, expectedEmail string) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/users/%d", baseURL, id))
	assert.NoError(t, err, "Request to get user should not fail")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return 200 OK status")

	var user domain.UserResponse
	err = json.NewDecoder(resp.Body).Decode(&user)
	assert.NoError(t, err, "Response should be valid JSON")

	assert.Equal(t, id, user.ID, "User ID should match")
	assert.Equal(t, expectedName, user.Name, "User name should match")
	assert.Equal(t, expectedEmail, user.Email, "User email should match")
	assert.False(t, user.CreatedAt.IsZero(), "Created timestamp should be set")
}
