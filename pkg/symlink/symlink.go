// Package symlink provides utilities for handling symbolic links.
package symlink

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SymlinkAction represents the action to take when a symlink is encountered.
type SymlinkAction int

const (
	// ActionCancel indicates the operation should be cancelled.
	ActionCancel SymlinkAction = iota
	// ActionReplaceTarget indicates the symlink target should be replaced.
	ActionReplaceTarget
	// ActionReplaceSymlink indicates the symlink itself should be replaced.
	ActionReplaceSymlink
)

// Info contains information about a symlink.
type Info struct {
	// IsSymlink indicates whether the path is a symbolic link.
	IsSymlink bool
	// Path is the original path (the symlink itself).
	Path string
	// Target is the resolved path that the symlink points to.
	Target string
}

// Check checks if the given path is a symbolic link and returns information about it.
func Check(path string) (*Info, error) {
	info := &Info{
		Path:      path,
		IsSymlink: false,
	}

	// Use Lstat to not follow symlinks.
	fileInfo, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist, not a symlink.
			return info, nil
		}
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	// Check if it's a symlink.
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		info.IsSymlink = true
		// Resolve the target.
		target, err := os.Readlink(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read symlink target: %w", err)
		}
		info.Target = target
	}

	return info, nil
}

// PromptAction prompts the user to choose how to handle a symlink.
// Returns the chosen action.
func PromptAction(symlinkPath, targetPath string) SymlinkAction {
	fmt.Printf("\nNote: %s is a symlink to %s\n\n", symlinkPath, targetPath)
	fmt.Println("How would you like to proceed?")
	fmt.Printf("  [1] Replace the symlink target (%s)\n", targetPath)
	fmt.Printf("  [2] Replace the symlink itself (%s)\n", symlinkPath)
	fmt.Println("  [3] Cancel")
	fmt.Print("\nChoice [1/2/3]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	switch response {
	case "1":
		return ActionReplaceTarget
	case "2":
		return ActionReplaceSymlink
	default:
		return ActionCancel
	}
}

// ErrorNonInteractive returns an error for when a symlink is encountered in non-interactive mode.
func ErrorNonInteractive(symlinkPath, targetPath string) error {
	return fmt.Errorf("%s is a symlink to %s\n       Cannot proceed in non-interactive mode.\n       Run without --yes to choose how to handle symlinks", symlinkPath, targetPath)
}

// ResolveTarget resolves the effective path to use based on the symlink action.
// For ActionReplaceTarget, returns the target path.
// For ActionReplaceSymlink, returns the symlink path.
func ResolveTarget(info *Info, action SymlinkAction) string {
	if action == ActionReplaceTarget {
		return info.Target
	}
	return info.Path
}
