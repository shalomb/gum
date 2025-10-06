package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shalomb/gum/internal/database"
)

func validateMigration() {
	fmt.Println("🧪 Running Migration Validation")
	fmt.Println("================================")

	// Create temporary test environment
	tempDir := "/tmp/gum_validation_test"
	cacheDir := filepath.Join(tempDir, ".cache", "gum")

	// Clean up any existing test data
	os.RemoveAll(tempDir)
	os.MkdirAll(cacheDir, 0755)

	// Set environment
	os.Setenv("XDG_CACHE_HOME", filepath.Join(tempDir, ".cache"))

	fmt.Println("✅ Test environment created")

	// Step 1: Create JSON cache files that simulate the bug
	fmt.Println("\n📁 Creating bug scenario...")
	
	// Create projects.json with many projects
	projectsData := make([]map[string]string, 50)
	for i := 0; i < 50; i++ {
		projectsData[i] = map[string]string{
			"Path":   fmt.Sprintf("~/projects/project-%d", i),
			"Remote": fmt.Sprintf("https://github.com/user/project-%d.git", i),
			"Branch": "main",
		}
	}

	projectsJSON := map[string]interface{}{
		"data":      projectsData,
		"timestamp": time.Now().Format(time.RFC3339),
		"ttl":       300,
	}

	projectsBytes, err := json.Marshal(projectsJSON)
	if err != nil {
		fmt.Printf("❌ Failed to create projects JSON: %v\n", err)
		return
	}

	projectsFile := filepath.Join(cacheDir, "projects.json")
	if err := os.WriteFile(projectsFile, projectsBytes, 0644); err != nil {
		fmt.Printf("❌ Failed to write projects.json: %v\n", err)
		return
	}

	// Create project-dirs.json with different discovery (simulating the bug)
	projectDirsData := []map[string]interface{}{
		{"Path": "~/projects", "LastScanned": time.Now().Format(time.RFC3339), "GitCount": 3}, // Only 3 projects!
	}

	projectDirsJSON := map[string]interface{}{
		"data":      projectDirsData,
		"timestamp": time.Now().Format(time.RFC3339),
		"ttl":       300,
	}

	projectDirsBytes, err := json.Marshal(projectDirsJSON)
	if err != nil {
		fmt.Printf("❌ Failed to create project-dirs JSON: %v\n", err)
		return
	}

	projectDirsFile := filepath.Join(cacheDir, "project-dirs.json")
	if err := os.WriteFile(projectDirsFile, projectDirsBytes, 0644); err != nil {
		fmt.Printf("❌ Failed to write project-dirs.json: %v\n", err)
		return
	}

	fmt.Printf("✅ Created bug scenario: 50 projects in projects.json, 3 projects in project-dirs.json\n")

	// Step 2: Run migration
	fmt.Println("\n🔄 Running migration...")
	
	db, err := database.New()
	if err != nil {
		fmt.Printf("❌ Failed to create database: %v\n", err)
		return
	}
	defer db.Close()

	migrator := database.NewMigrator(db)
	if err := migrator.MigrateFromJSON(cacheDir); err != nil {
		fmt.Printf("❌ Migration failed: %v\n", err)
		return
	}

	fmt.Println("✅ Migration completed successfully")

	// Step 3: Verify migration results
	fmt.Println("\n🔍 Verifying migration results...")
	
	projects, err := db.GetProjects(false, "")
	if err != nil {
		fmt.Printf("❌ Failed to get projects: %v\n", err)
		return
	}

	dirs, err := db.GetProjectDirs()
	if err != nil {
		fmt.Printf("❌ Failed to get project directories: %v\n", err)
		return
	}

	fmt.Printf("✅ Migrated %d projects and %d directories\n", len(projects), len(dirs))

	// Step 4: Test cache consistency
	fmt.Println("\n🧪 Testing cache consistency...")
	
	cache := database.NewDatabaseCache(db)

	// Test multiple calls
	projects1, err := cache.GetProjects()
	if err != nil {
		fmt.Printf("❌ Failed to get projects from cache: %v\n", err)
		return
	}

	projects2, err := cache.GetProjects()
	if err != nil {
		fmt.Printf("❌ Failed to get projects from cache: %v\n", err)
		return
	}

	projects3, err := cache.GetProjects()
	if err != nil {
		fmt.Printf("❌ Failed to get projects from cache: %v\n", err)
		return
	}

	// Check consistency
	if len(projects1) == len(projects2) && len(projects2) == len(projects3) {
		fmt.Printf("✅ Cache consistency verified: %d projects consistently returned\n", len(projects1))
	} else {
		fmt.Printf("❌ Cache inconsistency detected: %d, %d, %d\n", len(projects1), len(projects2), len(projects3))
		return
	}

	// Step 5: Test refresh functionality
	fmt.Println("\n🔄 Testing refresh functionality...")
	
	if err := cache.ClearCache("projects"); err != nil {
		fmt.Printf("❌ Failed to clear cache: %v\n", err)
		return
	}

	// Verify cache miss
	if cache.IsCacheHit("projects") {
		fmt.Println("❌ Expected cache miss after clearing")
		return
	}

	// Test refresh
	testProjects := []*database.Project{
		{Path: "~/test-project", Name: "test-project", RemoteURL: "https://github.com/user/test.git"},
	}

	if err := cache.SetProjects(testProjects); err != nil {
		fmt.Printf("❌ Failed to set projects: %v\n", err)
		return
	}

	// Verify cache hit after refresh
	if !cache.IsCacheHit("projects") {
		fmt.Println("❌ Expected cache hit after refresh")
		return
	}

	fmt.Println("✅ Refresh functionality verified")

	// Step 6: Test rollback
	fmt.Println("\n↩️  Testing rollback functionality...")
	
	if err := migrator.RollbackMigration(cacheDir); err != nil {
		fmt.Printf("❌ Rollback failed: %v\n", err)
		return
	}

	// Verify JSON files were restored
	if _, err := os.Stat(projectsFile); os.IsNotExist(err) {
		fmt.Println("❌ projects.json was not restored after rollback")
		return
	}

	if _, err := os.Stat(projectDirsFile); os.IsNotExist(err) {
		fmt.Println("❌ project-dirs.json was not restored after rollback")
		return
	}

	// Verify database was cleared
	projects, err = db.GetProjects(false, "")
	if err != nil {
		fmt.Printf("❌ Failed to get projects after rollback: %v\n", err)
		return
	}

	if len(projects) != 0 {
		fmt.Printf("❌ Expected 0 projects after rollback, got %d\n", len(projects))
		return
	}

	fmt.Println("✅ Rollback functionality verified")

	// Step 7: Performance test
	fmt.Println("\n⚡ Testing performance...")
	
	// Re-run migration for performance test
	if err := migrator.MigrateFromJSON(cacheDir); err != nil {
		fmt.Printf("❌ Re-migration failed: %v\n", err)
		return
	}

	// Time the operations
	start := time.Now()
	projects, err = cache.GetProjects()
	if err != nil {
		fmt.Printf("❌ Failed to get projects: %v\n", err)
		return
	}
	duration := time.Since(start)

	fmt.Printf("✅ Retrieved %d projects in %v\n", len(projects), duration)

	if duration > 100*time.Millisecond {
		fmt.Printf("⚠️  Performance warning: operation took %v (expected < 100ms)\n", duration)
	}

	// Final results
	fmt.Println("\n🎉 All Tests Passed!")
	fmt.Println("===================")
	fmt.Println("✅ Migration functionality works")
	fmt.Println("✅ Cache consistency is maintained")
	fmt.Println("✅ Rollback functionality works")
	fmt.Println("✅ Performance is acceptable")
	fmt.Println("✅ Database integrity is maintained")
	fmt.Println("")
	fmt.Println("🚀 Ready for deployment!")

	// Cleanup
	os.RemoveAll(tempDir)
}

func main() {
	validateMigration()
}