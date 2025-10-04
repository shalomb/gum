# Advanced Gum Usage

## Command Reference

### Projects Command
```bash
gum projects [flags]
```

#### Flags
- `--format string`: Output format (simple, json, table) (default "simple")
- `--refresh`: Force refresh of project discovery
- `--search string`: Search projects by name
- `--similar string`: Find projects similar to given name
- `--limit int`: Limit number of results (default 0 = no limit)

#### Examples
```bash
# Basic usage
gum projects

# JSON output for scripting
gum projects --format json

# Search for specific projects
gum projects --search "api"

# Find similar projects
gum projects --similar "web"

# Limit results
gum projects --limit 10
```

### Directory Management
```bash
gum dirs [flags]
```

#### Flags
- `--format string`: Output format (default, fzf, json, simple)
- `--verbose`: Show frecency scores and additional information
- `--refresh`: Force refresh cache with current processes
- `--clear-cache`: Clear directory cache
- `--demo`: Show frecency algorithm demonstration

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

### Cache Management
```bash
gum dirs-cache [flags]
```

### Crontab Automation
```bash
gum --crontab
```

Generate an ideal crontab configuration for automatic updates:

#### Features
- **Smart Detection**: Checks for existing crontab entries
- **Path Resolution**: Automatically finds gum executable path
- **Updatedb Integration**: Includes locate database updates
- **Customizable**: Provides commented options for advanced users

#### Example Output
```bash
# Ideal crontab configuration for gum
# Generated on: Sat  4 Oct 17:50:42 CEST 2025
#
# To install: crontab -e
# Copy the lines below (uncomment as needed)

# Update locate database daily at 2 AM
0 2 * * * /usr/bin/updatedb

# Refresh project cache daily at 3 AM (after updatedb)
0 3 * * * /home/user/.local/bin/gum projects --refresh

# Refresh directory cache every 2 hours
0 */2 * * * /home/user/.local/bin/gum dirs --refresh

# Update databases every 6 hours
0 */6 * * * /home/user/.local/bin/gum update
```

#### Installation
```bash
# Generate configuration
gum --crontab > ~/gum-crontab.txt

# Review and edit
vim ~/gum-crontab.txt

# Install
crontab ~/gum-crontab.txt
```

#### Flags
- `--clear`: Clear directory cache
- `--info`: Show cache information
- `--refresh`: Refresh cache

#### Examples
```bash
# Clear cache
gum dirs-cache --clear

# Show cache info
gum dirs-cache --info
```

## Installation and Setup

### Make Targets

**Available Targets**
```bash
make build          # Build gum binary
make install        # Smart installation (prefers user location)
make install-user   # Explicit user installation
make test           # Run unit tests
make test-integration # Run all tests including integration
make test-coverage  # Run tests with coverage reporting
```

**Installation Logic**
- `make install` tries `~/.local/bin` first, falls back to system locations
- `make install-user` forces user installation, creates directory if needed
- Both targets provide helpful PATH configuration messages

### Environment Setup

**XDG Compliance**
Gum follows XDG Base Directory specification:
- Config: `~/.config/gum/` (respects `XDG_CONFIG_HOME`)
- Cache: `~/.cache/gum/` (respects `XDG_CACHE_HOME`)
- Data: `~/.local/share/gum/` (respects `XDG_DATA_HOME`)
- State: `~/.local/state/gum/` (respects `XDG_STATE_HOME`)

**Custom Environment**
```bash
# Set custom XDG directories
export XDG_CONFIG_HOME="/custom/config"
export XDG_CACHE_HOME="/custom/cache"
export XDG_DATA_HOME="/custom/data"

# Install and run gum
make install
gum projects
```

## Integration Examples

### Shell Integration

#### Bash Function
```bash
# Add to ~/.bashrc
g() {
    local project=$(gum projects --search "$1" | head -1)
    if [ -n "$project" ]; then
        cd "$project"
    else
        echo "Project not found: $1"
    fi
}
```

#### Zsh Function
```bash
# Add to ~/.zshrc
function g() {
    local project=$(gum projects --search "$1" | head -1)
    if [ -n "$project" ]; then
        cd "$project"
    else
        echo "Project not found: $1"
    fi
}
```

