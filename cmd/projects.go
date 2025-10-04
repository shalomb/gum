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

	"github.com/spf13/cobra"
	"github.com/shalomb/gum/internal/cache"
	"github.com/shalomb/gum/internal/locate"
)

// projectsCmd represents the projects command
var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List Git projects from configured directories",
	Long: `Scan configured directories for Git repositories and list them with their
remote URLs or current branch information. This replaces the shell script
projects-list with better performance and Go-native implementation.`,

	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		refresh, _ := cmd.Flags().GetBool("refresh")
		clearCache, _ := cmd.Flags().GetBool("clear-cache")
		verbose, _ := cmd.Flags().GetBool("verbose")
		
		if clearCache {
			c := cache.New()
			if err := c.Clear("projects"); err != nil {
				fmt.Fprintf(os.Stderr, "Error clearing cache: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Projects cache cleared")
			return
		}
		
		doListProjects(format, refresh, verbose)
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)

	// Add flags for different output formats
	projectsCmd.Flags().StringP("format", "f", "default", "Output format: default, fzf, json, simple")
	projectsCmd.Flags().BoolP("refresh", "r", false, "Force refresh cache")
	projectsCmd.Flags().BoolP("clear-cache", "", false, "Clear cache and exit")
	projectsCmd.Flags().BoolP("verbose", "v", false, "Show verbose output including locate usage")
}

type Project struct {
	Path   string
	Remote string
	Branch string
}

var verboseMode bool

func doListProjects(format string, refresh bool, verbose bool) {
	verboseMode = verbose
	c := cache.New()
	var projects []Project
	
	// Try to get from cache first (unless refresh is requested)
	if !refresh {
		if c.Get("projects", &projects) {
			// Cache hit - use cached data
		} else {
			// Cache miss - fetch fresh data
			projects = fetchProjects()
			c.Set("projects", projects, cache.ProjectsCacheTTL)
		}
	} else {
		// Force refresh - fetch fresh data
		projects = fetchProjects()
		c.Set("projects", projects, cache.ProjectsCacheTTL)
	}
	
	// Sort projects by path
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Path < projects[j].Path
	})
	
	// Output based on format
	switch format {
	case "fzf":
		outputProjectsFzfFormat(projects)
	case "json":
		outputProjectsJsonFormat(projects)
	case "simple":
		outputProjectsSimpleFormat(projects)
	default:
		outputProjectsDefaultFormat(projects)
	}
}

func fetchProjects() []Project {
	// Get project directories from config
	projectDirs := getProjectDirs()
	
	// Find all Git repositories
	return findGitProjects(projectDirs)
}

func getProjectDirs() []string {
	c := cache.New()
	var cachedDirs []ProjectDir
	
	// Try to get from cache first
	if c.Get("project-dirs", &cachedDirs) {
		var dirs []string
		for _, dir := range cachedDirs {
			dirs = append(dirs, dir.Path)
		}
		return dirs
	}
	
	// Smart auto-discovery: find directories with Git repositories
	home := os.Getenv("HOME")
	var dirs []string
	
	// First, try to read from YAML config if it exists
	if yamlDirs := readYAMLConfig(); len(yamlDirs) > 0 {
		dirs = append(dirs, yamlDirs...)
	} else {
		// Smart discovery: scan common patterns and find directories with Git repos
		discoveredDirs := smartDiscoverProjectDirs(home)
		dirs = append(dirs, discoveredDirs...)
		
		// Generate config stub if user might want explicit control
		generateConfigStubIfNeeded(home, discoveredDirs)
	}
	
	
	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueDirs []string
	for _, dir := range dirs {
		if !seen[dir] {
			seen[dir] = true
			uniqueDirs = append(uniqueDirs, dir)
		}
	}
	
	return uniqueDirs
}

// readYAMLConfig reads project directories from YAML config
func readYAMLConfig() []string {
	home := os.Getenv("HOME")
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(home, ".config")
	}
	
	configFile := filepath.Join(configDir, "gum", "config.yaml")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil // No config file, that's fine
	}
	
	// Simple YAML parsing for projects section
	lines := strings.Split(string(data), "\n")
	var dirs []string
	inProjects := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "projects:") {
			inProjects = true
			continue
		}
		
		if inProjects {
			if strings.HasPrefix(line, "- ") {
				// Extract directory path
				dir := strings.TrimSpace(line[2:])
				if strings.HasPrefix(dir, "~/") {
					dir = filepath.Join(home, dir[2:])
				}
				dirs = append(dirs, dir)
			} else if line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				// End of projects section
				break
			}
		}
	}
	
	return dirs
}

