package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AdminUser represents an admin user in the system (for Auth0 admin panel)
type AdminUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name,omitempty"`
	Auth0Sub  string    `json:"auth0_sub"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AdminUserStore handles admin user database operations
type AdminUserStore struct {
	db *sql.DB
}

// NewAdminUserStore creates a new admin user store
func NewAdminUserStore(db *sql.DB) *AdminUserStore {
	return &AdminUserStore{db: db}
}

// GetByAuth0Sub retrieves an admin user by their Auth0 subject identifier
func (s *AdminUserStore) GetByAuth0Sub(auth0Sub string) (*AdminUser, error) {
	query := `
		SELECT id, email, name, auth0_sub, created_at, updated_at
		FROM admin_users
		WHERE auth0_sub = ?
	`

	var u AdminUser
	var name sql.NullString

	err := s.db.QueryRow(query, auth0Sub).Scan(
		&u.ID, &u.Email, &name, &u.Auth0Sub, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query admin user: %w", err)
	}

	if name.Valid {
		u.Name = name.String
	}

	return &u, nil
}

// GetByID retrieves an admin user by their ID
func (s *AdminUserStore) GetByID(id string) (*AdminUser, error) {
	query := `
		SELECT id, email, name, auth0_sub, created_at, updated_at
		FROM admin_users
		WHERE id = ?
	`

	var u AdminUser
	var name sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&u.ID, &u.Email, &name, &u.Auth0Sub, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin user not found")
		}
		return nil, fmt.Errorf("failed to query admin user: %w", err)
	}

	if name.Valid {
		u.Name = name.String
	}

	return &u, nil
}

// Create creates a new admin user
func (s *AdminUserStore) Create(email, name, auth0Sub string) (*AdminUser, error) {
	now := time.Now()
	user := &AdminUser{
		ID:        uuid.NewString(),
		Email:     email,
		Name:      name,
		Auth0Sub:  auth0Sub,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `
		INSERT INTO admin_users (id, email, name, auth0_sub, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var nameVal sql.NullString
	if name != "" {
		nameVal.String = name
		nameVal.Valid = true
	}

	_, err := s.db.Exec(query, user.ID, user.Email, nameVal, user.Auth0Sub, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	return user, nil
}

// Update updates an admin user's information
func (s *AdminUserStore) Update(id, email, name string) error {
	query := `
		UPDATE admin_users
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
		return fmt.Errorf("failed to update admin user: %w", err)
	}

	return nil
}
