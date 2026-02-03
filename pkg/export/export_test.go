package export

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

func createTestDB(t *testing.T) *comments.SQLiteStore {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := comments.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return store
}

func createTestData(t *testing.T, store *comments.SQLiteStore) (siteID, pageID, commentID string) {
	db := store.GetDB()

	// Create admin user
	_, err := db.Exec(`INSERT INTO admin_users (id, email, name, auth0_sub) VALUES (?, ?, ?, ?)`,
		"admin-1", "admin@example.com", "Admin User", "auth0|123")
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create site
	siteStore := models.NewSiteStore(db)
	site, err := siteStore.Create("admin-1", "Test Site", "test.example.com", "Test site for export")
	if err != nil {
		t.Fatalf("Failed to create site: %v", err)
	}
	siteID = site.ID

	// Create page
	pageStore := models.NewPageStore(db)
	page, err := pageStore.Create(siteID, "/test-page", "Test Page")
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}
	pageID = page.ID

	// Create user
	userStore := models.NewUserStore(db)
	user := &models.User{
		ID:     "user-1",
		SiteID: siteID,
		Name:   "Test User",
		Email:  "user@example.com",
	}
	if err := userStore.CreateOrUpdate(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create comment
	_, err = db.Exec(`
		INSERT INTO comments (id, site_id, page_id, author, author_id, author_email, text, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"comment-1", siteID, pageID, "Test User", user.ID, "user@example.com",
		"This is a test comment", "approved", time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}
	commentID = "comment-1"

	// Create allowed reaction
	reactionStore := models.NewAllowedReactionStore(db)
	allowedReaction, err := reactionStore.Create(siteID, "thumbs_up", "👍", "both")
	if err != nil {
		t.Fatalf("Failed to create allowed reaction: %v", err)
	}

	// Create reaction on comment
	_, err = db.Exec(`
		INSERT INTO reactions (id, comment_id, allowed_reaction_id, user_id, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		"reaction-1", commentID, allowedReaction.ID, user.ID, time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to create reaction: %v", err)
	}

	return siteID, pageID, commentID
}

func TestExporter_ExportToJSON(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, pageID, commentID := createTestData(t, store)

	exporter := NewExporter(store.GetDB())
	exportData, err := exporter.ExportToJSON(siteID)
	if err != nil {
		t.Fatalf("ExportToJSON failed: %v", err)
	}

	// Verify metadata
	if exportData.Metadata.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", exportData.Metadata.Version)
	}
	if exportData.Metadata.SiteID != siteID {
		t.Errorf("Expected site ID %s, got %s", siteID, exportData.Metadata.SiteID)
	}
	if exportData.Metadata.SiteName != "Test Site" {
		t.Errorf("Expected site name 'Test Site', got %s", exportData.Metadata.SiteName)
	}
	if exportData.Metadata.TotalPages != 1 {
		t.Errorf("Expected 1 page, got %d", exportData.Metadata.TotalPages)
	}
	if exportData.Metadata.TotalComments != 1 {
		t.Errorf("Expected 1 comment, got %d", exportData.Metadata.TotalComments)
	}
	if exportData.Metadata.TotalReactions != 1 {
		t.Errorf("Expected 1 reaction, got %d", exportData.Metadata.TotalReactions)
	}

	// Verify site
	if exportData.Site.ID != siteID {
		t.Errorf("Expected site ID %s, got %s", siteID, exportData.Site.ID)
	}

	// Verify pages
	if len(exportData.Pages) != 1 {
		t.Fatalf("Expected 1 page, got %d", len(exportData.Pages))
	}
	pageExport := exportData.Pages[0]
	if pageExport.Page.ID != pageID {
		t.Errorf("Expected page ID %s, got %s", pageID, pageExport.Page.ID)
	}

	// Verify comments
	if len(pageExport.Comments) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(pageExport.Comments))
	}
	comment := pageExport.Comments[0]
	if comment.ID != commentID {
		t.Errorf("Expected comment ID %s, got %s", commentID, comment.ID)
	}
	if comment.Text != "This is a test comment" {
		t.Errorf("Expected comment text 'This is a test comment', got %s", comment.Text)
	}

	// Verify reactions
	if len(comment.Reactions) != 1 {
		t.Fatalf("Expected 1 reaction, got %d", len(comment.Reactions))
	}
	reaction := comment.Reactions[0]
	if reaction.ReactionName != "thumbs_up" {
		t.Errorf("Expected reaction name 'thumbs_up', got %s", reaction.ReactionName)
	}
	if reaction.ReactionEmoji != "👍" {
		t.Errorf("Expected reaction emoji '👍', got %s", reaction.ReactionEmoji)
	}
}

