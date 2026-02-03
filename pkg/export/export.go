package export

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/saasuke-labs/kotomi/pkg/models"
)

// Exporter handles data export operations
type Exporter struct {
	db *sql.DB
}

// NewExporter creates a new Exporter
func NewExporter(db *sql.DB) *Exporter {
	return &Exporter{db: db}
}

// ExportToJSON exports site data to JSON format
func (e *Exporter) ExportToJSON(siteID string) (*models.ExportData, error) {
	// Get site information
	siteStore := models.NewSiteStore(e.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get site: %w", err)
	}

	// Get pages for the site
	pageStore := models.NewPageStore(e.db)
	pages, err := pageStore.GetBySite(siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}

	// Get allowed reactions for the site
	reactionStore := models.NewAllowedReactionStore(e.db)
	allowedReactions, err := reactionStore.GetBySite(siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowed reactions: %w", err)
	}

	exportData := &models.ExportData{
		Metadata: models.ExportMetadata{
			Version:    "1.0",
			ExportedAt: time.Now().UTC(),
			SiteID:     site.ID,
			SiteName:   site.Name,
		},
		Site:  *site,
		Pages: make([]models.PageExport, 0, len(pages)),
	}

	totalComments := 0
	totalReactions := 0

	// For each page, get comments and reactions
	for _, page := range pages {
		comments, err := e.getCommentsForPage(siteID, page.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get comments for page %s: %w", page.ID, err)
		}

		pageReactions, err := e.getPageReactions(page.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get page reactions for page %s: %w", page.ID, err)
		}

		totalComments += len(comments)
		totalReactions += len(pageReactions)

		for _, comment := range comments {
			totalReactions += len(comment.Reactions)
		}

		pageExport := models.PageExport{
			Page:             page,
			Comments:         comments,
			PageReactions:    pageReactions,
			AllowedReactions: allowedReactions,
		}

		exportData.Pages = append(exportData.Pages, pageExport)
	}

	exportData.Metadata.TotalPages = len(pages)
	exportData.Metadata.TotalComments = totalComments
	exportData.Metadata.TotalReactions = totalReactions

	return exportData, nil
}

// getCommentsForPage retrieves all comments for a page with their reactions
func (e *Exporter) getCommentsForPage(siteID, pageID string) ([]models.CommentExport, error) {
	query := `
		SELECT id, author, author_id, author_email, text, parent_id, status, 
		       moderated_by, moderated_at, created_at, updated_at
		FROM comments
		WHERE site_id = ? AND page_id = ?
		ORDER BY created_at ASC
	`

	rows, err := e.db.Query(query, siteID, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]models.CommentExport, 0)
	for rows.Next() {
		var c models.CommentExport
		var parentID, moderatedBy sql.NullString
		var authorEmail sql.NullString
		var moderatedAt sql.NullTime

		err := rows.Scan(&c.ID, &c.Author, &c.AuthorID, &authorEmail, &c.Text, &parentID,
			&c.Status, &moderatedBy, &moderatedAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			c.ParentID = parentID.String
		}
		if moderatedBy.Valid {
			c.ModeratedBy = moderatedBy.String
		}
		if moderatedAt.Valid {
			t := moderatedAt.Time
			c.ModeratedAt = &t
		}
		if authorEmail.Valid {
			c.AuthorEmail = authorEmail.String
		}

		// Get reactions for this comment
		reactions, err := e.getCommentReactions(c.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get reactions for comment %s: %w", c.ID, err)
		}
		c.Reactions = reactions

		comments = append(comments, c)
	}

	return comments, rows.Err()
}

