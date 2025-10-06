package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMigrationFromJSON(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test JSON cache files
	cacheDir := filepath.Join(tempDir, "cache")
	os.MkdirAll(cacheDir, 0755)

	// Test projects JSON
	projectsJSON := `{
		"data": [
			{
				"Path": "~/test-project",
				"Remote": "https://github.com/user/test-project.git",
				"Branch": "main"
			},
			{
				"Path": "~/another-project", 
				"Remote": "https://github.com/user/another-project.git",
				"Branch": "develop"
			}
		],
		"timestamp": "2025-10-06T14:00:00Z",
		"ttl": 300
	}`

	projectsFile := filepath.Join(cacheDir, "projects.json")
	err = os.WriteFile(projectsFile, []byte(projectsJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test projects JSON: %v", err)
	}

	// Test project-dirs JSON
	projectDirsJSON := `{
		"data": [
			{
				"Path": "~/projects",
				"LastScanned": "2025-10-06T14:00:00Z",
				"GitCount": 2
			}
		],
		"timestamp": "2025-10-06T14:00:00Z",
		"ttl": 300
	}`

	projectDirsFile := filepath.Join(cacheDir, "project-dirs.json")
	err = os.WriteFile(projectDirsFile, []byte(projectDirsJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test project-dirs JSON: %v", err)
	}

	// Run migration
	migrator := NewMigrator(db)
	err = migrator.MigrateFromJSON(cacheDir)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify projects were migrated
	projects, err := db.GetProjects(false, "")
	if err != nil {
		t.Fatalf("Failed to get projects: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}

	// Verify project directories were migrated
	dirs, err := db.GetProjectDirs()
	if err != nil {
		t.Fatalf("Failed to get project directories: %v", err)
	}

	if len(dirs) != 1 {
		t.Errorf("Expected 1 project directory, got %d", len(dirs))
	}

	// Verify JSON files were backed up
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
}

func TestLinkGitHubRepositories(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add test projects
	project1 := &Project{
		Path:      "~/test-project",
		Name:      "test-project",
		RemoteURL: "https://github.com/user/test-project.git",
		Branch:    "main",
	}
	project2 := &Project{
		Path:      "~/another-project",
		Name:      "another-project", 
		RemoteURL: "https://github.com/user/another-project.git",
		Branch:    "develop",
	}

	err = db.UpsertProject(project1)
	if err != nil {
		t.Fatalf("Failed to insert project1: %v", err)
	}

	err = db.UpsertProject(project2)
	if err != nil {
		t.Fatalf("Failed to insert project2: %v", err)
	}

	// Add test GitHub repositories
	githubRepo1 := &GitHubRepo{
		Name:     "test-project",
		FullName: "user/test-project",
		CloneURL: "https://github.com/user/test-project.git",
		Language: "Go",
		StarCount: 100,
	}
	githubRepo2 := &GitHubRepo{
		Name:     "another-project",
		FullName: "user/another-project", 
		CloneURL: "https://github.com/user/another-project.git",
		Language: "Python",
		StarCount: 50,
	}

	err = db.UpsertGitHubRepo(githubRepo1)
	if err != nil {
		t.Fatalf("Failed to insert github repo1: %v", err)
	}

	err = db.UpsertGitHubRepo(githubRepo2)
	if err != nil {
		t.Fatalf("Failed to insert github repo2: %v", err)
	}

	// Link projects to GitHub repositories
	migrator := NewMigrator(db)
	linked, err := migrator.LinkGitHubRepositories()
	if err != nil {
		t.Fatalf("Failed to link GitHub repositories: %v", err)
	}

	if linked != 2 {
		t.Errorf("Expected 2 linked projects, got %d", linked)
	}

	// Verify links were created
	projects, err := db.GetProjects(false, "")
	if err != nil {
		t.Fatalf("Failed to get projects: %v", err)
	}

	for _, project := range projects {
		if project.GitHubRepoID == 0 {
			t.Errorf("Project %s was not linked to GitHub repository", project.Name)
		}
	}
}

func TestCacheConsistency(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add test data
	project := &Project{
		Path:      "~/test-project",
		Name:      "test-project",
		RemoteURL: "https://github.com/user/test-project.git",
		Branch:    "main",
	}

	err = db.UpsertProject(project)
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}

	// Test cache consistency
	cache := NewDatabaseCache(db)
	
	// First call should populate cache
	projects1, err := cache.GetProjects()
	if err != nil {
		t.Fatalf("Failed to get projects from cache: %v", err)
	}

	if len(projects1) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects1))
	}

	// Second call should use cache
	projects2, err := cache.GetProjects()
	if err != nil {
		t.Fatalf("Failed to get projects from cache: %v", err)
	}

	if len(projects2) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects2))
	}

	// Verify cache hit
	if !cache.IsCacheHit("projects") {
		t.Error("Expected cache hit on second call")
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test concurrent reads and writes
	done := make(chan bool, 10)
	
	// Start 5 readers
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_, err := db.GetProjects(false, "")
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
				project := &Project{
					Path:      fmt.Sprintf("~/test-project-%d-%d", id, j),
					Name:      fmt.Sprintf("test-project-%d-%d", id, j),
					RemoteURL: fmt.Sprintf("https://github.com/user/test-project-%d-%d.git", id, j),
					Branch:    "main",
				}
				err := db.UpsertProject(project)
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

	expectedCount := 50 // 5 writers * 10 projects each
	if len(projects) != expectedCount {
		t.Errorf("Expected %d projects, got %d", expectedCount, len(projects))
	}
}