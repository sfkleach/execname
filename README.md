# Execman

Execman is a command-line tool for managing standalone executables from GitHub releases.

## Features

- **Install** executables directly from GitHub releases
- **Track** installed executables with version and origin information
- **List** all managed executables with details
- **Check** for available updates across all executables
- **Update** executables individually or all at once
- **Remove** executables and delete files
- **Forget** executables while keeping files on disk
- **Registry** maintains metadata for secure updates
- **Cross-platform** support for Linux, macOS, and Windows

## Building

```bash
go build -o bin/execman ./cmd/execman
```

Or using the Justfile:

```bash
just build
```

## Usage

### Install an executable

```bash
# Install latest version
execman install github.com/owner/repo

# Install specific version
execman install github.com/owner/repo@v1.2.3

# Install to custom directory
execman install github.com/owner/repo --into /usr/local/bin

# Skip confirmation prompts
execman install github.com/owner/repo --yes
```

### List managed executables

```bash
# Show all managed executables
execman list
execman ls

# Show specific executable
execman list myapp

# Show detailed information
execman list --long
execman ls -l

# Show specific executable with details
execman list myapp --long

# Output as JSON
execman list --json
```

### Check for updates

```bash
# Check all executables for updates
execman check

# Check specific executable
execman check myapp

# Show all executables including up-to-date ones
execman check --no-skip

# Output as JSON
execman check --json
```

### Update executables

```bash
# Update a specific executable
execman update myapp

# Update all executables
execman update --all

# Skip confirmation prompts
execman update --all --yes
```

### Remove an executable

```bash
# Remove executable and delete file
execman remove myapp

# Skip confirmation prompt
execman remove myapp --yes
```

### Forget an executable

```bash
# Stop tracking but keep the file
execman forget myapp

# Skip confirmation prompt
execman forget myapp --yes
```

### Show version

```bash
# Show version using flag
execman --version

# Show version using subcommand
execman version
```

### Get help

```bash
# Show all commands
execman

# Get help for any command
execman [command] --help
```

## Commands

- `version` - Print the version number of execman
- `install` - Install an executable from GitHub releases
- `list` (alias: `ls`) - List managed executables with optional filtering and detailed view
- `check` - Check for available updates
- `update` - Update executables to latest versions
- `remove` - Remove an executable and delete the file
- `forget` - Stop tracking an executable but keep the file
- `adopt` - Adopt an existing executable (TBD)

## Configuration

### Registry

Location: `~/.config/execman/registry.json`

Tracks all installed executables with version, source, checksum, and path information.

### Config (Optional)

Location: `~/.config/execman/config.json`

```json
{
  "default_install_dir": "/home/user/.local/bin",
  "include_prereleases": false
}
```

Defaults:
- `default_install_dir`: `~/.local/bin`
- `include_prereleases`: `false`

## Example Workflow

```bash
# Install some executables
execman install github.com/sfkleach/pathman --yes
execman install github.com/sfkleach/nutmeg-run --yes

# List what's installed
execman list

# Check for updates
execman check

# Update all executables
execman update --all --yes

# Remove an executable
execman remove nutmeg-run

# Stop tracking but keep the file
execman forget pathman
```

## Project Structure

```
execman/
├── cmd/
│   └── execman/
│       └── main.go          # Main entry point
├── pkg/
│   ├── archive/             # Archive extraction and checksums
│   ├── check/               # Check command implementation
│   ├── config/              # Configuration management
│   ├── forget/              # Forget command implementation
│   ├── github/              # GitHub API integration
│   ├── install/             # Install command implementation
│   ├── list/                # List command implementation
│   ├── registry/            # Registry management
│   ├── remove/              # Remove command implementation
│   ├── update/              # Update command implementation
│   └── version/             # Version information
├── go.mod
└── README.md
```
