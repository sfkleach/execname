package list

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
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
		if slices.Contains(names, filterName) {
			names = []string{filterName}
			found = true
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
		return outputJSON(reg, names)
	}

	return outputText(reg, names, longFormat)
}

func outputJSON(reg *registry.Registry, names []string) error {
	// Convert to output format. JSON always includes full details per spec.
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

		executables = append(executables, info)
	}

	output := ListOutput{Executables: executables}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputText(reg *registry.Registry, names []string, longFormat bool) error {
	if longFormat {
		return outputLongFormat(reg, names)
	}
	return outputCompactFormat(names)
}

// outputCompactFormat prints just the executable names, one per line.
// No headers, no extra info - suitable for piping to other commands.
func outputCompactFormat(names []string) error {
	for _, name := range names {
		fmt.Println(name)
	}
	return nil
}

// outputLongFormat prints detailed information with labeled key-value pairs.
// Each executable is separated by a blank line.
func outputLongFormat(reg *registry.Registry, names []string) error {
	const labelWidth = 16 // Width for left-aligned labels including colon.

	for i, name := range names {
		exec, ok := reg.Get(name)
		if !ok {
			continue
		}

		// Add blank line between entries (but not before first).
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("%-*s%s\n", labelWidth, "Name:", name)
		fmt.Printf("%-*s%s\n", labelWidth, "Source:", exec.Source)
		fmt.Printf("%-*s%s\n", labelWidth, "Version:", exec.Version)
		fmt.Printf("%-*s%s\n", labelWidth, "Path:", exec.Path)
		fmt.Printf("%-*s%s\n", labelWidth, "Installed at:", exec.InstalledAt.Format(time.RFC3339))
	}

	return nil
}
