package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// KotomiAuthUser represents a user authenticated through Kotomi's Auth0 integration
type KotomiAuthUser struct {
	ID         string    `json:"id"`
	SiteID     string    `json:"site_id"`
	Email      string    `json:"email"`
	Auth0Sub   string    `json:"auth0_sub"` // Auth0 subject identifier
	Name       string    `json:"name"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// KotomiAuthSession represents a user session with JWT tokens
type KotomiAuthSession struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	SiteID            string    `json:"site_id"`
	Token             string    `json:"token"`
	RefreshToken      string    `json:"refresh_token"`
	ExpiresAt         time.Time `json:"expires_at"`
	RefreshExpiresAt  time.Time `json:"refresh_expires_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// KotomiAuthStore handles database operations for kotomi authentication
type KotomiAuthStore struct {
	db *sql.DB
}

// NewKotomiAuthStore creates a new kotomi auth store
func NewKotomiAuthStore(db *sql.DB) *KotomiAuthStore {
	return &KotomiAuthStore{db: db}
}

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateOrUpdateUserFromAuth0 creates or updates a user from Auth0 user info
func (s *KotomiAuthStore) CreateOrUpdateUserFromAuth0(siteID string, userInfo *UserInfo) (*KotomiAuthUser, error) {
	// Check if user exists by auth0_sub
	existingUser, err := s.GetUserByAuth0Sub(siteID, userInfo.Sub)
	
	now := time.Now()
	
	if err != nil && err.Error() != "user not found" {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	
	if existingUser != nil {
		// Update existing user
		existingUser.Email = userInfo.Email
		existingUser.Name = userInfo.Name
		existingUser.AvatarURL = userInfo.Picture
		existingUser.IsVerified = userInfo.EmailVerified
		existingUser.UpdatedAt = now
		
		query := `
			UPDATE kotomi_auth_users
			SET email = ?, name = ?, avatar_url = ?, is_verified = ?, updated_at = ?
			WHERE id = ?
		`
		
		var avatarURL sql.NullString
		if existingUser.AvatarURL != "" {
			avatarURL.String = existingUser.AvatarURL
			avatarURL.Valid = true
		}
		
		_, err = s.db.Exec(query, existingUser.Email, existingUser.Name, avatarURL, 
			existingUser.IsVerified, existingUser.UpdatedAt, existingUser.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		
		return existingUser, nil
	}
	
	// Create new user
	user := &KotomiAuthUser{
		ID:         uuid.NewString(),
		SiteID:     siteID,
		Email:      userInfo.Email,
		Auth0Sub:   userInfo.Sub,
		Name:       userInfo.Name,
		AvatarURL:  userInfo.Picture,
		IsVerified: userInfo.EmailVerified,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	
	query := `
		INSERT INTO kotomi_auth_users (id, site_id, email, auth0_sub, name, avatar_url, is_verified, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	var avatarURL sql.NullString
	if user.AvatarURL != "" {
		avatarURL.String = user.AvatarURL
		avatarURL.Valid = true
	}
	
	_, err = s.db.Exec(query, user.ID, user.SiteID, user.Email, user.Auth0Sub, user.Name,
		avatarURL, user.IsVerified, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return user, nil
}

// GetUserByAuth0Sub retrieves a user by site ID and Auth0 subject
func (s *KotomiAuthStore) GetUserByAuth0Sub(siteID, auth0Sub string) (*KotomiAuthUser, error) {
	query := `
		SELECT id, site_id, email, auth0_sub, name, avatar_url, is_verified, created_at, updated_at
		FROM kotomi_auth_users
		WHERE site_id = ? AND auth0_sub = ?
	`

	var user KotomiAuthUser
	var avatarURL sql.NullString

	err := s.db.QueryRow(query, siteID, auth0Sub).Scan(
		&user.ID, &user.SiteID, &user.Email, &user.Auth0Sub, &user.Name, &avatarURL,
		&user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *KotomiAuthStore) GetUserByID(userID string) (*KotomiAuthUser, error) {
	query := `
		SELECT id, site_id, email, auth0_sub, name, avatar_url, is_verified, created_at, updated_at
		FROM kotomi_auth_users
		WHERE id = ?
	`

	var user KotomiAuthUser
	var avatarURL sql.NullString

	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.SiteID, &user.Email, &user.Auth0Sub, &user.Name, &avatarURL,
		&user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return &user, nil
}

// GenerateJWTToken generates a JWT token for a user
func GenerateJWTToken(user *KotomiAuthUser, siteID string, secret string, expirationMinutes int) (string, error) {
	// Create claims
	claims := jwt.MapClaims{
		"iss": "kotomi",
		"sub": user.ID,
		"aud": "kotomi",
		"exp": time.Now().Add(time.Duration(expirationMinutes) * time.Minute).Unix(),
		"iat": time.Now().Unix(),
		"kotomi_user": map[string]interface{}{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"verified": user.IsVerified,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// CreateSession creates a new session with JWT tokens
func (s *KotomiAuthStore) CreateSession(user *KotomiAuthUser, jwtSecret string) (*KotomiAuthSession, error) {
	// Generate access token (expires in 1 hour)
	accessToken, err := GenerateJWTToken(user, user.SiteID, jwtSecret, 60)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (random string, expires in 30 days)
	refreshToken, err := GenerateRandomToken()
	if err != nil {
		return nil, err
	}

	session := &KotomiAuthSession{
		ID:                uuid.NewString(),
		UserID:            user.ID,
		SiteID:            user.SiteID,
		Token:             accessToken,
		RefreshToken:      refreshToken,
		ExpiresAt:         time.Now().Add(1 * time.Hour),
		RefreshExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:         time.Now(),
	}

	query := `
		INSERT INTO kotomi_auth_sessions (id, user_id, site_id, token, refresh_token, expires_at, refresh_expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query, session.ID, session.UserID, session.SiteID, session.Token,
		session.RefreshToken, session.ExpiresAt, session.RefreshExpiresAt, session.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetSessionByToken retrieves a session by access token
func (s *KotomiAuthStore) GetSessionByToken(token string) (*KotomiAuthSession, error) {
	query := `
		SELECT id, user_id, site_id, token, refresh_token, expires_at, refresh_expires_at, created_at
		FROM kotomi_auth_sessions
		WHERE token = ?
	`

	var session KotomiAuthSession
	err := s.db.QueryRow(query, token).Scan(
		&session.ID, &session.UserID, &session.SiteID, &session.Token,
		&session.RefreshToken, &session.ExpiresAt, &session.RefreshExpiresAt, &session.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	return &session, nil
}

// GetSessionByRefreshToken retrieves a session by refresh token
func (s *KotomiAuthStore) GetSessionByRefreshToken(refreshToken string) (*KotomiAuthSession, error) {
	query := `
		SELECT id, user_id, site_id, token, refresh_token, expires_at, refresh_expires_at, created_at
		FROM kotomi_auth_sessions
		WHERE refresh_token = ?
	`

	var session KotomiAuthSession
	err := s.db.QueryRow(query, refreshToken).Scan(
		&session.ID, &session.UserID, &session.SiteID, &session.Token,
		&session.RefreshToken, &session.ExpiresAt, &session.RefreshExpiresAt, &session.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	return &session, nil
}

// DeleteSession deletes a session (logout)
func (s *KotomiAuthStore) DeleteSession(sessionID string) error {
	query := `DELETE FROM kotomi_auth_sessions WHERE id = ?`
	_, err := s.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteSessionByToken deletes a session by token
func (s *KotomiAuthStore) DeleteSessionByToken(token string) error {
	query := `DELETE FROM kotomi_auth_sessions WHERE token = ?`
	_, err := s.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// UpdateUser updates user information
func (s *KotomiAuthStore) UpdateUser(user *KotomiAuthUser) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE kotomi_auth_users
		SET name = ?, avatar_url = ?, is_verified = ?, updated_at = ?
		WHERE id = ?
	`

	var avatarURL sql.NullString
	if user.AvatarURL != "" {
		avatarURL.String = user.AvatarURL
		avatarURL.Valid = true
	}

	result, err := s.db.Exec(query, user.Name, avatarURL, user.IsVerified, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
