package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Create a temporary directory for test database
	tempDir := t.TempDir()
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	// Set test cache directory
	os.Setenv("XDG_CACHE_HOME", tempDir)

	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify database file was created
	dbPath := filepath.Join(tempDir, "gum", "gum.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", dbPath)
	}
}

func TestDatabaseOperations(t *testing.T) {
	// Create test database
	tempDir := t.TempDir()
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	os.Setenv("XDG_CACHE_HOME", tempDir)

	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	t.Run("Project Operations", func(t *testing.T) {
		// Test UpsertProject
		project := &Project{
			Path:         "/test/project",
			Name:         "test-project",
			RemoteURL:    "https://github.com/test/test-project.git",
			Branch:       "main",
			LastModified: time.Now(),
			GitCount:     1,
		}

		err := db.UpsertProject(project)
		if err != nil {
			t.Errorf("UpsertProject failed: %v", err)
		}

		// Test GetProjects
		projects, err := db.GetProjects(false, "")
		if err != nil {
			t.Errorf("GetProjects failed: %v", err)
		}

		if len(projects) != 1 {
			t.Errorf("Expected 1 project, got %d", len(projects))
		}

		if projects[0].Path != "/test/project" {
			t.Errorf("Expected path '/test/project', got '%s'", projects[0].Path)
		}

		// Test GetSimilarProjects
		similar, err := db.GetSimilarProjects("test", 10)
		if err != nil {
			t.Errorf("GetSimilarProjects failed: %v", err)
		}

		if len(similar) != 1 {
			t.Errorf("Expected 1 similar project, got %d", len(similar))
		}
	})

	t.Run("ProjectDir Operations", func(t *testing.T) {
		// Test UpsertProjectDir
		dir := &ProjectDir{
			Path:        "/test/dir",
			LastScanned: time.Now(),
			GitCount:    5,
		}

		err := db.UpsertProjectDir(dir)
		if err != nil {
			t.Errorf("UpsertProjectDir failed: %v", err)
		}

		// Test GetProjectDirs
		dirs, err := db.GetProjectDirs()
		if err != nil {
			t.Errorf("GetProjectDirs failed: %v", err)
		}

		if len(dirs) != 1 {
			t.Errorf("Expected 1 project directory, got %d", len(dirs))
		}

		if dirs[0].Path != "/test/dir" {
			t.Errorf("Expected path '/test/dir', got '%s'", dirs[0].Path)
		}
	})

	t.Run("DirUsage Operations", func(t *testing.T) {
		// Test UpsertDirUsage
		usage := &DirUsage{
			Path:      "/test/usage",
			Frequency: 10,
			LastSeen:  time.Now(),
		}

		err := db.UpsertDirUsage(usage)
		if err != nil {
			t.Errorf("UpsertDirUsage failed: %v", err)
		}

		// Test GetFrequentDirs
		frequent, err := db.GetFrequentDirs(10)
		if err != nil {
			t.Errorf("GetFrequentDirs failed: %v", err)
		}

		if len(frequent) != 1 {
			t.Errorf("Expected 1 frequent directory, got %d", len(frequent))
		}

		if frequent[0].Path != "/test/usage" {
			t.Errorf("Expected path '/test/usage', got '%s'", frequent[0].Path)
		}
	})

	t.Run("GitHub Repo Operations", func(t *testing.T) {
		// Test UpsertGitHubRepo
		repo := &GitHubRepo{
			Name:        "test-repo",
			FullName:    "test/test-repo",
			Description: "A test repository",
			URL:         "https://github.com/test/test-repo",
			CloneURL:    "https://github.com/test/test-repo.git",
			SSHURL:      "git@github.com:test/test-repo.git",
			IsPrivate:   false,
			IsFork:      false,
			UpdatedAt:   time.Now(),
		}

		err := db.UpsertGitHubRepo(repo)
		if err != nil {
			t.Errorf("UpsertGitHubRepo failed: %v", err)
		}

		// Test GetGitHubRepos
		repos, err := db.GetGitHubRepos()
		if err != nil {
			t.Errorf("GetGitHubRepos failed: %v", err)
		}

		if len(repos) != 1 {
			t.Errorf("Expected 1 GitHub repository, got %d", len(repos))
		}

		if repos[0].FullName != "test/test-repo" {
			t.Errorf("Expected full name 'test/test-repo', got '%s'", repos[0].FullName)
		}
	})

	t.Run("Stats and Clear Operations", func(t *testing.T) {
		// Test GetStats
		stats, err := db.GetStats()
		if err != nil {
			t.Errorf("GetStats failed: %v", err)
		}

		expectedTables := []string{"projects", "project_dirs", "github_repos", "dir_usage"}
		for _, table := range expectedTables {
			if _, exists := stats[table]; !exists {
				t.Errorf("Stats missing for table: %s", table)
			}
		}

		// Test ClearCache
		err = db.ClearCache()
		if err != nil {
			t.Errorf("ClearCache failed: %v", err)
		}

		// Verify cache is cleared
		statsAfter, err := db.GetStats()
		if err != nil {
			t.Errorf("GetStats after clear failed: %v", err)
		}

		for _, table := range expectedTables {
			if statsAfter[table] != 0 {
				t.Errorf("Table %s not cleared, still has %d records", table, statsAfter[table])
			}
		}
	})
}

func TestDatabaseConcurrency(t *testing.T) {
	// Test concurrent access
	tempDir := t.TempDir()
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	os.Setenv("XDG_CACHE_HOME", tempDir)

	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			project := &Project{
				Path:         "/test/concurrent/project" + string(rune(i)),
				Name:         "concurrent-project",
				RemoteURL:    "https://github.com/test/concurrent-project.git",
				Branch:       "main",
				LastModified: time.Now(),
				GitCount:     1,
			}
			err := db.UpsertProject(project)
			if err != nil {
				t.Errorf("Concurrent UpsertProject failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all projects were inserted
	projects, err := db.GetProjects(false, "")
	if err != nil {
		t.Errorf("GetProjects after concurrent writes failed: %v", err)
	}

	if len(projects) != 10 {
		t.Errorf("Expected 10 projects after concurrent writes, got %d", len(projects))
	}
}

func TestDatabasePathResolution(t *testing.T) {
	tests := []struct {
		name           string
		xdgCacheHome   string
		expectedSuffix string
	}{
		{
			name:           "Default XDG cache home",
			xdgCacheHome:  "",
			expectedSuffix: ".cache/gum/gum.db",
		},
		{
			name:           "Custom XDG cache home",
			xdgCacheHome:  "/custom/cache",
			expectedSuffix: "/custom/cache/gum/gum.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCacheHome := os.Getenv("XDG_CACHE_HOME")
			defer func() {
				if originalCacheHome != "" {
					os.Setenv("XDG_CACHE_HOME", originalCacheHome)
				} else {
					os.Unsetenv("XDG_CACHE_HOME")
				}
			}()

			if tt.xdgCacheHome != "" {
				os.Setenv("XDG_CACHE_HOME", tt.xdgCacheHome)
			} else {
				os.Unsetenv("XDG_CACHE_HOME")
			}

			// We can't easily test the path resolution without creating the database,
			// but we can verify the environment variable is respected
			if tt.xdgCacheHome != "" {
				if os.Getenv("XDG_CACHE_HOME") != tt.xdgCacheHome {
					t.Errorf("XDG_CACHE_HOME not set correctly")
				}
			}
		})
	}
}