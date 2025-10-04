package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// License represents GitHub repository license information
type License struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SpdxID string `json:"spdx_id"`
	URL    string `json:"url"`
}

// GitHubMetadata represents repository metadata from GitHub API
type GitHubMetadata struct {
	// Basic info
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	Topics      []string  `json:"topics"`
	Language    string    `json:"language"`
	Languages   map[string]int `json:"languages"`
	
	// Repository properties
	IsPrivate   bool      `json:"private"`
	IsArchived  bool      `json:"archived"`
	IsFork      bool      `json:"fork"`
	IsTemplate  bool      `json:"is_template"`
	Visibility  string    `json:"visibility"`
	DefaultBranch string  `json:"default_branch"`
	License     *License  `json:"license"`
	
	// Activity metrics
	StarCount   int       `json:"stargazers_count"`
	ForkCount   int       `json:"forks_count"`
	OpenIssues  int       `json:"open_issues_count"`
	Size        int       `json:"size"`
	
	// Timestamps
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PushedAt    time.Time `json:"pushed_at"`
	
	// Update tracking
	LastSynced  time.Time `json:"last_synced"`
}

// GitHubClient handles GitHub API interactions
type GitHubClient struct {
	token      string
	httpClient *http.Client
	rateLimiter *RateLimiter
}

// RateLimiter manages GitHub API rate limits
type RateLimiter struct {
	requestsPerHour int
	requestsUsed    int
	resetTime       time.Time
}

// NewGitHubClient creates a new GitHub client using gh CLI authentication
func NewGitHubClient() (*GitHubClient, error) {
	// Check if gh CLI is available and authenticated
	if err := checkGitHubAuth(); err != nil {
		return nil, fmt.Errorf("GitHub authentication required: %v", err)
	}
	
	// Get token from gh CLI
	token, err := getGitHubToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub token: %v", err)
	}
	
	// Create HTTP client with token
	httpClient := &http.Client{
		Transport: &githubTransport{
			token: token,
			base:  http.DefaultTransport,
		},
	}
	
	return &GitHubClient{
		token:      token,
		httpClient: httpClient,
		rateLimiter: &RateLimiter{
			requestsPerHour: 5000, // GitHub API limit
			requestsUsed:    0,
			resetTime:       time.Now().Add(time.Hour),
		},
	}, nil
}

// checkGitHubAuth verifies GitHub CLI authentication
func checkGitHubAuth() error {
	cmd := exec.Command("gh", "auth", "status")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("GitHub CLI not authenticated: %v", err)
	}
	
	// Check if authenticated
	if !strings.Contains(string(output), "Logged in to") {
		return fmt.Errorf("GitHub CLI not authenticated")
	}
	
	return nil
}

// getGitHubToken retrieves the GitHub token from gh CLI
func getGitHubToken() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub token: %v", err)
	}
	
	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("empty GitHub token")
	}
	
	return token, nil
}

// githubTransport adds authentication headers to requests
type githubTransport struct {
	token string
	base  http.RoundTripper
}

func (t *githubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "gum/1.0")
	
	return t.base.RoundTrip(req)
}

// DiscoverAllRepositories discovers all repositories accessible to the user
func (gc *GitHubClient) DiscoverAllRepositories(ctx context.Context) ([]*GitHubMetadata, error) {
	var allRepos []*GitHubMetadata
	
	// Get user's accessible repositories (includes org repos)
	repos, err := gc.getUserRepositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user repositories: %v", err)
	}
	
	// Process repositories in batches
	batchSize := 100
	for i := 0; i < len(repos); i += batchSize {
		end := i + batchSize
		if end > len(repos) {
			end = len(repos)
		}
		
		batch := repos[i:end]
		enrichedRepos, err := gc.enrichBatchMetadata(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to enrich batch metadata: %v", err)
		}
		
		allRepos = append(allRepos, enrichedRepos...)
		
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	
	return allRepos, nil
}

// getUserRepositories gets all repositories accessible to the user
func (gc *GitHubClient) getUserRepositories(ctx context.Context) ([]*GitHubMetadata, error) {
	var allRepos []*GitHubMetadata
	page := 1
	perPage := 100
	
	for {
		url := fmt.Sprintf("https://api.github.com/user/repos?page=%d&per_page=%d&sort=updated", page, perPage)
		
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		
		resp, err := gc.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
		}
		
		var repos []*GitHubMetadata
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			return nil, err
		}
		
		if len(repos) == 0 {
			break // No more repositories
		}
		
		allRepos = append(allRepos, repos...)
		page++
		
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	
	return allRepos, nil
}

// enrichBatchMetadata enriches a batch of repositories with additional metadata
func (gc *GitHubClient) enrichBatchMetadata(ctx context.Context, repos []*GitHubMetadata) ([]*GitHubMetadata, error) {
	var enrichedRepos []*GitHubMetadata
	
	for _, repo := range repos {
		// Add sync timestamp
		repo.LastSynced = time.Now()
		
		// License is already parsed as License struct
		
		enrichedRepos = append(enrichedRepos, repo)
		
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	
	return enrichedRepos, nil
}

// extractLicenseName extracts license name from license object
func extractLicenseName(license *License) string {
	if license == nil {
		return ""
	}
	return license.Name
}

// parseGitHubURL parses GitHub repository URL to extract owner and repo name
func parseGitHubURL(url string) (owner, repo string, err error) {
	patterns := []string{
		`github\.com/([^/]+)/([^/]+)`,
		`git@github\.com:([^/]+)/([^/]+)`,
		`https://github\.com/([^/]+)/([^/]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) == 3 {
			return matches[1], strings.TrimSuffix(matches[2], ".git"), nil
		}
	}
	
	return "", "", fmt.Errorf("invalid GitHub URL format: %s", url)
}

// GetRepositoryMetadata gets metadata for a specific repository
func (gc *GitHubClient) GetRepositoryMetadata(ctx context.Context, owner, repo string) (*GitHubMetadata, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("repository not found: %s/%s", owner, repo)
	}
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}
	
	var metadata GitHubMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}
	
	metadata.LastSynced = time.Now()
	return &metadata, nil
}