package comments

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore provides SQLite-based persistent storage for comments
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite-based comment store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create table and indexes if they don't exist
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		name TEXT,
		auth0_sub TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sites (
		id TEXT PRIMARY KEY,
		owner_id TEXT NOT NULL,
		name TEXT NOT NULL,
		domain TEXT,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_sites_owner ON sites(owner_id);

	CREATE TABLE IF NOT EXISTS pages (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		path TEXT NOT NULL,
		title TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
		UNIQUE(site_id, path)
	);

	CREATE INDEX IF NOT EXISTS idx_pages_site ON pages(site_id);

	CREATE TABLE IF NOT EXISTS comments (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		page_id TEXT NOT NULL,
		author TEXT NOT NULL,
		text TEXT NOT NULL,
		parent_id TEXT,
		status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'approved', 'rejected')),
		moderated_by TEXT,
		moderated_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_site_page ON comments(site_id, page_id);
	CREATE INDEX IF NOT EXISTS idx_parent ON comments(parent_id);
	CREATE INDEX IF NOT EXISTS idx_comments_status ON comments(status);

	CREATE TABLE IF NOT EXISTS allowed_reactions (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		name TEXT NOT NULL,
		emoji TEXT NOT NULL,
		reaction_type TEXT NOT NULL DEFAULT 'comment' CHECK(reaction_type IN ('page', 'comment', 'both')),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
		UNIQUE(site_id, name, reaction_type)
	);

	CREATE INDEX IF NOT EXISTS idx_allowed_reactions_site ON allowed_reactions(site_id);
	CREATE INDEX IF NOT EXISTS idx_allowed_reactions_type ON allowed_reactions(reaction_type);

	CREATE TABLE IF NOT EXISTS reactions (
		id TEXT PRIMARY KEY,
		page_id TEXT,
		comment_id TEXT,
		allowed_reaction_id TEXT NOT NULL,
		user_identifier TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE,
		FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
		FOREIGN KEY (allowed_reaction_id) REFERENCES allowed_reactions(id) ON DELETE CASCADE,
		CHECK ((page_id IS NOT NULL AND comment_id IS NULL) OR (page_id IS NULL AND comment_id IS NOT NULL)),
		UNIQUE(page_id, comment_id, allowed_reaction_id, user_identifier)
	);

	CREATE INDEX IF NOT EXISTS idx_reactions_page ON reactions(page_id);
	CREATE INDEX IF NOT EXISTS idx_reactions_comment ON reactions(comment_id);
	CREATE INDEX IF NOT EXISTS idx_reactions_allowed ON reactions(allowed_reaction_id);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

