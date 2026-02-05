package db

import (
	"context"
	"database/sql"

	"github.com/saasuke-labs/kotomi/pkg/comments"
)

// Provider represents a database provider type
type Provider string

const (
	// ProviderSQLite represents SQLite database
	ProviderSQLite Provider = "sqlite"
	// ProviderFirestore represents Google Cloud Firestore
	ProviderFirestore Provider = "firestore"
)

// Store is the main database interface that all implementations must satisfy
// This interface provides a unified way to interact with different database backends
// (SQLite, Firestore, etc.) without the caller needing to know the implementation details
type Store interface {
	// AddPageComment adds a comment to a specific page
	AddPageComment(ctx context.Context, site, page string, comment comments.Comment) error
	// GetPageComments retrieves all comments for a specific page
	GetPageComments(ctx context.Context, site, page string) ([]comments.Comment, error)
	// GetCommentsBySite retrieves comments for a site with optional status filter
	GetCommentsBySite(ctx context.Context, siteID string, status string) ([]comments.Comment, error)
	// GetCommentByID retrieves a specific comment by ID
	GetCommentByID(ctx context.Context, commentID string) (*comments.Comment, error)
	// UpdateCommentStatus updates a comment's status (pending, approved, rejected)
	UpdateCommentStatus(ctx context.Context, commentID, status, moderatorID string) error
	// UpdateCommentText updates a comment's text content
	UpdateCommentText(ctx context.Context, commentID, text string) error
	// DeleteComment deletes a comment by ID
	DeleteComment(ctx context.Context, commentID string) error
	// GetCommentSiteID retrieves the site ID for a comment
	GetCommentSiteID(ctx context.Context, commentID string) (string, error)
	// GetDB returns the underlying database connection (for SQLite) or nil for NoSQL databases
	GetDB() *sql.DB
	// Close closes the database connection
	Close() error
}
