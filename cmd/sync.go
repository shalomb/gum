package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shalomb/gum/internal/github"
)

var (
	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync repository metadata from GitHub",
		Long: `Sync repository metadata from GitHub API including:
- Repository names and descriptions
- Topics and languages
- Activity metrics (stars, forks, issues)
- Timestamps (created, updated, pushed)
- Repository properties (private, archived, fork)

This command is designed to be run daily via crontab to keep metadata fresh.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := doSync(); err != nil {
				log.Fatalf("Sync failed: %v", err)
			}
		},
	}
	
	syncType string
	dryRun   bool
)

func init() {
	rootCmd.AddCommand(syncCmd)
	
	syncCmd.Flags().StringVarP(&syncType, "type", "t", "full", "Sync type: full, incremental, metadata")
	syncCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be synced without making changes")
}

func doSync() error {
	// Initialize GitHub client
	client, err := github.NewGitHubClient()
	if err != nil {
		return fmt.Errorf("failed to initialize GitHub client: %v", err)
	}
	
	// Initialize database
	db, err := initDatabase()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	defer db.Close()
	
	// Create sync status record
	syncID, err := createSyncStatus(db, syncType)
	if err != nil {
		return fmt.Errorf("failed to create sync status: %v", err)
	}
	
	// Perform sync based on type
	var syncErr error
	switch syncType {
	case "full":
		syncErr = performFullSync(db, client, syncID)
	case "incremental":
		syncErr = performIncrementalSync(db, client, syncID)
	case "metadata":
		syncErr = performMetadataSync(db, client, syncID)
	default:
		syncErr = fmt.Errorf("invalid sync type: %s", syncType)
	}
	
	if syncErr != nil {
		updateSyncStatus(db, syncID, "failed", syncErr.Error())
		return syncErr
	} else {
		updateSyncStatus(db, syncID, "completed", "")
	}
	
	fmt.Printf("Sync completed successfully (%s)\n", syncType)
	return nil
}

func getCacheDir() string {
	if cacheHome := os.Getenv("XDG_CACHE_HOME"); cacheHome != "" {
		return fmt.Sprintf("%s/gum", cacheHome)
	}
	return fmt.Sprintf("%s/.cache/gum", os.Getenv("HOME"))
}

func initDatabase() (*sql.DB, error) {
	// Get cache directory
	cacheDir := getCacheDir()
	dbPath := fmt.Sprintf("%s/gum.db", cacheDir)
	
	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	
	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, err
	}
	
	return db, nil
}

func createTables(db *sql.DB) error {
	// Read schema file
	schema := `
	CREATE TABLE IF NOT EXISTS github_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		full_name TEXT UNIQUE NOT NULL,
		owner TEXT NOT NULL,
		repo TEXT NOT NULL,
		description TEXT,
		topics JSON,
		language TEXT,
		languages JSON,
		is_private BOOLEAN DEFAULT FALSE,
		is_archived BOOLEAN DEFAULT FALSE,
		is_fork BOOLEAN DEFAULT FALSE,
		is_template BOOLEAN DEFAULT FALSE,
		visibility TEXT,
		default_branch TEXT,
		license TEXT,
		star_count INTEGER DEFAULT 0,
		fork_count INTEGER DEFAULT 0,
		open_issues INTEGER DEFAULT 0,
		size INTEGER DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		pushed_at TIMESTAMP,
		last_synced TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		sync_count INTEGER DEFAULT 0
	);
	
	CREATE TABLE IF NOT EXISTS github_sync_status (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sync_type TEXT NOT NULL,
		started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP,
		status TEXT NOT NULL,
		repositories_processed INTEGER DEFAULT 0,
		repositories_total INTEGER DEFAULT 0,
		error_message TEXT
	);
	`
	
	_, err := db.Exec(schema)
	return err
}

func createSyncStatus(db *sql.DB, syncType string) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO github_sync_status (sync_type, status, started_at)
		VALUES (?, 'running', CURRENT_TIMESTAMP)
	`, syncType)
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

func updateSyncStatus(db *sql.DB, syncID int64, status, errorMsg string) error {
	_, err := db.Exec(`
		UPDATE github_sync_status 
		SET status = ?, completed_at = CURRENT_TIMESTAMP, error_message = ?
		WHERE id = ?
	`, status, errorMsg, syncID)
	return err
}

