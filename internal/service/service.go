package service

import (
	"context"
	"fmt"

	"github.com/dyrober/AgencyCRM/internal/domain"
	"github.com/dyrober/AgencyCRM/internal/repository"
)

// Service provides buisness logic operations
type Service struct {
	repo repository.UserRepository
}

// New Service creates a new service instance
func NewService(repo repository.UserRepository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetUser retreives user by id
func (s *Service) GetUser(ctx context.Context, id int) (*domain.UserResponse, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Service error - get user: %w", err)
	}
	return &domain.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

// Creates a new user
func (s *Service) CreateUser(ctx context.Context, req domain.CreateUserRequest) (int, error) {
	user := domain.User{
		Name:  req.Name,
		Email: req.Email,
	}

	id, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("Service error- create user: %w", err)
	}
	return id, nil
}
