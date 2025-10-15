package cmd

import (
	"os"
	"testing"
	"github.com/shalomb/gum/internal/database"
)

func TestProjectsCommandSimple(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
		if originalConfigHome != "" {
			os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	os.Setenv("XDG_CACHE_HOME", tempDir)
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	t.Run("Database initialization", func(t *testing.T) {
		// Test database initialization
		db, err := database.New()
		if err != nil {
			t.Fatalf("Failed to initialize database: %v", err)
		}
		defer db.Close()

		// Test cache initialization
		cache := database.NewDatabaseCache(db)
		if cache == nil {
			t.Error("Failed to initialize cache")
		}
	})

	t.Run("Cache behavior with cron-based approach", func(t *testing.T) {
		// Test database initialization
		db, err := database.New()
		if err != nil {
			t.Fatalf("Failed to initialize database: %v", err)
		}
		defer db.Close()

		cache := database.NewDatabaseCache(db)

		// Test that IsCacheValid always returns true
		if !cache.IsCacheValid("projects") {
			t.Error("IsCacheValid should always return true with cron-based approach")
		}

		// Test that IsCacheHit always returns true
		if !cache.IsCacheHit("projects") {
			t.Error("IsCacheHit should always return true with cron-based approach")
		}

		// Test that GetProjects works (even if database is empty)
		projects, err := cache.GetProjects()
		if err != nil {
			t.Errorf("GetProjects should work even with empty database: %v", err)
		}

		// Should return empty slice if no projects
		if projects == nil {
			t.Error("GetProjects should return empty slice, not nil")
		}
	})

	t.Run("Database operations", func(t *testing.T) {
		// Test database initialization
		db, err := database.New()
		if err != nil {
			t.Fatalf("Failed to initialize database: %v", err)
		}
		defer db.Close()

		// Test getting project directories
		dirs, err := db.GetProjectDirs()
		if err != nil {
			t.Errorf("GetProjectDirs failed: %v", err)
		}

		// Should return empty slice if no directories
		if dirs == nil {
			t.Error("GetProjectDirs should return empty slice, not nil")
		}
	})
}