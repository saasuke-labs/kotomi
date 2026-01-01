package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Site represents a site in the system
type Site struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Name        string    `json:"name"`
	Domain      string    `json:"domain,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SiteStore handles site database operations
type SiteStore struct {
	db *sql.DB
}

// NewSiteStore creates a new site store
func NewSiteStore(db *sql.DB) *SiteStore {
	return &SiteStore{db: db}
}

// GetByID retrieves a site by its ID
func (s *SiteStore) GetByID(id string) (*Site, error) {
	query := `
		SELECT id, owner_id, name, domain, description, created_at, updated_at
		FROM sites
		WHERE id = ?
	`

	var site Site
	var domain, description sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&site.ID, &site.OwnerID, &site.Name, &domain, &description, &site.CreatedAt, &site.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("site not found")
		}
		return nil, fmt.Errorf("failed to query site: %w", err)
	}

	if domain.Valid {
		site.Domain = domain.String
	}
	if description.Valid {
		site.Description = description.String
	}

	return &site, nil
}

// GetByOwner retrieves all sites owned by a user
func (s *SiteStore) GetByOwner(ownerID string) ([]Site, error) {
	query := `
		SELECT id, owner_id, name, domain, description, created_at, updated_at
		FROM sites
		WHERE owner_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sites: %w", err)
	}
	defer rows.Close()

	var sites []Site
	for rows.Next() {
		var site Site
		var domain, description sql.NullString

		err := rows.Scan(
			&site.ID, &site.OwnerID, &site.Name, &domain, &description, &site.CreatedAt, &site.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan site: %w", err)
		}

		if domain.Valid {
			site.Domain = domain.String
		}
		if description.Valid {
			site.Description = description.String
		}

		sites = append(sites, site)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sites: %w", err)
	}

	if sites == nil {
		sites = []Site{}
	}

	return sites, nil
}

// Create creates a new site
func (s *SiteStore) Create(ownerID, name, domain, description string) (*Site, error) {
	now := time.Now()
	site := &Site{
		ID:          uuid.NewString(),
		OwnerID:     ownerID,
		Name:        name,
		Domain:      domain,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := `
		INSERT INTO sites (id, owner_id, name, domain, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	var domainVal, descVal sql.NullString
	if domain != "" {
		domainVal.String = domain
		domainVal.Valid = true
	}
	if description != "" {
		descVal.String = description
		descVal.Valid = true
	}

	_, err := s.db.Exec(query, site.ID, site.OwnerID, site.Name, domainVal, descVal, site.CreatedAt, site.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create site: %w", err)
	}

	return site, nil
}

// Update updates a site's information
func (s *SiteStore) Update(id, name, domain, description string) error {
	query := `
		UPDATE sites
		SET name = ?, domain = ?, description = ?, updated_at = ?
		WHERE id = ?
	`

	var domainVal, descVal sql.NullString
	if domain != "" {
		domainVal.String = domain
		domainVal.Valid = true
	}
	if description != "" {
		descVal.String = description
		descVal.Valid = true
	}

	_, err := s.db.Exec(query, name, domainVal, descVal, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update site: %w", err)
	}

	return nil
}

// Delete deletes a site
func (s *SiteStore) Delete(id string) error {
	query := `DELETE FROM sites WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete site: %w", err)
	}

	return nil
}
