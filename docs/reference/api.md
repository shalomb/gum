# Gum API Reference

## Database Schema

### Tables

#### `projects`
Stores discovered Git projects.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER PRIMARY KEY | Unique project identifier |
| `name` | TEXT NOT NULL | Project name (directory name) |
| `path` | TEXT NOT NULL UNIQUE | Full path to project |
| `directory` | TEXT NOT NULL | Parent directory path |
| `last_modified` | DATETIME | Last modification time |
| `created_at` | DATETIME DEFAULT CURRENT_TIMESTAMP | Creation timestamp |

#### `project_dirs`
Stores configured project directories.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER PRIMARY KEY | Unique directory identifier |
| `path` | TEXT NOT NULL UNIQUE | Directory path |
| `enabled` | BOOLEAN DEFAULT 1 | Whether directory is enabled |
| `created_at` | DATETIME DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| `updated_at` | DATETIME DEFAULT CURRENT_TIMESTAMP | Last update timestamp |

#### `github_repos`
Stores GitHub repository information.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER PRIMARY KEY | Unique repository identifier |
| `name` | TEXT NOT NULL | Repository name |
| `full_name` | TEXT NOT NULL UNIQUE | Full repository name (owner/repo) |
| `html_url` | TEXT | GitHub URL |
| `clone_url` | TEXT | Clone URL |
| `description` | TEXT | Repository description |
| `language` | TEXT | Primary language |
| `stars` | INTEGER DEFAULT 0 | Star count |
| `forks` | INTEGER DEFAULT 0 | Fork count |
| `created_at` | DATETIME DEFAULT CURRENT_TIMESTAMP | Creation timestamp |
| `updated_at` | DATETIME DEFAULT CURRENT_TIMESTAMP | Last update timestamp |

#### `dir_usage`
Tracks directory usage statistics.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER PRIMARY KEY | Unique usage record identifier |
| `directory` | TEXT NOT NULL | Directory path |
| `access_count` | INTEGER DEFAULT 0 | Number of accesses |
| `last_accessed` | DATETIME | Last access time |
| `created_at` | DATETIME DEFAULT CURRENT_TIMESTAMP | Creation timestamp |

#### `similarity_cache`
Caches similarity calculations for performance.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER PRIMARY KEY | Unique cache entry identifier |
| `query` | TEXT NOT NULL | Search query |
| `project_name` | TEXT NOT NULL | Project name |
| `similarity_score` | REAL NOT NULL | Calculated similarity score |
| `created_at` | DATETIME DEFAULT CURRENT_TIMESTAMP | Creation timestamp |

### Indexes

```sql
-- Projects indexes
CREATE INDEX idx_projects_name ON projects(name);
CREATE INDEX idx_projects_directory ON projects(directory);
CREATE INDEX idx_projects_path ON projects(path);

-- Project directories indexes
CREATE INDEX idx_project_dirs_path ON project_dirs(path);
CREATE INDEX idx_project_dirs_enabled ON project_dirs(enabled);

-- GitHub repos indexes
CREATE INDEX idx_github_repos_name ON github_repos(name);
CREATE INDEX idx_github_repos_full_name ON github_repos(full_name);
CREATE INDEX idx_github_repos_language ON github_repos(language);

-- Directory usage indexes
CREATE INDEX idx_dir_usage_directory ON dir_usage(directory);
CREATE INDEX idx_dir_usage_last_accessed ON dir_usage(last_accessed);

-- Similarity cache indexes
CREATE INDEX idx_similarity_cache_query ON similarity_cache(query);
CREATE INDEX idx_similarity_cache_score ON similarity_cache(similarity_score);
```

### Triggers

```sql
-- Update timestamp trigger for project_dirs
CREATE TRIGGER update_project_dirs_timestamp 
    AFTER UPDATE ON project_dirs
    FOR EACH ROW
    BEGIN
        UPDATE project_dirs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

-- Update timestamp trigger for github_repos
CREATE TRIGGER update_github_repos_timestamp 
    AFTER UPDATE ON github_repos
    FOR EACH ROW
    BEGIN
        UPDATE github_repos SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
```

## Go API

### Database Operations

#### `New() (*Database, error)`
Creates a new database connection.

