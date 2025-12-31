package list

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sfkleach/execman/pkg/registry"
	"github.com/spf13/cobra"
)

// ListOutput represents the JSON output format for the list command.
type ListOutput struct {
	Executables []ExecutableInfo `json:"executables"`
}

// ExecutableInfo represents information about a single executable.
type ExecutableInfo struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Version     string `json:"version"`
	Path        string `json:"path"`
	Platform    string `json:"platform,omitempty"`
	Checksum    string `json:"checksum,omitempty"`
	InstalledAt string `json:"installed_at"`
}

// NewListCommand creates the list command.
func NewListCommand() *cobra.Command {
	var jsonOutput bool
	var longFormat bool

	cmd := &cobra.Command{
		Use:     "list [executable]",
		Aliases: []string{"ls"},
		Short:   "List managed executables",
		Long:    "Display executables managed by execman. Optionally filter by name.",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var filterName string
			if len(args) > 0 {
				filterName = args[0]
			}
			return runList(filterName, jsonOutput, longFormat)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVarP(&longFormat, "long", "l", false, "Show detailed information")

	return cmd
}

func runList(filterName string, jsonOutput bool, longFormat bool) error {
	// Load registry.
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Get all executable names.
	names := reg.List()

	// Filter by name if specified.
	if filterName != "" {
		found := false
		for _, name := range names {
			if name == filterName {
				names = []string{filterName}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("executable %q is not managed by execman", filterName)
		}
	}

	if len(names) == 0 {
		if jsonOutput {
			output := ListOutput{Executables: []ExecutableInfo{}}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(output)
		}
		fmt.Println("No managed executables.")
		return nil
	}

	// Sort by name.
	sort.Strings(names)

	if jsonOutput {
		return outputJSON(reg, names, longFormat)
	}

	return outputText(reg, names, longFormat)
}

func outputJSON(reg *registry.Registry, names []string, longFormat bool) error {
	// Convert to output format.
	executables := make([]ExecutableInfo, 0, len(names))
	for _, name := range names {
		exec, ok := reg.Get(name)
		if !ok {
			continue
		}

		info := ExecutableInfo{
			Name:        name,
			Source:      exec.Source,
			Version:     exec.Version,
			Path:        exec.Path,
			InstalledAt: exec.InstalledAt.Format(time.RFC3339),
		}

		if longFormat {
			info.Platform = exec.Platform
			info.Checksum = exec.Checksum
		}

		executables = append(executables, info)
	}

	output := ListOutput{Executables: executables}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputText(reg *registry.Registry, names []string, longFormat bool) error {
	homeDir, _ := os.UserHomeDir()

	// If showing a single executable in long format, use detailed view.
	if len(names) == 1 && longFormat {
		name := names[0]
		exec, ok := reg.Get(name)
		if !ok {
			return fmt.Errorf("executable %q not found", name)
		}

		fmt.Printf("%s\n\n", name)
		fmt.Printf("  Source:       %s\n", exec.Source)
		fmt.Printf("  Version:      %s\n", exec.Version)
		fmt.Printf("  Path:         %s\n", exec.Path)
		fmt.Printf("  Platform:     %s\n", exec.Platform)
		fmt.Printf("  Installed:    %s\n", exec.InstalledAt.Format(time.RFC3339))
		fmt.Printf("  Checksum:     %s\n", exec.Checksum)

		return nil
	}

	// Multiple executables or short format.
	if len(names) > 1 || !longFormat {
		fmt.Println("Managed executables:")
		fmt.Println()
	}

	for _, name := range names {
		exec, ok := reg.Get(name)
		if !ok {
			continue
		}

		// Display path with ~ for home directory.
		displayPath := exec.Path
		if homeDir != "" && strings.HasPrefix(exec.Path, homeDir) {
			displayPath = "~" + strings.TrimPrefix(exec.Path, homeDir)
		}

		// Extract repo path from source URL.
		source := strings.TrimPrefix(exec.Source, "https://")

		// Format installed_at timestamp.
		installedDate := exec.InstalledAt.Format("2006-01-02")

		// Get just the executable name from the path.
		execName := filepath.Base(exec.Path)

		// Print formatted output.
		fmt.Printf("  %-15s %-9s %s\n", execName, exec.Version, displayPath)
		fmt.Printf("  %-15s %-9s %s\n", "", "", source)

		if longFormat {
			fmt.Printf("  %-15s %-9s platform: %s\n", "", "", exec.Platform)
			fmt.Printf("  %-15s %-9s checksum: %s\n", "", "", exec.Checksum)
		}

		fmt.Printf("  %-15s %-9s installed %s\n", "", "", installedDate)
		fmt.Println()
	}

	if len(names) > 1 {
		count := len(names)
		if count == 1 {
			fmt.Println("1 executable managed")
		} else {
			fmt.Printf("%d executables managed\n", count)
		}
	}

	return nil
}
