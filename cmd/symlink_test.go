package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGitProjectsInDir_Symlinks(t *testing.T) {
	// Create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "gum-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	realRepo := filepath.Join(tmpDir, "real-repo")
	os.Mkdir(realRepo, 0755)
	os.Mkdir(filepath.Join(realRepo, ".git"), 0755)

	symlinkDir := filepath.Join(tmpDir, "symlink-dir")
	if err := os.Symlink(realRepo, symlinkDir); err != nil {
		t.Fatal(err)
	}

	// Run the function being tested
	projects := findGitProjectsInDir(symlinkDir)

	// Verify that the project was found despite being in a symlinked directory
	found := false
	for _, p := range projects {
		if p.Path == realRepo {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find project at %s, but did not", realRepo)
	}
}
