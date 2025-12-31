package symlink

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckRegularFile(t *testing.T) {
	// Create a temp file.
	tmpDir, err := os.MkdirTemp("", "symlink-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "regular-file")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	info, err := Check(filePath)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	if info.IsSymlink {
		t.Error("expected IsSymlink to be false for regular file")
	}
	if info.Path != filePath {
		t.Errorf("expected Path to be %q, got %q", filePath, info.Path)
	}
}

func TestCheckSymlink(t *testing.T) {
	// Create a temp directory.
	tmpDir, err := os.MkdirTemp("", "symlink-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create target file.
	targetPath := filepath.Join(tmpDir, "target")
	if err := os.WriteFile(targetPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}

	// Create symlink.
	symlinkPath := filepath.Join(tmpDir, "link")
	if err := os.Symlink(targetPath, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	info, err := Check(symlinkPath)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	if !info.IsSymlink {
		t.Error("expected IsSymlink to be true for symlink")
	}
	if info.Path != symlinkPath {
		t.Errorf("expected Path to be %q, got %q", symlinkPath, info.Path)
	}
	if info.Target != targetPath {
		t.Errorf("expected Target to be %q, got %q", targetPath, info.Target)
	}
}

func TestCheckNonExistent(t *testing.T) {
	info, err := Check("/non/existent/path")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	if info.IsSymlink {
		t.Error("expected IsSymlink to be false for non-existent path")
	}
}

func TestResolveTarget(t *testing.T) {
	info := &Info{
		IsSymlink: true,
		Path:      "/usr/local/bin/myapp",
		Target:    "/opt/myapp/v1.2.3/myapp",
	}

	// ActionReplaceTarget should return the target.
	if result := ResolveTarget(info, ActionReplaceTarget); result != info.Target {
		t.Errorf("ActionReplaceTarget: expected %q, got %q", info.Target, result)
	}

	// ActionReplaceSymlink should return the symlink path.
	if result := ResolveTarget(info, ActionReplaceSymlink); result != info.Path {
		t.Errorf("ActionReplaceSymlink: expected %q, got %q", info.Path, result)
	}
}

func TestErrorNonInteractive(t *testing.T) {
	err := ErrorNonInteractive("/usr/local/bin/myapp", "/opt/myapp/myapp")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "symlink") {
		t.Errorf("error message should mention 'symlink': %s", errStr)
	}
	if !strings.Contains(errStr, "non-interactive") {
		t.Errorf("error message should mention 'non-interactive': %s", errStr)
	}
}
