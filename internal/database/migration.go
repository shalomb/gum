/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Migrator handles migration from JSON caches to database
type Migrator struct {
	db *Database
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *Database) *Migrator {
	return &Migrator{db: db}
}

// JSONCacheEntry represents a cached entry from JSON files
type JSONCacheEntry struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       int         `json:"ttl"`
}

// JSONProject represents a project from the old JSON format
type JSONProject struct {
	Path   string `json:"Path"`
	Remote string `json:"Remote"`
	Branch string `json:"Branch"`
}

// JSONProjectDir represents a project directory from the old JSON format
type JSONProjectDir struct {
	Path        string    `json:"Path"`
	LastScanned time.Time `json:"LastScanned"`
	GitCount    int       `json:"GitCount"`
}

// MigrateFromJSON migrates all JSON cache files to the database
func (m *Migrator) MigrateFromJSON(cacheDir string) error {
	// Create backup directory
	backupDir := filepath.Join(cacheDir, "backup")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Migrate projects.json
	projectsFile := filepath.Join(cacheDir, "projects.json")
	if err := m.migrateProjects(projectsFile, backupDir); err != nil {
		return fmt.Errorf("failed to migrate projects: %v", err)
	}

	// Migrate project-dirs.json
	projectDirsFile := filepath.Join(cacheDir, "project-dirs.json")
	if err := m.migrateProjectDirs(projectDirsFile, backupDir); err != nil {
		return fmt.Errorf("failed to migrate project directories: %v", err)
	}

	// Update cache metadata
	if err := m.updateCacheMetadata("projects", time.Now()); err != nil {
		return fmt.Errorf("failed to update cache metadata: %v", err)
	}

	if err := m.updateCacheMetadata("project-dirs", time.Now()); err != nil {
		return fmt.Errorf("failed to update cache metadata: %v", err)
	}

	return nil
}

// migrateProjects migrates projects.json to the database
func (m *Migrator) migrateProjects(projectsFile, backupDir string) error {
	// Check if file exists
	if _, err := os.Stat(projectsFile); os.IsNotExist(err) {
		return nil // No projects file to migrate
	}

	// Read JSON file
	data, err := os.ReadFile(projectsFile)
	if err != nil {
		return fmt.Errorf("failed to read projects file: %v", err)
	}

	// Parse JSON
	var cacheEntry JSONCacheEntry
	if err := json.Unmarshal(data, &cacheEntry); err != nil {
		return fmt.Errorf("failed to parse projects JSON: %v", err)
	}

	// Convert to projects slice
	projectsData, ok := cacheEntry.Data.([]interface{})
	if !ok {
		return fmt.Errorf("invalid projects data format")
	}

	var projects []JSONProject
	for _, item := range projectsData {
		itemBytes, err := json.Marshal(item)
		if err != nil {
			continue // Skip invalid entries
		}

		var project JSONProject
		if err := json.Unmarshal(itemBytes, &project); err != nil {
			continue // Skip invalid entries
		}

		projects = append(projects, project)
	}

	// Migrate to database
	migrated := 0
	for _, jsonProject := range projects {
		// Convert to database format
		project := &Project{
			Path:      jsonProject.Path,
			Name:      filepath.Base(jsonProject.Path),
			RemoteURL: jsonProject.Remote,
			Branch:    jsonProject.Branch,
		}

		// Insert into database
		if err := m.db.UpsertProject(project); err != nil {
			continue // Skip failed entries
		}
		migrated++
	}

	// Backup original file
	backupFile := filepath.Join(backupDir, "projects.json")
	if err := os.Rename(projectsFile, backupFile); err != nil {
		return fmt.Errorf("failed to backup projects file: %v", err)
	}

	fmt.Printf("Migrated %d projects from JSON cache\n", migrated)
	return nil
}

// migrateProjectDirs migrates project-dirs.json to the database
func (m *Migrator) migrateProjectDirs(projectDirsFile, backupDir string) error {
	// Check if file exists
	if _, err := os.Stat(projectDirsFile); os.IsNotExist(err) {
		return nil // No project-dirs file to migrate
	}

	// Read JSON file
	data, err := os.ReadFile(projectDirsFile)
	if err != nil {
		return fmt.Errorf("failed to read project-dirs file: %v", err)
	}

	// Parse JSON
	var cacheEntry JSONCacheEntry
	if err := json.Unmarshal(data, &cacheEntry); err != nil {
		return fmt.Errorf("failed to parse project-dirs JSON: %v", err)
	}

	// Convert to project directories slice
	dirsData, ok := cacheEntry.Data.([]interface{})
	if !ok {
		return fmt.Errorf("invalid project-dirs data format")
	}

	var projectDirs []JSONProjectDir
	for _, item := range dirsData {
		itemBytes, err := json.Marshal(item)
		if err != nil {
			continue // Skip invalid entries
		}

		var projectDir JSONProjectDir
		if err := json.Unmarshal(itemBytes, &projectDir); err != nil {
			continue // Skip invalid entries
		}

		projectDirs = append(projectDirs, projectDir)
	}

	// Migrate to database
	migrated := 0
	for _, jsonDir := range projectDirs {
		// Convert to database format
		projectDir := &ProjectDir{
			Path:        jsonDir.Path,
			LastScanned: jsonDir.LastScanned,
			GitCount:    jsonDir.GitCount,
		}

		// Insert into database
		if err := m.db.UpsertProjectDir(projectDir); err != nil {
			continue // Skip failed entries
		}
		migrated++
	}

	// Backup original file
	backupFile := filepath.Join(backupDir, "project-dirs.json")
	if err := os.Rename(projectDirsFile, backupFile); err != nil {
		return fmt.Errorf("failed to backup project-dirs file: %v", err)
	}

	fmt.Printf("Migrated %d project directories from JSON cache\n", migrated)
	return nil
}

