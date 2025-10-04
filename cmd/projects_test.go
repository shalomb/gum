package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shalomb/gum/internal/cache"
)

func TestProjectsCommand(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
		if originalConfigHome != "" {
			os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	os.Setenv("XDG_CACHE_HOME", tempDir)
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	t.Run("GetProjectDirs with default directories", func(t *testing.T) {
		// Test default directory discovery
		dirs := getProjectDirs()
		
		// Should include at least some directories
		if len(dirs) == 0 {
			t.Error("Expected at least some directories")
		}

		// Check that we get some directories (exact directories depend on environment)
		home := os.Getenv("HOME")
		foundHomeDir := false
		for _, dir := range dirs {
			if filepath.HasPrefix(dir, home) {
				foundHomeDir = true
				break
			}
		}
		if !foundHomeDir {
			t.Errorf("Expected at least one directory under home (%s), got %v", home, dirs)
		}
	})


	t.Run("FindGitProjects", func(t *testing.T) {
		// Create a test directory structure
		testDir := filepath.Join(tempDir, "test-projects")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Create a mock .git directory
		gitDir := filepath.Join(testDir, "test-repo", ".git")
		err = os.MkdirAll(gitDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create .git directory: %v", err)
		}

		// Test finding Git projects
		projects := findGitProjects([]string{testDir})
		
		if len(projects) != 1 {
			t.Errorf("Expected 1 project, got %d", len(projects))
		}

		if projects[0].Path == "" {
			t.Error("Project path should not be empty")
		}
	})

	t.Run("GetProjectInfo", func(t *testing.T) {
		// Create a test project directory
		testDir := filepath.Join(tempDir, "test-project")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Test getting project info
		project := getProjectInfo(testDir)
		
		if project.Path == "" {
			t.Error("Project path should not be empty")
		}

		// Should convert absolute path to ~ notation when under home
		home := os.Getenv("HOME")
		if filepath.HasPrefix(testDir, home) {
			if !filepath.HasPrefix(project.Path, "~") {
				t.Errorf("Project path should start with ~ for home directory, got %s", project.Path)
			}
		}
	})

	t.Run("Output Formats", func(t *testing.T) {
		// Create test projects
		projects := []Project{
			{
				Path:   "~/test/project1",
				Remote: "https://github.com/test/project1.git",
				Branch: "",
			},
			{
				Path:   "~/test/project2",
				Remote: "",
				Branch: "main",
			},
		}

		// Test default format
		t.Run("Default Format", func(t *testing.T) {
			// This is hard to test without capturing stdout,
			// but we can at least ensure the function doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("outputProjectsDefaultFormat panicked: %v", r)
				}
			}()
			outputProjectsDefaultFormat(projects)
		})

		// Test FZF format
		t.Run("FZF Format", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("outputProjectsFzfFormat panicked: %v", r)
				}
			}()
			outputProjectsFzfFormat(projects)
		})

		// Test JSON format
		t.Run("JSON Format", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("outputProjectsJsonFormat panicked: %v", r)
				}
			}()
			outputProjectsJsonFormat(projects)
		})

		// Test simple format
		t.Run("Simple Format", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("outputProjectsSimpleFormat panicked: %v", r)
				}
			}()
			outputProjectsSimpleFormat(projects)
		})
	})

	t.Run("Similarity Functions", func(t *testing.T) {
		projects := []Project{
			{Path: "~/test/project1"},
			{Path: "~/test/project2"},
			{Path: "~/other/project"},
		}

		// Test similarity sorting
		sorted := sortProjectsBySimilarity(projects, "project1")
		
		if len(sorted) != len(projects) {
			t.Errorf("Expected %d projects, got %d", len(projects), len(sorted))
		}

		// The first project should be the most similar
		if sorted[0].Path != "~/test/project1" {
			t.Errorf("Expected most similar project to be first, got %s", sorted[0].Path)
		}

		// Test Levenshtein distance
		distance := levenshteinDistance("test", "test")
		if distance != 0 {
			t.Errorf("Expected distance 0 for identical strings, got %d", distance)
		}

		distance = levenshteinDistance("test", "testing")
		if distance != 3 {
			t.Errorf("Expected distance 3 for 'test' vs 'testing', got %d", distance)
		}
	})

	t.Run("Min3 Function", func(t *testing.T) {
		tests := []struct {
			a, b, c, expected int
		}{
			{1, 2, 3, 1},
			{3, 1, 2, 1},
			{2, 3, 1, 1},
			{1, 1, 1, 1},
			{5, 3, 4, 3},
		}

		for _, tt := range tests {
			result := min3(tt.a, tt.b, tt.c)
			if result != tt.expected {
				t.Errorf("min3(%d, %d, %d) = %d, expected %d", tt.a, tt.b, tt.c, result, tt.expected)
			}
		}
	})
}

func TestProjectsCommandIntegration(t *testing.T) {
	// Set up test environment
	tempDir := t.TempDir()
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	os.Setenv("XDG_CACHE_HOME", tempDir)

	t.Run("DoListProjects with refresh", func(t *testing.T) {
		// Test the main function with refresh flag
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("doListProjects panicked: %v", r)
			}
		}()

		// This will try to scan actual directories, which might not exist
		// but should not panic
		doListProjects("default", true)
	})

	t.Run("DoListProjects with different formats", func(t *testing.T) {
		formats := []string{"default", "fzf", "json", "simple"}
		
		for _, format := range formats {
			t.Run(format, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("doListProjects with format %s panicked: %v", format, r)
					}
				}()

				doListProjects(format, false)
			})
		}
	})
}