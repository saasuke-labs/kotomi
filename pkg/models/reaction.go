package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AllowedReaction represents a reaction type that is allowed on a site
type AllowedReaction struct {
	ID           string    `json:"id"`
	SiteID       string    `json:"site_id"`
	Name         string    `json:"name"`          // Unique name for admins/logging (e.g., "thumbs_up", "heart")
	Emoji        string    `json:"emoji"`         // The emoji to display (e.g., "👍", "❤️")
	ReactionType string    `json:"reaction_type"` // 'page', 'comment', or 'both'
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Reaction represents a user's reaction to a page or comment
type Reaction struct {
	ID                string    `json:"id"`
	PageID            string    `json:"page_id,omitempty"`    // Set for page reactions
	CommentID         string    `json:"comment_id,omitempty"` // Set for comment reactions
	AllowedReactionID string    `json:"allowed_reaction_id"`
	UserID            string    `json:"user_id"` // Authenticated user ID
	CreatedAt         time.Time `json:"created_at"`
}

// ReactionWithDetails includes the emoji and name from the allowed reaction
type ReactionWithDetails struct {
	ID        string    `json:"id"`
	PageID    string    `json:"page_id,omitempty"`
	CommentID string    `json:"comment_id,omitempty"`
	Name      string    `json:"name"`
	Emoji     string    `json:"emoji"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ReactionCount represents aggregated reaction counts for a comment
type ReactionCount struct {
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	Count int    `json:"count"`
}

// AllowedReactionStore handles allowed_reactions database operations
type AllowedReactionStore struct {
	db *sql.DB
}

// NewAllowedReactionStore creates a new allowed reaction store
func NewAllowedReactionStore(db *sql.DB) *AllowedReactionStore {
	return &AllowedReactionStore{db: db}
}

// GetBySite retrieves all allowed reactions for a site
func (s *AllowedReactionStore) GetBySite(siteID string) ([]AllowedReaction, error) {
	query := `
		SELECT id, site_id, name, emoji, reaction_type, created_at, updated_at
		FROM allowed_reactions
		WHERE site_id = ?
		ORDER BY reaction_type, created_at ASC
	`

	rows, err := s.db.Query(query, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query allowed reactions: %w", err)
	}
	defer rows.Close()

	var reactions []AllowedReaction
	for rows.Next() {
		var reaction AllowedReaction
		err := rows.Scan(
			&reaction.ID, &reaction.SiteID, &reaction.Name, &reaction.Emoji,
			&reaction.ReactionType, &reaction.CreatedAt, &reaction.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan allowed reaction: %w", err)
		}
		reactions = append(reactions, reaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating allowed reactions: %w", err)
	}

	if reactions == nil {
		reactions = []AllowedReaction{}
	}

	return reactions, nil
}

// GetBySiteAndType retrieves allowed reactions for a site filtered by type
func (s *AllowedReactionStore) GetBySiteAndType(siteID, reactionType string) ([]AllowedReaction, error) {
	query := `
		SELECT id, site_id, name, emoji, reaction_type, created_at, updated_at
		FROM allowed_reactions
		WHERE site_id = ? AND (reaction_type = ? OR reaction_type = 'both')
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, siteID, reactionType)
	if err != nil {
		return nil, fmt.Errorf("failed to query allowed reactions: %w", err)
	}
	defer rows.Close()

	var reactions []AllowedReaction
	for rows.Next() {
		var reaction AllowedReaction
		err := rows.Scan(
			&reaction.ID, &reaction.SiteID, &reaction.Name, &reaction.Emoji,
			&reaction.ReactionType, &reaction.CreatedAt, &reaction.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan allowed reaction: %w", err)
		}
		reactions = append(reactions, reaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating allowed reactions: %w", err)
	}

	if reactions == nil {
		reactions = []AllowedReaction{}
	}

	return reactions, nil
}

// GetByID retrieves an allowed reaction by its ID
func (s *AllowedReactionStore) GetByID(id string) (*AllowedReaction, error) {
	query := `
		SELECT id, site_id, name, emoji, reaction_type, created_at, updated_at
		FROM allowed_reactions
		WHERE id = ?
	`

	var reaction AllowedReaction
	err := s.db.QueryRow(query, id).Scan(
		&reaction.ID, &reaction.SiteID, &reaction.Name, &reaction.Emoji,
		&reaction.ReactionType, &reaction.CreatedAt, &reaction.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("allowed reaction not found")
		}
		return nil, fmt.Errorf("failed to query allowed reaction: %w", err)
	}

	return &reaction, nil
}

