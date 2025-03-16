-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Add sample data (optional)
INSERT INTO users (name, email, created_at, updated_at)
VALUES 
    ('John Doe', 'john@example.com', NOW(), NOW()),
    ('Jane Smith', 'jane@example.com', NOW(), NOW())
ON CONFLICT (email) DO NOTHING;