# Gum Database Reference

## Overview

Gum uses SQLite3 as its persistent storage backend, providing ACID compliance, zero configuration, and excellent performance for CLI tools. This document provides a complete reference for the database schema, operations, and integration patterns.

## Database Location

- **Path**: `~/.cache/gum/gum.db` (XDG compliant)
- **Environment**: Respects `XDG_CACHE_HOME` environment variable
- **Mode**: WAL (Write-Ahead Logging) for better concurrency
- **Foreign Keys**: Enabled for data integrity

## Schema

### Core Tables

#### `projects`
Stores discovered Git repositories with metadata.

```sql
CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    remote_url TEXT,
    branch TEXT,
    last_modified DATETIME,
    git_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**Fields:**
- `path`: Full filesystem path to the repository
- `name`: Repository name (basename of path)
- `remote_url`: Primary remote URL (origin)
- `branch`: Current branch name
- `last_modified`: Last filesystem modification time
- `git_count`: Number of Git repositories in this directory

#### `project_dirs`
Tracks directories that contain projects.

```sql
CREATE TABLE project_dirs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    last_scanned DATETIME,
    git_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**Fields:**
- `path`: Directory path to scan for projects
- `last_scanned`: Last time directory was scanned
- `git_count`: Number of Git repositories found

#### `github_repos`
GitHub repositories discovered via API.

```sql
CREATE TABLE github_repos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    full_name TEXT NOT NULL UNIQUE,
    description TEXT,
    url TEXT,
    clone_url TEXT,
    ssh_url TEXT,
    is_private BOOLEAN DEFAULT 0,
    is_fork BOOLEAN DEFAULT 0,
    updated_at DATETIME,
    last_discovered DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### `dir_usage`
Tracks directory usage for frequency scoring.

```sql
CREATE TABLE dir_usage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL,
    frequency INTEGER DEFAULT 1,
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(path)
);
```

#### `similarity_cache`
Pre-computed similarity scores for performance.

```sql
CREATE TABLE similarity_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_type TEXT NOT NULL,
    source_id INTEGER NOT NULL,
    target_type TEXT NOT NULL,
    target_id INTEGER NOT NULL,
    similarity_score REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_type, source_id, target_type, target_id)
);
```

## Indexes

Performance indexes are automatically created:

```sql
-- Projects indexes
CREATE INDEX idx_projects_path ON projects(path);
CREATE INDEX idx_projects_name ON projects(name);
CREATE INDEX idx_projects_remote ON projects(remote_url);
CREATE INDEX idx_projects_updated ON projects(updated_at);

-- Project directories indexes
CREATE INDEX idx_project_dirs_path ON project_dirs(path);
CREATE INDEX idx_project_dirs_scanned ON project_dirs(last_scanned);

-- GitHub repos indexes
CREATE INDEX idx_github_repos_full_name ON github_repos(full_name);
CREATE INDEX idx_github_repos_name ON github_repos(name);
CREATE INDEX idx_github_repos_updated ON github_repos(updated_at);
CREATE INDEX idx_github_repos_private ON github_repos(is_private);

-- Directory usage indexes
CREATE INDEX idx_dir_usage_path ON dir_usage(path);
CREATE INDEX idx_dir_usage_frequency ON dir_usage(frequency);
CREATE INDEX idx_dir_usage_last_seen ON dir_usage(last_seen);

-- Similarity cache indexes
CREATE INDEX idx_similarity_source ON similarity_cache(source_type, source_id);
CREATE INDEX idx_similarity_target ON similarity_cache(target_type, target_id);
CREATE INDEX idx_similarity_score ON similarity_cache(similarity_score);
```

## Triggers

Automatic timestamp updates:

```sql
CREATE TRIGGER update_projects_timestamp 
    AFTER UPDATE ON projects
    BEGIN
        UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER update_project_dirs_timestamp 
    AFTER UPDATE ON project_dirs
    BEGIN
        UPDATE project_dirs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

-- ... similar triggers for other tables
```

## Operations

### Project Management

```go
// Upsert a project (insert or update)
func (d *Database) UpsertProject(project *Project) error

// Get all projects, optionally sorted by similarity
func (d *Database) GetProjects(sortBySimilarity bool, targetPath string) ([]*Project, error)

// Get similar projects
func (d *Database) GetSimilarProjects(targetPath string, limit int) ([]*Project, error)
```

### Directory Management

```go
// Upsert a project directory
func (d *Database) UpsertProjectDir(dir *ProjectDir) error

// Get all project directories
func (d *Database) GetProjectDirs() ([]*ProjectDir, error)
```

### Usage Tracking

```go
// Track directory usage
func (d *Database) UpsertDirUsage(usage *DirUsage) error

// Get frequently used directories
func (d *Database) GetFrequentDirs(limit int) ([]*DirUsage, error)
```

### Maintenance

```go
// Clear all cached data
func (d *Database) ClearCache() error

// Get database statistics
func (d *Database) GetStats() (map[string]int, error)
```

## Migration from Legacy Files

### From `projects-dirs.list`

The tool automatically migrates existing `~/.config/projects-dirs.list` files:

```bash
# Format: One directory per line
~/projects/
~/oneTakeda/
~/projects/docker-images/
```

Migration process:
1. Read `projects-dirs.list` if it exists
2. Insert directories into `project_dirs` table
3. Mark as migrated to avoid re-processing

## Performance Considerations

### Caching Strategy

- **Projects**: 5-minute TTL (projects change less frequently)
- **Directories**: 30-second TTL (directories change more frequently)
- **Project Dirs**: 1-hour TTL (project directories change rarely)

### Query Optimization

- **Indexes**: All frequently queried columns are indexed
- **Similarity**: Pre-computed similarity scores cached
- **Batch Operations**: Bulk inserts/updates for efficiency

### Database Maintenance

```bash
# Analyze database performance
sqlite3 ~/.cache/gum/gum.db "ANALYZE;"

# Check database integrity
sqlite3 ~/.cache/gum/gum.db "PRAGMA integrity_check;"

# Vacuum to reclaim space
sqlite3 ~/.cache/gum/gum.db "VACUUM;"
```

## Troubleshooting

### Common Issues

1. **Database locked**: Usually resolves automatically with WAL mode
2. **Corrupted database**: Delete `~/.cache/gum/gum.db` to recreate
3. **Permission errors**: Ensure write access to `~/.cache/gum/`

### Debugging

```bash
# Inspect database directly
sqlite3 ~/.cache/gum/gum.db

# Check table schemas
.schema

# View recent projects
SELECT * FROM projects ORDER BY updated_at DESC LIMIT 10;

# Check project directories
SELECT * FROM project_dirs;
```

## Future Enhancements

- **Full-text search**: SQLite FTS for project search
- **Backup/restore**: Database export/import functionality
- **Replication**: Multi-machine synchronization
- **Compression**: Database compression for large datasets