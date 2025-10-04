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

Manage project directories for scanning.

#### Syntax
```bash
gum dirs [flags]
```

#### Flags
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--add` | string | "" | Add directory to scan list |
| `--remove` | string | "" | Remove directory from scan list |
| `--list` | bool | false | List configured directories |
| `--refresh` | bool | false | Refresh directory cache |

#### Examples
```bash
# List configured directories
gum dirs --list

# Add directory
gum dirs --add ~/new-projects

# Remove directory
gum dirs --remove ~/old-projects

# Refresh cache
gum dirs --refresh
```

### `gum dirs-cache`

Manage directory cache operations.

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