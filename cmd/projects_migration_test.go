package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shalomb/gum/internal/database"
)

func TestMigrationFixesCacheInconsistency(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	
	// Set up test environment
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()
	
	os.Setenv("XDG_CACHE_HOME", tempDir)
	
	// Create test JSON cache files that simulate the bug
	cacheDir := filepath.Join(tempDir, "gum")
	os.MkdirAll(cacheDir, 0755)
	
	// Create projects.json with many projects
	projectsJSON := `{
		"data": [
			{"Path": "~/test-project-1", "Remote": "https://github.com/user/project1.git", "Branch": "main"},
			{"Path": "~/test-project-2", "Remote": "https://github.com/user/project2.git", "Branch": "main"},
			{"Path": "~/test-project-3", "Remote": "https://github.com/user/project3.git", "Branch": "main"}
		],
		"timestamp": "2025-10-06T14:00:00Z",
		"ttl": 300
	}`
	
	projectsFile := filepath.Join(cacheDir, "projects.json")
	err := os.WriteFile(projectsFile, []byte(projectsJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test projects JSON: %v", err)
	}
	
	// Create project-dirs.json with different discovery (simulating the bug)
	projectDirsJSON := `{
		"data": [
			{"Path": "~/projects", "LastScanned": "2025-10-06T14:00:00Z", "GitCount": 1}
		],
		"timestamp": "2025-10-06T14:00:00Z",
		"ttl": 300
	}`
	
	projectDirsFile := filepath.Join(cacheDir, "project-dirs.json")
	err = os.WriteFile(projectDirsFile, []byte(projectDirsJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test project-dirs JSON: %v", err)
	}
	
	// Initialize database
	dbPath := filepath.Join(cacheDir, "gum.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()
	
	// Run migration
	migrator := database.NewMigrator(db)
	err = migrator.MigrateFromJSON(cacheDir)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
	
	// Test 1: Verify projects were migrated
	projects, err := db.GetProjects(false, "")
	if err != nil {
		t.Fatalf("Failed to get projects: %v", err)
	}
	
	if len(projects) != 3 {
		t.Errorf("Expected 3 projects after migration, got %d", len(projects))
	}
	
	// Test 2: Verify project directories were migrated
	dirs, err := db.GetProjectDirs()
	if err != nil {
		t.Fatalf("Failed to get project directories: %v", err)
	}
	
	if len(dirs) != 1 {
		t.Errorf("Expected 1 project directory after migration, got %d", len(dirs))
	}
	
	// Test 3: Verify cache consistency
	cache := database.NewDatabaseCache(db)
	
	// First call should populate cache
	projects1, err := cache.GetProjects()
	if err != nil {
		t.Fatalf("Failed to get projects from cache: %v", err)
	}
	
	if len(projects1) != 3 {
		t.Errorf("Expected 3 projects from cache, got %d", len(projects1))
	}
	
	// Second call should use cache (consistent)
	projects2, err := cache.GetProjects()
	if err != nil {
		t.Fatalf("Failed to get projects from cache: %v", err)
	}
	
	if len(projects2) != 3 {
		t.Errorf("Expected 3 projects from cache, got %d", len(projects2))
	}
	
	// Test 4: Verify cache hit
	if !cache.IsCacheHit("projects") {
		t.Error("Expected cache hit on second call")
	}
	
	// Test 5: Verify JSON files were backed up
	backupDir := filepath.Join(cacheDir, "backup")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		t.Error("Backup directory was not created")
	}
	
	if _, err := os.Stat(filepath.Join(backupDir, "projects.json")); os.IsNotExist(err) {
		t.Error("Projects JSON was not backed up")
	}
	
	if _, err := os.Stat(filepath.Join(backupDir, "project-dirs.json")); os.IsNotExist(err) {
		t.Error("Project-dirs JSON was not backed up")
	}
	
	// Test 6: Verify original JSON files were removed
	if _, err := os.Stat(projectsFile); err == nil {
		t.Error("Original projects.json should have been removed")
	}
	
	if _, err := os.Stat(projectDirsFile); err == nil {
		t.Error("Original project-dirs.json should have been removed")
	}
}

func TestConcurrentAccessAfterMigration(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	
	// Set up test environment
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()
	
	os.Setenv("XDG_CACHE_HOME", tempDir)
	
	// Initialize database
	cacheDir := filepath.Join(tempDir, "gum")
	os.MkdirAll(cacheDir, 0755)
	
	dbPath := filepath.Join(cacheDir, "gum.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()
	
	// Add test data
	project := &database.Project{
		Path:      "~/test-project",
		Name:      "test-project",
		RemoteURL: "https://github.com/user/test-project.git",
		Branch:    "main",
	}
	
	err = db.UpsertProject(project)
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}
	
	// Test concurrent access
	cache := database.NewDatabaseCache(db)
	done := make(chan bool, 10)
	
	// Start 5 readers
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_, err := cache.GetProjects()
				if err != nil {
					t.Errorf("Read failed: %v", err)
				}
			}
			done <- true
		}()
	}
	
	// Start 5 writers
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				testProject := &database.Project{
					Path:      fmt.Sprintf("~/test-project-%d-%d", id, j),
					Name:      fmt.Sprintf("test-project-%d-%d", id, j),
					RemoteURL: fmt.Sprintf("https://github.com/user/test-project-%d-%d.git", id, j),
					Branch:    "main",
				}
				err := db.UpsertProject(testProject)
				if err != nil {
					t.Errorf("Write failed: %v", err)
				}
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify database integrity
	projects, err := db.GetProjects(false, "")
	if err != nil {
		t.Fatalf("Failed to get projects: %v", err)
	}
	
	expectedCount := 51 // 1 original + 50 new projects
	if len(projects) != expectedCount {
		t.Errorf("Expected %d projects, got %d", expectedCount, len(projects))
	}
}

func TestMigrationRollback(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	
	// Set up test environment
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()
	
	os.Setenv("XDG_CACHE_HOME", tempDir)
	
	// Create test JSON cache files
	cacheDir := filepath.Join(tempDir, "gum")
	os.MkdirAll(cacheDir, 0755)
	
	projectsJSON := `{
		"data": [
			{"Path": "~/test-project", "Remote": "https://github.com/user/test-project.git", "Branch": "main"}
		],
		"timestamp": "2025-10-06T14:00:00Z",
		"ttl": 300
	}`
	
	projectsFile := filepath.Join(cacheDir, "projects.json")
	err := os.WriteFile(projectsFile, []byte(projectsJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test projects JSON: %v", err)
	}
	
	// Initialize database
	dbPath := filepath.Join(cacheDir, "gum.db")
	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()
	
	// Run migration
	migrator := database.NewMigrator(db)
	err = migrator.MigrateFromJSON(cacheDir)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
	
	// Verify migration worked
	projects, err := db.GetProjects(false, "")
	if err != nil {
		t.Fatalf("Failed to get projects: %v", err)
	}
	
	if len(projects) != 1 {
		t.Errorf("Expected 1 project after migration, got %d", len(projects))
	}
	
	// Run rollback
	err = migrator.RollbackMigration(cacheDir)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	
	// Verify JSON files were restored
	if _, err := os.Stat(projectsFile); os.IsNotExist(err) {
		t.Error("Projects JSON was not restored after rollback")
	}
	
	// Verify database was cleared
	projects, err = db.GetProjects(false, "")
	if err != nil {
		t.Fatalf("Failed to get projects after rollback: %v", err)
	}
	
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects after rollback, got %d", len(projects))
	}
}