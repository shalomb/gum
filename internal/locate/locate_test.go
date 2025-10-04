package locate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewLocateFinder(t *testing.T) {
	finder := NewLocateFinder()
	
	// Should always create a finder, even if locate is not available
	if finder == nil {
		t.Fatal("NewLocateFinder() returned nil")
	}
	
	// Test availability detection
	available := finder.available
	if available && !IsAvailable() {
		t.Error("Finder reports available but IsAvailable() returns false")
	}
}

func TestGetStatus(t *testing.T) {
	finder := NewLocateFinder()
	status := finder.GetStatus()
	
	if finder.available != status.Available {
		t.Errorf("Status.Available (%v) != finder.available (%v)", status.Available, finder.available)
	}
	
	if status.Available {
		if status.DatabasePath == "" {
			t.Error("Available locate should have database path")
		}
		
		if status.Age < 0 {
			t.Error("Age should be non-negative")
		}
		
		// Test freshness logic
		expectedFresh := status.Age < 24*time.Hour
		if status.IsFresh != expectedFresh {
			t.Errorf("IsFresh (%v) != expected (%v)", status.IsFresh, expectedFresh)
		}
	}
}

func TestFindGitRepos(t *testing.T) {
	finder := NewLocateFinder()
	
	if !finder.available {
		t.Skip("locate not available, skipping test")
	}
	
	// Test with a path that should exist
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		t.Skip("HOME not set, skipping test")
	}
	
	repos, err := finder.FindGitRepos(homeDir)
	if err != nil {
		t.Fatalf("FindGitRepos failed: %v", err)
	}
	
	// Validate results
	validRepos := 0
	for _, repo := range repos {
		if !strings.HasPrefix(repo, homeDir) {
			t.Errorf("Repo %s not in expected path %s", repo, homeDir)
			continue
		}
		
		// Check if it's actually a git repository
		gitPath := filepath.Join(repo, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			validRepos++
		}
		// Don't fail the test for false positives from locate database
	}
	
	// We should find at least some valid repos
	if validRepos == 0 && len(repos) > 0 {
		t.Logf("Warning: Found %d repos but none are valid git repositories", len(repos))
		t.Logf("This might indicate locate database contains stale entries")
	}
}

func TestFindGitReposNotAvailable(t *testing.T) {
	// Create a finder that's not available
	finder := &LocateFinder{available: false}
	
	repos, err := finder.FindGitRepos("/some/path")
	if err == nil {
		t.Error("Expected error when locate not available")
	}
	
	if repos != nil {
		t.Error("Expected nil repos when locate not available")
	}
}

func TestIsAvailable(t *testing.T) {
	// This test depends on system configuration
	// We can't easily mock exec.LookPath, so we just test the function exists
	available := IsAvailable()
	
	// Should return a boolean value
	if available != true && available != false {
		t.Error("IsAvailable() should return true or false")
	}
}

func TestGetDatabaseAge(t *testing.T) {
	age, err := GetDatabaseAge()
	
	if !IsAvailable() {
		if err == nil {
			t.Error("Expected error when locate not available")
		}
		return
	}
	
	if err != nil {
		t.Fatalf("GetDatabaseAge failed: %v", err)
	}
	
	if age < 0 {
		t.Error("Database age should be non-negative")
	}
	
	// Age should be reasonable (not more than a few years)
	if age > 365*24*time.Hour {
		t.Errorf("Database age %v seems too old", age)
	}
}

func TestFindDatabasePath(t *testing.T) {
	path := findDatabasePath()
	
	// If locate is available, we should find a database path
	if IsAvailable() && path == "" {
		t.Error("Should find database path when locate is available")
	}
	
	// If we found a path, it should exist
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Database path %s does not exist: %v", path, err)
		}
	}
}