package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dyrober/AgencyCRM/internal/domain"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id int) (*domain.User, error)
	GetUsers(ctx context.Context) ([]*domain.User, error)
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user domain.User) (int, error)

	// Close closes any resources used by the repository
	Close() error
}

// Repository is the concrete implementation of UserRepository using PostgreSQL
type Repository struct {
	db *sql.DB
}

// Ensure Repository implements UserRepository
var _ UserRepository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// DBConnection is an interface representing the database connection
// This allows us to easily mock the database in tests
type DBConnection interface {
	QueryRowContext(ctx context.Context, query string, args ...any) RowScanner
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
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

func (r *Repository) GetUsers(ctx context.Context) ([]*domain.User, error) {
	query := `SELECT id, name, email, created_at, updated_at FROM users ORDER BY id DESC LIMIT 100`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()
	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over user rows: %w", err)
	}
	return users, nil
}
