package db

import (
	"context"
	"database/sql"
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
type Store interface {
	// CommentStore provides comment operations
	CommentStore
	// GetDB returns the underlying database connection (for SQLite) or nil for NoSQL databases
	GetDB() *sql.DB
	// Close closes the database connection
	Close() error
}

// CommentStore defines the interface for comment operations
type CommentStore interface {
	// AddPageComment adds a comment to a specific page
	AddPageComment(ctx context.Context, site, page string, comment interface{}) error
	// GetPageComments retrieves all comments for a specific page
	GetPageComments(ctx context.Context, site, page string) ([]interface{}, error)
	// GetCommentsBySite retrieves comments for a site with optional status filter
	GetCommentsBySite(ctx context.Context, siteID string, status string) ([]interface{}, error)
	// GetCommentByID retrieves a specific comment by ID
	GetCommentByID(ctx context.Context, commentID string) (interface{}, error)
	// UpdateCommentStatus updates a comment's status (pending, approved, rejected)
	UpdateCommentStatus(ctx context.Context, commentID, status, moderatorID string) error
	// UpdateCommentText updates a comment's text content
	UpdateCommentText(ctx context.Context, commentID, text string) error
	// DeleteComment deletes a comment by ID
	DeleteComment(ctx context.Context, commentID string) error
	// GetCommentSiteID retrieves the site ID for a comment
	GetCommentSiteID(ctx context.Context, commentID string) (string, error)
}
