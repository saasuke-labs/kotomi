package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SiteAuthConfig represents authentication configuration for a site
type SiteAuthConfig struct {
	ID                    string    `json:"id"`
	SiteID                string    `json:"site_id"`
	AuthMode              string    `json:"auth_mode"` // "external" for Phase 1 (kotomi auth in future phases)
	JWTValidationType     string    `json:"jwt_validation_type,omitempty"` // "hmac", "rsa", "ecdsa", "jwks"
	JWTSecret             string    `json:"jwt_secret,omitempty"`          // For HMAC (symmetric) - stored encrypted
	JWTPublicKey          string    `json:"jwt_public_key,omitempty"`      // For RSA/ECDSA (asymmetric)
	JWKSEndpoint          string    `json:"jwks_endpoint,omitempty"`       // For JWKS (JSON Web Key Set) URL
	JWTIssuer             string    `json:"jwt_issuer,omitempty"`          // Expected issuer claim
	JWTAudience           string    `json:"jwt_audience,omitempty"`        // Expected audience claim
	TokenExpirationBuffer int       `json:"token_expiration_buffer"`       // Grace period in seconds
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// SiteAuthConfigStore handles site_auth_configs database operations
type SiteAuthConfigStore struct {
	db *sql.DB
}

// NewSiteAuthConfigStore creates a new site auth config store
func NewSiteAuthConfigStore(db *sql.DB) *SiteAuthConfigStore {
	return &SiteAuthConfigStore{db: db}
}

// GetBySiteID retrieves auth config for a specific site
func (s *SiteAuthConfigStore) GetBySiteID(ctx context.Context, siteID string) (*SiteAuthConfig, error) {
	query := `
		SELECT id, site_id, auth_mode, jwt_validation_type, jwt_secret, jwt_public_key,
		       jwks_endpoint, jwt_issuer, jwt_audience, token_expiration_buffer,
		       created_at, updated_at
		FROM site_auth_configs
		WHERE site_id = ?
	`

	var config SiteAuthConfig
	err := s.db.QueryRowContext(ctx, query, siteID).Scan(
		&config.ID, &config.SiteID, &config.AuthMode, &config.JWTValidationType,
		&config.JWTSecret, &config.JWTPublicKey, &config.JWKSEndpoint,
		&config.JWTIssuer, &config.JWTAudience, &config.TokenExpirationBuffer,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("site auth config not found")
		}
		return nil, fmt.Errorf("failed to query site auth config: %w", err)
	}

	return &config, nil
}

// Create creates a new site auth configuration
func (s *SiteAuthConfigStore) Create(ctx context.Context, config *SiteAuthConfig) error {
	if config.ID == "" {
		config.ID = uuid.NewString()
	}
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}
	if config.UpdatedAt.IsZero() {
		config.UpdatedAt = time.Now()
	}
	if config.TokenExpirationBuffer == 0 {
		config.TokenExpirationBuffer = 60 // Default 60 seconds
	}

	query := `
		INSERT INTO site_auth_configs (
			id, site_id, auth_mode, jwt_validation_type, jwt_secret, jwt_public_key,
			jwks_endpoint, jwt_issuer, jwt_audience, token_expiration_buffer,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		config.ID, config.SiteID, config.AuthMode, config.JWTValidationType,
		config.JWTSecret, config.JWTPublicKey, config.JWKSEndpoint,
		config.JWTIssuer, config.JWTAudience, config.TokenExpirationBuffer,
		config.CreatedAt, config.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create site auth config: %w", err)
	}

	return nil
}

// Update updates an existing site auth configuration
func (s *SiteAuthConfigStore) Update(ctx context.Context, config *SiteAuthConfig) error {
	config.UpdatedAt = time.Now()

	query := `
		UPDATE site_auth_configs
		SET auth_mode = ?, jwt_validation_type = ?, jwt_secret = ?, jwt_public_key = ?,
		    jwks_endpoint = ?, jwt_issuer = ?, jwt_audience = ?, token_expiration_buffer = ?,
		    updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		config.AuthMode, config.JWTValidationType, config.JWTSecret, config.JWTPublicKey,
		config.JWKSEndpoint, config.JWTIssuer, config.JWTAudience, config.TokenExpirationBuffer,
		config.UpdatedAt, config.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update site auth config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("site auth config not found")
	}

	return nil
}

// Delete deletes a site auth configuration
func (s *SiteAuthConfigStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM site_auth_configs WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete site auth config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("site auth config not found")
	}

	return nil
}

// KotomiUser represents user claims from JWT token
type KotomiUser struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Email      string   `json:"email,omitempty"`
	AvatarURL  string   `json:"avatar_url,omitempty"`
	ProfileURL string   `json:"profile_url,omitempty"`
	Verified   bool     `json:"verified"`
	Roles      []string `json:"roles,omitempty"`
}

// StandardClaims represents the standard JWT claims we expect
type StandardClaims struct {
	Issuer    string       `json:"iss"`
	Subject   string       `json:"sub"`
	Audience  string       `json:"aud"`
	ExpiresAt int64        `json:"exp"`
	IssuedAt  int64        `json:"iat"`
	KotomiUser *KotomiUser `json:"kotomi_user"`
}

// ToJSON converts KotomiUser to JSON string
func (u *KotomiUser) ToJSON() (string, error) {
	data, err := json.Marshal(u)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user to JSON: %w", err)
	}
	return string(data), nil
}

// FromJSON parses KotomiUser from JSON string
func (u *KotomiUser) FromJSON(data string) error {
	err := json.Unmarshal([]byte(data), u)
	if err != nil {
		return fmt.Errorf("failed to unmarshal user from JSON: %w", err)
	}
	return nil
}
