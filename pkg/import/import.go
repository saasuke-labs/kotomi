package importpkg

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/saasuke-labs/kotomi/pkg/models"
)

// DuplicateStrategy defines how to handle duplicate entries
type DuplicateStrategy string

const (
	StrategySkip   DuplicateStrategy = "skip"   // Skip duplicates
	StrategyUpdate DuplicateStrategy = "update" // Update existing entries
)

// ImportResult contains the results of an import operation
type ImportResult struct {
	CommentsImported   int      `json:"comments_imported"`
	CommentsSkipped    int      `json:"comments_skipped"`
	CommentsUpdated    int      `json:"comments_updated"`
	ReactionsImported  int      `json:"reactions_imported"`
	ReactionsSkipped   int      `json:"reactions_skipped"`
	PagesCreated       int      `json:"pages_created"`
	PagesSkipped       int      `json:"pages_skipped"`
	Errors             []string `json:"errors,omitempty"`
}

// Importer handles data import operations
type Importer struct {
	db       *sql.DB
	strategy DuplicateStrategy
}

// NewImporter creates a new Importer
func NewImporter(db *sql.DB, strategy DuplicateStrategy) *Importer {
	return &Importer{
		db:       db,
		strategy: strategy,
	}
}

// ImportFromJSON imports data from JSON format
func (i *Importer) ImportFromJSON(r io.Reader, siteID string) (*ImportResult, error) {
	var exportData models.ExportData
	if err := json.NewDecoder(r).Decode(&exportData); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Validate that the export is for the correct site
	if exportData.Metadata.SiteID != siteID {
		return nil, fmt.Errorf("export is for site %s, but importing to site %s",
			exportData.Metadata.SiteID, siteID)
	}

	result := &ImportResult{
		Errors: make([]string, 0),
	}

	// Start a transaction
	tx, err := i.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Import pages and their data
	for _, pageExport := range exportData.Pages {
		// Import or get existing page
		pageID, created, err := i.importPage(tx, siteID, &pageExport.Page)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Failed to import page %s: %v", pageExport.Page.Path, err))
			continue
		}

		if created {
			result.PagesCreated++
		} else {
			result.PagesSkipped++
		}

		// Import comments for this page
		for _, comment := range pageExport.Comments {
			imported, skipped, updated, err := i.importComment(tx, siteID, pageID, &comment)
			if err != nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("Failed to import comment %s: %v", comment.ID, err))
				continue
			}

			result.CommentsImported += imported
			result.CommentsSkipped += skipped
			result.CommentsUpdated += updated

			// Import reactions for this comment
			for _, reaction := range comment.Reactions {
				imported, skipped, err := i.importCommentReaction(tx, comment.ID, &reaction)
				if err != nil {
					result.Errors = append(result.Errors,
						fmt.Sprintf("Failed to import reaction: %v", err))
					continue
				}
				result.ReactionsImported += imported
				result.ReactionsSkipped += skipped
			}
		}

		// Import page reactions
		for _, reaction := range pageExport.PageReactions {
			imported, skipped, err := i.importPageReaction(tx, pageID, &reaction)
			if err != nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("Failed to import page reaction: %v", err))
				continue
			}
			result.ReactionsImported += imported
			result.ReactionsSkipped += skipped
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// importPage imports a page or returns existing page ID
func (i *Importer) importPage(tx *sql.Tx, siteID string, page *models.Page) (string, bool, error) {
	// Check if page already exists
	var existingID string
	err := tx.QueryRow(`SELECT id FROM pages WHERE site_id = ? AND path = ?`,
		siteID, page.Path).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Page doesn't exist, create it
		now := time.Now().UTC()
		_, err := tx.Exec(`
			INSERT INTO pages (id, site_id, path, title, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			page.ID, siteID, page.Path, page.Title, now, now)
		if err != nil {
			return "", false, err
		}
		return page.ID, true, nil
	} else if err != nil {
		return "", false, err
	}

	// Page exists, return existing ID
	return existingID, false, nil
}

// importComment imports a comment with duplicate handling
func (i *Importer) importComment(tx *sql.Tx, siteID, pageID string, comment *models.CommentExport) (imported, skipped, updated int, err error) {
	// Check if comment already exists
	var existingID string
	err = tx.QueryRow(`SELECT id FROM comments WHERE id = ?`, comment.ID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Comment doesn't exist, import it
		_, err = tx.Exec(`
			INSERT INTO comments (id, site_id, page_id, author, author_id, author_email, text, parent_id, 
			                      status, moderated_by, moderated_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			comment.ID, siteID, pageID, comment.Author, comment.AuthorID, nullString(comment.AuthorEmail),
			comment.Text, nullString(comment.ParentID), comment.Status,
			nullString(comment.ModeratedBy), nullTime(comment.ModeratedAt),
			comment.CreatedAt, comment.UpdatedAt)
		if err != nil {
			return 0, 0, 0, err
		}
		return 1, 0, 0, nil
	} else if err != nil {
		return 0, 0, 0, err
	}

	// Comment exists, handle based on strategy
	if i.strategy == StrategySkip {
		return 0, 1, 0, nil
	}

	// Strategy is Update
	_, err = tx.Exec(`
		UPDATE comments 
		SET author = ?, author_id = ?, author_email = ?, text = ?, parent_id = ?, 
		    status = ?, moderated_by = ?, moderated_at = ?, updated_at = ?
		WHERE id = ?`,
		comment.Author, comment.AuthorID, nullString(comment.AuthorEmail), comment.Text, nullString(comment.ParentID),
		comment.Status, nullString(comment.ModeratedBy), nullTime(comment.ModeratedAt),
		time.Now().UTC(), comment.ID)
	if err != nil {
		return 0, 0, 0, err
	}
	return 0, 0, 1, nil
}

