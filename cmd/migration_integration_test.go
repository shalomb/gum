package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shalomb/gum/internal/database"
)

// TestMigrationIntegration tests the complete migration process
func TestMigrationIntegration(t *testing.T) {
	// Create temporary test environment
	tempDir := t.TempDir()
	
	// Set up environment
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()
	
	os.Setenv("XDG_CACHE_HOME", tempDir)
	
	// Create test cache directory
	cacheDir := filepath.Join(tempDir, "gum")
	os.MkdirAll(cacheDir, 0755)
	
	// Test 1: Create JSON cache files that simulate the bug
	t.Run("CreateBugScenario", func(t *testing.T) {
		// Create projects.json with many projects
		projectsData := []map[string]string{
			{"Path": "~/projects/project-1", "Remote": "https://github.com/user/project-1.git", "Branch": "main"},
			{"Path": "~/projects/project-2", "Remote": "https://github.com/user/project-2.git", "Branch": "main"},
			{"Path": "~/projects/project-3", "Remote": "https://github.com/user/project-3.git", "Branch": "main"},
			{"Path": "~/projects/project-4", "Remote": "https://github.com/user/project-4.git", "Branch": "main"},
			{"Path": "~/projects/project-5", "Remote": "https://github.com/user/project-5.git", "Branch": "main"},
		}
		
		projectsJSON := map[string]interface{}{
			"data":      projectsData,
			"timestamp": time.Now().Format(time.RFC3339),
			"ttl":       300,
		}
		
		projectsBytes, err := json.Marshal(projectsJSON)
		if err != nil {
			t.Fatalf("Failed to marshal projects JSON: %v", err)
		}
		
		projectsFile := filepath.Join(cacheDir, "projects.json")
		if err := os.WriteFile(projectsFile, projectsBytes, 0644); err != nil {
			t.Fatalf("Failed to write projects.json: %v", err)
		}
		
		// Create project-dirs.json with different discovery (simulating the bug)
		projectDirsData := []map[string]interface{}{
			{"Path": "~/projects", "LastScanned": time.Now().Format(time.RFC3339), "GitCount": 2}, // Only 2 projects!
		}
		
		projectDirsJSON := map[string]interface{}{
			"data":      projectDirsData,
			"timestamp": time.Now().Format(time.RFC3339),
			"ttl":       300,
		}
		
		projectDirsBytes, err := json.Marshal(projectDirsJSON)
		if err != nil {
			t.Fatalf("Failed to marshal project-dirs JSON: %v", err)
		}
		
		projectDirsFile := filepath.Join(cacheDir, "project-dirs.json")
		if err := os.WriteFile(projectDirsFile, projectDirsBytes, 0644); err != nil {
			t.Fatalf("Failed to write project-dirs.json: %v", err)
		}
		
		t.Logf("Created bug scenario: 5 projects in projects.json, 2 projects in project-dirs.json")
	})
	
	// Test 2: Run migration
	t.Run("RunMigration", func(t *testing.T) {
		dbPath := filepath.Join(cacheDir, "gum.db")
		db, err := database.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()
		
		// Run migration
		migrator := database.NewMigrator(db)
		if err := migrator.MigrateFromJSON(cacheDir); err != nil {
			t.Fatalf("Migration failed: %v", err)
		}
		
		t.Log("Migration completed successfully")
	})
	
	// Test 3: Verify migration results
	t.Run("VerifyMigrationResults", func(t *testing.T) {
		dbPath := filepath.Join(cacheDir, "gum.db")
		db, err := database.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()
		
		// Check projects were migrated
		projects, err := db.GetProjects(false, "")
		if err != nil {
			t.Fatalf("Failed to get projects: %v", err)
		}
		
		if len(projects) != 5 {
			t.Errorf("Expected 5 projects after migration, got %d", len(projects))
		}
		
		// Check project directories were migrated
		dirs, err := db.GetProjectDirs()
		if err != nil {
			t.Fatalf("Failed to get project directories: %v", err)
		}
		
		if len(dirs) != 1 {
			t.Errorf("Expected 1 project directory after migration, got %d", len(dirs))
		}
		
		t.Logf("Migration verified: %d projects, %d directories", len(projects), len(dirs))
	})
	
	// Test 4: Test cache consistency
	t.Run("TestCacheConsistency", func(t *testing.T) {
		dbPath := filepath.Join(cacheDir, "gum.db")
		db, err := database.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()
		
		cache := database.NewDatabaseCache(db)
		
		// Test multiple cache calls
		projects1, err := cache.GetProjects()
		if err != nil {
			t.Fatalf("Failed to get projects from cache: %v", err)
		}
		
		projects2, err := cache.GetProjects()
		if err != nil {
			t.Fatalf("Failed to get projects from cache: %v", err)
		}
		
		projects3, err := cache.GetProjects()
		if err != nil {
			t.Fatalf("Failed to get projects from cache: %v", err)
		}
		
		// Verify consistency
		if len(projects1) != len(projects2) || len(projects2) != len(projects3) {
			t.Errorf("Cache inconsistency detected: %d, %d, %d", len(projects1), len(projects2), len(projects3))
		}
		
		// Verify cache hit
		if !cache.IsCacheHit("projects") {
			t.Error("Expected cache hit on subsequent calls")
		}
		
		t.Logf("Cache consistency verified: %d projects consistently returned", len(projects1))
	})
	
	// Test 5: Test refresh functionality
	t.Run("TestRefreshFunctionality", func(t *testing.T) {
		dbPath := filepath.Join(cacheDir, "gum.db")
		db, err := database.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()
		
		cache := database.NewDatabaseCache(db)
		
		// Clear cache
		if err := cache.ClearCache("projects"); err != nil {
			t.Fatalf("Failed to clear cache: %v", err)
		}
		
		// Verify cache miss
		if cache.IsCacheHit("projects") {
			t.Error("Expected cache miss after clearing")
		}
		
		// Test refresh
		projects := []*database.Project{
			{Path: "~/test-project", Name: "test-project", RemoteURL: "https://github.com/user/test.git"},
		}
		
		if err := cache.SetProjects(projects); err != nil {
			t.Fatalf("Failed to set projects: %v", err)
		}
		
		// Verify cache hit after refresh
		if !cache.IsCacheHit("projects") {
			t.Error("Expected cache hit after refresh")
		}
		
		t.Log("Refresh functionality verified")
	})
	
	// Test 6: Test rollback functionality
	t.Run("TestRollbackFunctionality", func(t *testing.T) {
		dbPath := filepath.Join(cacheDir, "gum.db")
		db, err := database.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()
		
		migrator := database.NewMigrator(db)
		
		// Test rollback
		if err := migrator.RollbackMigration(cacheDir); err != nil {
			t.Fatalf("Rollback failed: %v", err)
		}
		
		// Verify JSON files were restored
		projectsFile := filepath.Join(cacheDir, "projects.json")
		projectDirsFile := filepath.Join(cacheDir, "project-dirs.json")
		
		if _, err := os.Stat(projectsFile); os.IsNotExist(err) {
			t.Error("projects.json was not restored after rollback")
		}
		
		if _, err := os.Stat(projectDirsFile); os.IsNotExist(err) {
			t.Error("project-dirs.json was not restored after rollback")
		}
		
		// Verify database was cleared
		projects, err := db.GetProjects(false, "")
		if err != nil {
			t.Fatalf("Failed to get projects after rollback: %v", err)
		}
		
		if len(projects) != 0 {
			t.Errorf("Expected 0 projects after rollback, got %d", len(projects))
		}
		
		t.Log("Rollback functionality verified")
	})
	
	// Test 7: Test concurrent access
	t.Run("TestConcurrentAccess", func(t *testing.T) {
		dbPath := filepath.Join(cacheDir, "gum.db")
		db, err := database.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()
		
		cache := database.NewDatabaseCache(db)
		
		// Add test data
		projects := []*database.Project{
			{Path: "~/concurrent-test", Name: "concurrent-test", RemoteURL: "https://github.com/user/concurrent.git"},
		}
		
		if err := cache.SetProjects(projects); err != nil {
			t.Fatalf("Failed to set projects: %v", err)
		}
		
		// Test concurrent reads
		done := make(chan bool, 10)
		
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 10; j++ {
					_, err := cache.GetProjects()
					if err != nil {
						t.Errorf("Concurrent read failed: %v", err)
					}
				}
				done <- true
			}()
		}
		
		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
		
		t.Log("Concurrent access test completed")
	})
	
	// Test 8: Test performance
	t.Run("TestPerformance", func(t *testing.T) {
		dbPath := filepath.Join(cacheDir, "gum.db")
		db, err := database.New(dbPath)
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()
		
		cache := database.NewDatabaseCache(db)
		
		// Create large dataset
		var projects []*database.Project
		for i := 0; i < 1000; i++ {
			projects = append(projects, &database.Project{
				Path:      fmt.Sprintf("~/project-%d", i),
				Name:      fmt.Sprintf("project-%d", i),
				RemoteURL: fmt.Sprintf("https://github.com/user/project-%d.git", i),
			})
		}
		
		// Time the operation
		start := time.Now()
		if err := cache.SetProjects(projects); err != nil {
			t.Fatalf("Failed to set projects: %v", err)
		}
		setDuration := time.Since(start)
		
		start = time.Now()
		retrievedProjects, err := cache.GetProjects()
		if err != nil {
			t.Fatalf("Failed to get projects: %v", err)
		}
		getDuration := time.Since(start)
		
		if len(retrievedProjects) != 1000 {
			t.Errorf("Expected 1000 projects, got %d", len(retrievedProjects))
		}
		
		t.Logf("Performance test: Set 1000 projects in %v, Retrieved in %v", setDuration, getDuration)
		
		// Performance should be reasonable
		if setDuration > 5*time.Second {
			t.Errorf("Set operation too slow: %v", setDuration)
		}
		
		if getDuration > 100*time.Millisecond {
			t.Errorf("Get operation too slow: %v", getDuration)
		}
	})
}

