/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shalomb/gum/internal/cache"
	"github.com/spf13/cobra"
)

// ProjectDir represents a project directory entry
type ProjectDir struct {
	Path        string    `json:"path"`
	LastScanned time.Time `json:"last_scanned"`
	GitCount    int       `json:"git_count"`
}

var (
	dirsCacheRefresh bool
	dirsCacheClear   bool
	dirsCacheList    bool
)

// dirsCacheCmd represents the dirs-cache command
var dirsCacheCmd = &cobra.Command{
	Use:   "dirs-cache",
	Short: "Manage project directories cache",
	Long: `Manage the cache of project directories that gum scans for Git repositories.

This command allows you to:
- List currently cached project directories
- Refresh the cache by auto-discovering directories
- Clear the cache to force re-discovery

Auto-discovery looks for:
- ~/projects/ (default)
- ~/projects-* (glob pattern)
- Any directories in ~/.config/projects-dirs.list (legacy support)`,
	Run: func(cmd *cobra.Command, args []string) {
		c := cache.New()
		
		if dirsCacheClear {
			if err := c.Clear("project-dirs"); err != nil {
				fmt.Printf("Error clearing cache: %v\n", err)
				return
			}
			fmt.Println("Project directories cache cleared")
			return
		}
		
		if dirsCacheList {
			var dirs []ProjectDir
			if c.Get("project-dirs", &dirs) {
				fmt.Println("Cached project directories:")
				for _, dir := range dirs {
					fmt.Printf("  %s (last scanned: %s, %d git repos)\n", 
						dir.Path, dir.LastScanned.Format("2006-01-02 15:04:05"), dir.GitCount)
				}
			} else {
				fmt.Println("No cached project directories found")
			}
			return
		}
		
		if dirsCacheRefresh {
			dirs := discoverProjectDirs()
			if err := c.Set("project-dirs", dirs, cache.ProjectDirsCacheTTL); err != nil {
				fmt.Printf("Error updating cache: %v\n", err)
				return
			}
			fmt.Printf("Discovered %d project directories and cached them\n", len(dirs))
			for _, dir := range dirs {
				fmt.Printf("  %s (%d git repos)\n", dir.Path, dir.GitCount)
			}
			return
		}
		
		// Default: show status
		var dirs []ProjectDir
		if c.Get("project-dirs", &dirs) {
			fmt.Printf("Project directories cache is valid (%d directories)\n", len(dirs))
			fmt.Println("Use --list to see details, --refresh to update, --clear to reset")
		} else {
			fmt.Println("No project directories cache found")
			fmt.Println("Use --refresh to discover and cache project directories")
		}
	},
}

func init() {
	rootCmd.AddCommand(dirsCacheCmd)
	dirsCacheCmd.Flags().BoolVar(&dirsCacheRefresh, "refresh", false, "Refresh the cache by discovering directories")
	dirsCacheCmd.Flags().BoolVar(&dirsCacheClear, "clear", false, "Clear the project directories cache")
	dirsCacheCmd.Flags().BoolVar(&dirsCacheList, "list", false, "List cached project directories")
}

// discoverProjectDirs finds all project directories
func discoverProjectDirs() []ProjectDir {
	home := os.Getenv("HOME")
	var dirs []ProjectDir
	
	// Add default directories
	defaultDirs := []string{
		filepath.Join(home, "projects"),
		filepath.Join(home, "oneTakeda"),
	}
	
	// Add ~/projects-* directories
	projectsPattern := filepath.Join(home, "projects-*")
	if matches, err := filepath.Glob(projectsPattern); err == nil {
		for _, match := range matches {
			if stat, err := os.Stat(match); err == nil && stat.IsDir() {
				defaultDirs = append(defaultDirs, match)
			}
		}
	}
	
	// Add directories from legacy projects-dirs.list
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(home, ".config")
	}
	
	projectsDirsList := filepath.Join(configDir, "projects-dirs.list")
	if data, err := os.ReadFile(projectsDirsList); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				// Expand ~ to home directory
				if strings.HasPrefix(line, "~/") {
					line = filepath.Join(home, line[2:])
				}
				defaultDirs = append(defaultDirs, line)
			}
		}
	}
	
	// Remove duplicates and count git repos
	seen := make(map[string]bool)
	for _, dir := range defaultDirs {
		if !seen[dir] && dirExists(dir) {
			seen[dir] = true
			gitCount := countGitRepos(dir)
			dirs = append(dirs, ProjectDir{
				Path:        dir,
				LastScanned: time.Now(),
				GitCount:    gitCount,
			})
		}
	}
	
	return dirs
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

// countGitRepos counts .git directories in a path
func countGitRepos(path string) int {
	count := 0
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
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