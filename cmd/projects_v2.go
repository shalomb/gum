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
	"time"

	"github.com/spf13/cobra"
	"github.com/shalomb/gum/internal/database"
)

// projectsV2Cmd represents the new projects command using database
var projectsV2Cmd = &cobra.Command{
	Use:   "projects-v2",
	Short: "List Git projects from database (unified cache)",
	Long: `List Git projects using the unified database cache system.
This replaces the JSON-based caching with a robust SQLite-based solution
that links local projects with GitHub metadata.`,

	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		refresh, _ := cmd.Flags().GetBool("refresh")
		clearCache, _ := cmd.Flags().GetBool("clear-cache")
		verbose, _ := cmd.Flags().GetBool("verbose")
		withGithub, _ := cmd.Flags().GetBool("with-github")
		
		if clearCache {
			if err := clearProjectsCache(); err != nil {
				fmt.Fprintf(os.Stderr, "Error clearing cache: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Projects cache cleared")
			return
		}
		
		doListProjectsV2(format, refresh, verbose, withGithub)
	},
}

func init() {
	rootCmd.AddCommand(projectsV2Cmd)

	// Add flags for different output formats
	projectsV2Cmd.Flags().StringP("format", "f", "default", "Output format: default, fzf, json, simple")
	projectsV2Cmd.Flags().BoolP("refresh", "r", false, "Force refresh cache")
	projectsV2Cmd.Flags().BoolP("clear-cache", "", false, "Clear cache and exit")
	projectsV2Cmd.Flags().BoolP("verbose", "v", false, "Show verbose output including cache stats")
	projectsV2Cmd.Flags().BoolP("with-github", "", false, "Include GitHub metadata in output")
}

func doListProjectsV2(format string, refresh bool, verbose bool, withGithub bool) {
	// Initialize database
	db, err := database.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize cache
	cache := database.NewDatabaseCache(db)
	var projects []*database.Project
	
	// Try to get from cache first (unless refresh is requested)
	if !refresh {
		if cachedProjects, err := cache.GetProjects(); err == nil {
			projects = cachedProjects
			if verbose {
				fmt.Fprintf(os.Stderr, "Using cached projects (%d found)\n", len(projects))
			}
		} else {
			// Cache miss - fetch fresh data
			projects = fetchProjectsV2(db)
			cache.SetProjects(projects)
			if verbose {
				fmt.Fprintf(os.Stderr, "Cache miss - fetched fresh data (%d projects)\n", len(projects))
			}
		}
	} else {
		// Force refresh - fetch fresh data
		projects = fetchProjectsV2(db)
		cache.SetProjects(projects)
		if verbose {
			fmt.Fprintf(os.Stderr, "Force refresh - fetched fresh data (%d projects)\n", len(projects))
		}
	}
	
	// Sort projects by path
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Path < projects[j].Path
	})
	
	// Output based on format
	switch format {
	case "fzf":
		outputProjectsFzfFormatV2(projects, withGithub)
	case "json":
		outputProjectsJsonFormatV2(projects, withGithub)
	case "simple":
		outputProjectsSimpleFormatV2(projects)
	default:
		outputProjectsDefaultFormatV2(projects, withGithub)
	}

	// Show cache stats if verbose
	if verbose {
		stats, err := cache.GetCacheStats()
		if err == nil {
			fmt.Fprintf(os.Stderr, "Cache stats: %+v\n", stats)
		}
	}
}

func fetchProjectsV2(db *database.Database) []*database.Project {
	// Get project directories from database
	projectDirs, err := db.GetProjectDirs()
	if err != nil || len(projectDirs) == 0 {
		// Fallback to auto-discovery
		projectDirs = discoverProjectDirsV2()
		// Cache the discovered directories
		for _, dir := range projectDirs {
			db.UpsertProjectDir(dir)
		}
	}
	
	// Find all Git repositories in the directories
	var allProjects []*database.Project
	for _, dir := range projectDirs {
		projects := findGitProjectsInDir(dir.Path)
		allProjects = append(allProjects, projects...)
	}
	
	// Store projects in database
	for _, project := range allProjects {
		db.UpsertProject(project)
	}
	
	return allProjects
}

func discoverProjectDirsV2() []*database.ProjectDir {
	// Use the existing discovery logic but return database format
	home := os.Getenv("HOME")
	discoveredDirs := smartDiscoverProjectDirs(home)
	
	var projectDirs []*database.ProjectDir
	for _, dir := range discoveredDirs {
		gitCount := countGitReposInDir(dir)
		projectDir := &database.ProjectDir{
			Path:        dir,
			LastScanned: time.Now(),
			GitCount:    gitCount,
		}
		projectDirs = append(projectDirs, projectDir)
	}
	
	return projectDirs
}