// TestBugReproduction tests the exact bug scenario
func TestBugReproduction(t *testing.T) {
	// This test reproduces the exact bug described in the bug report
	tempDir := t.TempDir()
	
	// Set up environment
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()
	
	os.Setenv("XDG_CACHE_HOME", tempDir)
	
	cacheDir := filepath.Join(tempDir, "gum")
	os.MkdirAll(cacheDir, 0755)
	
	// Simulate the bug: projects.json has many projects, project-dirs.json has few
	projectsData := make([]map[string]string, 100)
	for i := 0; i < 100; i++ {
		projectsData[i] = map[string]string{
			"Path":   fmt.Sprintf("~/projects/project-%d", i),
			"Remote": fmt.Sprintf("https://github.com/user/project-%d.git", i),
			"Branch": "main",
		}
	}
	
	projectsJSON := map[string]interface{}{
		"data":      projectsData,
		"timestamp": time.Now().Format(time.RFC3339),
		"ttl":       300,
	}
	
	projectsBytes, _ := json.Marshal(projectsJSON)
	os.WriteFile(filepath.Join(cacheDir, "projects.json"), projectsBytes, 0644)
	
	// project-dirs.json only has 3 projects (simulating gum update interference)
	projectDirsData := []map[string]interface{}{
		{"Path": "~/projects", "LastScanned": time.Now().Format(time.RFC3339), "GitCount": 3},
	}
	
	projectDirsJSON := map[string]interface{}{
		"data":      projectDirsData,
		"timestamp": time.Now().Format(time.RFC3339),
		"ttl":       300,
	}
	
	projectDirsBytes, _ := json.Marshal(projectDirsJSON)
	os.WriteFile(filepath.Join(cacheDir, "project-dirs.json"), projectDirsBytes, 0644)
	
	// Now test the migration fixes this inconsistency
	dbPath := filepath.Join(cacheDir, "gum.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()
	
	// Run migration
	migrator := database.NewMigrator(db)
	if err := migrator.MigrateFromJSON(cacheDir); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
	
	// Verify the bug is fixed
	projects, err := db.GetProjects(false, "")
	if err != nil {
		t.Fatalf("Failed to get projects: %v", err)
	}
	
	// Should have all 100 projects, not just 3
	if len(projects) != 100 {
		t.Errorf("Bug not fixed: expected 100 projects, got %d", len(projects))
	}
	
	t.Logf("Bug reproduction test passed: %d projects migrated correctly", len(projects))
}