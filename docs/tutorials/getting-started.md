# Getting Started with Gum

## What is Gum?

Gum is a smart project discovery and management tool that automatically finds your Git repositories and provides fast access to them through a unified command-line interface.

## Quick Start

### Installation

```bash
# Install dependencies
sudo apt-get install sqlite3

# Build gum
git clone <repository>
cd gum
go build -o gum
sudo mv gum /usr/local/bin/
```

### First Run

```bash
# Discover your projects automatically
gum projects

# This will:
# 1. Scan common directories for Git repositories
# 2. Generate a config stub if needed
# 3. Show all your projects
```

### Example Output

```
~/oneTakeda/project1
~/oneTakeda/project2
~/projects-local/my-app
~/projects-local/website

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

### Search Projects
```bash
gum projects --search "api"    # Find projects containing "api"
gum projects --similar "web"   # Find projects similar to "web"
```

### Refresh Cache
```bash
gum projects --refresh         # Force refresh of project discovery
```

## Next Steps

- [Learn about configuration](how-to-guides/configuration.md)
- [Explore advanced features](how-to-guides/advanced-usage.md)
- [Understand the architecture](explanation/architecture.md)