// Create creates a new allowed reaction for a site
func (s *AllowedReactionStore) Create(siteID, name, emoji, reactionType string) (*AllowedReaction, error) {
	// Default to 'comment' if not specified
	if reactionType == "" {
		reactionType = "comment"
	}

	now := time.Now()
	reaction := &AllowedReaction{
		ID:           uuid.NewString(),
		SiteID:       siteID,
		Name:         name,
		Emoji:        emoji,
		ReactionType: reactionType,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	query := `
		INSERT INTO allowed_reactions (id, site_id, name, emoji, reaction_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, reaction.ID, reaction.SiteID, reaction.Name, reaction.Emoji,
		reaction.ReactionType, reaction.CreatedAt, reaction.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create allowed reaction: %w", err)
	}

	return reaction, nil
}

// Update updates an allowed reaction
func (s *AllowedReactionStore) Update(id, name, emoji, reactionType string) error {
	query := `
		UPDATE allowed_reactions
		SET name = ?, emoji = ?, reaction_type = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, name, emoji, reactionType, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update allowed reaction: %w", err)
	}

	return nil
}

// Delete deletes an allowed reaction
func (s *AllowedReactionStore) Delete(id string) error {
	query := `DELETE FROM allowed_reactions WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete allowed reaction: %w", err)
	}

	return nil
}

// ReactionStore handles reactions database operations
type ReactionStore struct {
	db *sql.DB
}

// NewReactionStore creates a new reaction store
func NewReactionStore(db *sql.DB) *ReactionStore {
	return &ReactionStore{db: db}
}

// AddReaction adds a reaction to a comment (or toggles it off if already exists)
func (s *ReactionStore) AddReaction(commentID, allowedReactionID, userID string) (*Reaction, error) {
	// Check if user already reacted with this reaction type
	existing, err := s.GetUserCommentReaction(commentID, allowedReactionID, userID)
	if err == nil && existing != nil {
		// User already reacted with this type - toggle it off (remove it)
		if err := s.RemoveReaction(existing.ID); err != nil {
			return nil, fmt.Errorf("failed to remove existing reaction: %w", err)
		}
		return nil, nil // Return nil to indicate removal
	}

	// Add new reaction
	now := time.Now()
	reaction := &Reaction{
		ID:                uuid.NewString(),
		CommentID:         commentID,
		AllowedReactionID: allowedReactionID,
		UserID:            userID,
		CreatedAt:         now,
	}

	query := `
		INSERT INTO reactions (id, page_id, comment_id, allowed_reaction_id, user_id, created_at)
		VALUES (?, NULL, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query, reaction.ID, reaction.CommentID, reaction.AllowedReactionID,
		reaction.UserID, reaction.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to add reaction: %w", err)
	}

	return reaction, nil
}

// AddPageReaction adds a reaction to a page (or toggles it off if already exists)
func (s *ReactionStore) AddPageReaction(pageID, allowedReactionID, userID string) (*Reaction, error) {
	// Check if user already reacted with this reaction type
	existing, err := s.GetUserPageReaction(pageID, allowedReactionID, userID)
	if err == nil && existing != nil {
		// User already reacted with this type - toggle it off (remove it)
		if err := s.RemoveReaction(existing.ID); err != nil {
			return nil, fmt.Errorf("failed to remove existing reaction: %w", err)
		}
		return nil, nil // Return nil to indicate removal
	}

	// Add new reaction
	now := time.Now()
	reaction := &Reaction{
		ID:                uuid.NewString(),
		PageID:            pageID,
		AllowedReactionID: allowedReactionID,
		UserID:            userID,
		CreatedAt:         now,
	}

	query := `
		INSERT INTO reactions (id, page_id, comment_id, allowed_reaction_id, user_id, created_at)
		VALUES (?, ?, NULL, ?, ?, ?)
	`

	_, err = s.db.Exec(query, reaction.ID, reaction.PageID, reaction.AllowedReactionID,
		reaction.UserID, reaction.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to add page reaction: %w", err)
	}

	return reaction, nil
}

// GetUserCommentReaction checks if a user has already reacted to a comment with a specific reaction type
func (s *ReactionStore) GetUserCommentReaction(commentID, allowedReactionID, userID string) (*Reaction, error) {
	query := `
		SELECT id, page_id, comment_id, allowed_reaction_id, user_id, created_at
		FROM reactions
		WHERE comment_id = ? AND allowed_reaction_id = ? AND user_id = ?
	`

	var reaction Reaction
	var pageID sql.NullString
	err := s.db.QueryRow(query, commentID, allowedReactionID, userID).Scan(
		&reaction.ID, &pageID, &reaction.CommentID, &reaction.AllowedReactionID,
		&reaction.UserID, &reaction.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reaction not found")
		}
		return nil, fmt.Errorf("failed to query reaction: %w", err)
	}

	if pageID.Valid {
		reaction.PageID = pageID.String
	}

	return &reaction, nil
}

// GetUserPageReaction checks if a user has already reacted to a page with a specific reaction type
func (s *ReactionStore) GetUserPageReaction(pageID, allowedReactionID, userID string) (*Reaction, error) {
	query := `
		SELECT id, page_id, comment_id, allowed_reaction_id, user_id, created_at
		FROM reactions
		WHERE page_id = ? AND allowed_reaction_id = ? AND user_id = ?
	`

	var reaction Reaction
	var commentID sql.NullString
	err := s.db.QueryRow(query, pageID, allowedReactionID, userID).Scan(
		&reaction.ID, &reaction.PageID, &commentID, &reaction.AllowedReactionID,
		&reaction.UserID, &reaction.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reaction not found")
		}
		return nil, fmt.Errorf("failed to query reaction: %w", err)
	}

	if commentID.Valid {
		reaction.CommentID = commentID.String
	}

	return &reaction, nil
}

