package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSystemLinkMatchesShim(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "cmdproxy")
	if err := os.WriteFile(bin, []byte{}, 0o755); err != nil {
		t.Fatal(err)
	}
	shim := filepath.Join(dir, "shims", "gh")
	if err := os.MkdirAll(filepath.Dir(shim), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(bin, shim); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "usr-local-gh")
	if err := os.Symlink(bin, dst); err != nil {
		t.Fatal(err)
	}
	ok, err := systemLinkMatchesShim(dst, shim)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected match when both resolve to same binary")
	}

	other := filepath.Join(dir, "other")
	if err := os.WriteFile(other, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst2 := filepath.Join(dir, "usr-local-other")
	if err := os.Symlink(other, dst2); err != nil {
		t.Fatal(err)
	}
	ok2, err := systemLinkMatchesShim(dst2, shim)
	if err != nil {
		t.Fatal(err)
	}
	if ok2 {
		t.Fatal("expected no match for different target")
	}
}