// smartDiscoverProjectDirs intelligently finds directories with Git repositories
func smartDiscoverProjectDirs(home string) []string {
	var dirs []string
	
	// Try locate first for speed
	if locateDirs := discoverWithLocate(home); len(locateDirs) > 0 {
		dirs = append(dirs, locateDirs...)
	}
	
	// Fallback to file system scanning for any missed directories
	fileSystemDirs := discoverWithFileSystem(home)
	dirs = append(dirs, fileSystemDirs...)
	
	// Remove duplicates
	dirs = removeDuplicateDirs(dirs)
	
	// If no directories found with Git repos, fall back to common defaults
	if len(dirs) == 0 {
		dirs = append(dirs, filepath.Join(home, "projects"))
		dirs = append(dirs, filepath.Join(home, "oneTakeda"))
	}
	
	return dirs
}

// discoverWithLocate uses the locate database for fast discovery
func discoverWithLocate(home string) []string {
	finder := locate.NewLocateFinder()
	if !finder.GetStatus().Available {
		if verboseMode {
			fmt.Fprintf(os.Stderr, "locate not available - using file system scanning\n")
		}
		return nil // No locate available
	}
	
	// Check if database is fresh enough
	status := finder.GetStatus()
	if verboseMode {
		fmt.Fprintf(os.Stderr, "Using locate database (last updated: %v)\n", status.LastUpdated.Format("2006-01-02 15:04:05"))
	}
	
	if !status.IsFresh {
		// Database is stale, but still use it as a starting point
		fmt.Fprintf(os.Stderr, "Warning: locate database is %v old - some recent projects may be missing\n", status.Age)
	}
	
	// Find git repos in home directory
	repos, err := finder.FindGitRepos(home)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: locate failed: %v\n", err)
		return nil
	}
	
	if verboseMode {
		fmt.Fprintf(os.Stderr, "Found %d git repositories via locate\n", len(repos))
	}
	
	// Group repos by parent directory
	dirMap := make(map[string]bool)
	for _, repo := range repos {
		parentDir := filepath.Dir(repo)
		// Only include directories that are direct children of common patterns
		if isCommonProjectDir(parentDir, home) {
			dirMap[parentDir] = true
		}
	}
	
	var dirs []string
	for dir := range dirMap {
		dirs = append(dirs, dir)
	}
	
	if verboseMode {
		fmt.Fprintf(os.Stderr, "Discovered %d project directories via locate\n", len(dirs))
	}
	
	return dirs
}

// discoverWithFileSystem uses traditional file system scanning
func discoverWithFileSystem(home string) []string {
	var dirs []string
	
	// Common patterns to check
	patterns := []string{
		filepath.Join(home, "projects"),
		filepath.Join(home, "oneTakeda"),
		filepath.Join(home, "projects-local"),
		filepath.Join(home, "code"),
		filepath.Join(home, "dev"),
		filepath.Join(home, "workspace"),
		filepath.Join(home, "repos"),
		filepath.Join(home, "repositories"),
	}
	
	// Add ~/projects-* directories
	projectsPattern := filepath.Join(home, "projects-*")
	if matches, err := filepath.Glob(projectsPattern); err == nil {
		patterns = append(patterns, matches...)
	}
	
	// Add ~/work-* directories
	workPattern := filepath.Join(home, "work-*")
	if matches, err := filepath.Glob(workPattern); err == nil {
		patterns = append(patterns, matches...)
	}
	
	// Check each pattern and only include directories that have Git repositories
	for _, pattern := range patterns {
		if stat, err := os.Stat(pattern); err == nil && stat.IsDir() {
			// Count Git repositories in this directory
			gitCount := countGitReposInDir(pattern)
			if gitCount > 0 {
				dirs = append(dirs, pattern)
			}
		}
	}
	
	return dirs
}

