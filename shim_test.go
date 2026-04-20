package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemovePathEntry(t *testing.T) {
	sep := string(os.PathListSeparator)
	a := filepath.Join("/tmp", "shims")
	b := filepath.Join("/usr", "bin")
	pathEnv := a + sep + b + sep + "/opt/bin"
	got := removePathEntry(pathEnv, a)
	want := b + sep + "/opt/bin"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
