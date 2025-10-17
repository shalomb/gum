/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shalomb/gum/internal/database"
	"gopkg.in/yaml.v3"
)

// getCacheDir returns the cache directory path
func getCacheDir() string {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		cacheDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(cacheDir, "gum")
}

// getDatabasePath returns the database file path
func getDatabasePath() string {
	cacheDir := getCacheDir()
	return filepath.Join(cacheDir, "gum.db")
}

// Config represents the gum configuration structure
type Config struct {
	Projects []string `yaml:"projects"`
}

// readYAMLConfig reads project directories from config.yaml
func readYAMLConfig() []string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	
	configFile := filepath.Join(configDir, "gum", "config.yaml")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil
	}
	
	// Expand ~ to home directory
	var expandedPaths []string
	for _, path := range config.Projects {
		if strings.HasPrefix(path, "~/") {
			home := os.Getenv("HOME")
			expandedPaths = append(expandedPaths, filepath.Join(home, path[2:]))
		} else {
			expandedPaths = append(expandedPaths, path)
		}
	}
	
	return expandedPaths
}

// loadProjectDirsFromConfig loads project directories from config.yaml and populates database
func loadProjectDirsFromConfig(db *database.Database) error {
	configDirs := readYAMLConfig()
	if len(configDirs) == 0 {
		return nil // No config or empty config
	}
	
	// Load each directory from config into database
	for _, dir := range configDirs {
		// Check if directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue // Skip non-existent directories
		}
		
		// Count git repositories in directory
		gitCount := countGitReposInDir(dir)
		
		// Create project directory entry
		projectDir := &database.ProjectDir{
			Path:        dir,
			LastScanned: time.Now(),
			GitCount:    gitCount,
		}
		
		// Upsert to database
		if err := db.UpsertProjectDir(projectDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to upsert project directory %s: %v\n", dir, err)
			continue
		}
	}
	
	return nil
}

// countGitReposInDir counts .git directories in a path
func countGitReposInDir(path string) int {
	count := 0
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() && info.Name() == ".git" {
			count++
			return filepath.SkipDir // Don't recurse into .git
		}
		return nil
	})
	return count
}