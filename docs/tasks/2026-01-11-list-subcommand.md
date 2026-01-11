# List subcommand

Display executables managed by execman. 
 - Optionally filters by name (exact match).
 - Optionally add complete details rather than just a summary. 
 - Optionally output in json, full details always given.

Usage:
  execman list [executable-name] [flags]

Aliases:
  list, ls

Flags:
  -h, --help              help for list
      --json              Output as JSON
  -l, --long              Show detailed information

Global Flags:
      --version   Print version information

## Output Format Specification

The general concept is that the format is reminiscent of the /usr/bin/ls command. This format specification should be replicated across execman, scriptman, and pathman for consistency.

### Compact Format (default)

The compact format lists only the executable names, one per line, with no headers or additional information.

Example output:
```
execman
pathman
```

**Characteristics:**
- No headers or footers
- No extra information (repo, version, etc.)
- Just executable names, one per line
- Suitable for piping to other commands
- Sorted alphabetically

### Long Format (--long, -l)

The long format shows all available information in a human-friendly labeled format. Each executable is separated by a blank line.

Example output:
```
Name:           execman
Source:         https://github.com/sfkleach/execman
Version:        v0.1.13
Path:           /home/sfkleach/.local/libexec/execman
Installed at:   2026-01-10T13:07:36Z

Name:           pathman
Source:         https://github.com/sfkleach/pathman
Version:        v0.1.0
Path:           /home/sfkleach/.local/libexec/pathman
Installed at:   2025-12-31T13:37:47Z
```

**Characteristics:**
- Left-aligned labels in a fixed width with colon separator
 for readability
- All available metadata displayed
- Blank line between entries

### JSON Format (--json)

The JSON format outputs the data structure for programmatic consumption. The
data is partitioned by type and within that sorted alphabetically.

Example output:
```json
{
  "executables": [
    {
      "name": "execman",
      "source": "https://github.com/sfkleach/execman",
      "version": "v0.1.13",
      "path": "/home/sfkleach/.local/libexec/execman",
      "installed_at": "2026-01-10T13:07:36Z"
    },
    {
      "name": "pathman",
      "source": "https://github.com/sfkleach/pathman",
      "version": "v0.1.0",
      "path": "/home/sfkleach/.local/libexec/pathman",
      "installed_at": "2025-12-31T13:37:47Z"
    }
  ]
}
```

**Characteristics:**
- Standard JSON formatting
- Complete data structure
- Suitable for parsing by other tools

### Filtering Behavior

When an executable name is provided as an argument, the output is filtered to show only exact matches i.e. not partial matches such as a prefix. The format remains the same (compact, long, or JSON) but only includes matching executables. 
