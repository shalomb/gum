# Installation Guide

## Overview

Gum provides multiple installation methods designed to be user-friendly and follow XDG conventions. The installation process prioritizes user directories over system-wide installation to avoid requiring sudo privileges.

## Quick Installation

### Prerequisites

**Required Dependencies**
- **Go 1.21+**: For building from source
- **SQLite3**: For database operations (usually pre-installed)

**Install SQLite3**
```bash
# Ubuntu/Debian
sudo apt-get install sqlite3

# macOS
brew install sqlite3

# CentOS/RHEL
sudo yum install sqlite3

# Arch Linux
sudo pacman -S sqlite3
```

### Standard Installation

```bash
# Clone repository
git clone https://github.com/shalomb/gum.git
cd gum

# Install (prefers user location)
make install
```

This will:
1. Build gum from source
2. Install to `~/.local/bin/gum` (if directory exists)
3. Fall back to `/usr/local/bin/gum` (may require sudo)
4. Set proper executable permissions
5. Provide PATH configuration guidance

## Installation Methods

### User Installation (Recommended)

**Explicit User Installation**
```bash
make install-user
```

This method:
- Creates `~/.local/bin/` if it doesn't exist
- Installs gum to user directory
- Never requires sudo
- Provides PATH setup instructions

**Manual User Installation**
```bash
make build
mkdir -p ~/.local/bin
cp gum ~/.local/bin/
chmod +x ~/.local/bin/gum
```

### System Installation

**Smart System Installation**
```bash
make install
```

This method:
- Tries user location first (`~/.local/bin`)
- Falls back to system location (`/usr/local/bin`)
- Uses sudo only if necessary
- Provides clear feedback about installation location

**Manual System Installation**
```bash
make build
sudo cp gum /usr/local/bin/
sudo chmod +x /usr/local/bin/gum
```

## PATH Configuration

### Check Current PATH

```bash
# Check if ~/.local/bin is in PATH
echo $PATH | grep -q ".local/bin" && echo "✓ ~/.local/bin in PATH" || echo "✗ ~/.local/bin not in PATH"

# Check where gum is installed
which gum
```

### Add to PATH

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

**System-wide** (`/etc/environment`)
```bash
PATH="/usr/local/bin:/usr/bin:/bin:/usr/local/sbin:/usr/sbin:/sbin"
```

### Reload Shell Configuration

```bash
# Reload current shell
source ~/.bashrc  # or ~/.zshrc

# Or start new shell session
exec $SHELL
```

## Verification

### Test Installation

```bash
# Check version
gum version

# Check help
gum help

# Test project discovery
gum projects --help
```

### Expected Output

```bash
$ gum version
version called

$ gum help
Gum - Smart project discovery and management

Usage:
  gum [command]

Available Commands:
  clone       Intelligently clone GitHub repositories
  completion  Generate the autocompletion script for the specified shell
  dirs        List frequently accessed directories
  dirs-cache  Manage project directories cache
  github      Discover and manage GitHub repositories
  help        Help about any command
  projects    List Git projects from configured directories
  update      Update the database
  version     A brief description of your command
```

## Troubleshooting

### Installation Issues

**Command not found after installation**
```bash
# Check installation location
ls -la ~/.local/bin/gum
ls -la /usr/local/bin/gum

# Check PATH
echo $PATH

# Add to PATH if missing
export PATH="$HOME/.local/bin:$PATH"
```

**Permission denied**
```bash
# Check permissions
ls -la ~/.local/bin/gum

# Fix permissions
chmod +x ~/.local/bin/gum
```

**Build failures**
```bash
# Check Go version
go version

# Check Go modules
go mod tidy

# Clean and rebuild
make clean  # if available
make build
```

### Runtime Issues

**SQLite not found**
```bash
# Check SQLite installation
sqlite3 --version

# Install if missing
sudo apt-get install sqlite3  # Ubuntu/Debian
brew install sqlite3          # macOS
```

**Database creation fails**
```bash
# Check directory permissions
ls -la ~/.cache/
ls -la ~/.local/share/

# Create directories if needed
mkdir -p ~/.cache/gum
mkdir -p ~/.local/share/gum
```

**Configuration issues**
```bash
# Check config directory
ls -la ~/.config/gum/

# Create if missing
mkdir -p ~/.config/gum
```

## Advanced Installation

### Custom Build

**Build with custom flags**
```bash
# Build with race detection
go build -race -o gum

# Build with optimizations
go build -ldflags="-s -w" -o gum

# Build for different architecture
GOOS=linux GOARCH=amd64 go build -o gum-linux-amd64
```

### Development Installation

**Development setup**
```bash
# Clone repository
git clone https://github.com/shalomb/gum.git
cd gum

# Install dependencies
go mod download

# Run tests
make test

# Build and install
make install-user
```

### Container Installation

**Docker**
```dockerfile
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache sqlite3-dev gcc musl-dev
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o gum

FROM alpine:latest
RUN apk add --no-cache sqlite3
COPY --from=builder /app/gum /usr/local/bin/gum
ENTRYPOINT ["gum"]
```

**Podman**
```bash
# Build container
podman build -t gum .

# Run container
podman run -v ~/.config:/root/.config -v ~/.cache:/root/.cache gum projects
```

## Uninstallation

### Remove Binary

```bash
# Remove from user location
rm ~/.local/bin/gum

# Remove from system location
sudo rm /usr/local/bin/gum
```

### Clean Configuration

```bash
# Remove configuration
rm -rf ~/.config/gum/

# Remove cache
rm -rf ~/.cache/gum/

# Remove data
rm -rf ~/.local/share/gum/

# Remove state
rm -rf ~/.local/state/gum/
```

### Complete Cleanup

```bash
# Remove everything
rm ~/.local/bin/gum
rm -rf ~/.config/gum/
rm -rf ~/.cache/gum/
rm -rf ~/.local/share/gum/
rm -rf ~/.local/state/gum/
```

## Best Practices

### Installation Location

- **Prefer user installation**: Avoids sudo requirements
- **Use `~/.local/bin`**: Follows XDG conventions
- **Check PATH**: Ensure installation directory is in PATH

### Environment Setup

- **XDG compliance**: Respects XDG environment variables
- **Shell integration**: Add to shell profile for persistence
- **Permission management**: Use appropriate file permissions

### Maintenance

- **Regular updates**: Pull latest changes and rebuild
- **Clean installation**: Remove old versions before upgrading
- **Backup configuration**: Save custom configurations before updates