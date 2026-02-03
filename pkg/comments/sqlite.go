package comments

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
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

	// Configure connection pool for production use
	// These settings help manage resources and prevent connection exhaustion
	db.SetMaxOpenConns(25)                        // Limit concurrent connections (SQLite is single-writer with WAL)
	db.SetMaxIdleConns(5)                         // Keep some connections warm for better performance
	db.SetConnMaxLifetime(5 * time.Minute)        // Recycle connections periodically to prevent stale connections
	db.SetConnMaxIdleTime(1 * time.Minute)        // Close idle connections to free resources

	// Verify database connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("database not responding: %w", err)
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Enable WAL (Write-Ahead Logging) mode for better concurrency
	// WAL mode allows multiple readers and one writer to work simultaneously
	// This is critical for handling concurrent HTTP requests
	_, err = db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Set busy timeout to handle lock contention gracefully
	// Instead of failing immediately when database is locked, retry for up to 5 seconds
	_, err = db.Exec("PRAGMA busy_timeout = 5000")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Set synchronous to NORMAL for better write performance
	// This is safe with WAL mode and provides good balance of performance and durability
	_, err = db.Exec("PRAGMA synchronous = NORMAL")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
	}

	// Increase cache size for better query performance (64MB)
	// Larger cache reduces disk I/O for frequently accessed data
	_, err = db.Exec("PRAGMA cache_size = -64000")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set cache size: %w", err)
	}

	// Use memory for temporary tables and indexes for better performance
	// This reduces disk I/O for sorting and temporary data operations
	_, err = db.Exec("PRAGMA temp_store = MEMORY")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set temp store: %w", err)
	}

	// Enable memory-mapped I/O for better read performance (256MB)
	// This is optional and will be ignored on systems that don't support it
	_, err = db.Exec("PRAGMA mmap_size = 268435456")
	if err != nil {
		// Log warning but don't fail - mmap not supported on all systems
		log.Printf("Warning: Could not enable mmap: %v", err)
	}

	// Log the configuration for debugging and verification
	log.Printf("SQLite database initialized with optimizations: WAL mode, 64MB cache, 25 max connections")

	// Create table and indexes if they don't exist
	schema := `
	CREATE TABLE IF NOT EXISTS admin_users (
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
		FOREIGN KEY (owner_id) REFERENCES admin_users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_sites_owner ON sites(owner_id);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT NOT NULL,
		site_id TEXT NOT NULL,
		name TEXT NOT NULL,
		email TEXT,
		avatar_url TEXT,
		profile_url TEXT,
		is_verified INTEGER DEFAULT 0,
		roles TEXT,
		reputation_score INTEGER DEFAULT 0,
		first_seen TIMESTAMP NOT NULL,
		last_seen TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
		PRIMARY KEY (site_id, id)
	);

	CREATE INDEX IF NOT EXISTS idx_users_site_id ON users(site_id);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(site_id, email);

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
		author_id TEXT NOT NULL,
		author_email TEXT,
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
	CREATE INDEX IF NOT EXISTS idx_comments_author ON comments(author_id);

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
		user_id TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE,
		FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
		FOREIGN KEY (allowed_reaction_id) REFERENCES allowed_reactions(id) ON DELETE CASCADE,
		CHECK ((page_id IS NOT NULL AND comment_id IS NULL) OR (page_id IS NULL AND comment_id IS NOT NULL)),
		UNIQUE(page_id, comment_id, allowed_reaction_id, user_id)
	);

	CREATE INDEX IF NOT EXISTS idx_reactions_page ON reactions(page_id);
	CREATE INDEX IF NOT EXISTS idx_reactions_comment ON reactions(comment_id);
	CREATE INDEX IF NOT EXISTS idx_reactions_allowed ON reactions(allowed_reaction_id);
	CREATE INDEX IF NOT EXISTS idx_reactions_user ON reactions(user_id);

	CREATE TABLE IF NOT EXISTS moderation_config (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL UNIQUE,
		enabled INTEGER DEFAULT 0,
		auto_reject_threshold REAL DEFAULT 0.85,
		auto_approve_threshold REAL DEFAULT 0.30,
		check_spam INTEGER DEFAULT 1,
		check_offensive INTEGER DEFAULT 1,
		check_aggressive INTEGER DEFAULT 1,
		check_off_topic INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_moderation_config_site ON moderation_config(site_id);

	CREATE TABLE IF NOT EXISTS site_auth_configs (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL UNIQUE,
		auth_mode TEXT NOT NULL DEFAULT 'external',
		jwt_validation_type TEXT,
		jwt_secret TEXT,
		jwt_public_key TEXT,
		jwks_endpoint TEXT,
		jwt_issuer TEXT,
		jwt_audience TEXT,
		token_expiration_buffer INTEGER DEFAULT 60,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_site_auth_configs_site ON site_auth_configs(site_id);

	CREATE TABLE IF NOT EXISTS kotomi_auth_users (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		email TEXT NOT NULL,
		auth0_sub TEXT NOT NULL,
		name TEXT,
		avatar_url TEXT,
		is_verified INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
		UNIQUE(site_id, auth0_sub)
	);

	CREATE INDEX IF NOT EXISTS idx_kotomi_auth_users_site ON kotomi_auth_users(site_id);
	CREATE INDEX IF NOT EXISTS idx_kotomi_auth_users_email ON kotomi_auth_users(site_id, email);
	CREATE INDEX IF NOT EXISTS idx_kotomi_auth_users_auth0 ON kotomi_auth_users(auth0_sub);

	CREATE TABLE IF NOT EXISTS kotomi_auth_sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		site_id TEXT NOT NULL,
		token TEXT NOT NULL UNIQUE,
		refresh_token TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMP NOT NULL,
		refresh_expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES kotomi_auth_users(id) ON DELETE CASCADE,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_kotomi_auth_sessions_user ON kotomi_auth_sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_kotomi_auth_sessions_token ON kotomi_auth_sessions(token);
	CREATE INDEX IF NOT EXISTS idx_kotomi_auth_sessions_refresh ON kotomi_auth_sessions(refresh_token);

	CREATE TABLE IF NOT EXISTS notification_settings (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL UNIQUE,
		enabled INTEGER DEFAULT 0,
		provider TEXT DEFAULT 'smtp',
		from_email TEXT NOT NULL,
		from_name TEXT NOT NULL,
		reply_to TEXT,
		smtp_host TEXT,
		smtp_port INTEGER,
		smtp_user TEXT,
		smtp_password TEXT,
		smtp_encryption TEXT,
		sendgrid_api_key TEXT,
		notify_new_comment INTEGER DEFAULT 1,
		notify_reply INTEGER DEFAULT 1,
		notify_moderation INTEGER DEFAULT 1,
		owner_email TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_notification_settings_site ON notification_settings(site_id);

	CREATE TABLE IF NOT EXISTS notification_queue (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		type TEXT NOT NULL,
		recipient TEXT NOT NULL,
		subject TEXT NOT NULL,
		body TEXT NOT NULL,
		data TEXT,
		status TEXT DEFAULT 'pending',
		attempts INTEGER DEFAULT 0,
		error TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		sent_at TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_notification_queue_site ON notification_queue(site_id);
	CREATE INDEX IF NOT EXISTS idx_notification_queue_status ON notification_queue(status);
	CREATE INDEX IF NOT EXISTS idx_notification_queue_created ON notification_queue(created_at);

	CREATE TABLE IF NOT EXISTS notification_log (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		type TEXT NOT NULL,
		recipient TEXT NOT NULL,
		subject TEXT NOT NULL,
		status TEXT NOT NULL,
		error TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		sent_at TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_notification_log_site ON notification_log(site_id);
	CREATE INDEX IF NOT EXISTS idx_notification_log_created ON notification_log(created_at);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// Run migrations for existing databases
	migrations := []string{
		// Phase 3: Add reputation_score to users table if it doesn't exist
		`ALTER TABLE users ADD COLUMN reputation_score INTEGER DEFAULT 0`,
	}

	for _, migration := range migrations {
		// Try to run migration, ignore only if column already exists
		_, err := db.Exec(migration)
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			// Log unexpected errors but don't fail - allows database to work
			log.Printf("Warning: Migration error (continuing anyway): %v", err)
		}
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

	// Auto-create site and page if they don't exist (for testing and standalone use without admin)
	// This allows the comment system to work without pre-creating sites/pages
	
	// First, ensure a system admin user exists (for auto-created sites)
	systemUserID := "system"
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO admin_users (id, email, name, auth0_sub, created_at, updated_at)
		VALUES (?, 'system@kotomi.local', 'System', 'system', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, systemUserID)
	if err != nil {
		return fmt.Errorf("failed to create system admin user: %w", err)
	}

	// Check if site exists, create if not
	var siteExists bool
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM sites WHERE id = ?)", site).Scan(&siteExists)
	if err != nil {
		return fmt.Errorf("failed to check site existence: %w", err)
	}
	if !siteExists {
		// Create a placeholder site owned by system user
		_, err = s.db.Exec(`
			INSERT OR IGNORE INTO sites (id, owner_id, name, created_at, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, site, systemUserID, site)
		if err != nil {
			return fmt.Errorf("failed to auto-create site: %w", err)
		}
	}

	// Check if page exists, create if not
	var pageExists bool
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM pages WHERE site_id = ? AND id = ?)", site, page).Scan(&pageExists)
	if err != nil {
		return fmt.Errorf("failed to check page existence: %w", err)
	}
	if !pageExists {
		// Create a placeholder page
		_, err = s.db.Exec(`
			INSERT OR IGNORE INTO pages (id, site_id, path, created_at, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, page, site, page)
		if err != nil {
			return fmt.Errorf("failed to auto-create page: %w", err)
		}
	}

	query := `
		INSERT INTO comments (id, site_id, page_id, author, author_id, author_email, text, parent_id, status, moderated_by, moderated_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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

	// Convert empty AuthorEmail to NULL
	var authorEmail sql.NullString
	if comment.AuthorEmail != "" {
		authorEmail.String = comment.AuthorEmail
		authorEmail.Valid = true
	}

	_, err = s.db.Exec(query,
		comment.ID,
		site,
		page,
		comment.Author,
		comment.AuthorID,
		authorEmail,
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
		SELECT c.id, c.author, c.author_id, c.author_email, c.text, c.parent_id, c.status, 
		       c.moderated_by, c.moderated_at, c.created_at, c.updated_at,
		       COALESCE(u.is_verified, 0) as author_verified,
		       COALESCE(u.reputation_score, 0) as author_reputation
		FROM comments c
		LEFT JOIN users u ON c.site_id = u.site_id AND c.author_id = u.id
		WHERE c.site_id = ? AND c.page_id = ?
		ORDER BY c.created_at ASC
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
		var authorEmail sql.NullString

		err := rows.Scan(&c.ID, &c.Author, &c.AuthorID, &authorEmail, &c.Text, &parentID, &c.Status, 
			&moderatedBy, &moderatedAt, &c.CreatedAt, &c.UpdatedAt, &c.AuthorVerified, &c.AuthorReputation)
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
		if authorEmail.Valid {
			c.AuthorEmail = authorEmail.String
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
		SELECT c.id, c.site_id, c.page_id, c.author, c.author_id, c.author_email, c.text, c.parent_id, 
		       c.status, c.moderated_by, c.moderated_at, c.created_at, c.updated_at,
		       COALESCE(u.is_verified, 0) as author_verified,
		       COALESCE(u.reputation_score, 0) as author_reputation
		FROM comments c
		LEFT JOIN users u ON c.site_id = u.site_id AND c.author_id = u.id
		WHERE c.site_id = ?
	`
	args := []interface{}{siteID}

	if status != "" {
		query += " AND c.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY c.created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		var pageID string // Scanned but not included in returned Comment struct
		var parentID sql.NullString
		var moderatedBy sql.NullString
		var moderatedAt sql.NullTime
		var authorEmail sql.NullString

		err := rows.Scan(&c.ID, &c.SiteID, &pageID, &c.Author, &c.AuthorID, &authorEmail, &c.Text, &parentID, 
			&c.Status, &moderatedBy, &moderatedAt, &c.CreatedAt, &c.UpdatedAt, &c.AuthorVerified, &c.AuthorReputation)
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
		if authorEmail.Valid {
			c.AuthorEmail = authorEmail.String
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
		SELECT id, site_id, page_id, author, author_id, author_email, text, parent_id, status, moderated_by, moderated_at, created_at, updated_at
		FROM comments
		WHERE id = ?
	`

	var c Comment
	var pageID string // Scanned but not included in returned Comment struct
	var authorEmail sql.NullString
	var parentID sql.NullString
	var moderatedBy sql.NullString
	var moderatedAt sql.NullTime

	err := s.db.QueryRow(query, commentID).Scan(
		&c.ID, &c.SiteID, &pageID, &c.Author, &c.AuthorID, &authorEmail, &c.Text, &parentID, &c.Status, &moderatedBy, &moderatedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to query comment: %w", err)
	}

	if authorEmail.Valid {
		c.AuthorEmail = authorEmail.String
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

// UpdateCommentText updates the text content of a comment
func (s *SQLiteStore) UpdateCommentText(commentID, text string) error {
	query := `
		UPDATE comments
		SET text = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	result, err := s.db.Exec(query, text, now, commentID)
	if err != nil {
		return fmt.Errorf("failed to update comment text: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comment not found")
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