func TestExporter_WriteJSON(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, _, _ := createTestData(t, store)

	exporter := NewExporter(store.GetDB())
	exportData, err := exporter.ExportToJSON(siteID)
	if err != nil {
		t.Fatalf("ExportToJSON failed: %v", err)
	}

	var buf bytes.Buffer
	if err := exporter.WriteJSON(&buf, exportData); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Verify JSON is valid
	var decoded models.ExportData
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if decoded.Metadata.SiteID != siteID {
		t.Errorf("Expected site ID %s, got %s", siteID, decoded.Metadata.SiteID)
	}
}

func TestExporter_ExportToCSV(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, _, _ := createTestData(t, store)

	exporter := NewExporter(store.GetDB())
	var buf bytes.Buffer
	if err := exporter.ExportToCSV(&buf, siteID); err != nil {
		t.Fatalf("ExportToCSV failed: %v", err)
	}

	csvContent := buf.String()
	// Verify CSV header
	if !bytes.Contains([]byte(csvContent), []byte("Comment ID")) {
		t.Error("CSV should contain 'Comment ID' header")
	}
	if !bytes.Contains([]byte(csvContent), []byte("Page Title")) {
		t.Error("CSV should contain 'Page Title' header")
	}

	// Verify comment data is present
	if !bytes.Contains([]byte(csvContent), []byte("This is a test comment")) {
		t.Error("CSV should contain comment text")
	}
}

func TestExporter_ExportReactionsToCSV(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, _, _ := createTestData(t, store)

	exporter := NewExporter(store.GetDB())
	var buf bytes.Buffer
	if err := exporter.ExportReactionsToCSV(&buf, siteID); err != nil {
		t.Fatalf("ExportReactionsToCSV failed: %v", err)
	}

	csvContent := buf.String()
	// Verify CSV header
	if !bytes.Contains([]byte(csvContent), []byte("Reaction ID")) {
		t.Error("CSV should contain 'Reaction ID' header")
	}
	if !bytes.Contains([]byte(csvContent), []byte("Target Type")) {
		t.Error("CSV should contain 'Target Type' header")
	}

	// Verify reaction data is present
	if !bytes.Contains([]byte(csvContent), []byte("thumbs_up")) {
		t.Error("CSV should contain reaction name")
	}
}

func TestGetExportFilename(t *testing.T) {
	tests := []struct {
		siteName string
		format   string
		want     string
	}{
		{"My Site", "json", "kotomi_export_my_site_"},
		{"Test-Site", "csv", "kotomi_export_test-site_"},
		{"Site With Spaces", "json", "kotomi_export_site_with_spaces_"},
	}

	for _, tt := range tests {
		t.Run(tt.siteName, func(t *testing.T) {
			filename := GetExportFilename(tt.siteName, tt.format)
			if !bytes.Contains([]byte(filename), []byte(tt.want)) {
				t.Errorf("GetExportFilename() = %v, want to contain %v", filename, tt.want)
			}
			if !bytes.Contains([]byte(filename), []byte("."+tt.format)) {
				t.Errorf("GetExportFilename() should end with .%s extension", tt.format)
			}
		})
	}
}

func TestExporter_ExportToJSON_NoData(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	db := store.GetDB()

	// Create only a site, no pages or comments
	_, err := db.Exec(`INSERT INTO admin_users (id, email, name, auth0_sub) VALUES (?, ?, ?, ?)`,
		"admin-1", "admin@example.com", "Admin User", "auth0|123")
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	siteStore := models.NewSiteStore(db)
	site, err := siteStore.Create("admin-1", "Empty Site", "empty.example.com", "")
	if err != nil {
		t.Fatalf("Failed to create site: %v", err)
	}

	exporter := NewExporter(db)
	exportData, err := exporter.ExportToJSON(site.ID)
	if err != nil {
		t.Fatalf("ExportToJSON failed: %v", err)
	}

	if exportData.Metadata.TotalPages != 0 {
		t.Errorf("Expected 0 pages, got %d", exportData.Metadata.TotalPages)
	}
	if exportData.Metadata.TotalComments != 0 {
		t.Errorf("Expected 0 comments, got %d", exportData.Metadata.TotalComments)
	}
	if len(exportData.Pages) != 0 {
		t.Errorf("Expected empty pages array, got %d pages", len(exportData.Pages))
	}
}

func TestExporter_ExportToJSON_InvalidSite(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	exporter := NewExporter(store.GetDB())
	_, err := exporter.ExportToJSON("non-existent-site")
	if err == nil {
		t.Error("Expected error for non-existent site, got nil")
	}
}
