// Package init provides functionality for initializing execman.
package init

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sfkleach/execman/pkg/config"
	"github.com/sfkleach/execman/pkg/install"
	"github.com/sfkleach/execman/pkg/registry"
	"github.com/spf13/cobra"
)

// Options represents the init command options.
type Options struct {
	Folder string
}

// NewInitCommand creates the init command.
func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <folder>",
		Short: "Initialize execman configuration and install execman itself",
		Long: `Initialize execman by creating configuration and registry files,
then install execman itself from GitHub with proper metadata.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := Options{
				Folder: args[0],
			}
			return Run(opts)
		},
	}

	return cmd
}

// Run executes the init command.
func Run(opts Options) error {
	// Get absolute path.
	absFolder, err := filepath.Abs(opts.Folder)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("Initializing execman in %s...\n\n", absFolder)

	// Create config.
	fmt.Println("Creating configuration...")
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.DefaultInstallDir = absFolder

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Println("✓ Configuration created")

	// Create registry.
	fmt.Println("Creating registry...")
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if err := reg.Save(); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}
	fmt.Println("✓ Registry created")

	// Install execman itself.
	fmt.Println("\nInstalling execman from GitHub...")
	installOpts := install.Options{
		Source:             "github.com/sfkleach/execman",
		Into:               absFolder,
		Yes:                true, // Skip prompts during init
		IncludePrereleases: false,
	}

	if err := install.Run(installOpts); err != nil {
		return fmt.Errorf("failed to install execman: %w", err)
	}

	fmt.Println("\n✓ Initialization complete!")
	fmt.Printf("\nExecman is installed at %s/execman\n", absFolder)

	// Remind user to delete bootstrap binary if running from elsewhere.
	currentExe, err := os.Executable()
	if err == nil {
		currentExeAbs, err := filepath.Abs(currentExe)
		if err == nil {
			installedExe := filepath.Join(absFolder, "execman")
			if currentExeAbs != installedExe {
				fmt.Printf("\nNote: You ran execman from %s\n", currentExeAbs)
				fmt.Println("You can now delete this bootstrap binary if you no longer need it.")
			}
		}
	}

	// Check if the folder is on PATH.
	pathEnv := os.Getenv("PATH")
	onPath := false
	for _, p := range filepath.SplitList(pathEnv) {
		if absPath, err := filepath.Abs(p); err == nil && absPath == absFolder {
			onPath = true
			break
		}
	}

	if !onPath {
		fmt.Printf("\nNote: %s is not on your $PATH.\n", absFolder)
		fmt.Println("You can run execman using the full path:")
		fmt.Printf("  %s/execman\n", absFolder)

		// Check if pathman is available.
		_, pathmanErr := os.Stat(filepath.Join(absFolder, "pathman"))
		if pathmanErr == nil {
			fmt.Println("\nOr use pathman to add it to your PATH:")
			fmt.Printf("  %s/pathman add %s\n", absFolder, absFolder)
		} else {
			fmt.Println("\nOr add it to your PATH by adding this line to your ~/.bashrc or ~/.profile:")
			fmt.Printf("  export PATH=\"%s:$PATH\"\n", absFolder)
		}
	}

	return nil
}
