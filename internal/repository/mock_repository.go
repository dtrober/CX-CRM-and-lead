package repository

import (
	"context"
	"time"

	"github.com/dyrober/AgencyCRM/internal/domain"
)

// MockRepository implements the UserRepository interface for testing
type MockRepository struct {
	users  map[int]*domain.User
	nextID int
}

// Ensure MockRepository implements UserRepository
var _ UserRepository = (*MockRepository)(nil)

// NewMockRepository creates a new mock repository instance
func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:  make(map[int]*domain.User),
		nextID: 1,
	}
}

// Close is a no-op for the mock
func (m *MockRepository) Close() error {
	return nil
}

// GetUser retrieves a user by ID from the in-memory map
func (m *MockRepository) GetUser(ctx context.Context, id int) (*domain.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, ErrNotFound
	}
	return user, nil
}

// CreateUser adds a new user to the in-memory map
func (m *MockRepository) CreateUser(ctx context.Context, user domain.User) (int, error) {
	// Assign an ID and timestamps
	id := m.nextID
	now := time.Now()

	m.users[id] = &domain.User{
		ID:        id,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m.nextID++
	return id, nil
}

// ErrNotFound is used to simulate database not found errors
var ErrNotFound = ErrorNotFound("record not found")

// ErrorNotFound is a custom error type for not found records
type ErrorNotFound string

func (e ErrorNotFound) Error() string {
	return string(e)
}
