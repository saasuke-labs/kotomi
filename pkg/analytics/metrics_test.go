package analytics

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Create temporary database using t.TempDir() for cross-platform compatibility
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/analytics_test.db"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create test schema
	schema := `
	CREATE TABLE IF NOT EXISTS sites (
		id TEXT PRIMARY KEY,
		owner_id TEXT NOT NULL,
		name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS pages (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		path TEXT NOT NULL,
		title TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT NOT NULL,
		site_id TEXT NOT NULL,
		name TEXT NOT NULL,
		email TEXT,
		first_seen TIMESTAMP NOT NULL,
		last_seen TIMESTAMP NOT NULL,
		PRIMARY KEY (site_id, id)
	);

	CREATE TABLE IF NOT EXISTS comments (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		page_id TEXT NOT NULL,
		author TEXT NOT NULL,
		author_id TEXT NOT NULL,
		author_email TEXT,
		text TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		moderated_by TEXT,
		moderated_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS allowed_reactions (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		name TEXT NOT NULL,
		emoji TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS reactions (
		id TEXT PRIMARY KEY,
		page_id TEXT,
		comment_id TEXT,
		allowed_reaction_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	cleanup := func() {
		db.Close()
		// TempDir is automatically cleaned up by Go testing framework
	}

	return db, cleanup
}

func insertTestData(t *testing.T, db *sql.DB) {
	siteID := "test-site-1"
	pageID := "test-page-1"
	userID := "test-user-1"

	// Insert site
	_, err := db.Exec("INSERT INTO sites (id, owner_id, name) VALUES (?, ?, ?)",
		siteID, "owner-1", "Test Site")
	if err != nil {
		t.Fatalf("Failed to insert test site: %v", err)
	}

	// Insert page
	_, err = db.Exec("INSERT INTO pages (id, site_id, path, title) VALUES (?, ?, ?, ?)",
		pageID, siteID, "/test", "Test Page")
	if err != nil {
		t.Fatalf("Failed to insert test page: %v", err)
	}

	// Insert user
	now := time.Now()
	_, err = db.Exec("INSERT INTO users (id, site_id, name, email, first_seen, last_seen) VALUES (?, ?, ?, ?, ?, ?)",
		userID, siteID, "Test User", "test@example.com", now, now)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert comments with different statuses
	comments := []struct {
		id          string
		status      string
		text        string
		createdAt   time.Time
		moderatedAt *time.Time
	}{
		{"comment-1", "approved", "Test comment 1", now.AddDate(0, 0, -5), &now},
		{"comment-2", "pending", "Test comment 2", now.AddDate(0, 0, -3), nil},
		{"comment-3", "rejected", "Spam comment", now.AddDate(0, 0, -2), &now},
		{"comment-4", "approved", "Test comment 4", now.AddDate(0, 0, -1), &now},
		{"comment-5", "approved", "Test comment 5", now, &now},
	}

	for _, c := range comments {
		query := `INSERT INTO comments (id, site_id, page_id, author, author_id, text, status, created_at, moderated_at) 
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = db.Exec(query, c.id, siteID, pageID, "Test User", userID, c.text, c.status, c.createdAt, c.moderatedAt)
		if err != nil {
			t.Fatalf("Failed to insert test comment: %v", err)
		}
	}

	// Insert allowed reactions
	reactionID := "reaction-1"
	_, err = db.Exec("INSERT INTO allowed_reactions (id, site_id, name, emoji) VALUES (?, ?, ?, ?)",
		reactionID, siteID, "thumbs_up", "👍")
	if err != nil {
		t.Fatalf("Failed to insert allowed reaction: %v", err)
	}

	// Insert reactions
	reactions := []struct {
		id        string
		commentID string
		createdAt time.Time
	}{
		{"react-1", "comment-1", now.AddDate(0, 0, -4)},
		{"react-2", "comment-1", now.AddDate(0, 0, -3)},
		{"react-3", "comment-4", now.AddDate(0, 0, -1)},
	}

	for _, r := range reactions {
		_, err = db.Exec("INSERT INTO reactions (id, comment_id, allowed_reaction_id, user_id, created_at) VALUES (?, ?, ?, ?, ?)",
			r.id, r.commentID, reactionID, userID, r.createdAt)
		if err != nil {
			t.Fatalf("Failed to insert reaction: %v", err)
		}
	}
}

