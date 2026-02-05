package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FirestoreStore provides Firestore-based persistent storage for comments
type FirestoreStore struct {
	client    *firestore.Client
	projectID string
}

// NewFirestoreStore creates a new Firestore-based comment store
func NewFirestoreStore(ctx context.Context, projectID string) (*FirestoreStore, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	log.Printf("Firestore database initialized for project: %s", projectID)

	store := &FirestoreStore{
		client:    client,
		projectID: projectID,
	}

	// Create indexes (note: Firestore indexes must be created via console or firestore.indexes.json)
	// Log recommendations for manual index creation
	log.Println("Firestore indexes should be configured in firestore.indexes.json")
	log.Println("Recommended composite indexes:")
	log.Println("  - comments: site_id, page_id, created_at")
	log.Println("  - comments: site_id, status, created_at")
	log.Println("  - comments: author_id, created_at")

	return store, nil
}

// AddPageComment adds a comment to a specific page
func (s *FirestoreStore) AddPageComment(ctx context.Context, site, page string, comment interface{}) error {
	c, ok := comment.(comments.Comment)
	if !ok {
		return fmt.Errorf("expected comments.Comment, got %T", comment)
	}

	// Set timestamps if not already set
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = time.Now()
	}
	// Set default status if not set
	if c.Status == "" {
		c.Status = "pending"
	}

	// Store comment in Firestore with optimized structure
	// Collection: comments/{commentID}
	// This allows direct access by ID and efficient queries
	_, err := s.client.Collection("comments").Doc(c.ID).Set(ctx, map[string]interface{}{
		"id":                c.ID,
		"site_id":           site,
		"page_id":           page,
		"author":            c.Author,
		"author_id":         c.AuthorID,
		"author_email":      c.AuthorEmail,
		"author_verified":   c.AuthorVerified,
		"author_reputation": c.AuthorReputation,
		"text":              c.Text,
		"parent_id":         c.ParentID,
		"status":            c.Status,
		"moderated_by":      c.ModeratedBy,
		"moderated_at":      c.ModeratedAt,
		"created_at":        c.CreatedAt,
		"updated_at":        c.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	// Create site document if it doesn't exist (auto-create for testing)
	siteRef := s.client.Collection("sites").Doc(site)
	_, err = siteRef.Get(ctx)
	if err != nil {
		// Check if it's a "not found" error vs other errors
		if status.Code(err) == codes.NotFound {
			// Site doesn't exist, create it
			_, err = siteRef.Set(ctx, map[string]interface{}{
				"id":         site,
				"owner_id":   "system",
				"name":       site,
				"created_at": time.Now(),
				"updated_at": time.Now(),
			})
			if err != nil {
				log.Printf("Warning: Could not auto-create site: %v", err)
			}
		} else {
			// Other errors (permission, network, etc.)
			log.Printf("Warning: Error checking site existence: %v", err)
		}
	}

	// Create page document if it doesn't exist
	pageRef := s.client.Collection("pages").Doc(page)
	_, err = pageRef.Get(ctx)
	if err != nil {
		// Check if it's a "not found" error vs other errors
		if status.Code(err) == codes.NotFound {
			// Page doesn't exist, create it
			_, err = pageRef.Set(ctx, map[string]interface{}{
				"id":         page,
				"site_id":    site,
				"path":       page,
				"created_at": time.Now(),
				"updated_at": time.Now(),
			})
			if err != nil {
				log.Printf("Warning: Could not auto-create page: %v", err)
			}
		} else {
			// Other errors (permission, network, etc.)
			log.Printf("Warning: Error checking page existence: %v", err)
		}
	}

	return nil
}

// GetPageComments retrieves all comments for a specific page
func (s *FirestoreStore) GetPageComments(ctx context.Context, site, page string) ([]interface{}, error) {
	// Query with composite index: site_id + page_id + created_at
	query := s.client.Collection("comments").
		Where("site_id", "==", site).
		Where("page_id", "==", page).
		OrderBy("created_at", firestore.Asc)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var result []interface{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate comments: %w", err)
		}

		comment := s.docToComment(doc)
		result = append(result, comment)
	}

	return result, nil
}