func performFullSync(db *sql.DB, client *github.GitHubClient, syncID int64) error {
	fmt.Println("Starting full sync...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	
	// Discover all repositories
	repos, err := client.DiscoverAllRepositories(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover repositories: %v", err)
	}
	
	fmt.Printf("Found %d repositories\n", len(repos))
	
	// Update sync status with total count
	_, err = db.Exec(`
		UPDATE github_sync_status 
		SET repositories_total = ?
		WHERE id = ?
	`, len(repos), syncID)
	if err != nil {
		return err
	}
	
	// Process repositories
	processed := 0
	for _, repo := range repos {
		if dryRun {
			fmt.Printf("Would sync: %s (%s)\n", repo.FullName, repo.Description)
		} else {
			if err := upsertRepository(db, repo); err != nil {
				log.Printf("Failed to sync %s: %v", repo.FullName, err)
				continue
			}
		}
		
		processed++
		
		// Update progress every 100 repositories
		if processed%100 == 0 {
			_, err = db.Exec(`
				UPDATE github_sync_status 
				SET repositories_processed = ?
				WHERE id = ?
			`, processed, syncID)
			if err != nil {
				log.Printf("Failed to update progress: %v", err)
			}
			
			fmt.Printf("Processed %d/%d repositories\n", processed, len(repos))
		}
		
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	
	fmt.Printf("Full sync completed: %d repositories processed\n", processed)
	return nil
}

func performIncrementalSync(db *sql.DB, client *github.GitHubClient, syncID int64) error {
	fmt.Println("Starting incremental sync...")
	
	// Get repositories that need updating (older than 24 hours)
	rows, err := db.Query(`
		SELECT full_name FROM github_metadata 
		WHERE last_synced < datetime('now', '-24 hours')
		ORDER BY last_synced ASC
		LIMIT 1000
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	var reposToUpdate []string
	for rows.Next() {
		var fullName string
		if err := rows.Scan(&fullName); err != nil {
			return err
		}
		reposToUpdate = append(reposToUpdate, fullName)
	}
	
	fmt.Printf("Found %d repositories needing update\n", len(reposToUpdate))
	
	// Update sync status
	_, err = db.Exec(`
		UPDATE github_sync_status 
		SET repositories_total = ?
		WHERE id = ?
	`, len(reposToUpdate), syncID)
	if err != nil {
		return err
	}
	
	// Process repositories
	processed := 0
	for _, fullName := range reposToUpdate {
		if dryRun {
			fmt.Printf("Would update: %s\n", fullName)
		} else {
			// Parse owner/repo from full_name
			parts := strings.Split(fullName, "/")
			if len(parts) != 2 {
				log.Printf("Invalid full_name format: %s", fullName)
				continue
			}
			
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			repo, err := client.GetRepositoryMetadata(ctx, parts[0], parts[1])
			cancel()
			
			if err != nil {
				log.Printf("Failed to get metadata for %s: %v", fullName, err)
				continue
			}
			
			if err := upsertRepository(db, repo); err != nil {
				log.Printf("Failed to update %s: %v", fullName, err)
				continue
			}
		}
		
		processed++
		
		// Update progress
		if processed%50 == 0 {
			_, err = db.Exec(`
				UPDATE github_sync_status 
				SET repositories_processed = ?
				WHERE id = ?
			`, processed, syncID)
			if err != nil {
				log.Printf("Failed to update progress: %v", err)
			}
			
			fmt.Printf("Updated %d/%d repositories\n", processed, len(reposToUpdate))
		}
	}
	
	fmt.Printf("Incremental sync completed: %d repositories updated\n", processed)
	return nil
}

func performMetadataSync(db *sql.DB, client *github.GitHubClient, syncID int64) error {
	fmt.Println("Starting metadata-only sync...")
	
	// This would sync only metadata fields (topics, descriptions, etc.)
	// without doing a full repository discovery
	// For now, we'll do a limited incremental sync
	return performIncrementalSync(db, client, syncID)
}

func upsertRepository(db *sql.DB, repo *github.GitHubMetadata) error {
	// Parse owner/repo from full_name
	parts := strings.Split(repo.FullName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid full_name format: %s", repo.FullName)
	}
	
	owner, repoName := parts[0], parts[1]
	
	// Convert topics to JSON
	topicsJSON := "[]"
	if len(repo.Topics) > 0 {
		topicsBytes, err := json.Marshal(repo.Topics)
		if err != nil {
			return err
		}
		topicsJSON = string(topicsBytes)
	}
	
	// Convert languages to JSON
	languagesJSON := "{}"
	if len(repo.Languages) > 0 {
		languagesBytes, err := json.Marshal(repo.Languages)
		if err != nil {
			return err
		}
		languagesJSON = string(languagesBytes)
	}
	
	// Extract license name
	licenseName := ""
	if repo.License != nil {
		licenseName = repo.License.Name
	}
	
	// Upsert repository
	_, err := db.Exec(`
		INSERT OR REPLACE INTO github_metadata (
			name, full_name, owner, repo, description, topics, language, languages,
			is_private, is_archived, is_fork, is_template, visibility, default_branch, license,
			star_count, fork_count, open_issues, size,
			created_at, updated_at, pushed_at, last_synced, sync_count
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, CURRENT_TIMESTAMP, COALESCE((SELECT sync_count FROM github_metadata WHERE full_name = ?), 0) + 1
		)
	`, 
		repo.Name, repo.FullName, owner, repoName, repo.Description, topicsJSON, repo.Language, languagesJSON,
		repo.IsPrivate, repo.IsArchived, repo.IsFork, repo.IsTemplate, repo.Visibility, repo.DefaultBranch, licenseName,
		repo.StarCount, repo.ForkCount, repo.OpenIssues, repo.Size,
		repo.CreatedAt, repo.UpdatedAt, repo.PushedAt, repo.FullName)
	
	return err
}