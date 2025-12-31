# Execman: Standalone Executable Manager

**Goal**: As a user, I can install, update, and manage standalone executables
from GitHub releases, with origin tracking for secure updates.

## Overview

Execman is a command-line tool that manages standalone executables installed
from GitHub releases. It maintains a registry of installed executables,
tracking their origin, version, and location, enabling secure updates without
relying on self-reported metadata from the executables themselves.

### Relationship to Pathman

| Tool | Responsibility |
|------|----------------|
| **pathman** | Manages `$PATH` - which folders are on the path |
| **execman** | Manages executables - installs/updates binaries in a folder |

These tools are complementary and can work together:

```bash
# Pathman ensures ~/.local/bin is on PATH
pathman add ~/.local/bin

# Execman installs applications there
execman install github.com/owner/repo --into ~/.local/bin
```

## Part 1: Registry and Configuration

### Registry File

Location: `~/.config/execman/registry.json`

```json
{
  "schema_version": 1,
  "executables": {
    "nutmeg-run": {
      "source": "https://github.com/sfkleach/nutmeg-run",
      "version": "v1.2.3",
      "installed_at": "2025-12-28T10:30:00Z",
      "path": "/home/user/.local/bin/nutmeg-run",
      "platform": "linux/amd64",
      "checksum": "sha256:abc123def456..."
    }
  }
}
```

### Required Fields per Executable

| Field | Type | Description |
|-------|------|-------------|
| `source` | string | GitHub repository HTTPS URL |
| `version` | string | Installed version (semver with v prefix) |
| `installed_at` | string | ISO 8601 timestamp of installation |
| `path` | string | Absolute path to installed executable |
| `platform` | string | OS/arch (e.g., "linux/amd64") |
| `checksum` | string | SHA256 checksum of installed binary |

### Configuration

Location: `~/.config/execman/config.json` (optional)

```json
{
  "default_install_dir": "/home/user/.local/bin",
  "include_prereleases": false
}
```

If no configuration exists, defaults are:
- `default_install_dir`: `~/.local/bin`
- `include_prereleases`: `false`

## Part 2: Install Command

```bash
execman install github.com/owner/repo
execman install github.com/owner/repo@v1.2.3
execman install github.com/owner/repo --into /usr/local/bin
```

### Interactive Flow

1. **Parse source**: Extract owner, repo, and optional version from argument
2. **Fetch release**: Query GitHub API for latest (or specified) release
3. **Confirm installation**: Show version, target directory, ask to proceed
4. **Download**: Download appropriate asset for platform with progress bar
5. **Verify checksum**: Download checksums.txt, verify SHA256
6. **Extract**: Extract binary from tar.gz archive
7. **Install**: Copy to target directory with 0755 permissions
8. **Register**: Add entry to registry with all metadata
9. **Confirm**: Show success message with installed path

### Options

| Option | Short | Description |
|--------|-------|-------------|
| `--into <dir>` | `-d` | Install to specified directory |
| `--yes` | `-y` | Skip confirmation prompts |
| `--include-prereleases` | | Allow installing prerelease versions |

### Asset Naming Convention

Execman expects assets named following common conventions:
- `{name}_{version}_{os}_{arch}.tar.gz`
- `{name}-{version}-{os}-{arch}.tar.gz`
- `{name}_{os}_{arch}.tar.gz`
- `{name}-{os}-{arch}.tar.gz`

Where:
- `os`: linux, darwin, windows
- `arch`: amd64, arm64, 386

### Error Handling

- **No matching asset**: List available assets, suggest correct platform
- **Checksum mismatch**: Abort installation, preserve download for debugging
- **Permission denied**: Explain which permissions are needed
- **Already installed**: Ask if user wants to reinstall/update

## Part 3: List Command

```bash
execman list
execman list --json
```

### Output Format (Text)

```
Managed executables:

  nutmeg-run      v1.2.3    ~/.local/bin/nutmeg-run
                            github.com/sfkleach/nutmeg-run
                            installed 2025-12-28

  pathman         v0.1.0    ~/.local/bin/pathman
                            github.com/sfkleach/pathman
                            installed 2025-12-20

2 executables managed
```

### Output Format (JSON)

```json
{
  "executables": [
    {
      "name": "nutmeg-run",
      "source": "https://github.com/sfkleach/nutmeg-run",
      "version": "v1.2.3",
      "path": "/home/user/.local/bin/nutmeg-run",
      "installed_at": "2025-12-28T10:30:00Z"
    }
  ]
}
```

