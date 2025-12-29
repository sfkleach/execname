package install

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
)

// Options represents the install command options.
type Options struct {
	Source             string
	Into               string
	Yes                bool
	IncludePrereleases bool
}

// Run executes the install command.
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

	// Parse source.
	owner, repo, version, err := github.ParseSource(opts.Source)
	if err != nil {
		return err
	}

	// Use config defaults if not specified.
	if opts.Into == "" {
		opts.Into = cfg.DefaultInstallDir
	}
	if !opts.IncludePrereleases {
		opts.IncludePrereleases = cfg.IncludePrereleases
	}

	// Fetch release.
	var release *github.Release
	if version != "" {
		fmt.Printf("Fetching release %s from %s/%s...\n", version, owner, repo)
		release, err = github.GetRelease(owner, repo, version)
	} else {
		fmt.Printf("Fetching latest release from %s/%s...\n", owner, repo)
		release, err = github.GetLatestRelease(owner, repo, opts.IncludePrereleases)
	}
	if err != nil {
		return err
	}

	// Set version from release tag if we fetched the latest.
	if version == "" {
		version = release.TagName
	}

	// Check if already installed.
	execName := repo
	if existing, found := reg.Get(execName); found {
		if existing.Version == version {
			fmt.Printf("Warning: %s version %s is already installed at %s\n", execName, version, existing.Path)
			if !opts.Yes {
				fmt.Print("Reinstall? (y/N): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					fmt.Println("Installation cancelled.")
					return nil
				}
			}
		}
	}

	// Confirm installation.
	targetPath := filepath.Join(opts.Into, execName)
	fmt.Printf("\nInstallation Details:\n")
	fmt.Printf("  Repository: %s\n", github.ToURL(owner, repo))
	fmt.Printf("  Version:    %s\n", version)
	fmt.Printf("  Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  Target:     %s\n", targetPath)

	if !opts.Yes {
		fmt.Print("\nProceed with installation? (Y/n): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response == "n" || response == "no" {
			fmt.Println("Installation cancelled.")
			return nil
		}
	}

	// Find matching asset.
	fmt.Println("\nFinding matching asset...")
	asset, err := github.FindAsset(release.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		fmt.Println("\nAvailable assets:")
		for _, a := range release.Assets {
			fmt.Printf("  - %s\n", a.Name)
		}
		return fmt.Errorf("no matching asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	fmt.Printf("Found: %s\n", asset.Name)

	// Create temp directory for download.
	tempDir, err := os.MkdirTemp("", "execman-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, asset.Name)

	// Download asset.
	fmt.Printf("\nDownloading %s...\n", asset.Name)
	if err := github.DownloadAsset(asset, archivePath); err != nil {
		return err
	}
	fmt.Println("Download complete.")

	// Try to download and verify checksum (optional, won't fail if not available).
	checksumPath := filepath.Join(tempDir, "checksums.txt")
	var expectedChecksum string
	for _, a := range release.Assets {
		if strings.Contains(strings.ToLower(a.Name), "checksum") ||
			strings.HasSuffix(strings.ToLower(a.Name), ".sha256") {
			fmt.Println("\nDownloading checksums...")
			if err := github.DownloadAsset(&a, checksumPath); err == nil {
				checksum, err := archive.FindChecksumInFile(checksumPath, asset.Name)
				if err == nil {
					expectedChecksum = checksum
					fmt.Println("Verifying checksum...")
					archiveChecksum, err := archive.CalculateChecksum(archivePath)
					if err != nil {
						return fmt.Errorf("failed to calculate checksum: %w", err)
					}
					if archiveChecksum != expectedChecksum {
						return fmt.Errorf("checksum verification failed")
					}
					fmt.Println("Checksum verified.")
				}
			}
			break
		}
	}

	// Ensure target directory exists.
	// #nosec G301 -- Install directory needs 0755 for executables to be accessible
	if err := os.MkdirAll(opts.Into, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract binary.
	fmt.Println("\nExtracting binary...")
	if err := archive.ExtractBinary(archivePath, targetPath); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	// Calculate checksum of installed binary.
	fmt.Println("Calculating checksum of installed binary...")
	checksum, err := archive.CalculateChecksum(targetPath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Register executable.
	fmt.Println("Updating registry...")
	platformStr := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	reg.Add(execName, &registry.Executable{
		Source:      github.ToURL(owner, repo),
		Version:     version,
		InstalledAt: time.Now(),
		Path:        targetPath,
		Platform:    platformStr,
		Checksum:    checksum,
	})

	if err := reg.Save(); err != nil {
		return fmt.Errorf("failed to save registry: %w", err)
	}

	fmt.Printf("\n✓ Successfully installed %s %s to %s\n", execName, version, targetPath)
	return nil
}
