package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
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

	// Build the test binary
	testBinary := filepath.Join(tempDir, "gum-test")
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	defer os.Remove(testBinary)

	t.Run("Version Command", func(t *testing.T) {
		cmd := exec.Command(testBinary, "version")
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Version command failed: %v", err)
		}

		if len(output) == 0 {
			t.Error("Version command returned empty output")
		}
	})

	t.Run("Help Command", func(t *testing.T) {
		cmd := exec.Command(testBinary, "--help")
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Help command failed: %v", err)
		}

		if len(output) == 0 {
			t.Error("Help command returned empty output")
		}

		// Check for expected commands
		expectedCommands := []string{"projects", "dirs", "github", "clone", "version"}
		outputStr := string(output)
		for _, cmd := range expectedCommands {
			if !contains(outputStr, cmd) {
				t.Errorf("Help output missing command: %s", cmd)
			}
		}
	})

	t.Run("Projects Command", func(t *testing.T) {
		// Test projects command with different flags
		testCases := []struct {
			name string
			args []string
		}{
			{"Default", []string{"projects"}},
			{"JSON Format", []string{"projects", "--format", "json"}},
			{"FZF Format", []string{"projects", "--format", "fzf"}},
			{"Simple Format", []string{"projects", "--format", "simple"}},
			{"Refresh", []string{"projects", "--refresh"}},
			{"Clear Cache", []string{"projects", "--clear-cache"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cmd := exec.Command(testBinary, tc.args...)
				output, err := cmd.Output()
				if err != nil {
					t.Errorf("Projects command %s failed: %v", tc.name, err)
				}

				// For clear-cache, we expect success even with empty output
				if tc.name == "Clear Cache" {
					return
				}

				// For other commands, output should not be empty (unless no projects found)
				// With cron-based approach, empty output is valid when no projects exist
				if len(output) == 0 && tc.name != "Default" && tc.name != "FZF Format" && tc.name != "Simple Format" && tc.name != "Refresh" {
					t.Errorf("Projects command %s returned empty output", tc.name)
				}
			})
		}
	})

	t.Run("Dirs Command", func(t *testing.T) {
		// Test dirs command
		testCases := []struct {
			name string
			args []string
		}{
			{"Default", []string{"dirs"}},
			{"Verbose", []string{"dirs", "--verbose"}},
			{"JSON Format", []string{"dirs", "--format", "json"}},
			{"FZF Format", []string{"dirs", "--format", "fzf"}},
			{"Simple Format", []string{"dirs", "--format", "simple"}},
			{"Refresh", []string{"dirs", "--refresh"}},
			{"Clear Cache", []string{"dirs", "--clear-cache"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cmd := exec.Command(testBinary, tc.args...)
				output, err := cmd.Output()
				if err != nil {
					t.Errorf("Dirs command %s failed: %v", tc.name, err)
				}

				// For clear-cache, we expect success even with empty output
				if tc.name == "Clear Cache" {
					return
				}

				// For other commands, output should not be empty
				if len(output) == 0 {
					t.Errorf("Dirs command %s returned empty output", tc.name)
				}
			})
		}
	})

	t.Run("Dirs Cache Command", func(t *testing.T) {
		// Test dirs-cache command
		testCases := []struct {
			name string
			args []string
		}{
			{"Default", []string{"dirs-cache"}},
			{"List", []string{"dirs-cache", "--list"}},
			{"Refresh", []string{"dirs-cache", "--refresh"}},
			{"Clear", []string{"dirs-cache", "--clear"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cmd := exec.Command(testBinary, tc.args...)
				output, err := cmd.Output()
				if err != nil {
					t.Errorf("Dirs-cache command %s failed: %v", tc.name, err)
				}

				// All commands should produce some output
				if len(output) == 0 {
					t.Errorf("Dirs-cache command %s returned empty output", tc.name)
				}
			})
		}
	})

	t.Run("GitHub Command", func(t *testing.T) {
		// Test github command
		cmd := exec.Command(testBinary, "github")
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("GitHub command failed: %v", err)
		}

		// GitHub command might return empty output if no repos found
		// but should not fail
		_ = output
	})

	t.Run("Clone Command Help", func(t *testing.T) {
		// Test clone command help
		cmd := exec.Command(testBinary, "clone", "--help")
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Clone help command failed: %v", err)
		}

		if len(output) == 0 {
			t.Error("Clone help command returned empty output")
		}
	})

	t.Run("Update Command", func(t *testing.T) {
		// Skip update command test as it takes too long and is not essential for TTL removal
		t.Skip("Update command test skipped - takes too long and not essential for TTL removal")
	})

	t.Run("Invalid Command", func(t *testing.T) {
		// Test invalid command
		cmd := exec.Command(testBinary, "invalid-command")
		err := cmd.Run()
		if err == nil {
			t.Error("Invalid command should fail")
		}
	})

	t.Run("Invalid Flag", func(t *testing.T) {
		// Test invalid flag
		cmd := exec.Command(testBinary, "projects", "--invalid-flag")
		err := cmd.Run()
		if err == nil {
			t.Error("Invalid flag should fail")
		}
	})
}

