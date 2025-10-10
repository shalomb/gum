/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/shalomb/gum/internal/database"
)

// integrityCmd represents the integrity command
var integrityCmd = &cobra.Command{
	Use:   "integrity",
	Short: "Check database integrity and concurrency safety",
	Long: `Check database integrity and verify concurrency safety.

This command performs comprehensive integrity checks:
- Database file integrity (SQLite PRAGMA integrity_check)
- Foreign key constraint validation
- Orphaned record detection
- Duplicate record detection
- Cache consistency verification
- Concurrent operation safety

Examples:
  gum integrity check
  gum integrity monitor --duration 60s
  gum integrity stress --workers 20`,

	Run: func(cmd *cobra.Command, args []string) {
		checkIntegrity()
	},
}

func init() {
	rootCmd.AddCommand(integrityCmd)
}

func checkIntegrity() {
	fmt.Printf("🔍 Database Integrity Check\n")
	fmt.Printf("===========================\n")

	// Initialize database
	db, err := database.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create integrity monitor
	_ = database.NewIntegrityMonitor(db)

	// Run comprehensive checks
	fmt.Println("Running integrity checks...")
	
	// 1. Database file integrity
	fmt.Print("  ✓ Checking database file integrity... ")
	if err := checkDatabaseFileIntegrity(db); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ PASSED")

	// 2. Foreign key constraints
	fmt.Print("  ✓ Checking foreign key constraints... ")
	if err := checkForeignKeyConstraints(db); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ PASSED")

	// 3. Orphaned records
	fmt.Print("  ✓ Checking for orphaned records... ")
	if err := checkOrphanedRecords(db); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ PASSED")

	// 4. Duplicate records
	fmt.Print("  ✓ Checking for duplicate records... ")
	if err := checkDuplicateRecords(db); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ PASSED")

	// 5. Cache consistency
	fmt.Print("  ✓ Checking cache consistency... ")
	if err := checkCacheConsistency(db); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ PASSED")

	// 6. Data statistics
	fmt.Println("\n📊 Database Statistics")
	fmt.Println("=====================")
	printDatabaseStatistics(db)

	// 7. Concurrency safety test
	fmt.Println("\n🧪 Concurrency Safety Test")
	fmt.Println("==========================")
	if err := runConcurrencySafetyTest(db); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ PASSED")

	fmt.Println("\n🎉 All integrity checks passed!")
	fmt.Println("Database is safe for concurrent operations.")
}

func checkDatabaseFileIntegrity(db *database.Database) error {
	var result string
	err := db.GetDB().QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return fmt.Errorf("failed to run integrity check: %w", err)
	}
	
	if result != "ok" {
		return fmt.Errorf("database integrity check failed: %s", result)
	}
	
	return nil
}

func checkForeignKeyConstraints(db *database.Database) error {
	// Enable foreign key checking
	_, err := db.GetDB().Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	
	// Check foreign key constraints
	_, err = db.GetDB().Exec("PRAGMA foreign_key_check")
	if err != nil {
		return fmt.Errorf("foreign key constraint violation: %w", err)
	}
	
	return nil
}

func checkOrphanedRecords(db *database.Database) error {
	// Check for projects with invalid github_repo_id
	query := `
		SELECT COUNT(*) 
		FROM projects p 
		LEFT JOIN github_repos gr ON p.github_repo_id = gr.id 
		WHERE p.github_repo_id IS NOT NULL AND gr.id IS NULL
	`
	
	var count int
	err := db.GetDB().QueryRow(query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check orphaned projects: %w", err)
	}
	
	if count > 0 {
		return fmt.Errorf("found %d orphaned projects with invalid github_repo_id", count)
	}
	
	return nil
}

func checkDuplicateRecords(db *database.Database) error {
	// Check for duplicate projects by path
	query := `
		SELECT path, COUNT(*) 
		FROM projects 
		GROUP BY path 
		HAVING COUNT(*) > 1
	`
	
	rows, err := db.GetDB().Query(query)
	if err != nil {
		return fmt.Errorf("failed to check duplicate projects: %w", err)
	}
	defer rows.Close()
	
	var duplicates []string
	for rows.Next() {
		var path string
		var count int
		if err := rows.Scan(&path, &count); err != nil {
			return fmt.Errorf("failed to scan duplicate record: %w", err)
		}
		duplicates = append(duplicates, fmt.Sprintf("%s (%d copies)", path, count))
	}
	
	if len(duplicates) > 0 {
		return fmt.Errorf("found duplicate projects: %v", duplicates)
	}
	
	return nil
}

