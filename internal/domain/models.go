package domain

import (
	"time"
)

// represents a user
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Lead represents a sales lead
type Lead struct {
	ID            int       `json:"id"`
	CompanyName   string    `json:"company_name"`
	PhoneNumber   string    `json:"phone_number"`
	Address       string    `json:"address"`
	Email         string    `json:"email"`
	Website       string    `json:"website"`
	WebsiteScore  int       `json:"website_score"`
	PreRenderSite bool      `json:"pre_render_site"`
	ReviewAvg     int       `json:"review_avg"`
	ReviewDate    time.Time `json:"review_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