func TestGetDefaultDateRange(t *testing.T) {
	dateRange := GetDefaultDateRange()
	
	now := time.Now()
	expectedFrom := now.AddDate(0, 0, -30)
	
	// Check that From is approximately 30 days ago
	diff := expectedFrom.Sub(dateRange.From)
	if diff > 1*time.Hour || diff < -1*time.Hour {
		t.Errorf("Expected From to be ~30 days ago, got %v", dateRange.From)
	}
	
	// Check that To is approximately now
	diff = now.Sub(dateRange.To)
	if diff > 1*time.Hour || diff < -1*time.Hour {
		t.Errorf("Expected To to be ~now, got %v", dateRange.To)
	}
}

func TestParseDateRange(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		to        string
		wantError bool
	}{
		{"Valid range", "2024-01-01", "2024-01-31", false},
		{"Empty dates", "", "", false},
		{"Invalid from", "invalid", "2024-01-31", true},
		{"Invalid to", "2024-01-01", "invalid", true},
		{"Only from", "2024-01-01", "", false},
		{"Only to", "", "2024-01-31", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dateRange, err := ParseDateRange(tt.from, tt.to)
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if dateRange.From.IsZero() || dateRange.To.IsZero() {
					t.Error("Expected valid date range")
				}
			}
		})
	}
}

func TestGetCommentMetrics(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	insertTestData(t, db)
	
	store := NewStore(db)
	dateRange := DateRange{
		From: time.Now().AddDate(0, 0, -10),
		To:   time.Now().AddDate(0, 0, 1),
	}
	
	metrics, err := store.GetCommentMetrics("test-site-1", dateRange)
	if err != nil {
		t.Fatalf("Failed to get comment metrics: %v", err)
	}
	
	if metrics.Total != 5 {
		t.Errorf("Expected 5 total comments, got %d", metrics.Total)
	}
	
	if metrics.Approved != 3 {
		t.Errorf("Expected 3 approved comments, got %d", metrics.Approved)
	}
	
	if metrics.Pending != 1 {
		t.Errorf("Expected 1 pending comment, got %d", metrics.Pending)
	}
	
	if metrics.Rejected != 1 {
		t.Errorf("Expected 1 rejected comment, got %d", metrics.Rejected)
	}
	
	expectedApprovalRate := 60.0 // 3/5 * 100
	if metrics.ApprovalRate < expectedApprovalRate-1 || metrics.ApprovalRate > expectedApprovalRate+1 {
		t.Errorf("Expected approval rate ~%.1f%%, got %.1f%%", expectedApprovalRate, metrics.ApprovalRate)
	}
}

func TestGetUserMetrics(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	insertTestData(t, db)
	
	store := NewStore(db)
	dateRange := DateRange{
		From: time.Now().AddDate(0, 0, -10),
		To:   time.Now().AddDate(0, 0, 1),
	}
	
	metrics, err := store.GetUserMetrics("test-site-1", dateRange)
	if err != nil {
		t.Fatalf("Failed to get user metrics: %v", err)
	}
	
	if metrics.TotalUsers != 1 {
		t.Errorf("Expected 1 total user, got %d", metrics.TotalUsers)
	}
	
	if len(metrics.TopContributors) == 0 {
		t.Error("Expected at least one top contributor")
	}
	
	if len(metrics.TopContributors) > 0 {
		if metrics.TopContributors[0].Name != "Test User" {
			t.Errorf("Expected top contributor name 'Test User', got '%s'", metrics.TopContributors[0].Name)
		}
		if metrics.TopContributors[0].CommentCount != 5 {
			t.Errorf("Expected 5 comments from top contributor, got %d", metrics.TopContributors[0].CommentCount)
		}
	}
}