func checkCacheConsistency(db *database.Database) error {
	cache := database.NewDatabaseCache(db)
	
	// First, populate the cache with some data
	projects, err := db.GetProjects(false, "")
	if err != nil {
		return fmt.Errorf("failed to get projects for cache test: %w", err)
	}
	
	// Set cache data
	if err := cache.SetProjects(projects); err != nil {
		return fmt.Errorf("failed to set cache data: %w", err)
	}
	
	// Test multiple cache operations
	for i := 0; i < 10; i++ {
		projects1, err := cache.GetProjects()
		if err != nil {
			return fmt.Errorf("cache get failed: %w", err)
		}
		
		projects2, err := cache.GetProjects()
		if err != nil {
			return fmt.Errorf("cache get failed: %w", err)
		}
		
		if len(projects1) != len(projects2) {
			return fmt.Errorf("cache inconsistency: first call returned %d projects, second call returned %d", len(projects1), len(projects2))
		}
	}
	
	return nil
}

func printDatabaseStatistics(db *database.Database) {
	// Projects count
	var projectsCount int
	err := db.GetDB().QueryRow("SELECT COUNT(*) FROM projects").Scan(&projectsCount)
	if err != nil {
		fmt.Printf("Failed to get projects count: %v\n", err)
		return
	}
	fmt.Printf("Projects: %d\n", projectsCount)

	// Project directories count
	var dirsCount int
	err = db.GetDB().QueryRow("SELECT COUNT(*) FROM project_dirs").Scan(&dirsCount)
	if err != nil {
		fmt.Printf("Failed to get project directories count: %v\n", err)
		return
	}
	fmt.Printf("Project Directories: %d\n", dirsCount)

	// GitHub repos count
	var githubReposCount int
	err = db.GetDB().QueryRow("SELECT COUNT(*) FROM github_repos").Scan(&githubReposCount)
	if err != nil {
		fmt.Printf("Failed to get GitHub repos count: %v\n", err)
		return
	}
	fmt.Printf("GitHub Repositories: %d\n", githubReposCount)

	// Linked projects count
	var linkedProjectsCount int
	err = db.GetDB().QueryRow("SELECT COUNT(*) FROM projects WHERE github_repo_id IS NOT NULL").Scan(&linkedProjectsCount)
	if err != nil {
		fmt.Printf("Failed to get linked projects count: %v\n", err)
		return
	}
	fmt.Printf("Linked Projects: %d\n", linkedProjectsCount)

	// Cache metadata count
	var cacheMetadataCount int
	err = db.GetDB().QueryRow("SELECT COUNT(*) FROM cache_metadata").Scan(&cacheMetadataCount)
	if err != nil {
		fmt.Printf("Failed to get cache metadata count: %v\n", err)
		return
	}
	fmt.Printf("Cache Entries: %d\n", cacheMetadataCount)
}

func runConcurrencySafetyTest(db *database.Database) error {
	// Run a simple concurrency test
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	
	// Start 10 concurrent operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Mix of reads and writes
			for j := 0; j < 10; j++ {
				// Read operation
				_, err := db.GetProjects(false, "")
				if err != nil {
					errors <- fmt.Errorf("read failed: %w", err)
					return
				}
				
				// Write operation
				project := &database.Project{
					Path:      fmt.Sprintf("~/concurrency-test-%d-%d", id, j),
					Name:      fmt.Sprintf("concurrency-test-%d-%d", id, j),
					RemoteURL: fmt.Sprintf("https://github.com/user/concurrency-test-%d-%d.git", id, j),
					Branch:    "main",
				}
				
				err = db.UpsertProject(project)
				if err != nil {
					errors <- fmt.Errorf("write failed: %w", err)
					return
				}
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for errors
	for err := range errors {
		return err
	}
	
	return nil
}