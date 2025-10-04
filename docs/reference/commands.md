# Gum Commands Reference

## Overview

Gum provides commands for project discovery, directory management, and cache operations.

## Commands

### `gum projects`

List and search Git projects from configured directories.

#### Syntax
```bash
gum projects [flags]
```

#### Flags
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | "simple" | Output format: simple, json, table |
| `--refresh` | bool | false | Force refresh of project discovery |
| `--search` | string | "" | Search projects by name |
| `--similar` | string | "" | Find projects similar to given name |
| `--limit` | int | 0 | Limit number of results (0 = no limit) |

#### Examples
```bash
# List all projects
gum projects

# JSON output
gum projects --format json

# Search for projects
gum projects --search "api"

# Find similar projects
gum projects --similar "web"

# Limit results
gum projects --limit 10

# Force refresh
gum projects --refresh
```

#### Output Formats

**Simple Format** (default):
```
~/projects/my-app
~/projects/website
~/oneTakeda/api-service
```

**JSON Format**:
```json
[
  {
    "name": "my-app",
    "path": "~/projects/my-app",
    "directory": "~/projects"
  },
  {
    "name": "website", 
    "path": "~/projects/website",
    "directory": "~/projects"
  }
]
```

**Table Format**:
```
NAME        PATH                    DIRECTORY
my-app      ~/projects/my-app      ~/projects
website     ~/projects/website      ~/projects
api-service ~/oneTakeda/api-service ~/oneTakeda
```

### `gum dirs`

List frequently accessed directories with intelligent frecency scoring.

#### Syntax
```bash
gum dirs [flags]
```

#### Flags
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | "default" | Output format: default, fzf, json, simple |
| `--verbose` | bool | false | Show frecency scores and additional information |
| `--refresh` | bool | false | Force refresh cache with current processes |
| `--clear-cache` | bool | false | Clear directory cache |
| `--demo` | bool | false | Show frecency algorithm demonstration |

#### Examples
```bash
# List frequently accessed directories
gum dirs

# Show frecency scores
gum dirs --verbose

# Demonstrate frecency algorithm
gum dirs --demo

# Force refresh with current processes
gum dirs --refresh

# Clear cache
gum dirs --clear-cache
```

#### Output Formats

**Default Format**:
```
~/oneTakeda/terraform-teams-chatbot
~/.config/dotfiles
~/oneTakeda/machi/machi-core
~/projects-local/gum
```

**Verbose Format**:
```
3218	~/oneTakeda/terraform-teams-chatbot
2708	~/.config/dotfiles
1791	~/oneTakeda/machi/machi-core
1791	~/projects-local/gum
```

**JSON Format**:
```json
[
  {
    "path": "~/oneTakeda/terraform-teams-chatbot",
    "score": 3218,
    "frequency": 23,
    "last_seen": "2025-10-04T17:32:38+02:00"
  }
]
```

### `gum dirs-cache`

Manage directory cache operations.

### `gum --crontab`

Generate ideal crontab configuration for automatic updates.

#### Syntax
```bash
gum --crontab
```

#### Description
Generates a complete crontab configuration optimized for gum automation:

- **Smart Detection**: Checks existing crontab entries to avoid duplicates
- **Path Resolution**: Automatically finds gum executable path
- **Updatedb Integration**: Includes locate database updates
- **Customizable Options**: Provides commented advanced features

#### Output Format
```bash
# Ideal crontab configuration for gum
# Generated on: [timestamp]
#
# To install: crontab -e
# Copy the lines below (uncomment as needed)

# Essential entries
0 2 * * * /usr/bin/updatedb                    # Update locate database
0 3 * * * /path/to/gum projects --refresh      # Refresh projects
0 */2 * * * /path/to/gum dirs --refresh        # Refresh directories
0 */6 * * * /path/to/gum update                # Update databases

# Optional: Data export, cleanup, monitoring
# [commented advanced options]
```

#### Examples
```bash
# Generate and review
gum --crontab > ~/gum-crontab.txt
vim ~/gum-crontab.txt

# Install directly
gum --crontab | crontab -

# Check current crontab
crontab -l | grep gum
```

#### Syntax
```bash
gum dirs-cache [flags]
```

#### Flags
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--clear` | bool | false | Clear directory cache |
| `--info` | bool | false | Show cache information |
| `--refresh` | bool | false | Refresh cache |

#### Examples
```bash
# Clear cache
gum dirs-cache --clear

# Show cache info
gum dirs-cache --info

# Refresh cache
gum dirs-cache --refresh
```

### `gum github`

GitHub integration commands.

#### Syntax
```bash
gum github [subcommand] [flags]
```

#### Subcommands
- `clone`: Clone GitHub repository
- `search`: Search GitHub repositories

#### Examples
```bash
# Clone repository
gum github clone owner/repo

# Search repositories
gum github search "golang cli"
```

### `gum clone`

Clone Git repositories.

#### Syntax
```bash
gum clone [repository] [flags]
```

#### Flags
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--directory` | string | "" | Target directory |
| `--depth` | int | 0 | Clone depth (0 = full clone) |

#### Examples
```bash
# Clone repository
gum clone https://github.com/owner/repo.git

# Clone to specific directory
gum clone https://github.com/owner/repo.git --directory ~/projects
```

### `gum update`

Update gum configuration and cache.

#### Syntax
```bash
gum update [flags]
```

#### Examples
```bash
# Update configuration
gum update
```

## Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--help` | bool | false | Show help information |
| `--version` | bool | false | Show version information |

## Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build gum binary from source |
| `make install` | Smart installation (prefers user location) |
| `make install-user` | Explicit user installation to `~/.local/bin` |
| `make test` | Run unit tests |
| `make test-integration` | Run all tests including integration |
| `make test-coverage` | Run tests with coverage reporting |
| `make test-clean` | Clean test artifacts |
| `make test-pkg PKG=<package>` | Run tests for specific package |
| `make test-race` | Run tests with race detection |
| `make test-bench` | Run benchmarks |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `XDG_CONFIG_HOME` | Configuration directory | `~/.config` |
| `XDG_CACHE_HOME` | Cache directory | `~/.cache` |
| `XDG_DATA_HOME` | Data directory | `~/.local/share` |
| `XDG_STATE_HOME` | State directory | `~/.local/state` |
| `GUM_DEBUG` | Enable debug logging | `false` |

## Configuration Files

### YAML Configuration
**Location**: `~/.config/gum/config.yaml`

```yaml
projects:
  - ~/projects
  - ~/oneTakeda
  - ~/projects-local
```

### Legacy Configuration
**Location**: `~/.config/projects-dirs.list`

```
# Project directories
~/projects
~/oneTakeda
~/projects-local
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Database error |
| 4 | Permission error |

## Error Messages

### Common Errors

**Configuration not found**:
```
gum: No configuration found
gum: Run 'gum projects --refresh' to generate config stub
```

**Permission denied**:
```
gum: Permission denied: /path/to/directory
gum: Check directory permissions
```

**Database error**:
```
gum: Database error: unable to open database file
gum: Check XDG_DATA_HOME permissions
```

**Invalid format**:
```
gum: Invalid output format: 'invalid'
gum: Supported formats: simple, json, table
```

## Performance Notes

- **Caching**: Results are cached for 1 hour by default
- **Scanning**: Only directories with Git repositories are scanned
- **Database**: SQLite database for fast lookups
- **Concurrency**: Parallel directory scanning for performance

## Security Considerations

- **Permissions**: Respects file system permissions
- **Paths**: Validates directory paths before scanning
- **Database**: SQLite database with appropriate permissions
- **Configuration**: Validates YAML configuration syntax