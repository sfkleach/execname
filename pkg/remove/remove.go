package remove

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sfkleach/execman/pkg/registry"
	"github.com/spf13/cobra"
)

// Options for the remove command.
type Options struct {
	Name string
	Yes  bool
}

// NewRemoveCommand creates the remove command.
func NewRemoveCommand() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "remove <executable>",
		Short: "Remove a managed executable",
		Long:  "Remove an executable from management and delete the file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := Options{
				Name: args[0],
				Yes:  yes,
			}
			return Remove(opts)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// Remove removes an executable from management.
func Remove(opts Options) error {
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
		fmt.Printf("Remove %s?\n\n", opts.Name)
		fmt.Printf("  Source:       %s\n", exec.Source)
		fmt.Printf("  Version:      %s\n", exec.Version)
		fmt.Printf("  Path:         %s\n", exec.Path)
		fmt.Printf("  Installed:    %s\n", exec.InstalledAt.Format("2006-01-02"))
		fmt.Println()
		fmt.Print("This will delete the executable file and remove it from management. Continue? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println("Removal cancelled.")
			return nil
		}
	}

	// Remove file.
	if err := os.Remove(exec.Path); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Warning: executable file not found at %s\n", exec.Path)
		} else {
			return fmt.Errorf("failed to remove executable: %w", err)
		}
	}

	// Remove from registry.
	reg.Remove(opts.Name)
	if err := reg.Save(); err != nil {
		return fmt.Errorf("failed to update registry: %w", err)
	}

	// Report success.
	fmt.Printf("\n%s removed successfully\n", opts.Name)

	return nil
}
