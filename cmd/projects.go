/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/shalomb/gum/internal/database"
)

// projectsCmd represents the projects command
var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List Git projects from configured directories",
	Long: `Scan configured directories for Git repositories and list them with their
remote URLs or current branch information. Uses unified database for consistent results.`,

	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		refresh, _ := cmd.Flags().GetBool("refresh")
		clearCache, _ := cmd.Flags().GetBool("clear-cache")
		verbose, _ := cmd.Flags().GetBool("verbose")
		withGithub, _ := cmd.Flags().GetBool("with-github")
		
		if clearCache {
			db, err := database.New()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
				os.Exit(1)
			}
			defer db.Close()
			
			cache := database.NewDatabaseCache(db)
			if err := cache.ClearCache("projects"); err != nil {
				fmt.Fprintf(os.Stderr, "Error clearing cache: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Projects cache cleared")
			return
		}
		
		doListProjects(format, refresh, verbose, withGithub)
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)

	// Add flags for different output formats
	projectsCmd.Flags().StringP("format", "f", "default", "Output format: default, fzf, json, simple")
	projectsCmd.Flags().BoolP("refresh", "r", false, "Force refresh cache")
	projectsCmd.Flags().BoolP("clear-cache", "", false, "Clear cache and exit")
	projectsCmd.Flags().BoolP("verbose", "v", false, "Show verbose output including cache stats")
	projectsCmd.Flags().BoolP("with-github", "g", false, "Include GitHub repository metadata")
}

type Project struct {
	Path   string
	Remote string
	Branch string
}

var verboseMode bool

func doListProjects(format string, refresh bool, verbose bool, withGithub bool) {
	verboseMode = verbose
	
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
	var err2 error
	
	if refresh {
		// Force refresh - discover projects and update cache
		projects, err2 = discoverAndCacheProjects(db, cache)
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "Failed to discover projects: %v\n", err2)
			os.Exit(1)
		}
	} else {
		// Get projects from database (cron jobs keep data fresh)
		projects, err2 = cache.GetProjects()
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "Failed to get projects: %v\n", err2)
			os.Exit(1)
		}
	}
	
	// Convert database projects to output format
	outputProjects := convertToOutputFormat(projects, withGithub)
	
	// Sort projects by path
	sort.Slice(outputProjects, func(i, j int) bool {
		return outputProjects[i].Path < outputProjects[j].Path
	})
	
	// Show cache stats if verbose
	if verbose {
		stats, err := cache.GetCacheStats()
		if err == nil {
			fmt.Fprintf(os.Stderr, "Cache stats: %+v\n", stats)
		}
	}
	
	// Output based on format
	switch format {
	case "fzf":
		outputProjectsFzfFormat(outputProjects)
	case "json":
		outputProjectsJsonFormat(outputProjects)
	case "simple":
		outputProjectsSimpleFormat(outputProjects)
	default:
		outputProjectsDefaultFormat(outputProjects)
	}
}

// discoverAndCacheProjects discovers projects and updates the cache
func discoverAndCacheProjects(db *database.Database, cache *database.DatabaseCache) ([]*database.Project, error) {
	// Load project directories from config.yaml first
	if err := loadProjectDirsFromConfig(db); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
	}
	
	// Get project directories from database
	projectDirs, err := db.GetProjectDirs()
	if err != nil {
		return nil, fmt.Errorf("failed to get project directories: %v", err)
	}
	
	// Discover projects from directories
	var allProjects []*database.Project
	for _, dir := range projectDirs {
		projects := findGitProjectsInDir(dir.Path)
		for _, project := range projects {
			// Convert to database format
			dbProject := &database.Project{
				Path:        project.Path,
				Name:        filepath.Base(project.Path),
				RemoteURL:   project.Remote,
				Branch:      project.Branch,
				LastModified: time.Now(),
				GitCount:    1,
			}
			
			// Upsert to database
			if err := db.UpsertProject(dbProject); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to upsert project %s: %v\n", project.Path, err)
				continue
			}
			
			allProjects = append(allProjects, dbProject)
		}
	}
	
	// Update cache
	if err := cache.SetProjects(allProjects); err != nil {
		return nil, fmt.Errorf("failed to update cache: %v", err)
	}
	
	return allProjects, nil
}

