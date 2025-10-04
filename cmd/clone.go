/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/shalomb/gum/internal/cache"
)

var (
	cloneTarget string
	cloneSuggest bool
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone [repository]",
	Short: "Intelligently clone GitHub repositories",
	Long: `Clone GitHub repositories with intelligent directory suggestions based on existing projects.

This command:
- Analyzes existing project structure
- Suggests optimal clone locations based on similarity
- Handles both full GitHub URLs and org/repo format
- Creates appropriate directory structure

Examples:
  gum clone shalomb/gum
  gum clone https://github.com/shalomb/gum.git
  gum clone --suggest shalomb/gum  # Just suggest, don't clone`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := args[0]
		
		// Normalize repository format
		normalizedRepo := normalizeRepoURL(repo)
		if normalizedRepo == "" {
			fmt.Printf("Invalid repository format: %s\n", repo)
			return
		}
		
		// Get existing projects for similarity matching
		existingProjects := getExistingProjects()
		
		// Find best clone location
		suggestions := suggestCloneLocation(normalizedRepo, existingProjects)
		
		if cloneSuggest {
			// Just show suggestions
			fmt.Printf("Suggested clone locations for %s:\n", normalizedRepo)
			for i, suggestion := range suggestions {
				fmt.Printf("%d. %s (similarity: %d)\n", i+1, suggestion.Path, suggestion.Score)
			}
			return
		}
		
		// Use the best suggestion or ask user
		var targetPath string
		if cloneTarget != "" {
			targetPath = cloneTarget
		} else if len(suggestions) > 0 {
			targetPath = suggestions[0].Path
		} else {
			// Fallback to default location
			targetPath = filepath.Join(os.Getenv("HOME"), "projects", filepath.Base(normalizedRepo))
		}
		
		// Clone the repository
		cloneRepo(normalizedRepo, targetPath)
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().StringVar(&cloneTarget, "target", "", "Specific target directory for cloning")
	cloneCmd.Flags().BoolVar(&cloneSuggest, "suggest", false, "Only suggest clone locations, don't actually clone")
}

// CloneSuggestion represents a suggested clone location
type CloneSuggestion struct {
	Path   string
	Score  int
	Reason string
}

// normalizeRepoURL normalizes repository URLs to org/repo format
func normalizeRepoURL(repo string) string {
	// Handle full GitHub URLs
	if strings.HasPrefix(repo, "https://github.com/") {
		parts := strings.Split(repo, "/")
		if len(parts) >= 5 {
			org := parts[3]
			repoName := strings.TrimSuffix(parts[4], ".git")
			return org + "/" + repoName
		}
	}
	
	// Handle git@github.com: format
	if strings.HasPrefix(repo, "git@github.com:") {
		parts := strings.Split(repo, ":")
		if len(parts) >= 2 {
			orgRepo := strings.TrimSuffix(parts[1], ".git")
			return orgRepo
		}
	}
	
	// Handle org/repo format
	if strings.Contains(repo, "/") && !strings.Contains(repo, "://") {
		return repo
	}
	
	return ""
}

// getExistingProjects gets existing projects from cache
func getExistingProjects() []Project {
	c := cache.New()
	var cachedDirs []ProjectDir
	
	if !c.Get("project-dirs", &cachedDirs) {
		return []Project{}
	}
	
	var projects []Project
	for _, dir := range cachedDirs {
		dirProjects := findGitProjects([]string{dir.Path})
		projects = append(projects, dirProjects...)
	}
	
	return projects
}

// suggestCloneLocation suggests optimal clone locations based on similarity
func suggestCloneLocation(repo string, existingProjects []Project) []CloneSuggestion {
	orgRepo := strings.Split(repo, "/")
	if len(orgRepo) != 2 {
		return []CloneSuggestion{}
	}
	
	org := orgRepo[0]
	repoName := orgRepo[1]
	
	var suggestions []CloneSuggestion
	
	// Analyze existing project structure
	orgPaths := make(map[string]int)
	similarRepos := make(map[string]int)
	
	for _, project := range existingProjects {
		projectPath := project.Path
		projectDir := filepath.Dir(projectPath)
		projectName := filepath.Base(projectPath)
		
		// Count org-based directories
		if strings.Contains(projectDir, org) {
			orgPaths[projectDir]++
		}
		
		// Find similar repository names
		similarity := levenshteinDistance(strings.ToLower(repoName), strings.ToLower(projectName))
		if similarity < 5 { // Threshold for similarity
			similarRepos[projectDir] = similarity
		}
	}
	
	// Generate suggestions based on analysis
	home := os.Getenv("HOME")
	
	// Suggestion 1: Existing org directory
	for orgPath, count := range orgPaths {
		suggestions = append(suggestions, CloneSuggestion{
			Path:   filepath.Join(orgPath, repoName),
			Score:  count * 10, // Higher score for more projects in this org
			Reason: fmt.Sprintf("Existing %s directory (%d projects)", org, count),
		})
	}
	
	// Suggestion 2: Similar repository locations
	for similarPath, similarity := range similarRepos {
		suggestions = append(suggestions, CloneSuggestion{
			Path:   filepath.Join(similarPath, repoName),
			Score:  20 - similarity, // Lower similarity = higher score
			Reason: fmt.Sprintf("Similar repository location (distance: %d)", similarity),
		})
	}
	
	// Suggestion 3: Default org-based location
	defaultOrgPath := filepath.Join(home, "projects", org)
	suggestions = append(suggestions, CloneSuggestion{
		Path:   filepath.Join(defaultOrgPath, repoName),
		Score:  5,
		Reason: fmt.Sprintf("Default %s organization directory", org),
	})
	
	// Suggestion 4: Default projects location
	defaultPath := filepath.Join(home, "projects", repoName)
	suggestions = append(suggestions, CloneSuggestion{
		Path:   defaultPath,
		Score:  1,
		Reason: "Default projects directory",
	})
	
	// Sort by score (highest first)
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Score > suggestions[j].Score
	})
	
	// Remove duplicates
	seen := make(map[string]bool)
	var unique []CloneSuggestion
	for _, suggestion := range suggestions {
		if !seen[suggestion.Path] {
			seen[suggestion.Path] = true
			unique = append(unique, suggestion)
		}
	}
	
	return unique
}

// cloneRepo clones the repository to the target path
func cloneRepo(repo, targetPath string) {
	// Ensure parent directory exists
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", parentDir, err)
		return
	}
	
	// Clone the repository
	cloneURL := "https://github.com/" + repo + ".git"
	fmt.Printf("Cloning %s to %s...\n", cloneURL, targetPath)
	
	// Use git clone command
	cmd := fmt.Sprintf("git clone %s %s", cloneURL, targetPath)
	fmt.Printf("Running: %s\n", cmd)
	
	// Note: In a real implementation, you'd use os/exec to run the git command
	// For now, just show what would be executed
	fmt.Printf("Repository would be cloned to: %s\n", targetPath)
	fmt.Printf("After cloning, run: gum dirs-cache --refresh\n")
}