// isCommonProjectDir checks if a directory matches common project patterns
func isCommonProjectDir(dir, home string) bool {
	// Check if it's a direct child of common project directories
	commonParents := []string{
		filepath.Join(home, "projects"),
		filepath.Join(home, "oneTakeda"),
		filepath.Join(home, "projects-local"),
		filepath.Join(home, "code"),
		filepath.Join(home, "dev"),
		filepath.Join(home, "workspace"),
		filepath.Join(home, "repos"),
		filepath.Join(home, "repositories"),
	}
	
	for _, parent := range commonParents {
		if strings.HasPrefix(dir, parent+string(filepath.Separator)) {
			return true
		}
	}
	
	// Check for projects-* and work-* patterns
	if strings.Contains(filepath.Base(dir), "projects-") || strings.Contains(filepath.Base(dir), "work-") {
		return true
	}
	
	return false
}

// removeDuplicateDirs removes duplicate directories from the slice
func removeDuplicateDirs(dirs []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, dir := range dirs {
		if !seen[dir] {
			seen[dir] = true
			result = append(result, dir)
		}
	}
	
	return result
}

// countGitReposInDir counts .git directories in a given directory
func countGitReposInDir(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() && info.Name() == ".git" {
			count++
			return filepath.SkipDir // Don't recurse into .git
		}
		return nil
	})
	return count
}

// generateConfigStubIfNeeded creates a config stub for users who want explicit control
func generateConfigStubIfNeeded(home string, discoveredDirs []string) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(home, ".config")
	}
	
	gumConfigDir := filepath.Join(configDir, "gum")
	configFile := filepath.Join(gumConfigDir, "config.yaml")
	
	// Only generate if config doesn't exist and we found multiple directories
	if _, err := os.Stat(configFile); err == nil {
		return // Config already exists
	}
	
	if len(discoveredDirs) < 2 {
		return // Not enough directories to warrant a config
	}
	
	// Create gum config directory
	if err := os.MkdirAll(gumConfigDir, 0755); err != nil {
		return // Can't create directory, skip
	}
	
	// Generate config stub
	configContent := generateConfigStub(home, discoveredDirs)
	
	// Write config file
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return // Can't write file, skip
	}
	
	// Show user-friendly message
	fmt.Fprintf(os.Stderr, "gum: Auto-discovered %d project directories\n", len(discoveredDirs))
	fmt.Fprintf(os.Stderr, "gum: Generated config stub at %s\n", configFile)
	fmt.Fprintf(os.Stderr, "gum: Edit the config to customize directory scanning\n")
}

// generateConfigStub creates a YAML config stub with discovered directories
func generateConfigStub(home string, discoveredDirs []string) string {
	var config strings.Builder
	
	config.WriteString("# Gum Configuration\n")
	config.WriteString("# This file was auto-generated based on discovered project directories\n")
	config.WriteString("# Edit this file to customize which directories gum scans for Git repositories\n\n")
	
	config.WriteString("projects:\n")
	
	// Remove duplicates before writing
	seen := make(map[string]bool)
	var uniqueDirs []string
	for _, dir := range discoveredDirs {
		if !seen[dir] {
			seen[dir] = true
			uniqueDirs = append(uniqueDirs, dir)
		}
	}
	
	for _, dir := range uniqueDirs {
		// Convert absolute path to ~ notation for readability
		displayPath := dir
		if strings.HasPrefix(dir, home) {
			displayPath = "~" + dir[len(home):]
		}
		
		// Count Git repos for user info
		gitCount := countGitReposInDir(dir)
		
		config.WriteString(fmt.Sprintf("  - %s  # %d Git repositories\n", displayPath, gitCount))
	}
	
	config.WriteString("\n# Additional directories you can add:\n")
	config.WriteString("# - ~/code\n")
	config.WriteString("# - ~/dev\n")
	config.WriteString("# - ~/workspace\n")
	config.WriteString("# - ~/repos\n")
	config.WriteString("# - ~/repositories\n")
	config.WriteString("# - /path/to/any/directory\n")
	
	config.WriteString("\n# Note: Directories with 0 Git repositories will be ignored\n")
	config.WriteString("# Remove directories from this list to exclude them from scanning\n")
	
	return config.String()
}

func findGitProjects(dirs []string) []Project {
	var projects []Project
	
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		
		// Find all .git directories
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors, continue walking
			}
			
			if info.IsDir() && info.Name() == ".git" {
				projectDir := filepath.Dir(path)
				project := getProjectInfo(projectDir)
				if project.Path != "" {
					projects = append(projects, project)
				}
				return filepath.SkipDir // Don't recurse into .git
			}
			
			return nil
		})
		
		if err != nil {
			continue
		}
	}
	
	return projects
}

