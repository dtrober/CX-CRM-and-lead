package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dyrober/AgencyCRM/internal/config"
	"github.com/dyrober/AgencyCRM/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// new postgres connection and db
func NewPostgresDB(cfg config.DBConfig) (*sql.DB, error) {
	connConfig, err := pgx.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("unable to parse DSN: %w", err)
	}

	//set con pool params
	connConfig.RuntimeParams["application_name"] = "CRMandLead"
	//convert to adapt
	db := stdlib.OpenDB(*connConfig)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}
	return db, nil
}

func NewRepository(db DBConnection) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Close() error {
	return r.db.Close()
}

// Get a user by ID
func (r *Repository) GetUser(ctx context.Context, id int) (*domain.User, error) {
	query := `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`
	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// create a user
func (r *Repository) CreateUser(ctx context.Context, user domain.User) (int, error) {
	query := `
	INSERT INTO users (name, email, created_at, updated_at)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`

	now := time.Now()
	var id int
	err := r.db.QueryRowContext(ctx, query,
		user.Name,
		user.Email,
		now,
		now).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create a user: %w", err)
	}

	return id, nil
}