func findGitProjectsInDir(dirPath string) []*database.Project {
	var projects []*database.Project
	
	// Expand tilde
	if strings.HasPrefix(dirPath, "~/") {
		home := os.Getenv("HOME")
		dirPath = filepath.Join(home, dirPath[2:])
	}
	
	// Find all .git directories
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}
		
		if info.IsDir() && info.Name() == ".git" {
			projectDir := filepath.Dir(path)
			project := getProjectInfoV2(projectDir)
			if project != nil {
				projects = append(projects, project)
			}
			return filepath.SkipDir // Don't recurse into .git
		}
		
		return nil
	})
	
	if err != nil {
		// Log error but continue
		fmt.Fprintf(os.Stderr, "Warning: failed to scan directory %s: %v\n", dirPath, err)
	}
	
	return projects
}

func getProjectInfoV2(projectDir string) *database.Project {
	// Convert absolute path to ~ notation
	home := os.Getenv("HOME")
	var displayPath string
	if strings.HasPrefix(projectDir, home) {
		displayPath = "~" + projectDir[len(home):]
	} else {
		displayPath = projectDir
	}
	
	project := &database.Project{
		Path: displayPath,
		Name: filepath.Base(projectDir),
	}
	
	// Get Git remote information
	if remotes := getGitRemotes(projectDir); len(remotes) > 0 {
		project.RemoteURL = remotes[0] // Use first remote
	} else {
		// No remotes, get current branch
		if branch := getCurrentBranch(projectDir); branch != "" {
			project.Branch = branch
		} else {
			project.Branch = "main" // Default branch
		}
	}
	
	return project
}

func outputProjectsDefaultFormatV2(projects []*database.Project, withGithub bool) {
	for _, project := range projects {
		if withGithub && project.GitHubRepoID != 0 {
			// TODO: Fetch GitHub metadata
			fmt.Printf("%s\t%s\t[GitHub: %d]\n", project.Path, project.RemoteURL, project.GitHubRepoID)
		} else if project.RemoteURL != "" {
			fmt.Printf("%s\t%s\n", project.Path, project.RemoteURL)
		} else {
			fmt.Printf("%s\t%s\n", project.Path, project.Branch)
		}
	}
}

func outputProjectsFzfFormatV2(projects []*database.Project, withGithub bool) {
	// Count stats
	totalProjects := len(projects)
	withRemotes := 0
	withBranches := 0
	withGithubLinked := 0
	
	for _, project := range projects {
		if project.RemoteURL != "" {
			withRemotes++
		} else if project.Branch != "" {
			withBranches++
		}
		if project.GitHubRepoID != 0 {
			withGithubLinked++
		}
	}
	
	// Get current working directory for similarity matching
	cwd, _ := os.Getwd()
	currentDir := filepath.Base(cwd)
	
	// Sort projects by similarity to current directory
	sortedProjects := sortProjectsBySimilarityV2(projects, currentDir)
	
	// Output projects
	for _, project := range sortedProjects {
		if withGithub && project.GitHubRepoID != 0 {
			fmt.Printf("%-60s %s [GitHub: %d]\n", 
				project.Path, project.RemoteURL, project.GitHubRepoID)
		} else if project.RemoteURL != "" {
			fmt.Printf("%-60s %s\n", 
				project.Path, project.RemoteURL)
		} else {
			fmt.Printf("%-60s %s\n", 
				project.Path, project.Branch)
		}
	}
	
	// Add stats separator and info
	fmt.Printf("\n")
	fmt.Printf("Stats: %d projects total | %d with remotes | %d local only | %d linked to GitHub\n", 
		totalProjects, withRemotes, withBranches, withGithubLinked)
}

func outputProjectsJsonFormatV2(projects []*database.Project, withGithub bool) {
	// TODO: Implement JSON output with GitHub metadata
	fmt.Println("JSON output not yet implemented for v2")
}

func outputProjectsSimpleFormatV2(projects []*database.Project) {
	for _, project := range projects {
		fmt.Printf("%s\n", project.Path)
	}
}

func sortProjectsBySimilarityV2(projects []*database.Project, currentDir string) []*database.Project {
	// Create a copy to avoid modifying the original slice
	sorted := make([]*database.Project, len(projects))
	copy(sorted, projects)
	
	// Sort by similarity to current directory
	sort.Slice(sorted, func(i, j int) bool {
		similarityI := calculateSimilarity(currentDir, filepath.Base(sorted[i].Path))
		similarityJ := calculateSimilarity(currentDir, filepath.Base(sorted[j].Path))
		return similarityI > similarityJ
	})
	
	return sorted
}

func clearProjectsCache() error {
	db, err := database.New()
	if err != nil {
		return err
	}
	defer db.Close()

	cache := database.NewDatabaseCache(db)
	return cache.ClearCache("projects")
}

func calculateSimilarity(s1, s2 string) int {
	// Simple similarity calculation based on common characters
	if s1 == s2 {
		return 100
	}
	
	common := 0
	for _, c1 := range s1 {
		for _, c2 := range s2 {
			if c1 == c2 {
				common++
				break
			}
		}
	}
	
	return common * 100 / len(s1)
}