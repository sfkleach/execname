package update

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sfkleach/execman/pkg/archive"
	"github.com/sfkleach/execman/pkg/config"
	"github.com/sfkleach/execman/pkg/github"
	"github.com/sfkleach/execman/pkg/registry"
	"github.com/sfkleach/execman/pkg/symlink"
	"github.com/spf13/cobra"
)

// Options represents the update command options.
type Options struct {
	Name               string
	All                bool
	Yes                bool
	IncludePrereleases bool
}

// NewUpdateCommand creates the update command.
func NewUpdateCommand() *cobra.Command {
	var all bool
	var yes bool
	var includePrereleases bool

	cmd := &cobra.Command{
		Use:   "update [executable]",
		Short: "Update an executable to the latest version",
		Long:  "Update one or all managed executables to their latest versions.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}

			if name != "" && all {
				return fmt.Errorf("cannot specify both executable name and --all flag")
			}

			if name == "" && !all {
				return fmt.Errorf("must specify either executable name or --all flag")
			}

			opts := Options{
				Name:               name,
				All:                all,
				Yes:                yes,
				IncludePrereleases: includePrereleases,
			}
			return Run(opts)
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Update all managed executables")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip all confirmation prompts")
	cmd.Flags().BoolVar(&includePrereleases, "include-prereleases", false, "Allow updating to prerelease versions")

	return cmd
}

// Run executes the update command.
func Run(opts Options) error {
	// Load registry and config.
	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !opts.IncludePrereleases {
		opts.IncludePrereleases = cfg.IncludePrereleases
	}

	if opts.All {
		return updateAll(reg, opts)
	}

	_, err = updateOne(reg, opts)
	return err
}

func updateAll(reg *registry.Registry, opts Options) error {
	names := reg.List()
	if len(names) == 0 {
		fmt.Println("No managed executables to update.")
		return nil
	}

	updatedCount := 0
	upToDateCount := 0
	failCount := 0

	for _, name := range names {
		fmt.Printf("\nUpdating %s...\n", name)
		opts.Name = name
		updated, err := updateOne(reg, opts)
		if err != nil {
			fmt.Printf("Failed to update %s: %v\n", name, err)
			failCount++
		} else if updated {
			updatedCount++
		} else {
			upToDateCount++
		}
	}

	fmt.Printf("\n%d updated, %d already up to date, %d failed.\n", updatedCount, upToDateCount, failCount)
	return nil
}

