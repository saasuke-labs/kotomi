package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// KotomiAuthUser represents a user in the kotomi authentication system
type KotomiAuthUser struct {
	ID                 string    `json:"id"`
	SiteID             string    `json:"site_id"`
	Email              string    `json:"email"`
	PasswordHash       string    `json:"-"` // Never expose in JSON
	Name               string    `json:"name"`
	AvatarURL          string    `json:"avatar_url,omitempty"`
	IsVerified         bool      `json:"is_verified"`
	VerificationToken  string    `json:"-"`
	ResetToken         string    `json:"-"`
	ResetTokenExpires  time.Time `json:"-"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
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

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a password with a hash
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateUser creates a new user in the kotomi auth system
func (s *KotomiAuthStore) CreateUser(siteID, email, password, name string) (*KotomiAuthUser, error) {
	// Hash password
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Generate verification token
	verificationToken, err := GenerateRandomToken()
	if err != nil {
		return nil, err
	}

	user := &KotomiAuthUser{
		ID:                uuid.NewString(),
		SiteID:            siteID,
		Email:             email,
		PasswordHash:      passwordHash,
		Name:              name,
		IsVerified:        false, // Email verification required by default
		VerificationToken: verificationToken,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	query := `
		INSERT INTO kotomi_auth_users (id, site_id, email, password_hash, name, is_verified, verification_token, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query, user.ID, user.SiteID, user.Email, user.PasswordHash, user.Name,
		user.IsVerified, user.VerificationToken, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by site ID and email
func (s *KotomiAuthStore) GetUserByEmail(siteID, email string) (*KotomiAuthUser, error) {
	query := `
		SELECT id, site_id, email, password_hash, name, avatar_url, is_verified, 
		       verification_token, reset_token, reset_token_expires, created_at, updated_at
		FROM kotomi_auth_users
		WHERE site_id = ? AND email = ?
	`

	var user KotomiAuthUser
	var avatarURL, verificationToken, resetToken sql.NullString
	var resetTokenExpires sql.NullTime

	err := s.db.QueryRow(query, siteID, email).Scan(
		&user.ID, &user.SiteID, &user.Email, &user.PasswordHash, &user.Name, &avatarURL,
		&user.IsVerified, &verificationToken, &resetToken, &resetTokenExpires,
		&user.CreatedAt, &user.UpdatedAt,
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
	if verificationToken.Valid {
		user.VerificationToken = verificationToken.String
	}
	if resetToken.Valid {
		user.ResetToken = resetToken.String
	}
	if resetTokenExpires.Valid {
		user.ResetTokenExpires = resetTokenExpires.Time
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *KotomiAuthStore) GetUserByID(userID string) (*KotomiAuthUser, error) {
	query := `
		SELECT id, site_id, email, password_hash, name, avatar_url, is_verified,
		       verification_token, reset_token, reset_token_expires, created_at, updated_at
		FROM kotomi_auth_users
		WHERE id = ?
	`

	var user KotomiAuthUser
	var avatarURL, verificationToken, resetToken sql.NullString
	var resetTokenExpires sql.NullTime

	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.SiteID, &user.Email, &user.PasswordHash, &user.Name, &avatarURL,
		&user.IsVerified, &verificationToken, &resetToken, &resetTokenExpires,
		&user.CreatedAt, &user.UpdatedAt,
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
	if verificationToken.Valid {
		user.VerificationToken = verificationToken.String
	}
	if resetToken.Valid {
		user.ResetToken = resetToken.String
	}
	if resetTokenExpires.Valid {
		user.ResetTokenExpires = resetTokenExpires.Time
	}

	return &user, nil
}

// AuthenticateUser validates email and password
func (s *KotomiAuthStore) AuthenticateUser(siteID, email, password string) (*KotomiAuthUser, error) {
	user, err := s.GetUserByEmail(siteID, email)
	if err != nil {
		return nil, err
	}

	// Check password
	if err := CheckPassword(password, user.PasswordHash); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
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

// SetPasswordResetToken sets a password reset token for a user
func (s *KotomiAuthStore) SetPasswordResetToken(userID, token string, expiresAt time.Time) error {
	query := `
		UPDATE kotomi_auth_users
		SET reset_token = ?, reset_token_expires = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, token, expiresAt, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to set reset token: %w", err)
	}

	return nil
}

// ResetPassword resets a user's password using a reset token
func (s *KotomiAuthStore) ResetPassword(token, newPassword string) error {
	// Find user by reset token
	query := `
		SELECT id FROM kotomi_auth_users
		WHERE reset_token = ? AND reset_token_expires > ?
	`

	var userID string
	err := s.db.QueryRow(query, token, time.Now()).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invalid or expired reset token")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Hash new password
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password and clear reset token
	updateQuery := `
		UPDATE kotomi_auth_users
		SET password_hash = ?, reset_token = NULL, reset_token_expires = NULL, updated_at = ?
		WHERE id = ?
	`

	_, err = s.db.Exec(updateQuery, passwordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
