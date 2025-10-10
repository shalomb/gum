# Gum Architecture Explanation

## Overview

Gum is a modern CLI tool built in Go that replaces legacy shell scripts with a robust, database-backed solution for managing Git projects and directories. This document explains the system architecture, design decisions, and implementation details.

## Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: Cobra
- **Database**: SQLite3 with WAL mode
- **Configuration**: Viper
- **Logging**: Logrus
- **XDG Compliance**: adrg/xdg

## Architecture Components

### 1. CLI Layer (`cmd/`)

- **`root.go`**: Main command setup and global flags
- **`projects.go`**: Project discovery and listing with hybrid auto-discovery
- **`dirs.go`**: Directory usage tracking with frecency scoring
- **`github.go`**: GitHub integration
- **`clone.go`**: Repository cloning
- **`version.go`**: Version information
- **`frecency_demo.go`**: Frecency algorithm demonstration

### 2. Database Layer (`internal/database/`)

- **`database.go`**: Connection management and schema initialization
- **`operations.go`**: CRUD operations for all entities
- **`schema.sql`**: Complete database schema definition
- **`cache.go`**: Database-backed caching with TTL support
- **`migration.go`**: JSON to SQLite migration utilities
- **`integrity.go`**: Database integrity monitoring and verification
- **`concurrency_test.go`**: Comprehensive concurrency testing

### 3. Cache Layer (`internal/cache/`)

- **`cache.go`**: File-based caching with TTL support
- **Legacy**: JSON-based caching (being replaced by database)

### 4. New Commands (`cmd/`)

- **`projects_v2.go`**: Database-backed project listing with unified cache
- **`migrate.go`**: Database migration management
- **`integrity.go`**: Database integrity verification
- **`stress_test.go`**: Concurrency stress testing

## Data Flow

```
User Command → CLI Layer → Database Layer → SQLite
                ↓
            Cache Layer (legacy)
```

## Database Design

### Entity Relationship

```
project_dirs (1) → (N) projects
projects (1) → (N) similarity_cache
dir_usage (standalone)
github_repos (standalone)
```

### Key Design Decisions

1. **SQLite Choice**: Zero-configuration, file-based, ACID compliant
2. **WAL Mode**: Better concurrency for CLI tool usage
3. **Foreign Keys**: Data integrity enforcement
4. **Indexes**: Performance optimization for common queries
5. **Triggers**: Automatic timestamp updates

## XDG Compliance

### Directory Usage

- **Config**: `~/.config/` - User configuration files
- **Cache**: `~/.cache/gum/` - Database and temporary data
- **State**: `~/.local/state/gum/` - Runtime state (planned)

### Environment Variables

- `XDG_CONFIG_HOME`: Configuration directory
- `XDG_CACHE_HOME`: Cache directory
- `XDG_STATE_HOME`: State directory (future use)

## Performance Optimizations

### Caching Strategy

- **Projects**: 5-minute TTL (stable data)
- **Directories**: 30-second TTL (frequently changing)
- **Project Dirs**: 1-hour TTL (rarely changing)

## Frecency Algorithm

The frecency algorithm combines frequency and recency to provide intelligent directory ranking:

### Algorithm Components

#### Frequency Component
- **Logarithmic Scaling**: `log(frequency + 1)` prevents high-frequency domination
- **Diminishing Returns**: Doubling frequency doesn't double the score
- **Prevents Inflation**: High-frequency directories don't dominate forever

#### Recency Component (Multi-Tier Decay)
- **Recent (0-1h)**: No decay (multiplier = 1.0)
- **Today (1-24h)**: Mild decay (exp(-0.1 * hours))
- **This Week (1-7d)**: Moderate decay (exp(-0.05 * hours) * 0.9)
- **This Month (1-30d)**: Stronger decay (exp(-0.02 * hours) * 0.5)
- **Older (30d+)**: Significant decay but never zero (min 1%)

### Formula
```
score = log(frequency + 1) * recency_multiplier * 1000
```

### Benefits
- **Natural Aging**: Recent directories float to the top
- **Frequency Balance**: High-frequency directories don't dominate forever
- **Smooth Transitions**: No hard cutoffs or sudden changes
- **Accessibility**: Minimum score ensures all directories remain accessible

## Locate Integration

Gum integrates with the system locate database for fast file discovery:

### Architecture
- **LocateFinder**: Wraps locate command with database detection
- **Hybrid Approach**: Uses locate for bulk discovery + file system for recent changes
- **Smart Fallback**: Graceful degradation when locate is unavailable