// GetCommentsBySite retrieves comments for a site with optional status filter
func (s *FirestoreStore) GetCommentsBySite(ctx context.Context, siteID string, status string) ([]interface{}, error) {
	query := s.client.Collection("comments").Where("site_id", "==", siteID)

	if status != "" {
		query = query.Where("status", "==", status)
	}

	query = query.OrderBy("created_at", firestore.Desc)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var result []interface{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate comments: %w", err)
		}

		comment := s.docToComment(doc)
		result = append(result, comment)
	}

	return result, nil
}

// GetCommentByID retrieves a specific comment by ID
func (s *FirestoreStore) GetCommentByID(ctx context.Context, commentID string) (interface{}, error) {
	doc, err := s.client.Collection("comments").Doc(commentID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	comment := s.docToComment(doc)
	return &comment, nil
}

// UpdateCommentStatus updates a comment's status
func (s *FirestoreStore) UpdateCommentStatus(ctx context.Context, commentID, status, moderatorID string) error {
	_, err := s.client.Collection("comments").Doc(commentID).Update(ctx, []firestore.Update{
		{Path: "status", Value: status},
		{Path: "moderated_by", Value: moderatorID},
		{Path: "moderated_at", Value: time.Now()},
		{Path: "updated_at", Value: time.Now()},
	})

	if err != nil {
		return fmt.Errorf("failed to update comment status: %w", err)
	}

	return nil
}

// UpdateCommentText updates a comment's text content
func (s *FirestoreStore) UpdateCommentText(ctx context.Context, commentID, text string) error {
	_, err := s.client.Collection("comments").Doc(commentID).Update(ctx, []firestore.Update{
		{Path: "text", Value: text},
		{Path: "updated_at", Value: time.Now()},
	})

	if err != nil {
		return fmt.Errorf("failed to update comment text: %w", err)
	}

	return nil
}

// DeleteComment deletes a comment by ID
func (s *FirestoreStore) DeleteComment(ctx context.Context, commentID string) error {
	_, err := s.client.Collection("comments").Doc(commentID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

// GetCommentSiteID retrieves the site ID for a comment
func (s *FirestoreStore) GetCommentSiteID(ctx context.Context, commentID string) (string, error) {
	doc, err := s.client.Collection("comments").Doc(commentID).Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get comment: %w", err)
	}

	siteID, ok := doc.Data()["site_id"].(string)
	if !ok {
		return "", fmt.Errorf("site_id not found in comment")
	}

	return siteID, nil
}

// GetDB returns nil for Firestore (no SQL database)
func (s *FirestoreStore) GetDB() *sql.DB {
	return nil
}

// Close closes the Firestore client
func (s *FirestoreStore) Close() error {
	return s.client.Close()
}

// docToComment converts a Firestore document to a Comment struct
func (s *FirestoreStore) docToComment(doc *firestore.DocumentSnapshot) comments.Comment {
	data := doc.Data()

	comment := comments.Comment{
		ID:         getString(data, "id"),
		SiteID:     getString(data, "site_id"),
		Author:     getString(data, "author"),
		AuthorID:   getString(data, "author_id"),
		Text:       getString(data, "text"),
		ParentID:   getString(data, "parent_id"),
		Status:     getString(data, "status"),
		CreatedAt:  getTime(data, "created_at"),
		UpdatedAt:  getTime(data, "updated_at"),
	}

	// Optional fields
	if email, ok := data["author_email"].(string); ok {
		comment.AuthorEmail = email
	}
	if verified, ok := data["author_verified"].(bool); ok {
		comment.AuthorVerified = verified
	}
	if reputation, ok := data["author_reputation"].(int64); ok {
		comment.AuthorReputation = int(reputation)
	}
	if moderatedBy, ok := data["moderated_by"].(string); ok {
		comment.ModeratedBy = moderatedBy
	}
	if moderatedAt := getTime(data, "moderated_at"); !moderatedAt.IsZero() {
		comment.ModeratedAt = moderatedAt
	}

	return comment
}

// Helper functions for type conversion
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getTime(data map[string]interface{}, key string) time.Time {
	if val, ok := data[key].(time.Time); ok {
		return val
	}
	return time.Time{}
}