// AddPageComment adds a comment to a specific page on a site
func (s *SQLiteStore) AddPageComment(site, page string, comment Comment) error {
	// Set timestamps if not already set
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = time.Now()
	}
	if comment.UpdatedAt.IsZero() {
		comment.UpdatedAt = time.Now()
	}
	// Set default status if not set
	if comment.Status == "" {
		comment.Status = "pending"
	}

	query := `
		INSERT INTO comments (id, site_id, page_id, author, text, parent_id, status, moderated_by, moderated_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Convert empty ParentID to NULL
	var parentID sql.NullString
	if comment.ParentID != "" {
		parentID.String = comment.ParentID
		parentID.Valid = true
	}

	// Convert empty ModeratedBy to NULL
	var moderatedBy sql.NullString
	if comment.ModeratedBy != "" {
		moderatedBy.String = comment.ModeratedBy
		moderatedBy.Valid = true
	}

	// Convert zero ModeratedAt to NULL
	var moderatedAt sql.NullTime
	if !comment.ModeratedAt.IsZero() {
		moderatedAt.Time = comment.ModeratedAt
		moderatedAt.Valid = true
	}

	_, err := s.db.Exec(query,
		comment.ID,
		site,
		page,
		comment.Author,
		comment.Text,
		parentID,
		comment.Status,
		moderatedBy,
		moderatedAt,
		comment.CreatedAt,
		comment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}

	return nil
}

// GetPageComments retrieves all comments for a specific page on a site
func (s *SQLiteStore) GetPageComments(site, page string) ([]Comment, error) {
	query := `
		SELECT id, author, text, parent_id, status, moderated_by, moderated_at, created_at, updated_at
		FROM comments
		WHERE site_id = ? AND page_id = ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, site, page)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		var parentID sql.NullString
		var moderatedBy sql.NullString
		var moderatedAt sql.NullTime

		err := rows.Scan(&c.ID, &c.Author, &c.Text, &parentID, &c.Status, &moderatedBy, &moderatedAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		if parentID.Valid {
			c.ParentID = parentID.String
		}
		if moderatedBy.Valid {
			c.ModeratedBy = moderatedBy.String
		}
		if moderatedAt.Valid {
			c.ModeratedAt = moderatedAt.Time
		}

		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	// Return empty slice if no comments found
	if comments == nil {
		comments = []Comment{}
	}

	return comments, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetDB returns the underlying database connection
func (s *SQLiteStore) GetDB() *sql.DB {
	return s.db
}

// GetCommentsBySite retrieves all comments for a specific site
func (s *SQLiteStore) GetCommentsBySite(siteID string, status string) ([]Comment, error) {
	query := `
		SELECT id, site_id, page_id, author, text, parent_id, status, moderated_by, moderated_at, created_at, updated_at
		FROM comments
		WHERE site_id = ?
	`
	args := []interface{}{siteID}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		var siteID, pageID string // Not used but needed for Scan
		var parentID sql.NullString
		var moderatedBy sql.NullString
		var moderatedAt sql.NullTime

		err := rows.Scan(&c.ID, &siteID, &pageID, &c.Author, &c.Text, &parentID, &c.Status, &moderatedBy, &moderatedAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		if parentID.Valid {
			c.ParentID = parentID.String
		}
		if moderatedBy.Valid {
			c.ModeratedBy = moderatedBy.String
		}
		if moderatedAt.Valid {
			c.ModeratedAt = moderatedAt.Time
		}

		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	if comments == nil {
		comments = []Comment{}
	}

	return comments, nil
}

// GetCommentByID retrieves a comment by its ID
func (s *SQLiteStore) GetCommentByID(commentID string) (*Comment, error) {
	query := `
		SELECT id, site_id, page_id, author, text, parent_id, status, moderated_by, moderated_at, created_at, updated_at
		FROM comments
		WHERE id = ?
	`

	var c Comment
	var siteID, pageID string // Not used but needed for Scan
	var parentID sql.NullString
	var moderatedBy sql.NullString
	var moderatedAt sql.NullTime

	err := s.db.QueryRow(query, commentID).Scan(
		&c.ID, &siteID, &pageID, &c.Author, &c.Text, &parentID, &c.Status, &moderatedBy, &moderatedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to query comment: %w", err)
	}

	if parentID.Valid {
		c.ParentID = parentID.String
	}
	if moderatedBy.Valid {
		c.ModeratedBy = moderatedBy.String
	}
	if moderatedAt.Valid {
		c.ModeratedAt = moderatedAt.Time
	}

	return &c, nil
}

// UpdateCommentStatus updates the status of a comment
func (s *SQLiteStore) UpdateCommentStatus(commentID, status, moderatorID string) error {
	query := `
		UPDATE comments
		SET status = ?, moderated_by = ?, moderated_at = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := s.db.Exec(query, status, moderatorID, now, now, commentID)
	if err != nil {
		return fmt.Errorf("failed to update comment status: %w", err)
	}

	return nil
}

// DeleteComment deletes a comment by its ID
func (s *SQLiteStore) DeleteComment(commentID string) error {
	query := `DELETE FROM comments WHERE id = ?`

	_, err := s.db.Exec(query, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

// GetCommentSiteID retrieves the site ID for a comment
func (s *SQLiteStore) GetCommentSiteID(commentID string) (string, error) {
	query := `SELECT site_id FROM comments WHERE id = ?`

	var siteID string
	err := s.db.QueryRow(query, commentID).Scan(&siteID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("comment not found")
		}
		return "", fmt.Errorf("failed to query comment site: %w", err)
	}

	return siteID, nil
}
