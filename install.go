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

// runLinkShims creates symlinks in linkDir (e.g. /usr/local/bin), one per name, pointing at
// the corresponding shim in the configured shims directory. Use when a tool ignores PATH
// (e.g. Claude Code’s Bash snapshot): on macOS, /usr/local/bin usually precedes
// /opt/homebrew/bin in the default PATH order, so `command -v gh` resolves to the shim.
func runLinkShims(linkDir string, names []string, stderr io.Writer) error {
	if len(names) == 0 {
		return fmt.Errorf("no shim names (use -names)")
	}
	linkDir = filepath.Clean(linkDir)
	shimDir, err := config.ShimDir()
	if err != nil {
		return err
	}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		src := filepath.Join(shimDir, name)
		if _, err := os.Lstat(src); err != nil {
			return fmt.Errorf("missing shim %q (run cmdproxy init first): %w", src, err)
		}
		srcAbs, err := filepath.Abs(src)
		if err != nil {
			return err
		}
		srcAbs, err = filepath.EvalSymlinks(srcAbs)
		if err != nil {
			return err
		}
		dst := filepath.Join(linkDir, name)
		_ = os.Remove(dst)
		if err := os.Symlink(srcAbs, dst); err != nil {
			return fmt.Errorf("symlink %s -> %s: %w (try sudo)", dst, srcAbs, err)
		}
		fmt.Fprintf(stderr, "linked %s -> %s\n", dst, src)
	}
	return nil
}