### Tmux Integration

#### tmux.conf
```bash
# Quick project switching
bind-key C-p new-window -c "$(gum projects --format simple | fzf)"
```

### Vim/Neovim Integration

#### Telescope Plugin
```lua
-- telescope.lua
local telescope = require('telescope')
telescope.setup({
  extensions = {
    gum = {
      command = 'gum projects --format json',
      parser = function(output)
        local projects = vim.json.decode(output)
        return vim.tbl_map(function(project)
          return {
            value = project.path,
            display = project.name,
            ordinal = project.name,
          }
        end, projects)
      end
    }
  }
})
```

## Scripting Examples

### Find and Open Project
```bash
#!/bin/bash
# open-project.sh

PROJECT_NAME="$1"
if [ -z "$PROJECT_NAME" ]; then
    echo "Usage: $0 <project-name>"
    exit 1
fi

PROJECT_PATH=$(gum projects --search "$PROJECT_NAME" --format simple | head -1)
if [ -n "$PROJECT_PATH" ]; then
    cd "$PROJECT_PATH"
    exec "$SHELL"
else
    echo "Project not found: $PROJECT_NAME"
    exit 1
fi
```

### Project Statistics
```bash
#!/bin/bash
# project-stats.sh

echo "Project Statistics:"
echo "=================="

TOTAL=$(gum projects --format simple | wc -l)
echo "Total projects: $TOTAL"

# Count by directory
echo ""
echo "Projects by directory:"
gum projects --format simple | sed 's|/[^/]*$||' | sort | uniq -c | sort -nr
```

### Automated Backup
```bash
#!/bin/bash
# backup-projects.sh

BACKUP_DIR="/backup/projects"
mkdir -p "$BACKUP_DIR"

gum projects --format simple | while read project; do
    if [ -d "$project" ]; then
        project_name=$(basename "$project")
        echo "Backing up: $project_name"
        tar -czf "$BACKUP_DIR/$project_name.tar.gz" -C "$(dirname "$project")" "$project_name"
    fi
done
```

## Performance Optimization

### Large Project Collections

#### Limit Scanning
```yaml
# ~/.config/gum/config.yaml
projects:
  - ~/active-projects    # Only scan active projects
  # - ~/archive-projects  # Comment out large archives
```

#### Cache Optimization
```bash
# Refresh cache periodically
gum projects --refresh

# Clear cache if corrupted
gum dirs-cache --clear
```

### Network Drives

#### Exclude Slow Paths
```yaml
# ~/.config/gum/config.yaml
projects:
  - ~/local-projects
  # - /mnt/slow-network-drive  # Exclude slow paths
```

## Troubleshooting

### Common Issues

#### Slow Performance
```bash
# Check cache status
gum dirs-cache --info

# Refresh cache
gum projects --refresh
```

#### Missing Projects
```bash
# Check configuration
cat ~/.config/gum/config.yaml

# Verify directories exist
ls -la ~/projects

# Force refresh
gum projects --refresh
```

#### Permission Errors
```bash
# Check directory permissions
ls -la ~/projects

# Fix permissions if needed
chmod 755 ~/projects
```

### Debug Mode

#### Verbose Output
```bash
# Enable debug logging
export GUM_DEBUG=1
gum projects --refresh
```

#### Check Database
```bash
# Inspect SQLite database
sqlite3 ~/.local/share/gum/gum.db "SELECT * FROM projects LIMIT 10;"
```

## Best Practices

### Configuration Management
1. **Use version control**: Track your `config.yaml`
2. **Document decisions**: Comment why directories are included
3. **Regular cleanup**: Remove unused directories
4. **Test changes**: Verify config changes work

### Performance
1. **Limit scope**: Only scan directories you use
2. **Use cache**: Let gum cache results
3. **Regular refresh**: Update cache periodically
4. **Monitor size**: Watch database growth

### Integration
1. **Shell functions**: Create convenient shortcuts
2. **Editor plugins**: Integrate with your editor
3. **Automation**: Use in scripts and workflows
4. **Documentation**: Document your customizations