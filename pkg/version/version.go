package version

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version information.
// These variables are set via ldflags during build.
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

// VersionOutput represents the JSON output format for the version command.
type VersionOutput struct {
	Version string `json:"version"`
	Source  string `json:"source"`
}

// NewVersionCommand creates the version command.
func NewVersionCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of execman",
		Long:  "Display the version of execman and its source repository.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ShowVersion(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

// ShowVersion displays version information.
func ShowVersion(jsonOutput bool) error {
	if jsonOutput {
		output := VersionOutput{
			Version: GetVersion(),
			Source:  "https://github.com/sfkleach/execman",
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	fmt.Printf("execman version %s\n", GetVersion())
	return nil
}

// GetVersion returns the current version.
func GetVersion() string {
	if Version == "dev" {
		return Version
	}
	// Check if version already has 'v' prefix (from git describe).
	if len(Version) > 0 && Version[0] == 'v' {
		return Version
	}
	return "v" + Version
}

// GetFullVersion returns version with build info.
func GetFullVersion() string {
	return Version + " (build: " + BuildDate + ", commit: " + GitCommit + ")"
}
