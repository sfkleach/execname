package check

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sfkleach/execman/pkg/archive"
	"github.com/sfkleach/execman/pkg/config"
	"github.com/sfkleach/execman/pkg/github"
	"github.com/sfkleach/execman/pkg/registry"
	"github.com/spf13/cobra"
)

// CheckOutput represents the JSON output format for the check command.
type CheckOutput struct {
	Executables      []ExecutableStatus `json:"executables"`
	UpdatesAvailable int                `json:"updates_available"`
	Missing          int                `json:"missing"`
	Modified         int                `json:"modified"`
}

// ExecutableStatus represents the update status of an executable.
type ExecutableStatus struct {
	Name            string `json:"name"`
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version,omitempty"`
	UpdateAvailable bool   `json:"update_available"`
	Status          string `json:"status"` // "ok", "missing", "modified"
}

// NewCheckCommand creates the check command.
func NewCheckCommand() *cobra.Command {
	var jsonOutput bool
	var includePrereleases bool
	var noSkip bool
	var verify bool

	cmd := &cobra.Command{
		Use:   "check [executable]",
		Short: "Check for available updates and integrity",
		Long:  "Check if updates are available for managed executables and verify file integrity.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			return runCheck(name, jsonOutput, includePrereleases, noSkip, verify)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&includePrereleases, "include-prereleases", false, "Include prerelease versions in check")
	cmd.Flags().BoolVar(&noSkip, "no-skip", false, "Show all executables, including up-to-date ones")
	cmd.Flags().BoolVar(&verify, "verify", false, "Verify checksums of installed executables")

	return cmd
}

func runCheck(name string, jsonOutput, includePrereleases, noSkip, verify bool) error {
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
	upToDateCount := 0
	missingCount := 0
	modifiedCount := 0

	for _, n := range names {
		exec, ok := reg.Get(n)
		if !ok {
			continue
		}

		// Check file integrity first.
		fileStatus := "ok"
		if _, err := os.Stat(exec.Path); os.IsNotExist(err) {
			fileStatus = "missing"
			missingCount++
		} else if verify {
			// Verify checksum if requested.
			actualChecksum, err := archive.CalculateChecksum(exec.Path)
			if err == nil && actualChecksum != exec.Checksum {
				fileStatus = "modified"
				modifiedCount++
			}
		}

		// If file is missing or modified, report that and skip update check.
		if fileStatus != "ok" {
			status := ExecutableStatus{
				Name:           n,
				CurrentVersion: exec.Version,
				Status:         fileStatus,
			}
			statuses = append(statuses, status)

			if !jsonOutput {
				if fileStatus == "missing" {
					fmt.Printf("  %-15s %-9s          MISSING\n", n, exec.Version)
				} else {
					fmt.Printf("  %-15s %-9s          MODIFIED\n", n, exec.Version)
				}
			}
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
		} else {
			upToDateCount++
		}

		status := ExecutableStatus{
			Name:            n,
			CurrentVersion:  exec.Version,
			LatestVersion:   latestVersion,
			UpdateAvailable: updateAvailable,
			Status:          "ok",
		}
		statuses = append(statuses, status)

		if !jsonOutput {
			if updateAvailable {
				fmt.Printf("  %-15s %s → %-9s update available\n", n, exec.Version, latestVersion)
			} else if noSkip {
				if verify {
					fmt.Printf("  %-15s %-9s          up to date (verified)\n", n, exec.Version)
				} else {
					fmt.Printf("  %-15s %-9s          up to date\n", n, exec.Version)
				}
			}
		}
	}

	if jsonOutput {
		output := CheckOutput{
			Executables:      statuses,
			UpdatesAvailable: updatesAvailable,
			Missing:          missingCount,
			Modified:         modifiedCount,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	// Text summary.
	fmt.Println()

	// Build summary parts.
	var parts []string
	if missingCount > 0 {
		parts = append(parts, fmt.Sprintf("%d missing", missingCount))
	}
	if modifiedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", modifiedCount))
	}
	parts = append(parts, fmt.Sprintf("%d up to date", upToDateCount))
	if updatesAvailable == 1 {
		parts = append(parts, "1 update available")
	} else {
		parts = append(parts, fmt.Sprintf("%d updates available", updatesAvailable))
	}

	fmt.Println(joinParts(parts) + ".")

	if missingCount > 0 || modifiedCount > 0 {
		fmt.Println("Run 'execman update <name>' to reinstall missing or modified executables.")
	} else if updatesAvailable > 0 {
		fmt.Println("Run 'execman update' to install updates.")
	}

	return nil
}

// joinParts joins parts with commas.
func joinParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(parts[0])
	for i := 1; i < len(parts); i++ {
		sb.WriteString(", ")
		sb.WriteString(parts[i])
	}
	return sb.String()
}
