package forget

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sfkleach/execman/pkg/registry"
	"github.com/spf13/cobra"
)

// Options for the forget command.
type Options struct {
	Name string
	Yes  bool
}

// NewForgetCommand creates the forget command.
func NewForgetCommand() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "forget <executable>",
		Short: "Stop tracking an executable without deleting it",
		Long:  "Remove an executable from management but keep the file on disk.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := Options{
				Name: args[0],
				Yes:  yes,
			}
			return Forget(opts)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// Forget removes an executable from management without deleting the file.
func Forget(opts Options) error {
	// Load registry.
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Check if executable exists.
	exec, ok := reg.Get(opts.Name)
	if !ok {
		return fmt.Errorf("executable %q is not managed by execman", opts.Name)
	}

	// Show details and confirm.
	if !opts.Yes {
		fmt.Printf("Forget %s?\n\n", opts.Name)
		fmt.Printf("  Source:       %s\n", exec.Source)
		fmt.Printf("  Version:      %s\n", exec.Version)
		fmt.Printf("  Path:         %s\n", exec.Path)
		fmt.Printf("  Installed:    %s\n", exec.InstalledAt.Format("2006-01-02"))
		fmt.Println()
		fmt.Print("This will stop tracking the executable but keep the file. Continue? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Remove from registry.
	reg.Remove(opts.Name)
	if err := reg.Save(); err != nil {
		return fmt.Errorf("failed to update registry: %w", err)
	}

	// Report success.
	fmt.Printf("\n%s forgotten (file kept at %s)\n", opts.Name, exec.Path)

	return nil
}