### Options

| Option | Description |
|--------|-------------|
| `--json` | Output as JSON |

## Part 4: Check Command

```bash
execman check
execman check nutmeg-run
execman check --json
execman check --no-skip
```

### Behavior

- With no arguments: Check all managed executables for updates
- With argument: Check specific executable for updates
- By default: Only shows executables with updates available
- With `--no-skip`: Shows all executables including up-to-date ones

### Output Format (Text)

```
Checking for updates...

  nutmeg-run      v1.2.3 → v1.3.0    update available

1 up to date, 1 update available. Run 'execman update' to install updates.
```

### Output Format (Text with --no-skip)

```
Checking for updates...

  nutmeg-run      v1.2.3 → v1.3.0    update available
  pathman         v0.1.0             up to date

1 up to date, 1 update available. Run 'execman update' to install updates.
```

### Output Format (JSON)

```json
{
  "executables": [
    {
      "name": "nutmeg-run",
      "current_version": "v1.2.3",
      "latest_version": "v1.3.0",
      "update_available": true
    },
    {
      "name": "pathman",
      "current_version": "v0.1.0",
      "latest_version": "v0.1.0",
      "update_available": false
    }
  ],
  "updates_available": 1
}
```

### Options

| Option | Description |
|--------|-------------|
| `--json` | Output as JSON |
| `--no-skip` | Show all executables, including up-to-date ones |
| `--include-prereleases` | Include prerelease versions in check |

## Part 5: Update Command

```bash
execman update nutmeg-run
execman update --all
execman update --all --yes
```

### Interactive Flow

1. **Check for updates**: Query GitHub for latest version
2. **Show comparison**: Display current vs latest version
3. **Confirm update**: Ask if user wants to proceed
4. **Backup prompt**: Ask if user wants to create backup
5. **Download**: Download new version with progress bar
6. **Verify checksum**: Verify SHA256 against checksums.txt
7. **Extract**: Extract binary to temporary location
8. **Permission check**: Verify ability to replace existing executable
9. **Backup**: Create backup if requested (e.g., `nutmeg-run.backup`)
10. **Replace**: Unlink old executable, install new one
11. **Update registry**: Update version, checksum, installed_at
12. **Cleanup prompt**: Ask if user wants to delete download archive
13. **Report**: Show success/failure status

### Options

| Option | Short | Description |
|--------|-------|-------------|
| `--all` | `-a` | Update all managed executables |
| `--yes` | `-y` | Skip all confirmation prompts |
| `--include-prereleases` | | Allow updating to prerelease versions |

### Error Handling

- **No update available**: Inform user, exit cleanly
- **Network error**: Show error, preserve current installation
- **Checksum mismatch**: Abort, preserve download for debugging
- **Permission denied**: Explain issue, suggest remediation
- **Replacement failed**: Rollback if possible, preserve download

## Part 6: Remove and Forget Commands

### Remove Command

```bash
execman remove nutmeg-run
execman remove nutmeg-run --yes
```

#### Interactive Flow

1. **Confirm removal**: Show executable details, ask to proceed
2. **Remove executable**: Delete the file
3. **Update registry**: Remove entry from registry
4. **Report**: Show success message

#### Options

| Option | Short | Description |
|--------|-------|-------------|
| `--yes` | `-y` | Skip confirmation prompt |

### Forget Command

```bash
execman forget nutmeg-run
execman forget nutmeg-run --yes
```

#### Interactive Flow

1. **Confirm forgetting**: Show executable details, ask to proceed
2. **Update registry**: Remove entry from registry
3. **Keep file**: Leave the executable file on disk
4. **Report**: Show success message

#### Options

| Option | Short | Description |
|--------|-------|-------------|
| `--yes` | `-y` | Skip confirmation prompt |

## Part 7: Info Command

```bash
execman info nutmeg-run
execman info nutmeg-run --json
```

### Output Format (Text)

```
nutmeg-run

  Source:       https://github.com/sfkleach/nutmeg-run
  Version:      v1.2.3
  Path:         /home/user/.local/bin/nutmeg-run
  Platform:     linux/amd64
  Installed:    2025-12-28T10:30:00Z
  Checksum:     sha256:abc123def456...

  Latest:       v1.3.0 (update available)
```

### Output Format (JSON)

```json
{
  "name": "nutmeg-run",
  "source": "https://github.com/sfkleach/nutmeg-run",
  "version": "v1.2.3",
  "path": "/home/user/.local/bin/nutmeg-run",
  "platform": "linux/amd64",
  "installed_at": "2025-12-28T10:30:00Z",
  "checksum": "sha256:abc123def456...",
  "latest_version": "v1.3.0",
  "update_available": true
}
```

