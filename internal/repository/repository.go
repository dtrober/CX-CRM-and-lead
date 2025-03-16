package repository

import (
	"context"

	"github.com/dyrober/AgencyCRM/internal/domain"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id int) (*domain.User, error)

	// CreateUser creates a new user
	CreateUser(ctx context.Context, user domain.User) (int, error)

	// Close closes any resources used by the repository
	Close() error
}

// Repository is the concrete implementation of UserRepository using PostgreSQL
type Repository struct {
	db DBConnection
}

// Ensure Repository implements UserRepository
var _ UserRepository = (*Repository)(nil)

// DBConnection is an interface representing the database connection
// This allows us to easily mock the database in tests
type DBConnection interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) RowScanner
	QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error)
	Close() error
}

// RowScanner is the interface that wraps the Scan method
type RowScanner interface {
	Scan(dest ...interface{}) error
}

// Rows is an interface for database rows
type Rows interface {
	Close() error
	Next() bool
	Err() error
	Scan(dest ...interface{}) error
}

// Result is an interface for database result
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}