// getCommentReactions retrieves all reactions for a comment
func (e *Exporter) getCommentReactions(commentID string) ([]models.ReactionExport, error) {
	query := `
		SELECT r.allowed_reaction_id, ar.name, ar.emoji, u.id as user_identifier, r.created_at
		FROM reactions r
		JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.comment_id = ?
		ORDER BY r.created_at ASC
	`

	rows, err := e.db.Query(query, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reactions := make([]models.ReactionExport, 0)
	for rows.Next() {
		var r models.ReactionExport
		var userIdentifier sql.NullString

		err := rows.Scan(&r.AllowedReactionID, &r.ReactionName, &r.ReactionEmoji,
			&userIdentifier, &r.CreatedAt)
		if err != nil {
			return nil, err
		}

		if userIdentifier.Valid {
			r.UserIdentifier = userIdentifier.String
		}

		reactions = append(reactions, r)
	}

	return reactions, rows.Err()
}

// getPageReactions retrieves all reactions for a page
func (e *Exporter) getPageReactions(pageID string) ([]models.ReactionExport, error) {
	query := `
		SELECT r.allowed_reaction_id, ar.name, ar.emoji, u.id as user_identifier, r.created_at
		FROM reactions r
		JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.page_id = ?
		ORDER BY r.created_at ASC
	`

	rows, err := e.db.Query(query, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reactions := make([]models.ReactionExport, 0)
	for rows.Next() {
		var r models.ReactionExport
		var userIdentifier sql.NullString

		err := rows.Scan(&r.AllowedReactionID, &r.ReactionName, &r.ReactionEmoji,
			&userIdentifier, &r.CreatedAt)
		if err != nil {
			return nil, err
		}

		if userIdentifier.Valid {
			r.UserIdentifier = userIdentifier.String
		}

		reactions = append(reactions, r)
	}

	return reactions, rows.Err()
}

// WriteJSON writes export data as JSON to a writer
func (e *Exporter) WriteJSON(w io.Writer, data *models.ExportData) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// ExportToCSV exports comments to CSV format
func (e *Exporter) ExportToCSV(w io.Writer, siteID string) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write CSV header
	header := []string{
		"Comment ID", "Page ID", "Page Title", "Author", "Author ID", "Author Email",
		"Text", "Parent ID", "Status", "Created At", "Updated At",
		"Reaction Count",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Get pages for the site
	pageStore := models.NewPageStore(e.db)
	pages, err := pageStore.GetBySite(siteID)
	if err != nil {
		return fmt.Errorf("failed to get pages: %w", err)
	}

	// For each page, get comments
	for _, page := range pages {
		comments, err := e.getCommentsForPage(siteID, page.ID)
		if err != nil {
			return fmt.Errorf("failed to get comments for page %s: %w", page.ID, err)
		}

		for _, comment := range comments {
			record := []string{
				comment.ID,
				page.ID,
				page.Title,
				comment.Author,
				comment.AuthorID,
				comment.AuthorEmail,
				comment.Text,
				comment.ParentID,
				comment.Status,
				comment.CreatedAt.Format(time.RFC3339),
				comment.UpdatedAt.Format(time.RFC3339),
				fmt.Sprintf("%d", len(comment.Reactions)),
			}
			if err := writer.Write(record); err != nil {
				return err
			}
		}
	}

	return nil
}

// ExportReactionsToCSV exports reactions to CSV format
func (e *Exporter) ExportReactionsToCSV(w io.Writer, siteID string) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write CSV header
	header := []string{
		"Reaction ID", "Target Type", "Target ID", "Reaction Name",
		"Reaction Emoji", "User Identifier", "Created At",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Get all reactions for the site through pages and comments
	query := `
		SELECT r.id, 
		       CASE WHEN r.page_id IS NOT NULL THEN 'page' ELSE 'comment' END as target_type,
		       COALESCE(r.page_id, r.comment_id) as target_id,
		       ar.name, ar.emoji, u.id as user_identifier, r.created_at
		FROM reactions r
		JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		LEFT JOIN users u ON r.user_id = u.id
		WHERE ar.site_id = ?
		ORDER BY r.created_at ASC
	`

	rows, err := e.db.Query(query, siteID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, targetType, targetID, name, emoji string
		var userIdentifier sql.NullString
		var createdAt time.Time

		err := rows.Scan(&id, &targetType, &targetID, &name, &emoji, &userIdentifier, &createdAt)
		if err != nil {
			return err
		}

		userID := ""
		if userIdentifier.Valid {
			userID = userIdentifier.String
		}

		record := []string{
			id,
			targetType,
			targetID,
			name,
			emoji,
			userID,
			createdAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return rows.Err()
}

// GetExportFilename generates a filename for the export
func GetExportFilename(siteName, format string) string {
	// Sanitize site name
	sanitized := strings.ReplaceAll(siteName, " ", "_")
	sanitized = strings.ToLower(sanitized)
	timestamp := time.Now().UTC().Format("20060102_150405")
	return fmt.Sprintf("kotomi_export_%s_%s.%s", sanitized, timestamp, format)
}