func getProjectInfo(projectDir string) Project {
	// Convert absolute path to ~ notation
	home := os.Getenv("HOME")
	var displayPath string
	if strings.HasPrefix(projectDir, home) {
		displayPath = "~" + projectDir[len(home):]
	} else {
		displayPath = projectDir
	}
	
	project := Project{Path: displayPath}
	
	// Get Git remote information
	if remotes := getGitRemotes(projectDir); len(remotes) > 0 {
		project.Remote = remotes[0] // Use first remote
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

func getGitRemotes(projectDir string) []string {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	
	var remotes []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[2] == "(fetch)" {
			remotes = append(remotes, fields[1])
		}
	}
	
	return remotes
}

func getCurrentBranch(projectDir string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(output))
}

func outputProjectsDefaultFormat(projects []Project) {
	for _, project := range projects {
		if project.Remote != "" {
			fmt.Printf("%s\t%s\n", project.Path, project.Remote)
		} else {
			fmt.Printf("%s\t%s\n", project.Path, project.Branch)
		}
	}
}

func outputProjectsFzfFormat(projects []Project) {
	// Count stats
	totalProjects := len(projects)
	withRemotes := 0
	withBranches := 0
	
	for _, project := range projects {
		if project.Remote != "" {
			withRemotes++
		} else if project.Branch != "" {
			withBranches++
		}
	}
	
	// Get current working directory for similarity matching
	cwd, _ := os.Getwd()
	currentDir := filepath.Base(cwd)
	
	// Sort projects by similarity to current directory
	sortedProjects := sortProjectsBySimilarity(projects, currentDir)
	
	// Output projects
	for _, project := range sortedProjects {
		if project.Remote != "" {
			fmt.Printf("%-60s %s\n", 
				project.Path, project.Remote)
		} else {
			fmt.Printf("%-60s %s\n", 
				project.Path, project.Branch)
		}
	}
	
	// Add stats separator and info
	fmt.Printf("\n")
	fmt.Printf("Stats: %d projects total | %d with remotes | %d local only\n", 
		totalProjects, withRemotes, withBranches)
}

func outputProjectsSimpleFormat(projects []Project) {
	for _, project := range projects {
		fmt.Printf("%s\n", project.Path)
	}
}

// sortProjectsBySimilarity sorts projects by similarity to current directory
func sortProjectsBySimilarity(projects []Project, currentDir string) []Project {
	// Create a copy to avoid modifying the original slice
	sorted := make([]Project, len(projects))
	copy(sorted, projects)
	
	// Sort by similarity score (lower distance = more similar)
	sort.Slice(sorted, func(i, j int) bool {
		scoreI := calculateSimilarityScore(sorted[i], currentDir)
		scoreJ := calculateSimilarityScore(sorted[j], currentDir)
		return scoreI < scoreJ
	})
	
	return sorted
}

// calculateSimilarityScore calculates similarity score between project and current directory
func calculateSimilarityScore(project Project, currentDir string) int {
	projectName := filepath.Base(project.Path)
	
	// Calculate Levenshtein distance
	distance := levenshteinDistance(strings.ToLower(currentDir), strings.ToLower(projectName))
	
	// Bonus for exact matches
	if strings.EqualFold(currentDir, projectName) {
		distance = 0
	}
	
	// Bonus for partial matches
	if strings.Contains(strings.ToLower(projectName), strings.ToLower(currentDir)) {
		distance = distance / 2
	}
	
	return distance
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	rows := len(r1) + 1
	cols := len(r2) + 1
	
	d := make([][]int, rows)
	for i := range d {
		d[i] = make([]int, cols)
	}
	
	for i := 1; i < rows; i++ {
		d[i][0] = i
	}
	for j := 1; j < cols; j++ {
		d[0][j] = j
	}
	
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

func outputProjectsJsonFormat(projects []Project) {
	fmt.Printf("[\n")
	for i, project := range projects {
		fmt.Printf("  {\n")
		fmt.Printf("    \"path\": \"%s\",\n", project.Path)
		if project.Remote != "" {
			fmt.Printf("    \"remote\": \"%s\",\n", project.Remote)
		} else {
			fmt.Printf("    \"branch\": \"%s\",\n", project.Branch)
		}
		if i < len(projects)-1 {
			fmt.Printf("  },\n")
		} else {
			fmt.Printf("  }\n")
		}
	}
	fmt.Printf("]\n")
}