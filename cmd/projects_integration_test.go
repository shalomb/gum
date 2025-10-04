package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProjectDiscoveryIntegration tests the complete project discovery workflow
func TestProjectDiscoveryIntegration(t *testing.T) {
	// Create temporary test environment
	testHome := t.TempDir()
	testConfigDir := filepath.Join(testHome, ".config")
	testGumConfigDir := filepath.Join(testConfigDir, "gum")
	
	// Set up test environment
	os.Setenv("HOME", testHome)
	os.Setenv("XDG_CONFIG_HOME", testConfigDir)
	defer os.Unsetenv("HOME")
	defer os.Unsetenv("XDG_CONFIG_HOME")
	
	// Create test directories with Git repositories
	testDirs := []string{
		filepath.Join(testHome, "projects"),
		filepath.Join(testHome, "oneTakeda"),
		filepath.Join(testHome, "projects-local"),
		filepath.Join(testHome, "empty-dir"),
	}
	
	for _, dir := range testDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}
	
	// Create Git repositories in some directories
	gitRepos := []string{
		filepath.Join(testHome, "projects", "repo1"),
		filepath.Join(testHome, "projects", "repo2"),
		filepath.Join(testHome, "oneTakeda", "repo3"),
		filepath.Join(testHome, "projects-local", "repo4"),
	}
	
	for _, repo := range gitRepos {
		if err := os.MkdirAll(repo, 0755); err != nil {
			t.Fatalf("Failed to create repo directory %s: %v", repo, err)
		}
		if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
			t.Fatalf("Failed to create .git directory %s: %v", repo, err)
		}
	}
	
	t.Run("AutoDiscovery", func(t *testing.T) {
		// Test auto-discovery without config
		dirs := getProjectDirs()
		
		// Should discover directories with Git repos
		expectedDirs := []string{
			filepath.Join(testHome, "projects"),
			filepath.Join(testHome, "oneTakeda"),
			filepath.Join(testHome, "projects-local"),
		}
		
		if len(dirs) < len(expectedDirs) {
			t.Errorf("Expected at least %d directories, got %d", len(expectedDirs), len(dirs))
		}
		
		// Check that empty directory is not included
		for _, dir := range dirs {
			if dir == filepath.Join(testHome, "empty-dir") {
				t.Errorf("Empty directory should not be included: %s", dir)
			}
		}
		
		// Verify we found the expected directories
		foundCount := 0
		for _, expectedDir := range expectedDirs {
			for _, dir := range dirs {
				if dir == expectedDir {
					foundCount++
					break
				}
			}
		}
		
		if foundCount < len(expectedDirs) {
			t.Errorf("Expected to find %d directories, found %d", len(expectedDirs), foundCount)
		}
	})
	
	t.Run("ConfigStubGeneration", func(t *testing.T) {
		// Clear any existing config
		os.RemoveAll(testGumConfigDir)
		
		// Run auto-discovery to trigger config stub generation
		dirs := getProjectDirs()
		
		// Verify we got some directories
		if len(dirs) == 0 {
			t.Error("Should have discovered some directories")
		}
		
		// Check if config stub was generated
		configFile := filepath.Join(testGumConfigDir, "config.yaml")
		if _, err := os.Stat(configFile); err != nil {
			t.Errorf("Config stub should have been generated: %v", err)
		}
		
		// Verify config content
		content, err := os.ReadFile(configFile)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}
		
		configStr := string(content)
		if !contains(configStr, "projects:") {
			t.Error("Config should contain 'projects:' section")
		}
		if !contains(configStr, "# Gum Configuration") {
			t.Error("Config should contain header comment")
		}
		if !contains(configStr, "# Additional directories you can add:") {
			t.Error("Config should contain helpful comments")
		}
	})
	
	t.Run("YAMLConfigOverride", func(t *testing.T) {
		// Create custom YAML config
		configContent := `projects:
  - ` + testHome + `/custom-projects
  - ` + testHome + `/work-repos
`
		
		if err := os.MkdirAll(testGumConfigDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}
		
		configFile := filepath.Join(testGumConfigDir, "config.yaml")
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		
		// Create custom directories with Git repos
		customDirs := []string{
			filepath.Join(testHome, "custom-projects"),
			filepath.Join(testHome, "work-repos"),
		}
		
		for _, dir := range customDirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create custom directory %s: %v", dir, err)
			}
			// Create a Git repo in each
			repoDir := filepath.Join(dir, "test-repo")
			if err := os.MkdirAll(repoDir, 0755); err != nil {
				t.Fatalf("Failed to create repo directory: %v", err)
			}
			if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0755); err != nil {
				t.Fatalf("Failed to create .git directory: %v", err)
			}
		}
		
		// Test that YAML config is used
		dirs := getProjectDirs()
		
		// Should include custom directories
		foundCustom := false
		for _, dir := range dirs {
			if dir == filepath.Join(testHome, "custom-projects") || 
			   dir == filepath.Join(testHome, "work-repos") {
				foundCustom = true
				break
			}
		}
		
		if !foundCustom {
			t.Error("YAML config should override auto-discovery")
		}
	})
	
	t.Run("LegacyProjectsDirsList", func(t *testing.T) {
		// Remove YAML config
		os.RemoveAll(testGumConfigDir)
		
		
		// Create legacy directories with Git repos
		legacyDirs := []string{
			filepath.Join(testHome, "legacy-projects"),
			filepath.Join(testHome, "old-repos"),
		}
		
		for _, dir := range legacyDirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create legacy directory %s: %v", dir, err)
			}
			// Create a Git repo in each
			repoDir := filepath.Join(dir, "legacy-repo")
			if err := os.MkdirAll(repoDir, 0755); err != nil {
				t.Fatalf("Failed to create repo directory: %v", err)
			}
			if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0755); err != nil {
				t.Fatalf("Failed to create .git directory: %v", err)
			}
		}
		
		// Test that legacy config is used
		dirs := getProjectDirs()
		
		// Should include legacy directories
		foundLegacy := false
		for _, dir := range dirs {
			if dir == filepath.Join(testHome, "legacy-projects") || 
			   dir == filepath.Join(testHome, "old-repos") {
				foundLegacy = true
				break
			}
		}
		
		if !foundLegacy {
			t.Error("Auto-discovery should find legacy directories")
		}
	})
}

