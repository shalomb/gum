/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
}

// Cache manages file-based caching for gum commands
type Cache struct {
	baseDir string
}

// New creates a new cache instance
func New() *Cache {
	cacheDir := getCacheDir()
	return &Cache{
		baseDir: cacheDir,
	}
}

// Get retrieves cached data if it exists and is not expired
func (c *Cache) Get(key string, target interface{}) bool {
	cacheFile := c.getCacheFile(key)
	
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return false
	}
	
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return false
	}
	
	// Check if cache is expired
	if time.Since(entry.Timestamp) > entry.TTL {
		os.Remove(cacheFile) // Clean up expired cache
		return false
	}
	
	// Unmarshal the actual data
	dataBytes, err := json.Marshal(entry.Data)
	if err != nil {
		return false
	}
	
	if err := json.Unmarshal(dataBytes, target); err != nil {
		return false
	}
	
	return true
}

// Set stores data in cache with TTL
func (c *Cache) Set(key string, data interface{}, ttl time.Duration) error {
	entry := CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
	
	// Ensure cache directory exists
	if err := os.MkdirAll(c.baseDir, 0755); err != nil {
		return err
	}
	
	cacheFile := c.getCacheFile(key)
	dataBytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	
	return os.WriteFile(cacheFile, dataBytes, 0644)
}

// Clear removes a specific cache entry
func (c *Cache) Clear(key string) error {
	cacheFile := c.getCacheFile(key)
	return os.Remove(cacheFile)
}

// ClearAll removes all cache entries
func (c *Cache) ClearAll() error {
	return os.RemoveAll(c.baseDir)
}

// getCacheFile returns the full path to a cache file
func (c *Cache) getCacheFile(key string) string {
	return filepath.Join(c.baseDir, key+".json")
}

// getCacheDir returns the cache directory path
func getCacheDir() string {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		cacheDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(cacheDir, "gum")
}

// Default TTL values
const (
	ProjectsCacheTTL = 5 * time.Minute  // Projects change less frequently
	DirsCacheTTL     = 30 * time.Second // Directories change more frequently
	ProjectDirsCacheTTL = 1 * time.Hour // Project directories change rarely
)