```go
db, err := database.New()
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

#### `UpsertProject(project Project) error`
Inserts or updates a project record.

```go
project := database.Project{
    Name:     "my-project",
    Path:     "/home/user/projects/my-project",
    Directory: "/home/user/projects",
}
err := db.UpsertProject(project)
```

#### `GetProjects() ([]Project, error)`
Retrieves all projects.

```go
projects, err := db.GetProjects()
if err != nil {
    log.Fatal(err)
}
```

#### `SearchProjects(query string) ([]Project, error)`
Searches projects by name.

```go
projects, err := db.SearchProjects("api")
if err != nil {
    log.Fatal(err)
}
```

#### `UpsertProjectDir(dir ProjectDir) error`
Inserts or updates a project directory.

```go
dir := database.ProjectDir{
    Path:    "/home/user/projects",
    Enabled: true,
}
err := db.UpsertProjectDir(dir)
```

#### `GetProjectDirs() ([]ProjectDir, error)`
Retrieves all project directories.

```go
dirs, err := db.GetProjectDirs()
if err != nil {
    log.Fatal(err)
}
```

### Cache Operations

#### `New() *Cache`
Creates a new cache instance.

```go
cache := cache.New()
```

#### `Set(key string, value interface{}, ttl time.Duration) error`
Sets a cache entry with TTL.

```go
err := cache.Set("projects", projects, time.Hour)
```

#### `Get(key string, value interface{}) bool`
Retrieves a cache entry.

```go
var projects []Project
if cache.Get("projects", &projects) {
    // Use cached projects
}
```

#### `Clear() error`
Clears all cache entries.

```go
err := cache.Clear()
```

### Configuration

#### YAML Structure
```go
type Config struct {
    Projects []string `yaml:"projects"`
}
```

#### Reading Configuration
```go
func readYAMLConfig() []string {
    configDir := os.Getenv("XDG_CONFIG_HOME")
    if configDir == "" {
        configDir = filepath.Join(os.Getenv("HOME"), ".config")
    }
    
    configFile := filepath.Join(configDir, "gum", "config.yaml")
    data, err := os.ReadFile(configFile)
    if err != nil {
        return nil
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil
    }
    
    return config.Projects
}
```

## File System Operations

### Project Discovery

#### `smartDiscoverProjectDirs(home string) []string`
Discovers project directories intelligently.

```go
dirs := smartDiscoverProjectDirs("/home/user")
// Returns: ["/home/user/projects", "/home/user/oneTakeda"]
```

#### `countGitReposInDir(dir string) int`
Counts Git repositories in a directory.

```go
count := countGitReposInDir("/home/user/projects")
// Returns: 42
```

#### `findGitProjects(dirs []string) []Project`
Finds all Git projects in given directories.

```go
projects := findGitProjects([]string{"/home/user/projects"})
// Returns: []Project{...}
```

### Configuration Generation

#### `generateConfigStub(home string, discoveredDirs []string) string`
Generates a YAML configuration stub.

```go
config := generateConfigStub("/home/user", []string{"/home/user/projects"})
// Returns: YAML configuration string
```

## Error Handling

### Error Types

```go
type ConfigError struct {
    File    string
    Message string
    Err     error
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("config error in %s: %s: %v", e.File, e.Message, e.Err)
}

type DatabaseError struct {
    Operation string
    Message   string
    Err       error
}

func (e *DatabaseError) Error() string {
    return fmt.Sprintf("database error during %s: %s: %v", e.Operation, e.Message, e.Err)
}
```

### Error Handling Patterns

```go
// Configuration errors
if err := readConfig(); err != nil {
    if configErr, ok := err.(*ConfigError); ok {
        log.Printf("Configuration issue: %v", configErr)
        return
    }
    log.Fatal(err)
}

// Database errors
if err := db.UpsertProject(project); err != nil {
    if dbErr, ok := err.(*DatabaseError); ok {
        log.Printf("Database issue: %v", dbErr)
        return
    }
    log.Fatal(err)
}
```

## Performance Considerations

### Caching Strategy
- **TTL**: 1 hour for project discovery
- **Storage**: File-based cache in XDG_CACHE_HOME
- **Invalidation**: Manual refresh or TTL expiration

### Database Optimization
- **Indexes**: On frequently queried columns
- **Batch Operations**: For bulk inserts/updates
- **Connection Pooling**: Single connection with proper cleanup

### Memory Management
- **Streaming**: Large result sets processed in chunks
- **Cleanup**: Proper resource disposal
- **Garbage Collection**: Minimal allocations in hot paths

## Security Considerations

### Path Validation
```go
func validatePath(path string) error {
    if !filepath.IsAbs(path) {
        return fmt.Errorf("path must be absolute: %s", path)
    }
    
    if strings.Contains(path, "..") {
        return fmt.Errorf("path contains parent directory reference: %s", path)
    }
    
    return nil
}
```

### Permission Checks
```go
func checkPermissions(path string) error {
    info, err := os.Stat(path)
    if err != nil {
        return err
    }
    
    if !info.IsDir() {
        return fmt.Errorf("path is not a directory: %s", path)
    }
    
    return nil
}
```

### Input Sanitization
```go
func sanitizeQuery(query string) string {
    // Remove potentially dangerous characters
    query = strings.ReplaceAll(query, "..", "")
    query = strings.ReplaceAll(query, "/", "")
    query = strings.ReplaceAll(query, "\\", "")
    return strings.TrimSpace(query)
}
```