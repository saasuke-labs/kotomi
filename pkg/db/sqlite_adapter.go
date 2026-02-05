package db

import (
	"context"
	"database/sql"
	"fmt"

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
func (a *SQLiteAdapter) AddPageComment(ctx context.Context, site, page string, comment interface{}) error {
	c, ok := comment.(comments.Comment)
	if !ok {
		return fmt.Errorf("invalid comment type: expected comments.Comment, got %T", comment)
	}
	return a.store.AddPageComment(ctx, site, page, c)
}

// GetPageComments retrieves all comments for a specific page
func (a *SQLiteAdapter) GetPageComments(ctx context.Context, site, page string) ([]interface{}, error) {
	commentsList, err := a.store.GetPageComments(ctx, site, page)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(commentsList))
	for i, c := range commentsList {
		result[i] = c
	}
	return result, nil
}

// GetCommentsBySite retrieves comments for a site with optional status filter
func (a *SQLiteAdapter) GetCommentsBySite(ctx context.Context, siteID string, status string) ([]interface{}, error) {
	commentsList, err := a.store.GetCommentsBySite(ctx, siteID, status)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(commentsList))
	for i, c := range commentsList {
		result[i] = c
	}
	return result, nil
}

// GetCommentByID retrieves a specific comment by ID
func (a *SQLiteAdapter) GetCommentByID(ctx context.Context, commentID string) (interface{}, error) {
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

// GetSQLiteStore returns the underlying SQLite store for compatibility
func (a *SQLiteAdapter) GetSQLiteStore() *comments.SQLiteStore {
	return a.store
}
