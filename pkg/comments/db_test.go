package comments

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewSitePagesIndex(t *testing.T) {
	index := NewSitePagesIndex()
	if index == nil {
		t.Fatal("NewSitePagesIndex returned nil")
	}
	if index.data == nil {
		t.Error("data map not initialized")
	}
}

func TestAddPageComment_SingleComment(t *testing.T) {
	index := NewSitePagesIndex()
	comment := Comment{
		ID:        "1",
		Author:    "John",
		Text:      "Test comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	index.AddPageComment("site1", "page1", comment)

	comments := index.GetPageComments("site1", "page1")
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	if comments[0].ID != "1" {
		t.Errorf("expected comment ID '1', got '%s'", comments[0].ID)
	}
	if comments[0].Author != "John" {
		t.Errorf("expected author 'John', got '%s'", comments[0].Author)
	}
	if comments[0].Text != "Test comment" {
		t.Errorf("expected text 'Test comment', got '%s'", comments[0].Text)
	}
}

func TestAddPageComment_MultipleComments(t *testing.T) {
	index := NewSitePagesIndex()

	comments := []Comment{
		{ID: "1", Author: "John", Text: "First", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "2", Author: "Jane", Text: "Second", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "3", Author: "Bob", Text: "Third", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, c := range comments {
		index.AddPageComment("site1", "page1", c)
	}

	retrieved := index.GetPageComments("site1", "page1")
	if len(retrieved) != 3 {
		t.Fatalf("expected 3 comments, got %d", len(retrieved))
	}

	for i, c := range retrieved {
		if c.ID != comments[i].ID {
			t.Errorf("comment %d: expected ID '%s', got '%s'", i, comments[i].ID, c.ID)
		}
	}
}

func TestGetPageComments_EmptySite(t *testing.T) {
	index := NewSitePagesIndex()
	comments := index.GetPageComments("nonexistent", "page1")

	if comments == nil {
		t.Error("expected non-nil slice")
	}
	if len(comments) != 0 {
		t.Errorf("expected empty slice, got %d comments", len(comments))
	}
}

func TestGetPageComments_EmptyPage(t *testing.T) {
	index := NewSitePagesIndex()
	
	// Add comment to site1/page1
	index.AddPageComment("site1", "page1", Comment{
		ID:        "1",
		Author:    "John",
		Text:      "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	// Query different page
	comments := index.GetPageComments("site1", "page2")

	if comments == nil {
		t.Error("expected non-nil slice")
	}
	if len(comments) != 0 {
		t.Errorf("expected empty slice, got %d comments", len(comments))
	}
}

func TestAddPageComment_MultipleSitesAndPages(t *testing.T) {
	index := NewSitePagesIndex()

	testCases := []struct {
		site    string
		page    string
		comment Comment
	}{
		{"site1", "page1", Comment{ID: "1", Author: "John", Text: "S1P1", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
		{"site1", "page2", Comment{ID: "2", Author: "Jane", Text: "S1P2", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
		{"site2", "page1", Comment{ID: "3", Author: "Bob", Text: "S2P1", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
		{"site2", "page2", Comment{ID: "4", Author: "Alice", Text: "S2P2", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
	}

	for _, tc := range testCases {
		index.AddPageComment(tc.site, tc.page, tc.comment)
	}

	for _, tc := range testCases {
		comments := index.GetPageComments(tc.site, tc.page)
		if len(comments) != 1 {
			t.Errorf("%s/%s: expected 1 comment, got %d", tc.site, tc.page, len(comments))
			continue
		}
		if comments[0].Text != tc.comment.Text {
			t.Errorf("%s/%s: expected text '%s', got '%s'", tc.site, tc.page, tc.comment.Text, comments[0].Text)
		}
	}
}

func TestAddPageComment_WithParentID(t *testing.T) {
	index := NewSitePagesIndex()

	parent := Comment{
		ID:        "1",
		Author:    "John",
		Text:      "Parent comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	index.AddPageComment("site1", "page1", parent)

	reply := Comment{
		ID:        "2",
		Author:    "Jane",
		Text:      "Reply comment",
		ParentID:  "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	index.AddPageComment("site1", "page1", reply)

	comments := index.GetPageComments("site1", "page1")
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	// Find the reply
	var foundReply *Comment
	for _, c := range comments {
		if c.ID == "2" {
			foundReply = &c
			break
		}
	}

	if foundReply == nil {
		t.Fatal("reply comment not found")
	}
	if foundReply.ParentID != "1" {
		t.Errorf("expected ParentID '1', got '%s'", foundReply.ParentID)
	}
}

func TestSitePagesIndex_ConcurrentAccess(t *testing.T) {
	index := NewSitePagesIndex()
	numGoroutines := 10
	commentsPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Write concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < commentsPerGoroutine; j++ {
				comment := Comment{
					ID:        fmt.Sprintf("g%d-c%d", goroutineID, j),
					Author:    fmt.Sprintf("Author%d", goroutineID),
					Text:      fmt.Sprintf("Text from goroutine %d, comment %d", goroutineID, j),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				index.AddPageComment("site1", "page1", comment)
			}
		}(i)
	}

	wg.Wait()

	// Verify all comments were added
	comments := index.GetPageComments("site1", "page1")
	expectedCount := numGoroutines * commentsPerGoroutine
	if len(comments) != expectedCount {
		t.Errorf("expected %d comments, got %d", expectedCount, len(comments))
	}
}

func TestSitePagesIndex_ConcurrentReadWrite(t *testing.T) {
	index := NewSitePagesIndex()
	var wg sync.WaitGroup

	// Add initial comments
	for i := 0; i < 5; i++ {
		index.AddPageComment("site1", "page1", Comment{
			ID:        fmt.Sprintf("%d", i),
			Author:    "Initial",
			Text:      "Initial comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	// Concurrent reads and writes
	wg.Add(2)

	// Writer goroutine
	go func() {
		defer wg.Done()
		for i := 5; i < 10; i++ {
			index.AddPageComment("site1", "page1", Comment{
				ID:        fmt.Sprintf("%d", i),
				Author:    "Writer",
				Text:      "New comment",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Reader goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			comments := index.GetPageComments("site1", "page1")
			if comments == nil {
				t.Error("GetPageComments returned nil")
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()

	wg.Wait()
}