// convertToOutputFormat converts database projects to output format
func convertToOutputFormat(projects []*database.Project, withGithub bool) []Project {
	var outputProjects []Project
	
	for _, dbProject := range projects {
		project := Project{
			Path:   dbProject.Path,
			Remote: dbProject.RemoteURL,
			Branch: dbProject.Branch,
		}
		
		// Add GitHub metadata if requested
		if withGithub && dbProject.GitHubRepoID > 0 {
			// TODO: Add GitHub metadata lookup
		}
		
		outputProjects = append(outputProjects, project)
	}
	
	return outputProjects
}

// findGitProjectsInDir finds Git projects in a directory
func findGitProjectsInDir(dir string) []Project {
	var projects []Project
	
	// Walk directory looking for .git folders
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		
		// Check if this is a .git directory
		if info.IsDir() && info.Name() == ".git" {
			// Get the parent directory (the project root)
			projectDir := filepath.Dir(path)
			
			// Skip if it's the root directory itself
			if projectDir == dir {
				return nil
			}
			
			// Get project info
			project := getProjectInfo(projectDir)
			projects = append(projects, project)
			
			// Don't recurse into .git directories
			return filepath.SkipDir
		}
		
		return nil
	})
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error walking directory %s: %v\n", dir, err)
	}
	
	return projects
}

// getProjectInfo extracts project information from a directory
func getProjectInfo(projectDir string) Project {
	project := Project{
		Path: projectDir,
	}
	
	// Get Git remotes
	remotes := getGitRemotes(projectDir)
	if len(remotes) > 0 {
		project.Remote = remotes[0] // Use first remote
	}
	
	// Get current branch
	project.Branch = getCurrentBranch(projectDir)
	
	return project
}

// getGitRemotes returns the list of Git remotes for a project
func getGitRemotes(projectDir string) []string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	
	remote := strings.TrimSpace(string(output))
	if remote == "" {
		return nil
	}
	
	return []string{remote}
}

// getCurrentBranch returns the current Git branch for a project
func getCurrentBranch(projectDir string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(output))
}

// Output functions
func outputProjectsDefaultFormat(projects []Project) {
	for _, project := range projects {
		if project.Remote != "" {
			fmt.Printf("%s (%s)\n", project.Path, project.Remote)
		} else if project.Branch != "" {
			fmt.Printf("%s [%s]\n", project.Path, project.Branch)
		} else {
			fmt.Printf("%s\n", project.Path)
		}
	}
}

func outputProjectsFzfFormat(projects []Project) {
	for _, project := range projects {
		if project.Remote != "" {
			fmt.Printf("%s\t%s\n", project.Path, project.Remote)
		} else if project.Branch != "" {
			fmt.Printf("%s\t[%s]\n", project.Path, project.Branch)
		} else {
			fmt.Printf("%s\t\n", project.Path)
		}
	}
}

func outputProjectsSimpleFormat(projects []Project) {
	for _, project := range projects {
		fmt.Printf("%s\n", project.Path)
	}
}

func outputProjectsJsonFormat(projects []Project) {
	fmt.Printf("[\n")
	for i, project := range projects {
		fmt.Printf("  {\n")
		fmt.Printf("    \"path\": \"%s\",\n", project.Path)
		if project.Remote != "" {
			fmt.Printf("    \"remote\": \"%s\",\n", project.Remote)
		}
		if project.Branch != "" {
			fmt.Printf("    \"branch\": \"%s\"\n", project.Branch)
		}
		if i < len(projects)-1 {
			fmt.Printf("  },\n")
		} else {
			fmt.Printf("  }\n")
		}
	}
	fmt.Printf("]\n")
}

// findGitProjects finds Git projects in multiple directories
func findGitProjects(dirs []string) []Project {
	var allProjects []Project
	
	for _, dir := range dirs {
		projects := findGitProjectsInDir(dir)
		allProjects = append(allProjects, projects...)
	}
	
	return allProjects
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	rows, cols := len(r1)+1, len(r2)+1
	
	d := make([][]int, rows)
	for i := range d {
		d[i] = make([]int, cols)
	}
	
	// Initialize first row and column
	for i := 0; i < rows; i++ {
		d[i][0] = i
	}
	for j := 0; j < cols; j++ {
		d[0][j] = j
	}
	
	// Fill the matrix
	for i := 1; i < rows; i++ {
		for j := 1; j < cols; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			
			d[i][j] = min3(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return d[rows-1][cols-1]
}

// min3 returns the minimum of three integers
func min3(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
