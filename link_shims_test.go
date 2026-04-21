package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"cmdproxy/internal/config"
)

func TestRunLinkShims(t *testing.T) {
	t.Setenv("CMDPROXY_CONFIG_DIR", t.TempDir())
	shimDir, err := config.ShimDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(shimDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fakeBin := filepath.Join(t.TempDir(), "cmdproxy")
	if err := os.WriteFile(fakeBin, []byte{}, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(fakeBin, filepath.Join(shimDir, "gh")); err != nil {
		t.Fatal(err)
	}
	linkDir := t.TempDir()
	if err := runLinkShims(linkDir, []string{"gh"}, io.Discard); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(linkDir, "gh")
	fi, err := os.Lstat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("want symlink at %s", dst)
	}
}
