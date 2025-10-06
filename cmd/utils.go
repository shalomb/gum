/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"os"
	"path/filepath"
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