// importCommentReaction imports a reaction for a comment
func (i *Importer) importCommentReaction(tx *sql.Tx, commentID string, reaction *models.ReactionExport) (imported, skipped int, err error) {
	// Check if this user already has this reaction on this comment
	var count int
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM reactions 
		WHERE comment_id = ? AND allowed_reaction_id = ? AND user_id IN (
			SELECT id FROM users WHERE id = ?
		)`, commentID, reaction.AllowedReactionID, reaction.UserIdentifier).Scan(&count)
	if err != nil {
		return 0, 0, err
	}

	if count > 0 {
		return 0, 1, nil // Skip duplicate
	}

	// Get or create user
	userID, err := i.getOrCreateUser(tx, reaction.UserIdentifier)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get/create user: %w", err)
	}

	// Generate a new ID for the reaction
	reactionID := fmt.Sprintf("reaction-%d", time.Now().UnixNano())

	// Import the reaction
	_, err = tx.Exec(`
		INSERT INTO reactions (id, comment_id, allowed_reaction_id, user_id, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		reactionID, commentID, reaction.AllowedReactionID, userID, reaction.CreatedAt)
	if err != nil {
		return 0, 0, err
	}

	return 1, 0, nil
}

// importPageReaction imports a reaction for a page
func (i *Importer) importPageReaction(tx *sql.Tx, pageID string, reaction *models.ReactionExport) (imported, skipped int, err error) {
	// Check if this user already has this reaction on this page
	var count int
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM reactions 
		WHERE page_id = ? AND allowed_reaction_id = ? AND user_id IN (
			SELECT id FROM users WHERE id = ?
		)`, pageID, reaction.AllowedReactionID, reaction.UserIdentifier).Scan(&count)
	if err != nil {
		return 0, 0, err
	}

	if count > 0 {
		return 0, 1, nil // Skip duplicate
	}

	// Get or create user
	userID, err := i.getOrCreateUser(tx, reaction.UserIdentifier)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get/create user: %w", err)
	}

	// Generate a new ID for the reaction
	reactionID := fmt.Sprintf("reaction-%d", time.Now().UnixNano())

	// Import the reaction
	_, err = tx.Exec(`
		INSERT INTO reactions (id, page_id, allowed_reaction_id, user_id, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		reactionID, pageID, reaction.AllowedReactionID, userID, reaction.CreatedAt)
	if err != nil {
		return 0, 0, err
	}

	return 1, 0, nil
}

// getOrCreateUser gets an existing user or creates a placeholder
func (i *Importer) getOrCreateUser(tx *sql.Tx, userIdentifier string) (string, error) {
	if userIdentifier == "" {
		return "", fmt.Errorf("user identifier is required")
	}

	// Check if user exists
	var userID string
	err := tx.QueryRow(`SELECT id FROM users WHERE id = ?`, userIdentifier).Scan(&userID)
	if err == nil {
		return userID, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	// User doesn't exist, we can't create it without site_id
	// Return the identifier as-is and let the caller handle it
	return userIdentifier, nil
}

// ValidateImportData validates import data before processing
func ValidateImportData(data *models.ExportData) error {
	if data.Metadata.Version == "" {
		return fmt.Errorf("export version is required")
	}

	if data.Metadata.SiteID == "" {
		return fmt.Errorf("site ID is required in export metadata")
	}

	// Additional validation can be added here
	return nil
}

// nullString converts a string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullTime converts a *time.Time to sql.NullTime
func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// ImportFromCSV imports comments from CSV format
func (i *Importer) ImportFromCSV(r io.Reader, siteID string) (*ImportResult, error) {
	reader := csv.NewReader(r)
	result := &ImportResult{
		Errors: make([]string, 0),
	}

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Validate header
	expectedHeader := []string{"Comment ID", "Page ID", "Page Title", "Author", "Author ID", "Author Email",
		"Text", "Parent ID", "Status", "Created At", "Updated At", "Reaction Count"}
	if len(header) < len(expectedHeader) {
		return nil, fmt.Errorf("invalid CSV header: expected %d columns, got %d", len(expectedHeader), len(header))
	}

	// Start transaction
	tx, err := i.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Read records
	lineNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Line %d: %v", lineNum, err))
			lineNum++
			continue
		}

		if len(record) < len(expectedHeader) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Line %d: expected %d columns, got %d", lineNum, len(expectedHeader), len(record)))
			lineNum++
			continue
		}

		// Parse record
		commentID := record[0]
		pageID := record[1]
		author := record[3]
		authorID := record[4]
		authorEmail := record[5]
		text := record[6]
		parentID := record[7]
		status := record[8]
		createdAtStr := record[9]
		updatedAtStr := record[10]

		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Line %d: invalid created_at format: %v", lineNum, err))
			lineNum++
			continue
		}

		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Line %d: invalid updated_at format: %v", lineNum, err))
			lineNum++
			continue
		}

		// Import comment
		comment := &models.CommentExport{
			ID:          commentID,
			Author:      author,
			AuthorID:    authorID,
			AuthorEmail: authorEmail,
			Text:        text,
			ParentID:    parentID,
			Status:      status,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}

		imported, skipped, updated, err := i.importComment(tx, siteID, pageID, comment)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Line %d: failed to import comment: %v", lineNum, err))
		} else {
			result.CommentsImported += imported
			result.CommentsSkipped += skipped
			result.CommentsUpdated += updated
		}

		lineNum++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}
