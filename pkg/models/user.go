package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name,omitempty"`
	Auth0Sub  string    `json:"auth0_sub"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserStore handles user database operations
type UserStore struct {
	db *sql.DB
}

// NewUserStore creates a new user store
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// GetByAuth0Sub retrieves a user by their Auth0 subject identifier
func (s *UserStore) GetByAuth0Sub(auth0Sub string) (*User, error) {
	query := `
		SELECT id, email, name, auth0_sub, created_at, updated_at
		FROM users
		WHERE auth0_sub = ?
	`

	var u User
	var name sql.NullString

	err := s.db.QueryRow(query, auth0Sub).Scan(
		&u.ID, &u.Email, &name, &u.Auth0Sub, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if name.Valid {
		u.Name = name.String
	}

	return &u, nil
}

// GetByID retrieves a user by their ID
func (s *UserStore) GetByID(id string) (*User, error) {
	query := `
		SELECT id, email, name, auth0_sub, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	var u User
	var name sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&u.ID, &u.Email, &name, &u.Auth0Sub, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if name.Valid {
		u.Name = name.String
	}

	return &u, nil
}

// Create creates a new user
func (s *UserStore) Create(email, name, auth0Sub string) (*User, error) {
	now := time.Now()
	user := &User{
		ID:        uuid.NewString(),
		Email:     email,
		Name:      name,
		Auth0Sub:  auth0Sub,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `
		INSERT INTO users (id, email, name, auth0_sub, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var nameVal sql.NullString
	if name != "" {
		nameVal.String = name
		nameVal.Valid = true
	}

	_, err := s.db.Exec(query, user.ID, user.Email, nameVal, user.Auth0Sub, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Update updates a user's information
func (s *UserStore) Update(id, email, name string) error {
	query := `
		UPDATE users
		SET email = ?, name = ?, updated_at = ?
		WHERE id = ?
	`

	var nameVal sql.NullString
	if name != "" {
		nameVal.String = name
		nameVal.Valid = true
	}

	_, err := s.db.Exec(query, email, nameVal, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