### Performance Benefits
- **34x Speed Improvement**: 4.3s → 0.125s for large directories
- **Database Detection**: Works with plocate, GNU locate, BSD locate
- **Freshness Monitoring**: Warns users about stale databases

### Implementation
```go
type LocateFinder struct {
    available    bool
    databasePath string
    lastUpdated  time.Time
}

func (lf *LocateFinder) FindGitRepos(basePath string) ([]string, error)
func (lf *LocateFinder) GetStatus() LocateStatus
```

## GitHub Integration

Gum integrates with GitHub API for rich repository metadata:

### Architecture
- **GitHubClient**: Wraps GitHub API with authentication via gh CLI
- **Metadata Sync**: Daily sync of repository metadata
- **Smart Caching**: Database-backed caching with TTL management
- **Cross-Org Access**: Discovers repositories across organizations

### Performance Benefits
- **Rich Metadata**: Topics, languages, activity metrics, timestamps
- **Smart Sync**: Full, incremental, and metadata-only sync modes
- **API Efficiency**: Batch processing with rate limit compliance
- **Offline Access**: Cached data available without network

### Implementation
```go
type GitHubClient struct {
    token      string
    httpClient *http.Client
    rateLimiter *RateLimiter
}

type GitHubMetadata struct {
    Name        string    `json:"name"`
    FullName    string    `json:"full_name"`
    Description string    `json:"description"`
    Topics      []string  `json:"topics"`
    Language    string    `json:"language"`
    StarCount   int       `json:"stargazers_count"`
    ForkCount   int       `json:"forks_count"`
    LastPushed  time.Time `json:"pushed_at"`
}

func (gc *GitHubClient) DiscoverAllRepositories() ([]*GitHubMetadata, error)
func (gc *GitHubClient) GetRepositoryMetadata(owner, repo string) (*GitHubMetadata, error)
```

### Database Optimizations

- **Indexes**: All frequently queried columns
- **Batch Operations**: Bulk inserts/updates
- **Similarity Cache**: Pre-computed similarity scores
- **WAL Mode**: Better concurrency

## Migration Strategy

### From Legacy Files

1. **`projects-dirs.list`** → `project_dirs` table
2. **JSON cache files** → Database tables
3. **Shell scripts** → Go commands

### Backward Compatibility

- Automatic migration of existing files
- Graceful fallback to legacy behavior
- Clear deprecation warnings

## Security Considerations

### Database Security

- **File Permissions**: Database file restricted to user
- **SQL Injection**: Parameterized queries only
- **Path Validation**: Sanitized file paths

### XDG Security

- **Isolation**: User-specific directories
- **Permissions**: Proper file/directory permissions
- **Environment**: Respects XDG environment variables

## Error Handling

### Database Errors

- **Connection Issues**: Graceful degradation
- **Schema Errors**: Automatic migration
- **Corruption**: Recreate database

### CLI Errors

- **Validation**: Input validation and sanitization
- **User Feedback**: Clear error messages
- **Recovery**: Automatic retry mechanisms

## Testing Strategy

### Unit Tests

- **Database Operations**: CRUD operations
- **CLI Commands**: Command execution
- **Cache Layer**: TTL and persistence

### Integration Tests

- **End-to-End**: Full command execution
- **Database**: Schema and data integrity
- **Migration**: Legacy file migration

## Future Enhancements

### Planned Features

1. **Full-Text Search**: SQLite FTS integration
2. **Backup/Restore**: Database export/import
3. **Multi-Machine Sync**: Cross-device synchronization
4. **Plugin System**: Extensible command system
5. **Web Interface**: Optional web UI

### Performance Improvements

1. **Connection Pooling**: Database connection management
2. **Query Optimization**: Advanced indexing strategies
3. **Compression**: Database compression
4. **Replication**: Read replicas for large datasets

## Development Guidelines

### Code Organization

- **Packages**: Clear separation of concerns
- **Interfaces**: Dependency injection ready
- **Error Handling**: Consistent error patterns
- **Logging**: Structured logging throughout

### Database Guidelines

- **Migrations**: Version-controlled schema changes
- **Indexes**: Performance-first indexing
- **Queries**: Parameterized queries only
- **Transactions**: Proper transaction management

### CLI Guidelines

- **Commands**: Single responsibility principle
- **Flags**: Consistent flag naming
- **Output**: Structured output formats
- **Help**: Comprehensive help text