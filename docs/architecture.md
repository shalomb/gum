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
- **`projects.go`**: Project discovery and listing
- **`dirs.go`**: Directory usage tracking
- **`github.go`**: GitHub integration
- **`clone.go`**: Repository cloning
- **`version.go`**: Version information

### 2. Database Layer (`internal/database/`)

- **`database.go`**: Connection management and schema initialization
- **`operations.go`**: CRUD operations for all entities
- **`schema.sql`**: Complete database schema definition

### 3. Cache Layer (`internal/cache/`)

- **`cache.go`**: File-based caching with TTL support
- **Legacy**: JSON-based caching (being replaced by database)

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