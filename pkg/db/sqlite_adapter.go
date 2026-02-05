package db

import (
	"context"
	"database/sql"

	"github.com/saasuke-labs/kotomi/pkg/comments"
)

// SQLiteAdapter adapts the SQLiteStore to the Store interface
type SQLiteAdapter struct {
	store *comments.SQLiteStore
}

// NewSQLiteAdapter creates a new SQLite adapter
func NewSQLiteAdapter(dbPath string) (*SQLiteAdapter, error) {
	store, err := comments.NewSQLiteStore(dbPath)
	if err != nil {
		return nil, err
	}
	return &SQLiteAdapter{store: store}, nil
}

// AddPageComment adds a comment to a specific page
func (a *SQLiteAdapter) AddPageComment(ctx context.Context, site, page string, comment comments.Comment) error {
	return a.store.AddPageComment(ctx, site, page, comment)
}

// GetPageComments retrieves all comments for a specific page
func (a *SQLiteAdapter) GetPageComments(ctx context.Context, site, page string) ([]comments.Comment, error) {
	return a.store.GetPageComments(ctx, site, page)
}

// GetCommentsBySite retrieves comments for a site with optional status filter
func (a *SQLiteAdapter) GetCommentsBySite(ctx context.Context, siteID string, status string) ([]comments.Comment, error) {
	return a.store.GetCommentsBySite(ctx, siteID, status)
}

// GetCommentByID retrieves a specific comment by ID
func (a *SQLiteAdapter) GetCommentByID(ctx context.Context, commentID string) (*comments.Comment, error) {
	return a.store.GetCommentByID(ctx, commentID)
}

// UpdateCommentStatus updates a comment's status
func (a *SQLiteAdapter) UpdateCommentStatus(ctx context.Context, commentID, status, moderatorID string) error {
	return a.store.UpdateCommentStatus(ctx, commentID, status, moderatorID)
}

// UpdateCommentText updates a comment's text content
func (a *SQLiteAdapter) UpdateCommentText(ctx context.Context, commentID, text string) error {
	return a.store.UpdateCommentText(ctx, commentID, text)
}

// DeleteComment deletes a comment by ID
func (a *SQLiteAdapter) DeleteComment(ctx context.Context, commentID string) error {
	return a.store.DeleteComment(ctx, commentID)
}

// GetCommentSiteID retrieves the site ID for a comment
func (a *SQLiteAdapter) GetCommentSiteID(ctx context.Context, commentID string) (string, error) {
	return a.store.GetCommentSiteID(ctx, commentID)
}

// GetDB returns the underlying database connection
func (a *SQLiteAdapter) GetDB() *sql.DB {
	return a.store.GetDB()
}

// Close closes the database connection
func (a *SQLiteAdapter) Close() error {
	return a.store.Close()
}