func updateOne(reg *registry.Registry, opts Options) (bool, error) {
	// Get current installation.
	exec, ok := reg.Get(opts.Name)
	if !ok {
		return false, fmt.Errorf("executable %q is not managed by execman", opts.Name)
	}

	// Check if executable file exists and if it's a symlink.
	executableMissing := false
	var symlinkInfo *symlink.Info
	var symlinkAction symlink.SymlinkAction
	effectivePath := exec.Path

	if _, err := os.Stat(exec.Path); os.IsNotExist(err) {
		executableMissing = true
	} else {
		// Check for symlink.
		symlinkInfo, err = symlink.Check(exec.Path)
		if err != nil {
			return false, fmt.Errorf("failed to check path: %w", err)
		}

		if symlinkInfo.IsSymlink {
			if opts.Yes {
				// Non-interactive mode with symlink - error out.
				return false, symlink.ErrorNonInteractive(symlinkInfo.Path, symlinkInfo.Target)
			}
			// Interactive mode - ask user.
			symlinkAction = symlink.PromptAction(symlinkInfo.Path, symlinkInfo.Target)
			if symlinkAction == symlink.ActionCancel {
				fmt.Println("Update cancelled.")
				return false, nil
			}
			effectivePath = symlink.ResolveTarget(symlinkInfo, symlinkAction)
		}
	}

	// Parse source.
	owner, repo, _, err := github.ParseSource(exec.Source)
	if err != nil {
		return false, err
	}

	// Fetch latest release.
	fmt.Printf("Checking for updates from %s/%s...\n", owner, repo)
	release, err := github.GetLatestRelease(owner, repo, opts.IncludePrereleases)
	if err != nil {
		return false, err
	}

	latestVersion := release.TagName

	// Handle missing executable.
	if executableMissing {
		fmt.Printf("Executable file is MISSING at %s\n", exec.Path)
		fmt.Printf("Recorded version:  %s\n", exec.Version)
		fmt.Printf("Latest version:    %s\n", latestVersion)
		fmt.Println()

		if !opts.Yes {
			if exec.Version == latestVersion {
				fmt.Printf("Reinstall %s %s? [y/N]: ", opts.Name, exec.Version)
			} else {
				fmt.Printf("Install %s? [r]ecorded %s / [l]atest %s / [N]o: ", opts.Name, exec.Version, latestVersion)
			}
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.ToLower(strings.TrimSpace(response))

			if exec.Version == latestVersion {
				if response != "y" && response != "yes" {
					fmt.Println("Reinstall cancelled.")
					return false, nil
				}
			} else {
				switch response {
				case "r", "recorded":
					// Use recorded version - need to fetch that specific release.
					release, err = github.GetRelease(owner, repo, exec.Version)
					if err != nil {
						return false, fmt.Errorf("failed to fetch recorded version %s: %w", exec.Version, err)
					}
					latestVersion = exec.Version
				case "l", "latest":
					// Use latest - already have it.
				default:
					fmt.Println("Reinstall cancelled.")
					return false, nil
				}
			}
		}
		// Continue with installation using selected version.
	} else {
		// Normal update flow - check if update is needed.
		if exec.Version == latestVersion {
			fmt.Printf("%s is already up to date (%s).\n", opts.Name, exec.Version)
			return false, nil
		}

		// Show comparison.
		fmt.Printf("Current version: %s\n", exec.Version)
		fmt.Printf("Latest version:  %s\n", latestVersion)
		fmt.Println()

		// Confirm update.
		if !opts.Yes {
			fmt.Printf("Update %s to %s? [y/N]: ", opts.Name, latestVersion)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Update cancelled.")
				return false, nil
			}
		}
	}

	// Ask about backup (only if file exists).
	createBackup := false
	if !executableMissing && !opts.Yes {
		fmt.Print("Create backup of current executable? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		createBackup = response == "y" || response == "yes"
	}

	// Find matching asset.
	asset, err := github.FindAsset(release.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return false, err
	}

	// Create temporary directory for download.
	tmpDir, err := os.MkdirTemp("", "execman-update-*")
	if err != nil {
		return false, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download asset.
	archivePath := filepath.Join(tmpDir, asset.Name)
	fmt.Printf("Downloading %s...\n", asset.Name)
	if err := github.DownloadAsset(asset, archivePath); err != nil {
		return false, err
	}

	// Extract binary to temp location.
	binaryPath := filepath.Join(tmpDir, "binary")
	fmt.Println("Extracting...")
	if err := archive.ExtractBinary(archivePath, binaryPath); err != nil {
		return false, err
	}

	// Calculate checksum.
	checksum, err := archive.CalculateChecksum(binaryPath)
	if err != nil {
		return false, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Check permissions on target.
	targetDir := filepath.Dir(effectivePath)
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return false, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Create backup if requested.
	if createBackup {
		backupPath := effectivePath + ".backup"
		fmt.Printf("Creating backup at %s...\n", backupPath)
		if err := copyFile(effectivePath, backupPath); err != nil {
			return false, fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Replace executable.
	fmt.Println("Installing...")
	if err := os.Remove(effectivePath); err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to remove old executable: %w", err)
	}

	if err := copyFile(binaryPath, effectivePath); err != nil {
		return false, fmt.Errorf("failed to install new executable: %w", err)
	}

	// Set executable permissions.
	// #nosec G302 -- Executables need 0755 permissions
	if err := os.Chmod(effectivePath, 0755); err != nil {
		return false, fmt.Errorf("failed to set executable permissions: %w", err)
	}

	// Update registry - if we replaced the symlink itself, update the path.
	if symlinkInfo != nil && symlinkInfo.IsSymlink && symlinkAction == symlink.ActionReplaceSymlink {
		exec.Path = effectivePath
	}
	exec.Version = latestVersion
	exec.Checksum = checksum
	exec.InstalledAt = time.Now()

	reg.Add(opts.Name, exec)
	if err := reg.Save(); err != nil {
		return false, fmt.Errorf("failed to update registry: %w", err)
	}

	fmt.Printf("\nSuccessfully updated %s to %s\n", opts.Name, latestVersion)

	// Ask about cleanup.
	if !opts.Yes {
		fmt.Print("Delete download archive? [Y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "" || response == "y" || response == "yes" {
			// tmpDir cleanup is handled by defer
			fmt.Println("Archive deleted.")
		} else {
			// Move archive to current directory before tmpDir cleanup.
			finalPath := filepath.Join(".", asset.Name)
			if err := copyFile(archivePath, finalPath); err == nil {
				fmt.Printf("Archive saved to %s\n", finalPath)
			}
		}
	}

	return true, nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	// #nosec G304 -- Reading from controlled temp directory and registry paths
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	// #nosec G306 -- Executables need 0755 permissions
	return os.WriteFile(dst, data, 0755)
}
