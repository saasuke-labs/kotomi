package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Page represents a page in a site
type Page struct {
	ID        string    `json:"id"`
	SiteID    string    `json:"site_id"`
	Path      string    `json:"path"`
	Title     string    `json:"title,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PageStore handles page database operations
type PageStore struct {
	db *sql.DB
}

// NewPageStore creates a new page store
func NewPageStore(db *sql.DB) *PageStore {
	return &PageStore{db: db}
}

// GetByID retrieves a page by its ID
func (s *PageStore) GetByID(id string) (*Page, error) {
	query := `
		SELECT id, site_id, path, title, created_at, updated_at
		FROM pages
		WHERE id = ?
	`

	var page Page
	var title sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&page.ID, &page.SiteID, &page.Path, &title, &page.CreatedAt, &page.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("page not found")
		}
		return nil, fmt.Errorf("failed to query page: %w", err)
	}

	if title.Valid {
		page.Title = title.String
	}

	return &page, nil
}

// GetBySite retrieves all pages for a site
func (s *PageStore) GetBySite(siteID string) ([]Page, error) {
	query := `
		SELECT id, site_id, path, title, created_at, updated_at
		FROM pages
		WHERE site_id = ?
		ORDER BY path ASC
	`

	rows, err := s.db.Query(query, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pages: %w", err)
	}
	defer rows.Close()

	var pages []Page
	for rows.Next() {
		var page Page
		var title sql.NullString

		err := rows.Scan(
			&page.ID, &page.SiteID, &page.Path, &title, &page.CreatedAt, &page.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan page: %w", err)
		}

		if title.Valid {
			page.Title = title.String
		}

		pages = append(pages, page)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pages: %w", err)
	}

	if pages == nil {
		pages = []Page{}
	}

	return pages, nil
}

// GetBySitePath retrieves a page by site ID and path
func (s *PageStore) GetBySitePath(siteID, path string) (*Page, error) {
	query := `
		SELECT id, site_id, path, title, created_at, updated_at
		FROM pages
		WHERE site_id = ? AND path = ?
	`

	var page Page
	var title sql.NullString

	err := s.db.QueryRow(query, siteID, path).Scan(
		&page.ID, &page.SiteID, &page.Path, &title, &page.CreatedAt, &page.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query page: %w", err)
	}

	if title.Valid {
		page.Title = title.String
	}

	return &page, nil
}

// Create creates a new page
func (s *PageStore) Create(siteID, path, title string) (*Page, error) {
	now := time.Now()
	page := &Page{
		ID:        uuid.NewString(),
		SiteID:    siteID,
		Path:      path,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `
		INSERT INTO pages (id, site_id, path, title, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	var titleVal sql.NullString
	if title != "" {
		titleVal.String = title
		titleVal.Valid = true
	}

	_, err := s.db.Exec(query, page.ID, page.SiteID, page.Path, titleVal, page.CreatedAt, page.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	return page, nil
}

// Update updates a page's information
func (s *PageStore) Update(id, path, title string) error {
	query := `
		UPDATE pages
		SET path = ?, title = ?, updated_at = ?
		WHERE id = ?
	`

	var titleVal sql.NullString
	if title != "" {
		titleVal.String = title
		titleVal.Valid = true
	}

	_, err := s.db.Exec(query, path, titleVal, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update page: %w", err)
	}

	return nil
}

// Delete deletes a page
func (s *PageStore) Delete(id string) error {
	query := `DELETE FROM pages WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete page: %w", err)
	}

	return nil
}