// GetUserReaction checks if a user has already reacted with a specific reaction type (deprecated, use GetUserCommentReaction)
func (s *ReactionStore) GetUserReaction(commentID, allowedReactionID, userID string) (*Reaction, error) {
	return s.GetUserCommentReaction(commentID, allowedReactionID, userID)
}

// RemoveReaction removes a reaction by its ID
func (s *ReactionStore) RemoveReaction(reactionID string) error {
	query := `DELETE FROM reactions WHERE id = ?`

	result, err := s.db.Exec(query, reactionID)
	if err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("reaction not found")
	}

	return nil
}

// GetReactionsByComment retrieves all reactions for a comment with details
func (s *ReactionStore) GetReactionsByComment(commentID string) ([]ReactionWithDetails, error) {
	query := `
		SELECT r.id, r.page_id, r.comment_id, ar.name, ar.emoji, r.user_id, r.created_at
		FROM reactions r
		JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		WHERE r.comment_id = ?
		ORDER BY r.created_at ASC
	`

	rows, err := s.db.Query(query, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reactions: %w", err)
	}
	defer rows.Close()

	var reactions []ReactionWithDetails
	for rows.Next() {
		var reaction ReactionWithDetails
		var pageID sql.NullString
		err := rows.Scan(
			&reaction.ID, &pageID, &reaction.CommentID, &reaction.Name, &reaction.Emoji,
			&reaction.UserID, &reaction.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reaction: %w", err)
		}
		if pageID.Valid {
			reaction.PageID = pageID.String
		}
		reactions = append(reactions, reaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reactions: %w", err)
	}

	if reactions == nil {
		reactions = []ReactionWithDetails{}
	}

	return reactions, nil
}

// GetReactionsByPage retrieves all reactions for a page with details
func (s *ReactionStore) GetReactionsByPage(pageID string) ([]ReactionWithDetails, error) {
	query := `
		SELECT r.id, r.page_id, r.comment_id, ar.name, ar.emoji, r.user_id, r.created_at
		FROM reactions r
		JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		WHERE r.page_id = ?
		ORDER BY r.created_at ASC
	`

	rows, err := s.db.Query(query, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reactions: %w", err)
	}
	defer rows.Close()

	var reactions []ReactionWithDetails
	for rows.Next() {
		var reaction ReactionWithDetails
		var commentID sql.NullString
		err := rows.Scan(
			&reaction.ID, &reaction.PageID, &commentID, &reaction.Name, &reaction.Emoji,
			&reaction.UserID, &reaction.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reaction: %w", err)
		}
		if commentID.Valid {
			reaction.CommentID = commentID.String
		}
		reactions = append(reactions, reaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reactions: %w", err)
	}

	if reactions == nil {
		reactions = []ReactionWithDetails{}
	}

	return reactions, nil
}

// GetReactionCounts retrieves aggregated reaction counts for a comment
func (s *ReactionStore) GetReactionCounts(commentID string) ([]ReactionCount, error) {
	query := `
		SELECT ar.name, ar.emoji, COUNT(*) as count
		FROM reactions r
		JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		WHERE r.comment_id = ?
		GROUP BY ar.name, ar.emoji
		ORDER BY count DESC, ar.name ASC
	`

	rows, err := s.db.Query(query, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reaction counts: %w", err)
	}
	defer rows.Close()

	var counts []ReactionCount
	for rows.Next() {
		var count ReactionCount
		err := rows.Scan(&count.Name, &count.Emoji, &count.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reaction count: %w", err)
		}
		counts = append(counts, count)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reaction counts: %w", err)
	}

	if counts == nil {
		counts = []ReactionCount{}
	}

	return counts, nil
}

// GetPageReactionCounts retrieves aggregated reaction counts for a page
func (s *ReactionStore) GetPageReactionCounts(pageID string) ([]ReactionCount, error) {
	query := `
		SELECT ar.name, ar.emoji, COUNT(*) as count
		FROM reactions r
		JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		WHERE r.page_id = ?
		GROUP BY ar.name, ar.emoji
		ORDER BY count DESC, ar.name ASC
	`

	rows, err := s.db.Query(query, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reaction counts: %w", err)
	}
	defer rows.Close()

	var counts []ReactionCount
	for rows.Next() {
		var count ReactionCount
		err := rows.Scan(&count.Name, &count.Emoji, &count.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reaction count: %w", err)
		}
		counts = append(counts, count)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reaction counts: %w", err)
	}

	if counts == nil {
		counts = []ReactionCount{}
	}

	return counts, nil
}
