/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package database

import (
	"fmt"
	"time"
)

// DatabaseCache provides caching functionality using the database
type DatabaseCache struct {
	db *Database
}

// NewDatabaseCache creates a new database cache instance
func NewDatabaseCache(db *Database) *DatabaseCache {
	return &DatabaseCache{db: db}
}

// GetProjects retrieves projects from cache or database
func (c *DatabaseCache) GetProjects() ([]*Project, error) {
	// Check if cache is valid
	if c.IsCacheValid("projects") {
		return c.db.GetProjects(false, "")
	}

	// Cache miss - this should trigger a refresh
	return nil, fmt.Errorf("cache miss")
}

// GetProjectDirs retrieves project directories from cache or database
func (c *DatabaseCache) GetProjectDirs() ([]*ProjectDir, error) {
	// Check if cache is valid
	if c.IsCacheValid("project-dirs") {
		return c.db.GetProjectDirs()
	}

	// Cache miss - this should trigger a refresh
	return nil, fmt.Errorf("cache miss")
}

// SetProjects updates the projects cache
func (c *DatabaseCache) SetProjects(projects []*Project) error {
	// Clear existing projects
	if err := c.clearTable("projects"); err != nil {
		return err
	}

	// Insert new projects
	for _, project := range projects {
		if err := c.db.UpsertProject(project); err != nil {
			return err
		}
	}

	// Update cache metadata
	return c.updateCacheMetadata("projects", time.Now())
}

// SetProjectDirs updates the project directories cache
func (c *DatabaseCache) SetProjectDirs(dirs []*ProjectDir) error {
	// Clear existing project directories
	if err := c.clearTable("project_dirs"); err != nil {
		return err
	}

	// Insert new project directories
	for _, dir := range dirs {
		if err := c.db.UpsertProjectDir(dir); err != nil {
			return err
		}
	}

	// Update cache metadata
	return c.updateCacheMetadata("project-dirs", time.Now())
}

// IsCacheValid checks if the cache is valid for the given key
func (c *DatabaseCache) IsCacheValid(key string) bool {
	query := `
		SELECT last_updated, ttl_seconds 
		FROM cache_metadata 
		WHERE cache_key = ?
	`

	var lastUpdated time.Time
	var ttlSeconds int

	err := c.db.db.QueryRow(query, key).Scan(&lastUpdated, &ttlSeconds)
	if err != nil {
		return false // Cache miss
	}

	// Check if cache has expired
	ttl := time.Duration(ttlSeconds) * time.Second
	return time.Since(lastUpdated) < ttl
}

// IsCacheHit checks if the last operation was a cache hit
func (c *DatabaseCache) IsCacheHit(key string) bool {
	return c.IsCacheValid(key)
}

// ClearCache clears the cache for the given key
func (c *DatabaseCache) ClearCache(key string) error {
	switch key {
	case "projects":
		return c.clearTable("projects")
	case "project-dirs":
		return c.clearTable("project_dirs")
	case "dirs":
		return c.clearTable("dir_usage")
	case "all":
		return c.clearAllCaches()
	default:
		return c.clearTable(key)
	}
}

// clearTable clears a specific table
func (c *DatabaseCache) clearTable(tableName string) error {
	query := "DELETE FROM " + tableName
	_, err := c.db.db.Exec(query)
	return err
}

// clearAllCaches clears all cache tables
func (c *DatabaseCache) clearAllCaches() error {
	tables := []string{"projects", "project_dirs", "cache_metadata"}
	
	for _, table := range tables {
		if err := c.clearTable(table); err != nil {
			return err
		}
	}

	return nil
}

// updateCacheMetadata updates the cache metadata
func (c *DatabaseCache) updateCacheMetadata(key string, lastUpdated time.Time) error {
	query := `
		INSERT INTO cache_metadata (cache_key, last_updated, ttl_seconds)
		VALUES (?, ?, ?)
		ON CONFLICT(cache_key) DO UPDATE SET
			last_updated = excluded.last_updated,
			ttl_seconds = excluded.ttl_seconds
	`

	// Set TTL based on cache type
	var ttlSeconds int
	switch key {
	case "projects":
		ttlSeconds = 300 // 5 minutes
	case "project-dirs":
		ttlSeconds = 1800 // 30 minutes
	case "dirs":
		ttlSeconds = 30 // 30 seconds (frequently changing)
	default:
		ttlSeconds = 300 // 5 minutes default
	}

	_, err := c.db.db.Exec(query, key, lastUpdated, ttlSeconds)
	return err
}

// GetDirs retrieves directory usage from cache or database
func (c *DatabaseCache) GetDirs() ([]*DirUsage, error) {
	// Check if cache is valid
	if c.IsCacheValid("dirs") {
		return c.db.GetFrequentDirs(1000) // Get top 1000 dirs
	}

	// Cache miss - this should trigger a refresh
	return nil, fmt.Errorf("cache miss")
}

// SetDirs updates the directory usage cache
func (c *DatabaseCache) SetDirs(dirs []*DirUsage) error {
	// Upsert directory usage entries
	for _, dir := range dirs {
		if err := c.db.UpsertDirUsage(dir); err != nil {
			return err
		}
	}

	// Update cache metadata
	return c.updateCacheMetadata("dirs", time.Now())
}

// GetCacheStats returns cache statistics
func (c *DatabaseCache) GetCacheStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get projects count
	var projectsCount int
	err := c.db.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&projectsCount)
	if err != nil {
		return nil, err
	}
	stats["projects_count"] = projectsCount

	// Get project directories count
	var dirsCount int
	err = c.db.db.QueryRow("SELECT COUNT(*) FROM project_dirs").Scan(&dirsCount)
	if err != nil {
		return nil, err
	}
	stats["project_dirs_count"] = dirsCount

	// Get directory usage count
	var dirUsageCount int
	err = c.db.db.QueryRow("SELECT COUNT(*) FROM dir_usage").Scan(&dirUsageCount)
	if err != nil {
		return nil, err
	}
	stats["dir_usage_count"] = dirUsageCount

	// Get GitHub repos count
	var githubReposCount int
	err = c.db.db.QueryRow("SELECT COUNT(*) FROM github_repos").Scan(&githubReposCount)
	if err != nil {
		return nil, err
	}
	stats["github_repos_count"] = githubReposCount

	// Get linked projects count
	var linkedProjectsCount int
	err = c.db.db.QueryRow("SELECT COUNT(*) FROM projects WHERE github_repo_id IS NOT NULL").Scan(&linkedProjectsCount)
	if err != nil {
		return nil, err
	}
	stats["linked_projects_count"] = linkedProjectsCount

	// Get cache metadata
	rows, err := c.db.db.Query(`
		SELECT cache_key, last_updated, ttl_seconds 
		FROM cache_metadata
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cacheInfo := make(map[string]interface{})
	for rows.Next() {
		var key string
		var lastUpdated time.Time
		var ttlSeconds int

		if err := rows.Scan(&key, &lastUpdated, &ttlSeconds); err != nil {
			continue
		}

		cacheInfo[key] = map[string]interface{}{
			"last_updated": lastUpdated,
			"ttl_seconds":  ttlSeconds,
			"is_valid":     c.IsCacheValid(key),
		}
	}
	stats["cache_info"] = cacheInfo

	return stats, nil
}