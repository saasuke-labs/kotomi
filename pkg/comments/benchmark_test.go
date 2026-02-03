package comments

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// BenchmarkConcurrentReads measures performance of concurrent read operations
func BenchmarkConcurrentReads(b *testing.B) {
	dbPath := "/tmp/benchmark_reads_" + uuid.New().String() + ".db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-wal")
	defer os.Remove(dbPath + "-shm")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Seed with some data
	siteID := "benchmark-site"
	pageID := "benchmark-page"
	for i := 0; i < 100; i++ {
		comment := Comment{
			ID:       fmt.Sprintf("comment-%d", i),
			Author:   fmt.Sprintf("User %d", i),
			AuthorID: fmt.Sprintf("user-%d", i),
			Text:     fmt.Sprintf("This is test comment number %d", i),
			Status:   "approved",
		}
		if err := store.AddPageComment(context.Background(), siteID, pageID, comment); err != nil {
			b.Fatalf("Failed to add comment: %v", err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := store.GetPageComments(context.Background(), siteID, pageID)
			if err != nil {
				b.Fatalf("Failed to get comments: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentWrites measures performance of concurrent write operations
func BenchmarkConcurrentWrites(b *testing.B) {
	dbPath := "/tmp/benchmark_writes_" + uuid.New().String() + ".db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-wal")
	defer os.Remove(dbPath + "-shm")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	siteID := "benchmark-site"
	pageID := "benchmark-page"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			comment := Comment{
				ID:       uuid.New().String(),
				Author:   fmt.Sprintf("User %d", i),
				AuthorID: fmt.Sprintf("user-%d", i),
				Text:     fmt.Sprintf("This is benchmark comment number %d", i),
				Status:   "approved",
			}
			if err := store.AddPageComment(context.Background(), siteID, pageID, comment); err != nil {
				b.Fatalf("Failed to add comment: %v", err)
			}
			i++
		}
	})
}

// BenchmarkMixedReadWrite measures performance of mixed read/write workload
func BenchmarkMixedReadWrite(b *testing.B) {
	dbPath := "/tmp/benchmark_mixed_" + uuid.New().String() + ".db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-wal")
	defer os.Remove(dbPath + "-shm")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	siteID := "benchmark-site"
	pageID := "benchmark-page"

	// Seed with initial data
	for i := 0; i < 50; i++ {
		comment := Comment{
			ID:       fmt.Sprintf("comment-%d", i),
			Author:   fmt.Sprintf("User %d", i),
			AuthorID: fmt.Sprintf("user-%d", i),
			Text:     fmt.Sprintf("This is initial comment number %d", i),
			Status:   "approved",
		}
		if err := store.AddPageComment(context.Background(), siteID, pageID, comment); err != nil {
			b.Fatalf("Failed to add comment: %v", err)
		}
	}

	b.ResetTimer()
	
	// 80% reads, 20% writes (typical web workload)
	var wg sync.WaitGroup
	
	// Readers
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < b.N/8; j++ {
				_, err := store.GetPageComments(context.Background(), siteID, pageID)
				if err != nil {
					b.Errorf("Failed to get comments: %v", err)
					return
				}
			}
		}()
	}

	// Writers
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < b.N/2; j++ {
				comment := Comment{
					ID:       uuid.New().String(),
					Author:   fmt.Sprintf("Writer %d", writerID),
					AuthorID: fmt.Sprintf("writer-%d", writerID),
					Text:     fmt.Sprintf("Benchmark write %d-%d", writerID, j),
					Status:   "approved",
				}
				if err := store.AddPageComment(context.Background(), siteID, pageID, comment); err != nil {
					b.Errorf("Failed to add comment: %v", err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

// BenchmarkUpdateOperations measures performance of update operations
func BenchmarkUpdateOperations(b *testing.B) {
	dbPath := "/tmp/benchmark_updates_" + uuid.New().String() + ".db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-wal")
	defer os.Remove(dbPath + "-shm")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Seed with data to update
	siteID := "benchmark-site"
	pageID := "benchmark-page"
	commentIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		commentID := fmt.Sprintf("comment-%d", i)
		commentIDs[i] = commentID
		comment := Comment{
			ID:       commentID,
			Author:   fmt.Sprintf("User %d", i),
			AuthorID: fmt.Sprintf("user-%d", i),
			Text:     fmt.Sprintf("Original text %d", i),
			Status:   "pending",
		}
		if err := store.AddPageComment(context.Background(), siteID, pageID, comment); err != nil {
			b.Fatalf("Failed to add comment: %v", err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			commentID := commentIDs[i%len(commentIDs)]
			if err := store.UpdateCommentStatus(context.Background(), commentID, "approved", "moderator-1"); err != nil {
				b.Fatalf("Failed to update comment: %v", err)
			}
			i++
		}
	})
}

// BenchmarkConnectionPoolEfficiency measures efficiency of connection pool
func BenchmarkConnectionPoolEfficiency(b *testing.B) {
	dbPath := "/tmp/benchmark_pool_" + uuid.New().String() + ".db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-wal")
	defer os.Remove(dbPath + "-shm")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	siteID := "benchmark-site"
	pageID := "benchmark-page"

	// Seed with initial data
	for i := 0; i < 10; i++ {
		comment := Comment{
			ID:       fmt.Sprintf("comment-%d", i),
			Author:   fmt.Sprintf("User %d", i),
			AuthorID: fmt.Sprintf("user-%d", i),
			Text:     fmt.Sprintf("Test comment %d", i),
			Status:   "approved",
		}
		if err := store.AddPageComment(context.Background(), siteID, pageID, comment); err != nil {
			b.Fatalf("Failed to add comment: %v", err)
		}
	}

	b.ResetTimer()

	// Simulate burst of concurrent requests (more than max pool size)
	var wg sync.WaitGroup
	concurrentRequests := 50 // More than MaxOpenConns (25)

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < b.N/concurrentRequests; j++ {
				_, err := store.GetPageComments(context.Background(), siteID, pageID)
				if err != nil {
					b.Errorf("Failed to get comments: %v", err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrencyStress is a stress test (not a benchmark) to verify robustness
func TestConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	dbPath := "/tmp/stress_test_" + uuid.New().String() + ".db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-wal")
	defer os.Remove(dbPath + "-shm")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	siteID := "stress-site"
	pageID := "stress-page"

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// 50 concurrent writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				comment := Comment{
					ID:       uuid.New().String(),
					Author:   fmt.Sprintf("Writer %d", writerID),
					AuthorID: fmt.Sprintf("writer-%d", writerID),
					Text:     fmt.Sprintf("Stress test comment %d-%d", writerID, j),
					Status:   "approved",
				}
				if err := store.AddPageComment(context.Background(), siteID, pageID, comment); err != nil {
					errors <- fmt.Errorf("writer %d failed: %w", writerID, err)
					return
				}
				time.Sleep(1 * time.Millisecond) // Small delay to prevent overwhelming
			}
		}(i)
	}

	// 100 concurrent readers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_, err := store.GetPageComments(context.Background(), siteID, pageID)
				if err != nil {
					errors <- fmt.Errorf("reader %d failed: %w", readerID, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	var errorCount int
	for err := range errors {
		t.Errorf("Stress test error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("Stress test failed with %d errors", errorCount)
	}

	// Verify final data integrity
	comments, err := store.GetPageComments(context.Background(), siteID, pageID)
	if err != nil {
		t.Fatalf("Failed to get final comments: %v", err)
	}

	expectedComments := 50 * 20 // 50 writers * 20 comments each
	if len(comments) != expectedComments {
		t.Errorf("Expected %d comments, got %d", expectedComments, len(comments))
	}

	t.Logf("Stress test passed: %d comments written by 50 concurrent writers, read by 100 concurrent readers", len(comments))
}