// LinkGitHubRepositories links local projects to GitHub repositories
func (m *Migrator) LinkGitHubRepositories() (int, error) {
	// Get all projects
	projects, err := m.db.GetProjects(false, "")
	if err != nil {
		return 0, fmt.Errorf("failed to get projects: %v", err)
	}

	// Get all GitHub repositories
	githubRepos, err := m.db.GetGitHubRepos()
	if err != nil {
		return 0, fmt.Errorf("failed to get GitHub repositories: %v", err)
	}

	// Create a map of clone URLs to GitHub repo IDs
	cloneURLMap := make(map[string]int64)
	for _, repo := range githubRepos {
		if repo.CloneURL != "" {
			cloneURLMap[repo.CloneURL] = repo.ID
		}
	}

	// Link projects to GitHub repositories
	linked := 0
	for _, project := range projects {
		if project.RemoteURL == "" {
			continue
		}

		// Try to find matching GitHub repository
		if githubRepoID, exists := cloneURLMap[project.RemoteURL]; exists {
			// Update project with GitHub repo ID
			project.GitHubRepoID = githubRepoID
			if err := m.db.UpsertProject(project); err != nil {
				continue // Skip failed updates
			}
			linked++
		}
	}

	fmt.Printf("Linked %d projects to GitHub repositories\n", linked)
	return linked, nil
}

// updateCacheMetadata updates the cache metadata table
func (m *Migrator) updateCacheMetadata(cacheKey string, lastUpdated time.Time) error {
	query := `
		INSERT INTO cache_metadata (cache_key, last_updated, ttl_seconds)
		VALUES (?, ?, ?)
		ON CONFLICT(cache_key) DO UPDATE SET
			last_updated = excluded.last_updated,
			ttl_seconds = excluded.ttl_seconds
	`

	_, err := m.db.db.Exec(query, cacheKey, lastUpdated, 300) // 5 minutes TTL
	return err
}

// RollbackMigration restores JSON cache files from backup
func (m *Migrator) RollbackMigration(cacheDir string) error {
	backupDir := filepath.Join(cacheDir, "backup")
	
	// Restore projects.json
	projectsBackup := filepath.Join(backupDir, "projects.json")
	projectsFile := filepath.Join(cacheDir, "projects.json")
	if _, err := os.Stat(projectsBackup); err == nil {
		if err := os.Rename(projectsBackup, projectsFile); err != nil {
			return fmt.Errorf("failed to restore projects.json: %v", err)
		}
	}

	// Restore project-dirs.json
	projectDirsBackup := filepath.Join(backupDir, "project-dirs.json")
	projectDirsFile := filepath.Join(cacheDir, "project-dirs.json")
	if _, err := os.Stat(projectDirsBackup); err == nil {
		if err := os.Rename(projectDirsBackup, projectDirsFile); err != nil {
			return fmt.Errorf("failed to restore project-dirs.json: %v", err)
		}
	}

	// Clear database tables
	if err := m.clearDatabaseTables(); err != nil {
		return fmt.Errorf("failed to clear database tables: %v", err)
	}

	fmt.Println("Migration rolled back successfully")
	return nil
}

// clearDatabaseTables clears the migrated tables
func (m *Migrator) clearDatabaseTables() error {
	tables := []string{"projects", "project_dirs", "cache_metadata"}
	
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		if _, err := m.db.db.Exec(query); err != nil {
			return fmt.Errorf("failed to clear table %s: %v", table, err)
		}
	}

	return nil
}

// BackupDatabase creates a backup of the database
func (m *Migrator) BackupDatabase(backupPath string) error {
	// Get the current database path
	dbPath := m.db.dbPath
	
	// Copy database file to backup location
	sourceFile, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer destFile.Close()

	// Copy file contents
	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return fmt.Errorf("failed to copy database: %v", err)
	}

	fmt.Printf("Database backed up to %s\n", backupPath)
	return nil
}

// RestoreDatabase restores the database from a backup
func (m *Migrator) RestoreDatabase(backupPath string) error {
	// Close current database
	if err := m.db.Close(); err != nil {
		return fmt.Errorf("failed to close current database: %v", err)
	}

	// Copy backup to current database location
	sourceFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(m.db.dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database file: %v", err)
	}
	defer destFile.Close()

	// Copy file contents
	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return fmt.Errorf("failed to copy backup: %v", err)
	}

	// Reopen database
	newDB, err := New(m.db.dbPath)
	if err != nil {
		return fmt.Errorf("failed to reopen database: %v", err)
	}
	m.db = newDB

	fmt.Printf("Database restored from %s\n", backupPath)
	return nil
}