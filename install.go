package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cmdproxy/internal/config"
)

func runInstallShims(binDir string, names []string, stderr io.Writer) error {
	if strings.TrimSpace(binDir) == "" {
		var err error
		binDir, err = config.ShimDir()
		if err != nil {
			return err
		}
	} else {
		binDir = filepath.Clean(binDir)
	}
	if len(names) == 0 {
		return fmt.Errorf("no shim names (use -names)")
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		target := filepath.Join(binDir, name)
		_ = os.Remove(target)
		if err := os.Symlink(self, target); err != nil {
			return fmt.Errorf("symlink %s: %w", target, err)
		}
		fmt.Fprintf(stderr, "installed shim %s -> %s\n", target, self)
	}
	printPATHSnippet(binDir, stderr)
	fmt.Fprintf(stderr, "Override only if needed: CMDPROXY_SHIM_DIR=%q\n", binDir)
	return nil
}

func printPATHSnippet(shimDir string, stderr io.Writer) {
	q := shellQuoteIfNeeded(shimDir)
	fmt.Fprintf(stderr, "\nPrepend shims to PATH (e.g. in ~/.zshrc for agents):\n")
	fmt.Fprintf(stderr, "  export PATH=%s:\"$PATH\"\n", q)
}

func shellQuoteIfNeeded(s string) string {
	if strings.ContainsAny(s, ` "'\$!`) {
		return fmt.Sprintf("%q", s)
	}
	return s
}
