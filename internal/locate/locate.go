// Package locate provides integration with the locate database for fast file discovery
package locate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// LocateFinder provides integration with the locate database
type LocateFinder struct {
	available    bool
	databasePath string
	lastUpdated  time.Time
}

// LocateStatus represents the status of the locate database
type LocateStatus struct {
	Available   bool
	DatabasePath string
	LastUpdated  time.Time
	Age         time.Duration
	IsFresh     bool
}

// NewLocateFinder creates a new LocateFinder instance
func NewLocateFinder() *LocateFinder {
	finder := &LocateFinder{}
	
	// Check if locate command is available
	if _, err := exec.LookPath("locate"); err != nil {
		return finder // available = false
	}
	
	finder.available = true
	
	// Find database path
	finder.databasePath = findDatabasePath()
	
	// Check database freshness
	if finder.databasePath != "" {
		if stat, err := os.Stat(finder.databasePath); err == nil {
			finder.lastUpdated = stat.ModTime()
		}
	}
	
	return finder
}

// GetStatus returns the current status of the locate database
func (lf *LocateFinder) GetStatus() LocateStatus {
	if !lf.available {
		return LocateStatus{Available: false}
	}
	
	age := time.Since(lf.lastUpdated)
	return LocateStatus{
		Available:    true,
		DatabasePath: lf.databasePath,
		LastUpdated:  lf.lastUpdated,
		Age:          age,
		IsFresh:      age < 24*time.Hour,
	}
}

// FindGitRepos finds git repositories using the locate database
func (lf *LocateFinder) FindGitRepos(basePath string) ([]string, error) {
	if !lf.available {
		return nil, fmt.Errorf("locate not available")
	}
	
	// Use basic regex that works across implementations
	cmd := exec.Command("locate", "-r", `\.git$`)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("locate command failed: %w", err)
	}
	
	var repos []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Filter for our base path
		if strings.HasPrefix(line, basePath) {
			// Extract parent directory (remove /.git)
			repoPath := filepath.Dir(line)
			repos = append(repos, repoPath)
		}
	}
	
	return repos, nil
}

// findDatabasePath finds the locate database path
func findDatabasePath() string {
	// Common database locations across implementations
	candidates := []string{
		"/var/lib/plocate/plocate.db",  // plocate
		"/var/lib/mlocate/mlocate.db",  // GNU locate
		"/var/db/locate.database",     // BSD locate
	}
	
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	
	return "" // Use default
}

// IsAvailable checks if locate is available on the system
func IsAvailable() bool {
	_, err := exec.LookPath("locate")
	return err == nil
}

// GetDatabaseAge returns the age of the locate database
func GetDatabaseAge() (time.Duration, error) {
	if !IsAvailable() {
		return 0, fmt.Errorf("locate not available")
	}
	
	finder := NewLocateFinder()
	status := finder.GetStatus()
	if !status.Available {
		return 0, fmt.Errorf("locate database not found")
	}
	
	return status.Age, nil
}