package database

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrentUpserts tests multiple goroutines upserting the same project
func TestConcurrentUpserts(t *testing.T) {
	// Create temporary database
	
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test concurrent upserts of the same project
	project := &Project{
		Path:      "~/test-project",
		Name:      "test-project",
		RemoteURL: "https://github.com/user/test.git",
		Branch:    "main",
	}

	// Run 100 concurrent upserts
	var wg sync.WaitGroup
	errors := make(chan error, 100)
	
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Modify project slightly for each goroutine
			testProject := *project
			testProject.Name = fmt.Sprintf("test-project-%d", id)
			
			if err := db.UpsertProject(&testProject); err != nil {
				errors <- err
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Errorf("Upsert failed: %v", err)
	}
	
	// Verify final state
	projects, err := db.GetProjects(false, "")
	if err != nil {
		t.Fatalf("Failed to get projects: %v", err)
	}
	
	// Should have exactly 1 project (last one wins due to ON CONFLICT)
	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}
}

// TestConcurrentReadsAndWrites tests mixed read/write operations
func TestConcurrentReadsAndWrites(t *testing.T) {
	// Create temporary database
	
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add initial data
	initialProject := &Project{
		Path:      "~/initial-project",
		Name:      "initial-project",
		RemoteURL: "https://github.com/user/initial.git",
		Branch:    "main",
	}
	
	if err := db.UpsertProject(initialProject); err != nil {
		t.Fatalf("Failed to insert initial project: %v", err)
	}

	// Run concurrent reads and writes
	var wg sync.WaitGroup
	readErrors := make(chan error, 50)
	writeErrors := make(chan error, 50)
	
	// 50 readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for j := 0; j < 10; j++ {
				projects, err := db.GetProjects(false, "")
				if err != nil {
					readErrors <- err
					return
				}
				
				// Verify we get consistent results
				if len(projects) < 1 {
					readErrors <- fmt.Errorf("expected at least 1 project, got %d", len(projects))
					return
				}
				
				time.Sleep(1 * time.Millisecond) // Small delay
			}
		}()
	}
	
	// 50 writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < 10; j++ {
				project := &Project{
					Path:      fmt.Sprintf("~/project-%d-%d", id, j),
					Name:      fmt.Sprintf("project-%d-%d", id, j),
					RemoteURL: fmt.Sprintf("https://github.com/user/project-%d-%d.git", id, j),
					Branch:    "main",
				}
				
				if err := db.UpsertProject(project); err != nil {
					writeErrors <- err
					return
				}
				
				time.Sleep(1 * time.Millisecond) // Small delay
			}
		}(i)
	}
	
	wg.Wait()
	close(readErrors)
	close(writeErrors)
	
	// Check for errors
	for err := range readErrors {
		t.Errorf("Read failed: %v", err)
	}
	
	for err := range writeErrors {
		t.Errorf("Write failed: %v", err)
	}
	
	// Verify final state
	projects, err := db.GetProjects(false, "")
	if err != nil {
		t.Fatalf("Failed to get final projects: %v", err)
	}
	
	// Should have 1 initial + 500 new projects
	expectedCount := 501
	if len(projects) != expectedCount {
		t.Errorf("Expected %d projects, got %d", expectedCount, len(projects))
	}
}

// TestTransactionIntegrity tests that transactions maintain consistency
func TestTransactionIntegrity(t *testing.T) {
	// Create temporary database
	
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test transaction rollback
	tx, err := db.db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	
	// Insert project in transaction
	project := &Project{
		Path:      "~/transaction-test",
		Name:      "transaction-test",
		RemoteURL: "https://github.com/user/transaction.git",
		Branch:    "main",
	}
	
	query := `
		INSERT INTO projects (path, name, remote_url, branch, last_modified, git_count)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	_, err = tx.Exec(query, project.Path, project.Name, project.RemoteURL, 
		project.Branch, project.LastModified, project.GitCount)
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}
	
	// Rollback transaction
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}
	
	// Verify project was not committed
	projects, err := db.GetProjects(false, "")
	if err != nil {
		t.Fatalf("Failed to get projects: %v", err)
	}
	
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects after rollback, got %d", len(projects))
	}
}

// TestCacheConsistencyUnderLoad tests cache consistency under concurrent load
func TestCacheConsistencyUnderLoad(t *testing.T) {
	// Create temporary database
	
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	cache := NewDatabaseCache(db)
	
	// Add test data
	projects := []*Project{
		{Path: "~/project-1", Name: "project-1", RemoteURL: "https://github.com/user/project-1.git"},
		{Path: "~/project-2", Name: "project-2", RemoteURL: "https://github.com/user/project-2.git"},
		{Path: "~/project-3", Name: "project-3", RemoteURL: "https://github.com/user/project-3.git"},
	}
	
	if err := cache.SetProjects(projects); err != nil {
		t.Fatalf("Failed to set projects: %v", err)
	}
	
	// Run concurrent cache operations
	var wg sync.WaitGroup
	results := make(chan int, 100)
	errors := make(chan error, 100)
	
	// 100 concurrent cache reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			cachedProjects, err := cache.GetProjects()
			if err != nil {
				errors <- err
				return
			}
			
			results <- len(cachedProjects)
		}()
	}
	
	wg.Wait()
	close(results)
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Errorf("Cache operation failed: %v", err)
	}
	
	// Verify all results are consistent
	projectCounts := make(map[int]int)
	for count := range results {
		projectCounts[count]++
	}
	
	// All reads should return the same count
	if len(projectCounts) != 1 {
		t.Errorf("Inconsistent cache results: %v", projectCounts)
	}
	
	// Should be 3 projects
	for count := range projectCounts {
		if count != 3 {
			t.Errorf("Expected 3 projects, got %d", count)
		}
	}
}

// TestDatabaseIntegrityAfterConcurrentOperations tests database integrity
func TestDatabaseIntegrityAfterConcurrentOperations(t *testing.T) {
	// Create temporary database
	
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Run heavy concurrent operations
	var wg sync.WaitGroup
	
	// 10 goroutines each doing 100 operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < 100; j++ {
				project := &Project{
					Path:      fmt.Sprintf("~/project-%d-%d", id, j),
					Name:      fmt.Sprintf("project-%d-%d", id, j),
					RemoteURL: fmt.Sprintf("https://github.com/user/project-%d-%d.git", id, j),
					Branch:    "main",
				}
				
				db.UpsertProject(project)
				
				// Also do some reads
				db.GetProjects(false, "")
			}
		}(i)
	}
	
	wg.Wait()
	
	// Check database integrity
	var integrityCheck string
	err = db.db.QueryRow("PRAGMA integrity_check").Scan(&integrityCheck)
	if err != nil {
		t.Fatalf("Failed to check database integrity: %v", err)
	}
	
	if integrityCheck != "ok" {
		t.Errorf("Database integrity check failed: %s", integrityCheck)
	}
	
	// Verify final state
	projects, err := db.GetProjects(false, "")
	if err != nil {
		t.Fatalf("Failed to get final projects: %v", err)
	}
	
	// Should have 1000 projects (10 * 100)
	expectedCount := 1000
	if len(projects) != expectedCount {
		t.Errorf("Expected %d projects, got %d", expectedCount, len(projects))
	}
}