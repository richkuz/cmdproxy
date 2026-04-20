package main

import (
	"os"
	"path/filepath"
	"testing"

	"cmdproxy/internal/config"
)

func TestNeedsInitEmpty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CMDPROXY_CONFIG_DIR", dir)
	ok, err := needsInit()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected need init when shim dir missing")
	}
}

func TestNeedsInitWithShim(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CMDPROXY_CONFIG_DIR", dir)
	shimDir, err := config.ShimDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(shimDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(shimDir, "placeholder"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := needsInit()
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected no init when shim dir non-empty")
	}
}
