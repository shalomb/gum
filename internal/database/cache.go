/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package database

import (
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

// GetProjects retrieves projects from database
func (c *DatabaseCache) GetProjects() ([]*Project, error) {
	// Always return database data - cron jobs keep it fresh
	return c.db.GetProjects(false, "")
}

// GetProjectDirs retrieves project directories from database
func (c *DatabaseCache) GetProjectDirs() ([]*ProjectDir, error) {
	// Always return database data - cron jobs keep it fresh
	return c.db.GetProjectDirs()
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

	// No cache metadata needed - cron jobs handle freshness
	return nil
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

	// No cache metadata needed - cron jobs handle freshness
	return nil
}

// IsCacheValid - DEPRECATED: TTL-based cache validation removed
// Always returns true since cron jobs keep data fresh
func (c *DatabaseCache) IsCacheValid(key string) bool {
	return true
}

// IsCacheHit - DEPRECATED: TTL-based cache validation removed
// Always returns true since cron jobs keep data fresh
func (c *DatabaseCache) IsCacheHit(key string) bool {
	return true
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
	tables := []string{"projects", "project_dirs"}
	
	for _, table := range tables {
		if err := c.clearTable(table); err != nil {
			return err
		}
	}

	return nil
}

// updateCacheMetadata - DEPRECATED: TTL-based cache metadata removed
// Cron jobs handle data freshness, no metadata tracking needed
func (c *DatabaseCache) updateCacheMetadata(key string, lastUpdated time.Time) error {
	// No-op: Cron jobs handle freshness
	return nil
}

// GetDirs retrieves directory usage from database
func (c *DatabaseCache) GetDirs() ([]*DirUsage, error) {
	// Always return database data - cron jobs keep it fresh
	return c.db.GetFrequentDirs(1000) // Get top 1000 dirs
}

// SetDirs updates the directory usage cache
func (c *DatabaseCache) SetDirs(dirs []*DirUsage) error {
	// Upsert directory usage entries
	for _, dir := range dirs {
		if err := c.db.UpsertDirUsage(dir); err != nil {
			return err
		}
	}

	// No cache metadata needed - cron jobs handle freshness
	return nil
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

	// Cache info - simplified for cron-based updates
	stats["cache_info"] = map[string]interface{}{
		"note": "TTL-based cache validation removed - cron jobs handle freshness",
		"cron_based": true,
	}

	return stats, nil
}