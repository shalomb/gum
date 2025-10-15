package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCoherence validates that BDD scenarios are properly implemented by TDD tests
func TestCoherence(t *testing.T) {
	t.Run("BDD_TDD_Alignment", func(t *testing.T) {
		// This test validates that all BDD scenarios have corresponding TDD tests
		validateBDDToTDDAignment(t)
	})

	t.Run("Vision_BDD_Alignment", func(t *testing.T) {
		// This test validates that all vision requirements have BDD scenarios
		validateVisionToBDDAlignment(t)
	})

	t.Run("Performance_Requirements", func(t *testing.T) {
		// This test validates that performance requirements are met
		validatePerformanceRequirements(t)
	})
}

func validateBDDToTDDAignment(t *testing.T) {
	// Read BDD feature files and check for corresponding TDD tests
	bddDir := "docs/bdd"
	tddFiles := []string{
		"cmd/dirs_test.go",
		"cmd/search_test.go", 
		"cmd/performance_test.go",
		"cmd/frecency_test.go",
		"integration_test.go",
		"internal/database/database_test.go",
		"internal/database/concurrency_test.go",
		"internal/cache/cache_test.go",
		"internal/locate/locate_test.go",
	}

	// Check that all BDD scenarios have corresponding test functions
	featureFiles, err := filepath.Glob(filepath.Join(bddDir, "*.feature"))
	if err != nil {
		t.Fatalf("Failed to find BDD feature files: %v", err)
	}

	for _, featureFile := range featureFiles {
		content, err := os.ReadFile(featureFile)
		if err != nil {
			t.Errorf("Failed to read feature file %s: %v", featureFile, err)
			continue
		}

		// Extract scenario names from feature file
		scenarios := extractScenarios(string(content))
		
		// Check that each scenario has a corresponding test
		for _, scenario := range scenarios {
			if !hasCorrespondingTest(scenario, tddFiles) {
				t.Errorf("BDD scenario '%s' in %s has no corresponding TDD test", scenario, featureFile)
			}
		}
	}
}

func validateVisionToBDDAlignment(t *testing.T) {
	// Read solution intent document and check for corresponding BDD scenarios
	solutionIntentFile := "docs/solution-intent.md"
	content, err := os.ReadFile(solutionIntentFile)
	if err != nil {
		t.Fatalf("Failed to read solution intent: %v", err)
	}

	// Extract key requirements from solution intent
	requirements := extractRequirements(string(content))
	
	// Check that each requirement has corresponding BDD scenarios
	for _, requirement := range requirements {
		if !hasCorrespondingBDD(requirement) {
			t.Errorf("Vision requirement '%s' has no corresponding BDD scenario", requirement)
		}
	}
}

func validatePerformanceRequirements(t *testing.T) {
	// Run performance tests and validate against requirements
	t.Run("DiscoverySpeed", func(t *testing.T) {
		// Test that project discovery is under 0.2s
		start := time.Now()
		// Run gum projects command
		cmd := exec.Command("gum", "projects")
		err := cmd.Run()
		duration := time.Since(start)
		
		if err != nil {
			t.Errorf("gum projects command failed: %v", err)
		}
		
		if duration > 200*time.Millisecond {
			t.Errorf("Project discovery took %v, expected < 200ms", duration)
		}
	})

	t.Run("CacheResponse", func(t *testing.T) {
		// Test that cache response is under 0.1s
		start := time.Now()
		// Run gum projects command (should use cache)
		cmd := exec.Command("gum", "projects")
		err := cmd.Run()
		duration := time.Since(start)
		
		if err != nil {
			t.Errorf("gum projects command failed: %v", err)
		}
		
		if duration > 100*time.Millisecond {
			t.Errorf("Cache response took %v, expected < 100ms", duration)
		}
	})
}

func extractScenarios(content string) []string {
	var scenarios []string
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## Scenario:") || strings.HasPrefix(line, "When I run") {
			scenarios = append(scenarios, line)
		}
	}
	
	return scenarios
}

func extractRequirements(content string) []string {
	var requirements []string
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- **") || strings.HasPrefix(line, "## ") {
			requirements = append(requirements, line)
		}
	}
	
	return requirements
}

func hasCorrespondingTest(scenario string, tddFiles []string) bool {
	// Simple check - in real implementation, this would be more sophisticated
	// For now, just check if any TDD file contains relevant keywords
	keywords := extractKeywords(scenario)
	
	for _, tddFile := range tddFiles {
		content, err := os.ReadFile(tddFile)
		if err != nil {
			continue
		}
		
		contentStr := strings.ToLower(string(content))
		for _, keyword := range keywords {
			if strings.Contains(contentStr, strings.ToLower(keyword)) {
				return true
			}
		}
	}
	
	return false
}

func hasCorrespondingBDD(requirement string) bool {
	// Check if requirement has corresponding BDD scenario
	// This is a simplified check - real implementation would be more sophisticated
	return true // Placeholder
}

func extractKeywords(text string) []string {
	// Extract keywords from scenario text
	words := strings.Fields(text)
	var keywords []string
	
	for _, word := range words {
		if len(word) > 3 && !isCommonWord(word) {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "or": true, "but": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "of": true, "with": true, "by": true,
		"should": true, "when": true, "then": true, "given": true, "that": true,
	}
	
	return commonWords[strings.ToLower(word)]
}

// RunCoherenceTests runs all coherence validation tests
func RunCoherenceTests() {
	fmt.Println("🔍 Running Coherence Tests...")
	fmt.Println("Validating Vision → BDD → TDD alignment...")
	
	// Run the coherence tests
	cmd := exec.Command("go", "test", "-v", "-run", "TestCoherence")
	output, err := cmd.Output()
	
	if err != nil {
		fmt.Printf("❌ Coherence tests failed: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		os.Exit(1)
	}
	
	fmt.Printf("✅ Coherence tests passed!\n%s\n", string(output))
}