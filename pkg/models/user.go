package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// User represents a JWT-authenticated commenter/reactor user (Phase 2)
type User struct {
	ID         string    `json:"id"`          // User ID from JWT
	SiteID     string    `json:"site_id"`     // Site this user belongs to
	Name       string    `json:"name"`        // Display name
	Email      string    `json:"email,omitempty"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	ProfileURL string    `json:"profile_url,omitempty"`
	IsVerified bool      `json:"is_verified"`
	Roles      []string  `json:"roles,omitempty"` // JSON array of roles
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UserStore handles JWT user database operations
type UserStore struct {
	db *sql.DB
}

// NewUserStore creates a new user store
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// GetBySiteAndID retrieves a user by site and user ID
func (s *UserStore) GetBySiteAndID(siteID, userID string) (*User, error) {
	query := `
		SELECT id, site_id, name, email, avatar_url, profile_url, is_verified, roles, 
		       first_seen, last_seen, created_at, updated_at
		FROM users
		WHERE site_id = ? AND id = ?
	`

	var u User
	var email, avatarURL, profileURL sql.NullString
	var rolesJSON sql.NullString

	err := s.db.QueryRow(query, siteID, userID).Scan(
		&u.ID, &u.SiteID, &u.Name, &email, &avatarURL, &profileURL,
		&u.IsVerified, &rolesJSON, &u.FirstSeen, &u.LastSeen, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if email.Valid {
		u.Email = email.String
	}
	if avatarURL.Valid {
		u.AvatarURL = avatarURL.String
	}
	if profileURL.Valid {
		u.ProfileURL = profileURL.String
	}
	if rolesJSON.Valid && rolesJSON.String != "" {
		if err := json.Unmarshal([]byte(rolesJSON.String), &u.Roles); err != nil {
			return nil, fmt.Errorf("failed to unmarshal roles: %w", err)
		}
	}

	return &u, nil
}

// ListBySite retrieves all users for a specific site
func (s *UserStore) ListBySite(siteID string) ([]*User, error) {
	query := `
		SELECT id, site_id, name, email, avatar_url, profile_url, is_verified, roles,
		       first_seen, last_seen, created_at, updated_at
		FROM users
		WHERE site_id = ?
		ORDER BY last_seen DESC
	`

	rows, err := s.db.Query(query, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		var email, avatarURL, profileURL sql.NullString
		var rolesJSON sql.NullString

		err := rows.Scan(
			&u.ID, &u.SiteID, &u.Name, &email, &avatarURL, &profileURL,
			&u.IsVerified, &rolesJSON, &u.FirstSeen, &u.LastSeen, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if email.Valid {
			u.Email = email.String
		}
		if avatarURL.Valid {
			u.AvatarURL = avatarURL.String
		}
		if profileURL.Valid {
			u.ProfileURL = profileURL.String
		}
		if rolesJSON.Valid && rolesJSON.String != "" {
			if err := json.Unmarshal([]byte(rolesJSON.String), &u.Roles); err != nil {
				return nil, fmt.Errorf("failed to unmarshal roles: %w", err)
			}
		}

		users = append(users, &u)
	}

	return users, nil
}

// CreateOrUpdate creates a new user or updates if exists
func (s *UserStore) CreateOrUpdate(user *User) error {
	now := time.Now()
	
	// Serialize roles to JSON
	var rolesJSON sql.NullString
	if len(user.Roles) > 0 {
		rolesBytes, err := json.Marshal(user.Roles)
		if err != nil {
			return fmt.Errorf("failed to marshal roles: %w", err)
		}
		rolesJSON.String = string(rolesBytes)
		rolesJSON.Valid = true
	}

	// Check if user exists
	existing, err := s.GetBySiteAndID(user.SiteID, user.ID)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing user
		query := `
			UPDATE users
			SET name = ?, email = ?, avatar_url = ?, profile_url = ?, 
			    is_verified = ?, roles = ?, last_seen = ?, updated_at = ?
			WHERE site_id = ? AND id = ?
		`

		var email, avatarURL, profileURL sql.NullString
		if user.Email != "" {
			email.String = user.Email
			email.Valid = true
		}
		if user.AvatarURL != "" {
			avatarURL.String = user.AvatarURL
			avatarURL.Valid = true
		}
		if user.ProfileURL != "" {
			profileURL.String = user.ProfileURL
			profileURL.Valid = true
		}

		_, err = s.db.Exec(query, user.Name, email, avatarURL, profileURL,
			user.IsVerified, rolesJSON, user.LastSeen, now, user.SiteID, user.ID)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
	} else {
		// Create new user
		if user.FirstSeen.IsZero() {
			user.FirstSeen = now
		}
		if user.LastSeen.IsZero() {
			user.LastSeen = now
		}
		if user.CreatedAt.IsZero() {
			user.CreatedAt = now
		}
		user.UpdatedAt = now

		query := `
			INSERT INTO users (id, site_id, name, email, avatar_url, profile_url, 
			                   is_verified, roles, first_seen, last_seen, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		var email, avatarURL, profileURL sql.NullString
		if user.Email != "" {
			email.String = user.Email
			email.Valid = true
		}
		if user.AvatarURL != "" {
			avatarURL.String = user.AvatarURL
			avatarURL.Valid = true
		}
		if user.ProfileURL != "" {
			profileURL.String = user.ProfileURL
			profileURL.Valid = true
		}

		_, err = s.db.Exec(query, user.ID, user.SiteID, user.Name, email, avatarURL, profileURL,
			user.IsVerified, rolesJSON, user.FirstSeen, user.LastSeen, user.CreatedAt, user.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	return nil
}

// UpdateLastSeen updates the last_seen timestamp for a user
func (s *UserStore) UpdateLastSeen(siteID, userID string) error {
	query := `
		UPDATE users
		SET last_seen = ?, updated_at = ?
		WHERE site_id = ? AND id = ?
	`

	now := time.Now()
	_, err := s.db.Exec(query, now, now, siteID, userID)
	if err != nil {
		return fmt.Errorf("failed to update last_seen: %w", err)
	}

	return nil
}

// Delete removes a user and all their comments/reactions
func (s *UserStore) Delete(siteID, userID string) error {
	// Note: Foreign key constraints will cascade delete comments and reactions
	query := `
		DELETE FROM users
		WHERE site_id = ? AND id = ?
	`

	_, err := s.db.Exec(query, siteID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