// TestConfigStubContent tests the content of generated config stubs
func TestConfigStubContent(t *testing.T) {
	testHome := "/home/testuser"
	discoveredDirs := []string{
		"/home/testuser/projects",
		"/home/testuser/oneTakeda",
		"/home/testuser/projects-local",
	}
	
	config := generateConfigStub(testHome, discoveredDirs)
	
	// Test that config contains expected sections
	if !contains(config, "# Gum Configuration") {
		t.Error("Config should contain header comment")
	}
	
	if !contains(config, "projects:") {
		t.Error("Config should contain projects section")
	}
	
	if !contains(config, "~/projects") {
		t.Error("Config should contain ~/projects")
	}
	
	if !contains(config, "~/oneTakeda") {
		t.Error("Config should contain ~/oneTakeda")
	}
	
	if !contains(config, "~/projects-local") {
		t.Error("Config should contain ~/projects-local")
	}
	
	if !contains(config, "# Additional directories you can add:") {
		t.Error("Config should contain helpful comments")
	}
	
	if !contains(config, "# Note: Directories with 0 Git repositories will be ignored") {
		t.Error("Config should contain usage notes")
	}
}

// TestSmartDiscovery tests the smart discovery algorithm
func TestSmartDiscovery(t *testing.T) {
	testHome := t.TempDir()
	
	// Create test directories
	testDirs := []string{
		filepath.Join(testHome, "projects"),
		filepath.Join(testHome, "oneTakeda"),
		filepath.Join(testHome, "projects-local"),
		filepath.Join(testHome, "empty-dir"),
		filepath.Join(testHome, "nonexistent"),
	}
	
	for _, dir := range testDirs[:3] { // Create first 3 directories
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}
	
	// Create Git repositories
	gitRepos := []string{
		filepath.Join(testHome, "projects", "repo1"),
		filepath.Join(testHome, "projects", "repo2"),
		filepath.Join(testHome, "oneTakeda", "repo3"),
	}
	
	for _, repo := range gitRepos {
		if err := os.MkdirAll(repo, 0755); err != nil {
			t.Fatalf("Failed to create repo directory %s: %v", repo, err)
		}
		if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
			t.Fatalf("Failed to create .git directory %s: %v", repo, err)
		}
	}
	
	// Test smart discovery
	dirs := smartDiscoverProjectDirs(testHome)
	
	// Should find directories with Git repos
	expectedDirs := []string{
		filepath.Join(testHome, "projects"),
		filepath.Join(testHome, "oneTakeda"),
	}
	
	if len(dirs) < len(expectedDirs) {
		t.Errorf("Expected at least %d directories, got %d", len(expectedDirs), len(dirs))
	}
	
	// Check that empty directory is not included
	for _, dir := range dirs {
		if dir == filepath.Join(testHome, "empty-dir") {
			t.Errorf("Empty directory should not be included: %s", dir)
		}
		if dir == filepath.Join(testHome, "nonexistent") {
			t.Errorf("Nonexistent directory should not be included: %s", dir)
		}
	}
}

// TestGitRepoCounting tests the Git repository counting function
func TestGitRepoCounting(t *testing.T) {
	testDir := t.TempDir()
	
	// Create test structure
	testRepos := []string{
		filepath.Join(testDir, "repo1"),
		filepath.Join(testDir, "repo2"),
		filepath.Join(testDir, "subdir", "repo3"),
		filepath.Join(testDir, "subdir", "repo4"),
	}
	
	for _, repo := range testRepos {
		if err := os.MkdirAll(repo, 0755); err != nil {
			t.Fatalf("Failed to create repo directory %s: %v", repo, err)
		}
		if err := os.MkdirAll(filepath.Join(repo, ".git"), 0755); err != nil {
			t.Fatalf("Failed to create .git directory %s: %v", repo, err)
		}
	}
	
	// Test counting
	count := countGitReposInDir(testDir)
	expectedCount := 4
	
	if count != expectedCount {
		t.Errorf("Expected %d Git repositories, got %d", expectedCount, count)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     contains(s[1:], substr)))
}