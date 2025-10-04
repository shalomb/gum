package cmd

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Find the gum binary
	gumPath, err := filepath.Abs("../gum")
	if err != nil {
		t.Fatalf("Failed to find gum binary: %v", err)
	}
	
	// Test basic version output
	cmd := exec.Command(gumPath, "version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	outputStr := string(output)
	
	// Check that all required fields are present
	requiredFields := []string{
		"gum version",
		"Git commit:",
		"Build date:",
		"Go version:",
		"OS/Arch:",
	}

	for _, field := range requiredFields {
		if !strings.Contains(outputStr, field) {
			t.Errorf("version output missing required field: %s", field)
		}
	}

	// Check that version is not empty
	if !strings.Contains(outputStr, "version ") {
		t.Error("version output should contain version information")
	}
}

func TestVersionCommandVerbose(t *testing.T) {
	// Find the gum binary
	gumPath, err := filepath.Abs("../gum")
	if err != nil {
		t.Fatalf("Failed to find gum binary: %v", err)
	}
	
	// Test verbose version output
	cmd := exec.Command(gumPath, "version", "--verbose")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("version --verbose command failed: %v", err)
	}

	outputStr := string(output)
	
	// Check that verbose fields are present
	verboseFields := []string{
		"Runtime:",
		"NumCPU:",
		"GOMAXPROCS:",
	}

	for _, field := range verboseFields {
		if !strings.Contains(outputStr, field) {
			t.Errorf("verbose version output missing field: %s", field)
		}
	}

	// Check that basic fields are still present
	basicFields := []string{
		"gum version",
		"Git commit:",
		"Build date:",
		"Go version:",
		"OS/Arch:",
	}

	for _, field := range basicFields {
		if !strings.Contains(outputStr, field) {
			t.Errorf("verbose version output missing basic field: %s", field)
		}
	}
}

func TestVersionCommandShortFlag(t *testing.T) {
	// Find the gum binary
	gumPath, err := filepath.Abs("../gum")
	if err != nil {
		t.Fatalf("Failed to find gum binary: %v", err)
	}
	
	// Test short verbose flag
	cmd := exec.Command(gumPath, "version", "-v")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("version -v command failed: %v", err)
	}

	outputStr := string(output)
	
	// Should include verbose information
	if !strings.Contains(outputStr, "Runtime:") {
		t.Error("version -v should include verbose information")
	}
}

func TestVersionCommandHelp(t *testing.T) {
	// Find the gum binary
	gumPath, err := filepath.Abs("../gum")
	if err != nil {
		t.Fatalf("Failed to find gum binary: %v", err)
	}
	
	// Test help output
	cmd := exec.Command(gumPath, "version", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("version --help command failed: %v", err)
	}

	outputStr := string(output)
	
	// Check that help contains expected information
	expectedHelp := []string{
		"Show detailed version information",
		"--verbose",
		"-v",
		"Show verbose runtime information",
	}

	for _, field := range expectedHelp {
		if !strings.Contains(outputStr, field) {
			t.Errorf("version help missing expected field: %s", field)
		}
	}
}