func TestGetReactionMetrics(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	insertTestData(t, db)
	
	store := NewStore(db)
	dateRange := DateRange{
		From: time.Now().AddDate(0, 0, -10),
		To:   time.Now().AddDate(0, 0, 1),
	}
	
	metrics, err := store.GetReactionMetrics("test-site-1", dateRange)
	if err != nil {
		t.Fatalf("Failed to get reaction metrics: %v", err)
	}
	
	if metrics.Total != 3 {
		t.Errorf("Expected 3 total reactions, got %d", metrics.Total)
	}
	
	if len(metrics.ByType) == 0 {
		t.Error("Expected at least one reaction type")
	}
	
	if len(metrics.ByType) > 0 {
		if metrics.ByType[0].Name != "thumbs_up" {
			t.Errorf("Expected reaction name 'thumbs_up', got '%s'", metrics.ByType[0].Name)
		}
		if metrics.ByType[0].Count != 3 {
			t.Errorf("Expected 3 reactions of type thumbs_up, got %d", metrics.ByType[0].Count)
		}
	}
}

func TestGetModerationMetrics(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	insertTestData(t, db)
	
	store := NewStore(db)
	dateRange := DateRange{
		From: time.Now().AddDate(0, 0, -10),
		To:   time.Now().AddDate(0, 0, 1),
	}
	
	metrics, err := store.GetModerationMetrics("test-site-1", dateRange)
	if err != nil {
		t.Fatalf("Failed to get moderation metrics: %v", err)
	}
	
	// 4 comments have been moderated (approved or rejected)
	if metrics.TotalModerated != 4 {
		t.Errorf("Expected 4 total moderated, got %d", metrics.TotalModerated)
	}
	
	// Should have some average moderation time
	if metrics.AverageModerationSec == 0 {
		t.Error("Expected non-zero average moderation time")
	}
}

func TestGetCommentsTrend(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	insertTestData(t, db)
	
	store := NewStore(db)
	dateRange := DateRange{
		From: time.Now().AddDate(0, 0, -10),
		To:   time.Now().AddDate(0, 0, 1),
	}
	
	trend, err := store.GetCommentsTrend("test-site-1", dateRange)
	if err != nil {
		t.Fatalf("Failed to get comments trend: %v", err)
	}
	
	if len(trend.Labels) == 0 {
		t.Error("Expected non-empty labels")
	}
	
	if len(trend.Values) == 0 {
		t.Error("Expected non-empty values")
	}
	
	if len(trend.Labels) != len(trend.Values) {
		t.Errorf("Expected labels and values to have same length, got %d labels and %d values",
			len(trend.Labels), len(trend.Values))
	}
}

func TestGetAnalyticsDashboard(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	
	insertTestData(t, db)
	
	store := NewStore(db)
	dateRange := DateRange{
		From: time.Now().AddDate(0, 0, -10),
		To:   time.Now().AddDate(0, 0, 1),
	}
	
	dashboard, err := store.GetAnalyticsDashboard("test-site-1", dateRange)
	if err != nil {
		t.Fatalf("Failed to get analytics dashboard: %v", err)
	}
	
	if dashboard.SiteID != "test-site-1" {
		t.Errorf("Expected site ID 'test-site-1', got '%s'", dashboard.SiteID)
	}
	
	if dashboard.Comments.Total == 0 {
		t.Error("Expected non-zero comment metrics")
	}
	
	if dashboard.Users.TotalUsers == 0 {
		t.Error("Expected non-zero user metrics")
	}
	
	if dashboard.Reactions.Total == 0 {
		t.Error("Expected non-zero reaction metrics")
	}
	
	if len(dashboard.CommentsTrend.Labels) == 0 {
		t.Error("Expected non-empty comments trend")
	}
}