### Options

| Option | Description |
|--------|-------------|
| `--json` | Output as JSON |
| `--include-prereleases` | Include prereleases when checking latest |

## Part 8: Adopt Command

For executables not installed via execman but that the user wants to manage:

```bash
execman adopt /usr/local/bin/pathman --source github.com/sfkleach/pathman
execman adopt /usr/local/bin/pathman --source github.com/sfkleach/pathman --version v0.1.0
```

### Interactive Flow

1. **Verify executable exists**: Check file exists and is executable
2. **Confirm adoption**: Show details, ask to proceed
3. **Detect version**: If `--version` not provided, try running
   `executable --version` to detect current version
4. **Compute checksum**: Calculate SHA256 of existing executable
5. **Add to registry**: Create registry entry with all metadata
6. **Report**: Show success message

### Options

| Option | Description |
|--------|-------------|
| `--source <url>` | GitHub repository URL (required) |
| `--version <ver>` | Current version (optional, auto-detected if possible) |
| `--yes` | Skip confirmation prompt |

### Version Detection

If `--version` is not provided, execman attempts to detect it by running:

```bash
/path/to/executable --version
```

And parsing common output formats:
- `appname version v1.2.3`
- `appname v1.2.3`
- `v1.2.3`
- `1.2.3`

If detection fails, prompt user to provide version manually.

## Part 9: Symlink Handling

When execman encounters a symlink (during update, remove, or adopt):

### Interactive Mode

```
Note: /usr/local/bin/myapp is a symlink to /opt/myapp/v1.2.3/myapp

How would you like to proceed?
  [1] Replace the symlink target (/opt/myapp/v1.2.3/myapp)
  [2] Replace the symlink itself (/usr/local/bin/myapp)
  [3] Cancel

Choice [1/2/3]:
```

### Non-Interactive Mode (--yes)

```
Error: /usr/local/bin/myapp is a symlink to /opt/myapp/v1.2.3/myapp
       Cannot proceed in non-interactive mode.
       Run without --yes to choose how to handle symlinks.
```

Exit with non-zero status code.

## Part 10: Version Command

```bash
execman version
execman version --json
```

### Output Format (Text)

```
execman version v0.1.0
```

### Output Format (JSON)

```json
{
  "version": "v0.1.0",
  "source": "https://github.com/sfkleach/execman"
}
```

### Options

| Option | Description |
|--------|-------------|
| `--json` | Output as JSON |

## Security Considerations

### Origin Trust Model

1. **Origin recorded at install time**: The `source` field is set by execman
   during installation, not self-reported by the executable
2. **Immutable after installation**: An executable cannot change its recorded
   source
3. **Checksum verification**: All downloads are verified against checksums.txt
4. **Checksum stored**: The installed binary's checksum is recorded, allowing
   integrity verification

### Adopt Command Security

When using `execman adopt`, the user explicitly provides the source URL. This
is a trust decision made by the user - they are asserting that the executable
at the given path came from the given source.

### Update Security

Updates only fetch from the recorded source URL. A compromised executable
cannot redirect updates to a malicious repository.

## Implementation Notes

### Shared Code with Pathman

The following components can be adapted from pathman:
- GitHub API client (pkg/github)
- TUI patterns using Bubbletea
- Progress bar rendering
- Checksum verification
- Archive extraction (tar.gz)
- Executable replacement logic

### Platform Support

- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64) - with .exe handling and .zip archives

### Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - TUI styling
- `golang.org/x/mod/semver` - Version comparison

### Build Configuration

Use GoReleaser for releases, baking in:
- Version via ldflags
- Source URL via ldflags

## Future Considerations

### Additional Git Hosts

Future versions could support:
- GitLab (`gitlab.com/owner/repo`)
- Bitbucket (`bitbucket.org/owner/repo`)
- Gitea instances (`gitea.example.com/owner/repo`)

### Signature Verification

Future versions could verify GPG signatures on releases for additional
security.

### Asset Naming Configuration

For projects that don't follow standard naming conventions, allow
configuration:

```json
{
  "executables": {
    "unusual-app": {
      "asset_pattern": "unusual-app-{version}.{os}.{arch}.tar.gz"
    }
  }
}
```

### Shell Completions

Provide shell completion scripts for bash, zsh, fish, and PowerShell.
