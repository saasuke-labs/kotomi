package importpkg

import (
	"context"
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
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

func createTestSite(t *testing.T, store *comments.SQLiteStore) (siteID, pageID string) {
	db := store.GetDB()

	// Create admin user
	_, err := db.Exec(`INSERT INTO admin_users (id, email, name, auth0_sub) VALUES (?, ?, ?, ?)`,
		"admin-1", "admin@example.com", "Admin User", "auth0|123")
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create site
	siteStore := models.NewSiteStore(db)
	site, err := siteStore.Create(context.Background(), "admin-1", "Test Site", "test.example.com", "Test site for import")
	if err != nil {
		t.Fatalf("Failed to create site: %v", err)
	}
	siteID = site.ID

	// Create page
	pageStore := models.NewPageStore(db)
	page, err := pageStore.Create(context.Background(), siteID, "/test-page", "Test Page")
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}
	pageID = page.ID

	// Create allowed reaction
	reactionStore := models.NewAllowedReactionStore(db)
	_, err = reactionStore.Create(context.Background(), siteID, "thumbs_up", "👍", "both")
	if err != nil {
		t.Fatalf("Failed to create allowed reaction: %v", err)
	}

	return siteID, pageID
}

func createTestExportData(siteID, pageID string) *models.ExportData {
	now := time.Now().UTC()
	return &models.ExportData{
		Metadata: models.ExportMetadata{
			Version:        "1.0",
			ExportedAt:     now,
			SiteID:         siteID,
			SiteName:       "Test Site",
			TotalPages:     1,
			TotalComments:  1,
			TotalReactions: 0,
		},
		Site: models.Site{
			ID:          siteID,
			OwnerID:     "admin-1",
			Name:        "Test Site",
			Domain:      "test.example.com",
			Description: "Test site",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Pages: []models.PageExport{
			{
				Page: models.Page{
					ID:        pageID,
					SiteID:    siteID,
					Path:      "/test-page",
					Title:     "Test Page",
					CreatedAt: now,
					UpdatedAt: now,
				},
				Comments: []models.CommentExport{
					{
						ID:          "comment-1",
						Author:      "Test User",
						AuthorID:    "user-1",
						AuthorEmail: "user@example.com",
						Text:        "This is a test comment",
						ParentID:    "",
						Status:      "approved",
						CreatedAt:   now,
						UpdatedAt:   now,
						Reactions:   []models.ReactionExport{},
					},
				},
				PageReactions:    []models.ReactionExport{},
				AllowedReactions: []models.AllowedReaction{},
			},
		},
	}
}

func TestImporter_ImportFromJSON_Skip(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, pageID := createTestSite(t, store)
	exportData := createTestExportData(siteID, pageID)

	// First import
	importer := NewImporter(store.GetDB(), StrategySkip)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(exportData); err != nil {
		t.Fatalf("Failed to encode export data: %v", err)
	}

	result, err := importer.ImportFromJSON(&buf, siteID)
	if err != nil {
		t.Fatalf("ImportFromJSON failed: %v", err)
	}

	if result.CommentsImported != 1 {
		t.Errorf("Expected 1 comment imported, got %d", result.CommentsImported)
	}
	if result.CommentsSkipped != 0 {
		t.Errorf("Expected 0 comments skipped, got %d", result.CommentsSkipped)
	}

	// Second import with same data (should skip)
	buf.Reset()
	if err := json.NewEncoder(&buf).Encode(exportData); err != nil {
		t.Fatalf("Failed to encode export data: %v", err)
	}

	result, err = importer.ImportFromJSON(&buf, siteID)
	if err != nil {
		t.Fatalf("Second ImportFromJSON failed: %v", err)
	}

	if result.CommentsImported != 0 {
		t.Errorf("Expected 0 comments imported on second import, got %d", result.CommentsImported)
	}
	if result.CommentsSkipped != 1 {
		t.Errorf("Expected 1 comment skipped on second import, got %d", result.CommentsSkipped)
	}
}

func TestImporter_ImportFromJSON_Update(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, pageID := createTestSite(t, store)
	exportData := createTestExportData(siteID, pageID)

	// First import
	importer := NewImporter(store.GetDB(), StrategyUpdate)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(exportData); err != nil {
		t.Fatalf("Failed to encode export data: %v", err)
	}

	result, err := importer.ImportFromJSON(&buf, siteID)
	if err != nil {
		t.Fatalf("ImportFromJSON failed: %v", err)
	}

	if result.CommentsImported != 1 {
		t.Errorf("Expected 1 comment imported, got %d", result.CommentsImported)
	}

	// Modify the comment text
	exportData.Pages[0].Comments[0].Text = "Updated comment text"

	// Second import with updated data
	buf.Reset()
	if err := json.NewEncoder(&buf).Encode(exportData); err != nil {
		t.Fatalf("Failed to encode export data: %v", err)
	}

	result, err = importer.ImportFromJSON(&buf, siteID)
	if err != nil {
		t.Fatalf("Second ImportFromJSON failed: %v", err)
	}

	if result.CommentsUpdated != 1 {
		t.Errorf("Expected 1 comment updated, got %d", result.CommentsUpdated)
	}

	// Verify the comment was updated
	db := store.GetDB()
	var text string
	err = db.QueryRow(`SELECT text FROM comments WHERE id = ?`, "comment-1").Scan(&text)
	if err != nil {
		t.Fatalf("Failed to query comment: %v", err)
	}
	if text != "Updated comment text" {
		t.Errorf("Expected comment text 'Updated comment text', got '%s'", text)
	}
}

func TestImporter_ImportFromJSON_WrongSite(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, pageID := createTestSite(t, store)
	exportData := createTestExportData(siteID, pageID)

	importer := NewImporter(store.GetDB(), StrategySkip)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(exportData); err != nil {
		t.Fatalf("Failed to encode export data: %v", err)
	}

	// Try to import to a different site
	_, err := importer.ImportFromJSON(&buf, "different-site-id")
	if err == nil {
		t.Error("Expected error when importing to wrong site, got nil")
	}
	if !strings.Contains(err.Error(), "but importing to site") {
		t.Errorf("Expected error about site mismatch, got: %v", err)
	}
}

func TestImporter_ImportFromJSON_NewPage(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, _ := createTestSite(t, store)

	// Create export data with a new page
	exportData := createTestExportData(siteID, "new-page-id")
	exportData.Pages[0].Page.Path = "/new-page"

	importer := NewImporter(store.GetDB(), StrategySkip)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(exportData); err != nil {
		t.Fatalf("Failed to encode export data: %v", err)
	}

	result, err := importer.ImportFromJSON(&buf, siteID)
	if err != nil {
		t.Fatalf("ImportFromJSON failed: %v", err)
	}

	if result.PagesCreated != 1 {
		t.Errorf("Expected 1 page created, got %d", result.PagesCreated)
	}
	if result.CommentsImported != 1 {
		t.Errorf("Expected 1 comment imported, got %d", result.CommentsImported)
	}
}

func TestImporter_ImportFromCSV(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, pageID := createTestSite(t, store)

	// Create CSV data
	now := time.Now().UTC()
	csvData := `Comment ID,Page ID,Page Title,Author,Author ID,Author Email,Text,Parent ID,Status,Created At,Updated At,Reaction Count
csv-comment-1,` + pageID + `,Test Page,CSV User,user-1,csv@example.com,CSV comment text,,approved,` + now.Format(time.RFC3339) + `,` + now.Format(time.RFC3339) + `,0`

	importer := NewImporter(store.GetDB(), StrategySkip)
	result, err := importer.ImportFromCSV(strings.NewReader(csvData), siteID)
	if err != nil {
		t.Fatalf("ImportFromCSV failed: %v", err)
	}

	if result.CommentsImported != 1 {
		t.Errorf("Expected 1 comment imported, got %d", result.CommentsImported)
	}

	// Verify the comment was imported
	db := store.GetDB()
	var text string
	err = db.QueryRow(`SELECT text FROM comments WHERE id = ?`, "csv-comment-1").Scan(&text)
	if err != nil {
		t.Fatalf("Failed to query comment: %v", err)
	}
	if text != "CSV comment text" {
		t.Errorf("Expected comment text 'CSV comment text', got '%s'", text)
	}
}

func TestImporter_ImportFromCSV_InvalidHeader(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, _ := createTestSite(t, store)

	csvData := `Invalid,Header
value1,value2`

	importer := NewImporter(store.GetDB(), StrategySkip)
	_, err := importer.ImportFromCSV(strings.NewReader(csvData), siteID)
	if err == nil {
		t.Error("Expected error for invalid CSV header, got nil")
	}
}

func TestImporter_ImportFromCSV_InvalidDate(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, pageID := createTestSite(t, store)

	csvData := `Comment ID,Page ID,Page Title,Author,Author ID,Author Email,Text,Parent ID,Status,Created At,Updated At,Reaction Count
csv-comment-1,` + pageID + `,Test Page,CSV User,user-1,csv@example.com,CSV comment text,,approved,invalid-date,invalid-date,0`

	importer := NewImporter(store.GetDB(), StrategySkip)
	result, err := importer.ImportFromCSV(strings.NewReader(csvData), siteID)
	if err != nil {
		t.Fatalf("ImportFromCSV failed: %v", err)
	}

	// Should have errors but not fail entirely
	if len(result.Errors) == 0 {
		t.Error("Expected errors for invalid date format")
	}
	if result.CommentsImported != 0 {
		t.Errorf("Expected 0 comments imported with invalid date, got %d", result.CommentsImported)
	}
}

func TestValidateImportData(t *testing.T) {
	tests := []struct {
		name    string
		data    *models.ExportData
		wantErr bool
	}{
		{
			name: "valid data",
			data: &models.ExportData{
				Metadata: models.ExportMetadata{
					Version: "1.0",
					SiteID:  "site-1",
				},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			data: &models.ExportData{
				Metadata: models.ExportMetadata{
					SiteID: "site-1",
				},
			},
			wantErr: true,
		},
		{
			name: "missing site ID",
			data: &models.ExportData{
				Metadata: models.ExportMetadata{
					Version: "1.0",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImportData(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImportData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImporter_ImportFromJSON_EmptyData(t *testing.T) {
	store := createTestDB(t)
	defer store.Close()

	siteID, _ := createTestSite(t, store)

	// Create export data with no pages
	exportData := &models.ExportData{
		Metadata: models.ExportMetadata{
			Version:        "1.0",
			ExportedAt:     time.Now().UTC(),
			SiteID:         siteID,
			SiteName:       "Test Site",
			TotalPages:     0,
			TotalComments:  0,
			TotalReactions: 0,
		},
		Site: models.Site{
			ID:      siteID,
			OwnerID: "admin-1",
			Name:    "Test Site",
		},
		Pages: []models.PageExport{},
	}

	importer := NewImporter(store.GetDB(), StrategySkip)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(exportData); err != nil {
		t.Fatalf("Failed to encode export data: %v", err)
	}

	result, err := importer.ImportFromJSON(&buf, siteID)
	if err != nil {
		t.Fatalf("ImportFromJSON failed: %v", err)
	}

	if result.CommentsImported != 0 {
		t.Errorf("Expected 0 comments imported, got %d", result.CommentsImported)
	}
	if result.PagesCreated != 0 {
		t.Errorf("Expected 0 pages created, got %d", result.PagesCreated)
	}
}
