/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/shalomb/gum/internal/cache"
)

// GitHubRepo represents a GitHub repository
type GitHubRepo struct {
	Name        string `json:"name"`
	FullName    string `json:"nameWithOwner"`
	Description string `json:"description"`
	URL         string `json:"url"`
	CloneURL    string `json:"sshUrl"`
	UpdatedAt   string `json:"updatedAt"`
	Private     bool   `json:"isPrivate"`
	Fork        bool   `json:"isFork"`
}

// GitHubDiscovery represents cached GitHub discovery data
type GitHubDiscovery struct {
	Repos      []GitHubRepo `json:"repos"`
	LastUpdate time.Time    `json:"last_update"`
	User       string       `json:"user"`
}

var (
	githubRefresh bool
	githubClear   bool
	githubList    bool
)

// githubCmd represents the github command
var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "Discover and manage GitHub repositories",
	Long: `Discover GitHub repositories visible to the current user using the GitHub CLI.

This command:
- Discovers repositories from the current user and organizations
- Caches results for performance (12-hour TTL)
- Provides similarity matching for repository discovery
- Updates twice daily via cron job

The discovery includes:
- User's own repositories
- Organization repositories the user has access to
- Both public and private repositories`,
	Run: func(cmd *cobra.Command, args []string) {
		c := cache.New()
		
		if githubClear {
			if err := c.Clear("github-repos"); err != nil {
				fmt.Printf("Error clearing cache: %v\n", err)
				return
			}
			fmt.Println("GitHub repositories cache cleared")
			return
		}
		
		if githubList {
			var discovery GitHubDiscovery
			if c.Get("github-repos", &discovery) {
				fmt.Printf("Cached GitHub repositories (%d total, last updated: %s):\n", 
					len(discovery.Repos), discovery.LastUpdate.Format("2006-01-02 15:04:05"))
				for _, repo := range discovery.Repos {
					visibility := "public"
					if repo.Private {
						visibility = "private"
					}
					fmt.Printf("  %s (%s) - %s\n", repo.FullName, visibility, repo.Description)
				}
			} else {
				fmt.Println("No cached GitHub repositories found")
			}
			return
		}
		
		if githubRefresh {
			repos := discoverGitHubRepos()
			if len(repos) == 0 {
				fmt.Println("No GitHub repositories found or gh CLI not authenticated")
				return
			}
			
			discovery := GitHubDiscovery{
				Repos:      repos,
				LastUpdate: time.Now(),
				User:       getCurrentUser(),
			}
			
			if err := c.Set("github-repos", discovery, 12*time.Hour); err != nil {
				fmt.Printf("Error updating cache: %v\n", err)
				return
			}
			
			fmt.Printf("Discovered %d GitHub repositories and cached them\n", len(repos))
			for _, repo := range repos[:min(5, len(repos))] {
				fmt.Printf("  %s - %s\n", repo.FullName, repo.Description)
			}
			if len(repos) > 5 {
				fmt.Printf("  ... and %d more\n", len(repos)-5)
			}
			return
		}
		
		// Default: show status
		var discovery GitHubDiscovery
		if c.Get("github-repos", &discovery) {
			fmt.Printf("GitHub repositories cache is valid (%d repositories, user: %s)\n", 
				len(discovery.Repos), discovery.User)
			fmt.Println("Use --list to see details, --refresh to update, --clear to reset")
		} else {
			fmt.Println("No GitHub repositories cache found")
			fmt.Println("Use --refresh to discover GitHub repositories")
		}
	},
}

func init() {
	rootCmd.AddCommand(githubCmd)
	githubCmd.Flags().BoolVar(&githubRefresh, "refresh", false, "Refresh the cache by discovering repositories")
	githubCmd.Flags().BoolVar(&githubClear, "clear", false, "Clear the GitHub repositories cache")
	githubCmd.Flags().BoolVar(&githubList, "list", false, "List cached GitHub repositories")
}

// discoverGitHubRepos discovers repositories using gh CLI
func discoverGitHubRepos() []GitHubRepo {
	// Get user's repositories
	userRepos := getReposFromGh("repo list --json name,nameWithOwner,description,url,sshUrl,updatedAt,isPrivate,isFork --limit 100")
	
	// Get organization repositories
	orgRepos := getReposFromGh("repo list --json name,nameWithOwner,description,url,sshUrl,updatedAt,isPrivate,isFork --limit 100 --source")
	
	// Combine and deduplicate
	allRepos := append(userRepos, orgRepos...)
	uniqueRepos := deduplicateRepos(allRepos)
	
	// Sort by update time (most recent first)
	sort.Slice(uniqueRepos, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, uniqueRepos[i].UpdatedAt)
		timeJ, _ := time.Parse(time.RFC3339, uniqueRepos[j].UpdatedAt)
		return timeI.After(timeJ)
	})
	
	return uniqueRepos
}

// getReposFromGh executes gh command and parses JSON output
func getReposFromGh(command string) []GitHubRepo {
	cmd := exec.Command("gh", strings.Fields(command)...)
	output, err := cmd.Output()
	if err != nil {
		return []GitHubRepo{}
	}
	
	var repos []GitHubRepo
	if err := json.Unmarshal(output, &repos); err != nil {
		return []GitHubRepo{}
	}
	
	return repos
}

// deduplicateRepos removes duplicate repositories
func deduplicateRepos(repos []GitHubRepo) []GitHubRepo {
	seen := make(map[string]bool)
	var unique []GitHubRepo
	
	for _, repo := range repos {
		if !seen[repo.FullName] {
			seen[repo.FullName] = true
			unique = append(unique, repo)
		}
	}
	
	return unique
}

// getCurrentUser gets the current GitHub user
func getCurrentUser() string {
	cmd := exec.Command("gh", "api", "user", "--jq", ".login")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}