package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/dyrober/AgencyCRM/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUsers(ctx context.Context) ([]*domain.User, error) {
	args := m.Called(ctx)

	// Handle the first return value, which should be []*domain.User
	users, ok := args.Get(0).([]*domain.User)
	if !ok && args.Get(0) != nil {
		// This will happen if the wrong type was returned from the mock setup
		return nil, fmt.Errorf("unexpected type for users return value")
	}

	// Handle the second return value, which should be an error
	var err error
	if errArg := args.Error(1); errArg != nil {
		err = errArg
	}

	return users, err
}

// Mock implementation of GetUser
func (m *MockUserRepository) GetUser(ctx context.Context, id int) (*domain.User, error) {
	args := m.Called(ctx, id)

	// If we're expected to return a user, return the value provided to the mock
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}

	// Otherwise return nil and the error
	return nil, args.Error(1)
}

// Mock implementation of CreateUser
func (m *MockUserRepository) CreateUser(ctx context.Context, user domain.User) (int, error) {
	args := m.Called(ctx, user)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestGetUser(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		userID      int
		mockUser    *domain.User
		mockErr     error
		expected    *domain.UserResponse
		expectedErr bool
	}{
		{
			name:   "success",
			userID: 1,
			mockUser: &domain.User{
				ID:        1,
				Name:      "John Doe",
				Email:     "john@example.com",
				CreatedAt: time.Now(),
			},
			mockErr:     nil,
			expected:    &domain.UserResponse{ID: 1, Name: "John Doe", Email: "john@example.com"},
			expectedErr: false,
		},
		{
			name:        "user not found",
			userID:      2,
			mockUser:    nil,
			mockErr:     errors.New("user not found"),
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockUserRepository)

			// Set expectations on mock
			mockRepo.On("GetUser", mock.Anything, tc.userID).Return(tc.mockUser, tc.mockErr)

			// Create service with mock repo
			service := NewService(mockRepo)

			// Call the method being tested
			user, err := service.GetUser(context.Background(), tc.userID)

			// Assert results
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.ID, user.ID)
				assert.Equal(t, tc.expected.Name, user.Name)
				assert.Equal(t, tc.expected.Email, user.Email)
				// Note: CreatedAt is dynamic so we check if it exists but don't compare exact values
				assert.Equal(t, tc.mockUser.CreatedAt, user.CreatedAt)
			}

			// Verify all mock expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateUser(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		request     domain.CreateUserRequest
		mockID      int
		mockErr     error
		expectedID  int
		expectedErr bool
	}{
		{
			name:        "success",
			request:     domain.CreateUserRequest{Name: "John Doe", Email: "john@example.com"},
			mockID:      1,
			mockErr:     nil,
			expectedID:  1,
			expectedErr: false,
		},
		{
			name:        "repository error",
			request:     domain.CreateUserRequest{Name: "Jane Doe", Email: "jane@example.com"},
			mockID:      0,
			mockErr:     errors.New("database error"),
			expectedID:  0,
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockUserRepository)

			// Create the expected user that would be passed to the repository
			expectedUser := domain.User{
				Name:  tc.request.Name,
				Email: tc.request.Email,
			}

			// Set expectations on mock
			mockRepo.On("CreateUser", mock.Anything, expectedUser).Return(tc.mockID, tc.mockErr)

			// Create service with mock repo
			service := NewService(mockRepo)

			// Call the method being tested
			id, err := service.CreateUser(context.Background(), tc.request)

			// Assert results
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedID, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedID, id)
			}

			// Verify all mock expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}
