package models

import "time"

// ExportData represents the complete export structure
type ExportData struct {
	Metadata ExportMetadata `json:"metadata"`
	Site     Site           `json:"site"`
	Pages    []PageExport   `json:"pages"`
}

// ExportMetadata contains information about the export
type ExportMetadata struct {
	Version     string    `json:"version"`
	ExportedAt  time.Time `json:"exported_at"`
	SiteID      string    `json:"site_id"`
	SiteName    string    `json:"site_name"`
	TotalPages  int       `json:"total_pages"`
	TotalComments int     `json:"total_comments"`
	TotalReactions int    `json:"total_reactions"`
}

// PageExport represents a page with all its comments and reactions
type PageExport struct {
	Page              Page                `json:"page"`
	Comments          []CommentExport     `json:"comments"`
	PageReactions     []ReactionExport    `json:"page_reactions"`
	AllowedReactions  []AllowedReaction   `json:"allowed_reactions"`
}

// CommentExport represents a comment with its reactions
type CommentExport struct {
	ID            string             `json:"id"`
	Author        string             `json:"author"`
	AuthorID      string             `json:"author_id"`
	AuthorEmail   string             `json:"author_email,omitempty"`
	Text          string             `json:"text"`
	ParentID      string             `json:"parent_id,omitempty"`
	Status        string             `json:"status"`
	ModeratedBy   string             `json:"moderated_by,omitempty"`
	ModeratedAt   *time.Time         `json:"moderated_at,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	Reactions     []ReactionExport   `json:"reactions"`
}

// ReactionExport represents a reaction for export
type ReactionExport struct {
	AllowedReactionID string    `json:"allowed_reaction_id"`
	ReactionName      string    `json:"reaction_name"`
	ReactionEmoji     string    `json:"reaction_emoji"`
	UserIdentifier    string    `json:"user_identifier,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}
