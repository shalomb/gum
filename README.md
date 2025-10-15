# Gum

A modern Go-based CLI tool for managing Git projects and directories, replacing legacy shell scripts with better performance and native database storage.

## Features

- **Project Discovery**: Automatically scan and catalog Git repositories
- **Locate Integration**: 34x faster discovery using system locate database
- **GitHub Sync**: Daily sync of repository metadata (topics, languages, activity)
- **Directory Management**: Track frequently accessed directories
- **GitHub Integration**: Discover and manage GitHub repositories
- **Database Storage**: SQLite-based persistent storage with XDG compliance
- **Performance**: Fast, native Go implementation with intelligent caching

## Performance

Gum delivers exceptional performance through multiple optimization strategies:

### Locate Integration
- **34x Speed Improvement**: Project discovery in 0.125s vs 4.3s with file system scanning
- **Automatic Detection**: Works with plocate, GNU locate, and BSD locate
- **Smart Fallback**: Graceful degradation when locate is unavailable
- **Database Monitoring**: Warns users about stale locate databases

### Technical Optimizations
- **Native Go**: Compiled binary with no runtime dependencies
- **Cron-Based Updates**: Background processes keep data fresh without user delays
- **Database Optimization**: SQLite with WAL mode for concurrent access
- **Parallel Processing**: Concurrent directory scanning
- **Hybrid Approach**: Combines locate bulk discovery with file system accuracy
- **Instant Response**: Always returns cached data immediately

## Installation

### Prerequisites

- **Go 1.21+**: Required for building from source
- **SQLite3**: Required for database operations (usually pre-installed on most systems)

### Quick Install

```bash
git clone https://github.com/shalomb/gum.git
cd gum
make install

# Verify installation
gum version
```

This automatically:
- Builds gum from source
- Installs to `~/.local/bin/gum` (user location)
- Sets proper permissions
- Provides PATH configuration guidance

### Alternative Installation Methods

**User Installation (Recommended)**
```bash
make install-user
# Creates ~/.local/bin if needed
# No sudo required
```

**System Installation**
```bash
make install
# Tries user location first, falls back to system
# May require sudo for system-wide installation
```

**Manual Installation**
```bash
make build
cp gum ~/.local/bin/
chmod +x ~/.local/bin/gum
```

### PATH Configuration

If `~/.local/bin` is not in your PATH, add this to your shell profile:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Dependencies

Gum uses SQLite3 for persistent storage. The database is stored in:
- **Location**: `~/.cache/gum/gum.db` (XDG compliant)
- **Driver**: `github.com/mattn/go-sqlite3`

## Usage

### Project Management

```bash
# List all discovered Git projects
gum projects

# List projects with different output formats
gum projects --format json
gum projects --format fzf
gum projects --format simple

# Force refresh project cache
gum projects --refresh

# Clear project cache
gum projects --clear-cache
```

### Directory Management

```bash
# List frequently accessed directories
gum dirs

# List directories with verbose output
gum dirs --verbose

# Force refresh directory cache
gum dirs --refresh
```

### GitHub Integration

```bash
# List GitHub repositories
gum github

# Clone a GitHub repository
gum clone <repository-url>
```

## Configuration

### Project Directories

Gum automatically discovers project directories by scanning:
- `~/projects/` (default)
- `~/projects-*` (glob pattern)
- `~/code/` (default)

### Database Schema

Gum uses SQLite with the following main tables:
- **`projects`**: Git repositories found in project directories
- **`project_dirs`**: Directories that contain projects
- **`github_repos`**: Repositories discovered via GitHub API
- **`dir_usage`**: Directory usage tracking for frequency scoring

See `internal/database/schema.sql` for the complete schema.

### XDG Compliance

Gum follows XDG Base Directory specification:
- **Config**: `~/.config/` (for configuration files)
- **Cache**: `~/.cache/gum/` (for database and temporary data)
- **State**: `~/.local/state/gum/` (for runtime state, planned)

### Cron Job Setup

For optimal performance, set up cron jobs to keep data fresh:

```bash
# Add to crontab (crontab -e)
0 */6 * * * gum projects --refresh  # Refresh projects every 6 hours
0 * * * * gum dirs --refresh        # Refresh directories every hour
```

This ensures data is always fresh without impacting user response times.

## Development

### Building

```bash
go build -o gum
```

### Testing

```bash
go test ./...
```

### Database Operations

The database is automatically initialized on first run. You can inspect it directly:

```bash
sqlite3 ~/.cache/gum/gum.db
```

## Troubleshooting

### Installation Issues

**Command not found**
```bash
# Check if gum is installed
which gum

# If not found, check PATH
echo $PATH | grep -q ".local/bin" || echo "~/.local/bin not in PATH"

# Add to PATH if needed
export PATH="$HOME/.local/bin:$PATH"
```

**Permission denied**
```bash
# Check permissions
ls -la ~/.local/bin/gum

# Fix permissions if needed
chmod +x ~/.local/bin/gum
```

**SQLite not found**
```bash
# Install SQLite3
sudo apt-get install sqlite3  # Ubuntu/Debian
brew install sqlite3          # macOS
```

### First Run Issues

**No projects found**
```bash
# Check if directories exist
ls -la ~/projects ~/code ~/dev

# Force refresh discovery
gum projects --refresh
```

**Config generation issues**
```bash
# Check config directory
ls -la ~/.config/gum/

# Manually create if needed
mkdir -p ~/.config/gum
```

## Migration from Legacy Scripts

Gum replaces the following legacy shell scripts:
- `projects-list` → `gum projects`
- `cwds-list` → `gum dirs`

The tool automatically migrates existing `~/.config/projects-dirs.list` configuration.

## License

See [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Architecture

- **CLI**: Cobra-based command structure
- **Database**: SQLite with WAL mode for concurrency
- **Caching**: Intelligent TTL-based caching system
- **XDG**: Full XDG Base Directory compliance
- **Performance**: Native Go with optimized database queries