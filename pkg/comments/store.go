package comments

import (
	"context"
	"database/sql"
)

// Store defines the interface for comment storage operations
// This interface can be implemented by different database backends (SQLite, PostgreSQL, etc.)
type Store interface {
	// Comment operations
	AddPageComment(ctx context.Context, site, page string, comment Comment) error
	GetPageComments(ctx context.Context, site, page string) ([]Comment, error)
	GetCommentsBySite(ctx context.Context, siteID string, status string) ([]Comment, error)
	GetCommentByID(ctx context.Context, commentID string) (*Comment, error)
	UpdateCommentStatus(ctx context.Context, commentID, status, moderatorID string) error
	UpdateCommentText(ctx context.Context, commentID, text string) error
	DeleteComment(ctx context.Context, commentID string) error
	GetCommentSiteID(ctx context.Context, commentID string) (string, error)

	// Database operations
	GetDB() *sql.DB
	Close() error
}
