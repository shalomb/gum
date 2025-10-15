# Getting Started with Gum

## What is Gum?

Gum is a smart project discovery and management tool that automatically finds your Git repositories and provides fast access to them through a unified command-line interface.

## Quick Start

### Installation

#### Prerequisites
```bash
# Install SQLite3 (required dependency)
sudo apt-get install sqlite3
# or on macOS: brew install sqlite3
```

#### Quick Install
```bash
# Clone and install gum
git clone <repository>
cd gum
make install
```

This will automatically:
- Build gum from source
- Install to `~/.local/bin/gum` (user location)
- Set proper permissions
- Provide PATH configuration guidance

#### Alternative Installation Methods

**User Installation (Recommended)**
```bash
make install-user
# Creates ~/.local/bin if needed
# Installs to user directory (no sudo required)
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

#### PATH Configuration

If `~/.local/bin` is not in your PATH, add this to your shell profile:

**Bash** (`~/.bashrc` or `~/.bash_profile`)
```bash
export PATH="$HOME/.local/bin:$PATH"
```

**Zsh** (`~/.zshrc`)
```bash
export PATH="$HOME/.local/bin:$PATH"
```

**Fish** (`~/.config/fish/config.fish`)
```fish
set -gx PATH $HOME/.local/bin $PATH
```

### First Run

```bash
# Discover your projects automatically
gum projects

# This will:
# 1. Scan common directories for Git repositories
# 2. Generate a config stub if needed
# 3. Show all your projects
# 4. Provide helpful feedback about discovery
```

#### Verification

```bash
# Check installation
gum version

# See available commands
gum help

# Test project discovery
gum projects --format simple
```

### Example Output

```
~/code/project1
~/code/project2
~/projects/my-app
~/projects/website

gum: Auto-discovered 2 project directories
gum: Generated config stub at ~/.config/gum/config.yaml
gum: Edit the config to customize directory scanning
```

## Basic Commands

### List Projects
```bash
gum projects                    # List all projects
gum projects --format simple   # Simple format (one per line)
gum projects --format json     # JSON output
```

### List Directories
```bash
gum dirs                        # List frequently accessed directories
gum dirs --verbose             # Show frecency scores
gum dirs --demo                # Demonstrate frecency algorithm
gum dirs --refresh             # Force refresh with current processes
```

### Search Projects
```bash
gum projects --search "api"    # Find projects containing "api"
gum projects --similar "web"   # Find projects similar to "web"
```

### Refresh Cache
```bash
gum projects --refresh         # Force refresh of project discovery
gum dirs --refresh             # Force refresh of directory tracking
```

### Crontab Automation
```bash
gum --crontab                  # Generate ideal crontab configuration
# Copy the output to your crontab: crontab -e
```

### Performance Optimization
```bash
gum projects --verbose          # Show locate database usage
# Automatically uses locate for 34x faster discovery when available
```

### Version Information
```bash
gum version                     # Show basic version info
gum version --verbose           # Show detailed build information
```

### GitHub Repository Sync
```bash
gum sync --dry-run             # Test what would be synced
gum sync --type full           # Sync all repositories (initial setup)
gum sync --type incremental    # Sync only stale repositories (daily)
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

## Next Steps

- [Learn about configuration](how-to-guides/configuration.md)
- [Explore advanced features](how-to-guides/advanced-usage.md)
- [Understand the architecture](explanation/architecture.md)