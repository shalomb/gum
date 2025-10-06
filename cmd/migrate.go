/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/shalomb/gum/internal/database"
)

var (
	migrateRollback bool
	migrateBackup   string
	migrateRestore  string
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate from JSON caches to unified database",
	Long: `Migrate from the old JSON-based caching system to the new unified SQLite database.

This command will:
- Migrate projects.json and project-dirs.json to the database
- Link local projects to GitHub repositories
- Backup original JSON files
- Update the database schema to v2

After migration, all gum commands will use the database instead of JSON files.`,

	Run: func(cmd *cobra.Command, args []string) {
		if migrateRollback {
			if err := rollbackMigration(); err != nil {
				fmt.Fprintf(os.Stderr, "Rollback failed: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if migrateRestore != "" {
			if err := restoreFromBackup(migrateRestore); err != nil {
				fmt.Fprintf(os.Stderr, "Restore failed: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if migrateBackup != "" {
			if err := createBackup(migrateBackup); err != nil {
				fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Default: run migration
		if err := runMigration(); err != nil {
			fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().BoolVar(&migrateRollback, "rollback", false, "Rollback migration and restore JSON caches")
	migrateCmd.Flags().StringVar(&migrateBackup, "backup", "", "Create database backup to specified path")
	migrateCmd.Flags().StringVar(&migrateRestore, "restore", "", "Restore database from backup file")
}

func runMigration() error {
	fmt.Println("Starting migration from JSON caches to unified database...")

	// Get cache directory
	cacheDir := getCacheDir()
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return fmt.Errorf("cache directory does not exist: %s", cacheDir)
	}

	// Initialize database
	db, err := database.New()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Check if migration is needed
	if !needsMigration(cacheDir) {
		fmt.Println("No JSON cache files found - migration not needed")
		return nil
	}

	// Create migrator
	migrator := database.NewMigrator(db)

	// Run migration
	if err := migrator.MigrateFromJSON(cacheDir); err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}

	// Link GitHub repositories
	fmt.Println("Linking projects to GitHub repositories...")
	linked, err := migrator.LinkGitHubRepositories()
	if err != nil {
		fmt.Printf("Warning: Failed to link GitHub repositories: %v\n", err)
	} else {
		fmt.Printf("Linked %d projects to GitHub repositories\n", linked)
	}

	fmt.Println("Migration completed successfully!")
	fmt.Println("You can now use 'gum projects-v2' to test the new system.")
	fmt.Println("Once confirmed working, you can replace 'gum projects' with the new implementation.")

	return nil
}

func rollbackMigration() error {
	fmt.Println("Rolling back migration...")

	cacheDir := getCacheDir()

	// Initialize database
	db, err := database.New()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create migrator
	migrator := database.NewMigrator(db)

	// Rollback migration
	if err := migrator.RollbackMigration(cacheDir); err != nil {
		return fmt.Errorf("rollback failed: %v", err)
	}

	fmt.Println("Migration rolled back successfully!")
	fmt.Println("JSON cache files have been restored.")

	return nil
}

func createBackup(backupPath string) error {
	fmt.Printf("Creating database backup to %s...\n", backupPath)

	db, err := database.New()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create migrator
	migrator := database.NewMigrator(db)

	// Create backup
	if err := migrator.BackupDatabase(backupPath); err != nil {
		return fmt.Errorf("backup failed: %v", err)
	}

	fmt.Println("Database backup created successfully!")
	return nil
}

func restoreFromBackup(backupPath string) error {
	fmt.Printf("Restoring database from %s...\n", backupPath)

	// Check if backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	db, err := database.New()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create migrator
	migrator := database.NewMigrator(db)

	// Restore from backup
	if err := migrator.RestoreDatabase(backupPath); err != nil {
		return fmt.Errorf("restore failed: %v", err)
	}

	fmt.Println("Database restored successfully!")
	return nil
}

func needsMigration(cacheDir string) bool {
	projectsFile := filepath.Join(cacheDir, "projects.json")
	projectDirsFile := filepath.Join(cacheDir, "project-dirs.json")

	// Check if either file exists
	_, projectsExists := os.Stat(projectsFile)
	_, dirsExists := os.Stat(projectDirsFile)

	return projectsExists == nil || dirsExists == nil
}
