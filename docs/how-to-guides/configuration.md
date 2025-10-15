# How to Configure Gum

## Configuration Overview

Gum uses a hybrid approach to project discovery:
1. **Auto-discovery**: Automatically finds directories with Git repositories
2. **YAML configuration**: Explicit control when needed
3. **Legacy support**: Works with existing `projects-dirs.list` files

## Auto-Discovery (Default)

Gum automatically scans these common directories:
- `~/projects`
- `~/code`
- `~/dev`
- `~/workspace`
- `~/repos`
- `~/repositories`

Only directories containing Git repositories are included.

## YAML Configuration

### Location
```
~/.config/gum/config.yaml
```

### Basic Structure
```yaml
projects:
  - ~/my-projects
  - ~/work/repos
  - /path/to/any/directory
```

### Example Configuration
```yaml
# Gum Configuration
# Edit this file to customize which directories gum scans for Git repositories

projects:
  - ~/code               # 599 Git repositories
  - ~/projects           # 13 Git repositories
  - ~/work/projects      # Work repositories
  - ~/personal/code      # Personal projects

# Additional directories you can add:
# - ~/code
# - ~/dev
# - ~/workspace
# - ~/repos
# - ~/repositories
# - /path/to/any/directory

# Note: Directories with 0 Git repositories will be ignored
# Remove directories from this list to exclude them from scanning
```

## Configuration Generation

Gum automatically generates a config stub when:
- No existing config file exists
- Multiple project directories are discovered
- You want explicit control over scanning

### Generated Config Features
- **Pre-populated**: Contains discovered directories
- **Documented**: Includes helpful comments
- **Customizable**: Easy to modify or extend

## Legacy Configuration

### projects-dirs.list Support
Gum still supports the legacy `~/.config/projects-dirs.list` format:

```
# Project directories
~/projects
~/code
~/dev
```

### Migration
To migrate from `projects-dirs.list` to YAML:
1. Run `gum projects --refresh` to generate config stub
2. Copy entries from `projects-dirs.list` to `config.yaml`
3. Remove `projects-dirs.list` when ready

## Configuration Priority

1. **YAML config** (`~/.config/gum/config.yaml`) - highest priority
2. **Auto-discovery** - when no YAML config exists
3. **Legacy list** (`~/.config/projects-dirs.list`) - fallback

## Troubleshooting

### Config Not Loading
```bash
# Check config file exists
ls -la ~/.config/gum/config.yaml

# Validate YAML syntax
gum projects --refresh
```

### Wrong Directories Scanned
```bash
# Edit config to exclude directories
vim ~/.config/gum/config.yaml

# Remove unwanted entries:
# - ~/unwanted-dir  # Remove this line
```

### Performance Issues
```bash
# Limit scanning to specific directories
# Edit config to include only needed paths
vim ~/.config/gum/config.yaml
```

## Advanced Configuration

### Environment Variables
```bash
export XDG_CONFIG_HOME=/custom/config/path
export XDG_CACHE_HOME=/custom/cache/path
export XDG_DATA_HOME=/custom/data/path
```

### Custom Cache Location
```bash
# Set custom cache directory
export XDG_CACHE_HOME=/fast/ssd/cache
gum projects --refresh
```

## Best Practices

1. **Start with auto-discovery**: Let gum find your projects automatically
2. **Generate config when needed**: Use config stub for customization
3. **Keep config minimal**: Only specify directories you need
4. **Use ~ notation**: Makes config portable across systems
5. **Comment your config**: Document why directories are included