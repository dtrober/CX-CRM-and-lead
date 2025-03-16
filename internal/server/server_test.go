package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/dyrober/AgencyCRM/internal/config"
	"github.com/dyrober/AgencyCRM/internal/domain"
	"github.com/dyrober/AgencyCRM/internal/repository"
	"github.com/dyrober/AgencyCRM/internal/service"
	"github.com/go-chi/chi/v5"
)

func setupTestServer() (*Server, *repository.MockRepository) {
	// Create a mock repository
	mockRepo := repository.NewMockRepository()

	// Create a service with the mock repository
	svc := service.NewService(mockRepo)

	// Create a minimal config for testing
	cfg := &config.Config{
		ServerAddress:      ":8080",
		ServerReadTimeout:  10 * time.Second,
		ServerWriteTimeout: 10 * time.Second,
	}

	// Create a server with the service
	srv := NewServer(cfg, svc)

	return srv, mockRepo
}

func TestHealthCheck(t *testing.T) {
	srv, _ := setupTestServer()

	// Create a request to send to the handler
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.healthCheck)

	// Send the request to the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	var response map[string]string
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	if response["status"] != "ok" {
		t.Errorf("handler returned unexpected status: got %v want %v",
			response["status"], "ok")
	}
}

func TestGetUser(t *testing.T) {
	srv, mockRepo := setupTestServer()

	// Create a test user in the mock repository
	testUser := domain.User{
		Name:  "Test User",
		Email: "test@example.com",
	}
	userID, err := mockRepo.CreateUser(context.Background(), testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a request with the user ID in the URL path
	req, err := http.NewRequest("GET", "/users/"+strconv.Itoa(userID), nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new router for this test
	router := chi.NewRouter()
	router.Get("/users/{id}", srv.getUser)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Reassign URL parameters for the request
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey,
		&chi.Context{URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{strconv.Itoa(userID)}}}))

	// Send the request to the router
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	var response domain.UserResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	if response.Name != testUser.Name {
		t.Errorf("handler returned unexpected name: got %v want %v",
			response.Name, testUser.Name)
	}

	if response.Email != testUser.Email {
		t.Errorf("handler returned unexpected email: got %v want %v",
			response.Email, testUser.Email)
	}
}

func TestCreateUser(t *testing.T) {
	srv, _ := setupTestServer()

	// Create a request body
	createReq := domain.CreateUserRequest{
		Name:  "New User",
		Email: "new@example.com",
	}
	body, err := json.Marshal(createReq)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new request
	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.createUser)

	// Send the request to the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// Check the response body contains an ID
	var response map[string]int
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	if _, exists := response["id"]; !exists {
		t.Errorf("handler response does not contain id field")
	}
}
