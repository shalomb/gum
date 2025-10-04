/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package database

import (
	"fmt"
	"strings"
)

// Projects operations

// UpsertProject inserts or updates a project
func (d *Database) UpsertProject(project *Project) error {
	query := `
		INSERT INTO projects (path, name, remote_url, branch, last_modified, git_count)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			name = excluded.name,
			remote_url = excluded.remote_url,
			branch = excluded.branch,
			last_modified = excluded.last_modified,
			git_count = excluded.git_count,
			updated_at = CURRENT_TIMESTAMP
	`
	
	_, err := d.db.Exec(query, project.Path, project.Name, project.RemoteURL, 
		project.Branch, project.LastModified, project.GitCount)
	return err
}

// GetProjects returns all projects, optionally sorted by similarity
func (d *Database) GetProjects(sortBySimilarity bool, targetPath string) ([]*Project, error) {
	var query string
	var args []interface{}
	
	if sortBySimilarity && targetPath != "" {
		// Use similarity scoring with Levenshtein distance
		query = `
			SELECT id, path, name, remote_url, branch, last_modified, git_count, created_at, updated_at
			FROM projects
			ORDER BY 
				CASE 
					WHEN LOWER(name) = LOWER(?) THEN 0
					WHEN LOWER(name) LIKE '%' || LOWER(?) || '%' THEN 1
					ELSE 2
				END,
				updated_at DESC
		`
		args = []interface{}{targetPath, targetPath}
	} else {
		query = `
			SELECT id, path, name, remote_url, branch, last_modified, git_count, created_at, updated_at
			FROM projects
			ORDER BY updated_at DESC
		`
	}
	
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var projects []*Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.Path, &p.Name, &p.RemoteURL, &p.Branch, 
			&p.LastModified, &p.GitCount, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}
	
	return projects, nil
}

// ProjectDirs operations

// UpsertProjectDir inserts or updates a project directory
func (d *Database) UpsertProjectDir(dir *ProjectDir) error {
	query := `
		INSERT INTO project_dirs (path, last_scanned, git_count)
		VALUES (?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			last_scanned = excluded.last_scanned,
			git_count = excluded.git_count,
			updated_at = CURRENT_TIMESTAMP
	`
	
	_, err := d.db.Exec(query, dir.Path, dir.LastScanned, dir.GitCount)
	return err
}

// GetProjectDirs returns all project directories
func (d *Database) GetProjectDirs() ([]*ProjectDir, error) {
	query := `
		SELECT id, path, last_scanned, git_count, created_at, updated_at
		FROM project_dirs
		ORDER BY path
	`
	
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var dirs []*ProjectDir
	for rows.Next() {
		var d ProjectDir
		err := rows.Scan(&d.ID, &d.Path, &d.LastScanned, &d.GitCount, 
			&d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, &d)
	}
	
	return dirs, nil
}

// GitHub repos operations

// UpsertGitHubRepo inserts or updates a GitHub repository
func (d *Database) UpsertGitHubRepo(repo *GitHubRepo) error {
	query := `
		INSERT INTO github_repos (name, full_name, description, url, clone_url, ssh_url, is_private, is_fork, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(full_name) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			url = excluded.url,
			clone_url = excluded.clone_url,
			ssh_url = excluded.ssh_url,
			is_private = excluded.is_private,
			is_fork = excluded.is_fork,
			updated_at = excluded.updated_at,
			last_discovered = CURRENT_TIMESTAMP
	`
	
	_, err := d.db.Exec(query, repo.Name, repo.FullName, repo.Description, 
		repo.URL, repo.CloneURL, repo.SSHURL, repo.IsPrivate, repo.IsFork, repo.UpdatedAt)
	return err
}

// GetGitHubRepos returns all GitHub repositories
func (d *Database) GetGitHubRepos() ([]*GitHubRepo, error) {
	query := `
		SELECT id, name, full_name, description, url, clone_url, ssh_url, is_private, is_fork, updated_at, last_discovered, created_at
		FROM github_repos
		ORDER BY updated_at DESC
	`
	
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var repos []*GitHubRepo
	for rows.Next() {
		var r GitHubRepo
		err := rows.Scan(&r.ID, &r.Name, &r.FullName, &r.Description, 
			&r.URL, &r.CloneURL, &r.SSHURL, &r.IsPrivate, &r.IsFork, 
			&r.UpdatedAt, &r.LastDiscovered, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		repos = append(repos, &r)
	}
	
	return repos, nil
}

// DirUsage operations

// UpsertDirUsage inserts or updates directory usage tracking
func (d *Database) UpsertDirUsage(usage *DirUsage) error {
	query := `
		INSERT INTO dir_usage (path, frequency, last_seen)
		VALUES (?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			frequency = frequency + 1,
			last_seen = excluded.last_seen,
			updated_at = CURRENT_TIMESTAMP
	`
	
	_, err := d.db.Exec(query, usage.Path, usage.Frequency, usage.LastSeen)
	return err
}

// GetFrequentDirs returns directories sorted by frequency and recency
func (d *Database) GetFrequentDirs(limit int) ([]*DirUsage, error) {
	query := `
		SELECT id, path, frequency, last_seen, created_at, updated_at
		FROM dir_usage
		ORDER BY 
			frequency * (1.0 / (1.0 + (julianday('now') - julianday(last_seen)) / 1.0)) DESC,
			last_seen DESC
		LIMIT ?
	`
	
	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var dirs []*DirUsage
	for rows.Next() {
		var d DirUsage
		err := rows.Scan(&d.ID, &d.Path, &d.Frequency, &d.LastSeen, 
			&d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, &d)
	}
	
	return dirs, nil
}

// Similarity operations

// GetSimilarProjects returns projects similar to the given path
func (d *Database) GetSimilarProjects(targetPath string, limit int) ([]*Project, error) {
	// Extract the target name for similarity matching
	targetName := targetPath
	if lastSlash := strings.LastIndex(targetPath, "/"); lastSlash != -1 {
		targetName = targetPath[lastSlash+1:]
	}
	
	query := `
		SELECT id, path, name, remote_url, branch, last_modified, git_count, created_at, updated_at
		FROM projects
		WHERE 
			LOWER(name) LIKE '%' || LOWER(?) || '%' OR
			LOWER(path) LIKE '%' || LOWER(?) || '%'
		ORDER BY 
			CASE 
				WHEN LOWER(name) = LOWER(?) THEN 0
				WHEN LOWER(name) LIKE '%' || LOWER(?) || '%' THEN 1
				ELSE 2
			END,
			updated_at DESC
		LIMIT ?
	`
	
	rows, err := d.db.Query(query, targetName, targetName, targetName, targetName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var projects []*Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.Path, &p.Name, &p.RemoteURL, &p.Branch, 
			&p.LastModified, &p.GitCount, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}
	
	return projects, nil
}

// ClearCache clears all cached data
func (d *Database) ClearCache() error {
	tables := []string{"projects", "project_dirs", "github_repos", "dir_usage", "similarity_cache"}
	
	for _, table := range tables {
		_, err := d.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("failed to clear table %s: %w", table, err)
		}
	}
	
	return nil
}

// GetStats returns database statistics
func (d *Database) GetStats() (map[string]int, error) {
	stats := make(map[string]int)
	
	tables := []string{"projects", "project_dirs", "github_repos", "dir_usage"}
	
	for _, table := range tables {
		var count int
		err := d.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			return nil, err
		}
		stats[table] = count
	}
	
	return stats, nil
}