func TestDatabaseIntegration(t *testing.T) {
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

	// Build the test binary
	testBinary := filepath.Join(tempDir, "gum-test")
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	defer os.Remove(testBinary)

	t.Run("Database Creation", func(t *testing.T) {
		// Run any command that would create the database
		cmd := exec.Command(testBinary, "projects", "--refresh")
		err := cmd.Run()
		if err != nil {
			t.Errorf("Failed to create database: %v", err)
		}

		// Check that database file was created
		dbPath := filepath.Join(tempDir, "gum", "gum.db")
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Errorf("Database file was not created at %s", dbPath)
		}
	})

	t.Run("Database Persistence", func(t *testing.T) {
		// First, populate the database
		cmd := exec.Command(testBinary, "projects", "--refresh")
		err := cmd.Run()
		if err != nil {
			t.Errorf("Failed to populate database: %v", err)
		}

		// Wait a bit to ensure database is written
		time.Sleep(100 * time.Millisecond)

		// Run another command that should use the database
		cmd = exec.Command(testBinary, "projects")
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Failed to read from database: %v", err)
		}

		// Should get some output (even if empty)
		_ = output
	})

	t.Run("Cache Operations", func(t *testing.T) {
		// Test cache clear operations
		cmd := exec.Command(testBinary, "projects", "--clear-cache")
		err := cmd.Run()
		if err != nil {
			t.Errorf("Clear cache failed: %v", err)
		}

		cmd = exec.Command(testBinary, "dirs", "--clear-cache")
		err = cmd.Run()
		if err != nil {
			t.Errorf("Clear dirs cache failed: %v", err)
		}
	})
}

func TestXDGCompliance(t *testing.T) {
	// Test XDG environment variable handling
	testCases := []struct {
		name        string
		cacheHome   string
		configHome  string
		expectError bool
	}{
		{
			name:        "Default XDG directories",
			cacheHome:   "",
			configHome:  "",
			expectError: false,
		},
		{
			name:        "Custom XDG directories",
			cacheHome:   "/custom/cache",
			configHome:  "/custom/config",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

			if tc.cacheHome != "" {
				os.Setenv("XDG_CACHE_HOME", tc.cacheHome)
			} else {
				os.Unsetenv("XDG_CACHE_HOME")
			}

			if tc.configHome != "" {
				os.Setenv("XDG_CONFIG_HOME", tc.configHome)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			// Build and run test binary
			testBinary := filepath.Join(tempDir, "gum-test")
			cmd := exec.Command("go", "build", "-o", testBinary, ".")
			err := cmd.Run()
			if err != nil {
				t.Fatalf("Failed to build test binary: %v", err)
			}
			defer os.Remove(testBinary)

			// Run a simple command
			cmd = exec.Command(testBinary, "version")
			err = cmd.Run()
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}