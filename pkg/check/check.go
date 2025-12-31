package check

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/sfkleach/execman/pkg/config"
	"github.com/sfkleach/execman/pkg/github"
	"github.com/sfkleach/execman/pkg/registry"
	"github.com/spf13/cobra"
)

// CheckOutput represents the JSON output format for the check command.
type CheckOutput struct {
	Executables      []ExecutableStatus `json:"executables"`
	UpdatesAvailable int                `json:"updates_available"`
}

// ExecutableStatus represents the update status of an executable.
type ExecutableStatus struct {
	Name            string `json:"name"`
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
}

// NewCheckCommand creates the check command.
func NewCheckCommand() *cobra.Command {
	var jsonOutput bool
	var includePrereleases bool

	cmd := &cobra.Command{
		Use:   "check [executable]",
		Short: "Check for available updates",
		Long:  "Check if updates are available for managed executables.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			return runCheck(name, jsonOutput, includePrereleases)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&includePrereleases, "include-prereleases", false, "Include prerelease versions in check")

	return cmd
}

func runCheck(name string, jsonOutput, includePrereleases bool) error {
	// Load registry.
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Load config for prerelease default.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !includePrereleases {
		includePrereleases = cfg.IncludePrereleases
	}

	// Get executables to check.
	var names []string
	if name != "" {
		// Check specific executable.
		if _, ok := reg.Get(name); !ok {
			return fmt.Errorf("executable %q is not managed by execman", name)
		}
		names = []string{name}
	} else {
		// Check all executables.
		names = reg.List()
		if len(names) == 0 {
			if jsonOutput {
				output := CheckOutput{Executables: []ExecutableStatus{}, UpdatesAvailable: 0}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(output)
			}
			fmt.Println("No managed executables.")
			return nil
		}
	}

	// Sort by name for consistent output.
	sort.Strings(names)

	// Check each executable.
	if !jsonOutput {
		fmt.Println("Checking for updates...")
		fmt.Println()
	}

	statuses := make([]ExecutableStatus, 0, len(names))
	updatesAvailable := 0

	for _, n := range names {
		exec, ok := reg.Get(n)
		if !ok {
			continue
		}

		// Parse source to get owner/repo.
		owner, repo, _, err := github.ParseSource(exec.Source)
		if err != nil {
			if !jsonOutput {
				fmt.Printf("  %-15s error: %v\n", n, err)
			}
			continue
		}

		// Fetch latest release.
		release, err := github.GetLatestRelease(owner, repo, includePrereleases)
		if err != nil {
			if !jsonOutput {
				fmt.Printf("  %-15s error: %v\n", n, err)
			}
			continue
		}

		latestVersion := release.TagName
		updateAvailable := exec.Version != latestVersion

		if updateAvailable {
			updatesAvailable++
		}

		status := ExecutableStatus{
			Name:            n,
			CurrentVersion:  exec.Version,
			LatestVersion:   latestVersion,
			UpdateAvailable: updateAvailable,
		}
		statuses = append(statuses, status)

		if !jsonOutput {
			if updateAvailable {
				fmt.Printf("  %-15s %s → %-9s update available\n", n, exec.Version, latestVersion)
			} else {
				fmt.Printf("  %-15s %-9s          up to date\n", n, exec.Version)
			}
		}
	}

	if jsonOutput {
		output := CheckOutput{
			Executables:      statuses,
			UpdatesAvailable: updatesAvailable,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	// Text summary.
	fmt.Println()
	if updatesAvailable == 0 {
		fmt.Println("All executables are up to date.")
	} else if updatesAvailable == 1 {
		fmt.Println("1 update available. Run 'execman update' to install updates.")
	} else {
		fmt.Printf("%d updates available. Run 'execman update' to install updates.\n", updatesAvailable)
	}

	return nil
}
