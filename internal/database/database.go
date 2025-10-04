/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents the gum SQLite database
type Database struct {
	db *sql.DB
}

// New creates a new database connection
func New() (*Database, error) {
	dbPath := getDatabasePath()
	
	// Ensure database directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	
	// Initialize schema
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}
	
	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// getDatabasePath returns the path to the SQLite database file
func getDatabasePath() string {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		cacheDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(cacheDir, "gum", "gum.db")
}

// initSchema initializes the database schema
func initSchema(db *sql.DB) error {
	schemaSQL := `
	-- Projects table: Git repositories found in project directories
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		remote_url TEXT,
		branch TEXT,
		last_modified DATETIME,
		git_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Project directories table: Directories that contain projects
	CREATE TABLE IF NOT EXISTS project_dirs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL UNIQUE,
		last_scanned DATETIME,
		git_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- GitHub repositories table: Repositories discovered via GitHub API
	CREATE TABLE IF NOT EXISTS github_repos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		full_name TEXT NOT NULL UNIQUE,
		description TEXT,
		url TEXT,
		clone_url TEXT,
		ssh_url TEXT,
		is_private BOOLEAN DEFAULT 0,
		is_fork BOOLEAN DEFAULT 0,
		updated_at DATETIME,
		last_discovered DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Directory usage tracking: For frecently used directories
	CREATE TABLE IF NOT EXISTS dir_usage (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL,
		frequency INTEGER DEFAULT 1,
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(path)
	);

	-- Similarity cache: Pre-computed similarity scores
	CREATE TABLE IF NOT EXISTS similarity_cache (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_type TEXT NOT NULL,
		source_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		target_id INTEGER NOT NULL,
		similarity_score REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source_type, source_id, target_type, target_id)
	);

	-- Indexes for performance
	CREATE INDEX IF NOT EXISTS idx_projects_path ON projects(path);
	CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
	CREATE INDEX IF NOT EXISTS idx_projects_remote ON projects(remote_url);
	CREATE INDEX IF NOT EXISTS idx_projects_updated ON projects(updated_at);

	CREATE INDEX IF NOT EXISTS idx_project_dirs_path ON project_dirs(path);
	CREATE INDEX IF NOT EXISTS idx_project_dirs_scanned ON project_dirs(last_scanned);

	CREATE INDEX IF NOT EXISTS idx_github_repos_full_name ON github_repos(full_name);
	CREATE INDEX IF NOT EXISTS idx_github_repos_name ON github_repos(name);
	CREATE INDEX IF NOT EXISTS idx_github_repos_updated ON github_repos(updated_at);
	CREATE INDEX IF NOT EXISTS idx_github_repos_private ON github_repos(is_private);

	CREATE INDEX IF NOT EXISTS idx_dir_usage_path ON dir_usage(path);
	CREATE INDEX IF NOT EXISTS idx_dir_usage_frequency ON dir_usage(frequency);
	CREATE INDEX IF NOT EXISTS idx_dir_usage_last_seen ON dir_usage(last_seen);

	CREATE INDEX IF NOT EXISTS idx_similarity_source ON similarity_cache(source_type, source_id);
	CREATE INDEX IF NOT EXISTS idx_similarity_target ON similarity_cache(target_type, target_id);
	CREATE INDEX IF NOT EXISTS idx_similarity_score ON similarity_cache(similarity_score);
	`
	
	_, err := db.Exec(schemaSQL)
	return err
}

// Project represents a Git project
type Project struct {
	ID           int       `json:"id"`
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	RemoteURL    string    `json:"remote_url"`
	Branch       string    `json:"branch"`
	LastModified time.Time `json:"last_modified"`
	GitCount     int       `json:"git_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ProjectDir represents a project directory
type ProjectDir struct {
	ID          int       `json:"id"`
	Path        string    `json:"path"`
	LastScanned time.Time `json:"last_scanned"`
	GitCount    int       `json:"git_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GitHubRepo represents a GitHub repository
type GitHubRepo struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	FullName       string    `json:"full_name"`
	Description    string    `json:"description"`
	URL            string    `json:"url"`
	CloneURL       string    `json:"clone_url"`
	SSHURL         string    `json:"ssh_url"`
	IsPrivate      bool      `json:"is_private"`
	IsFork         bool      `json:"is_fork"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastDiscovered time.Time `json:"last_discovered"`
	CreatedAt      time.Time `json:"created_at"`
}

// DirUsage represents directory usage tracking
type DirUsage struct {
	ID        int       `json:"id"`
	Path      string    `json:"path"`
	Frequency int       `json:"frequency"`
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}