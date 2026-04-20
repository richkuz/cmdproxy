package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"cmdproxy/internal/config"
)

// initOptions tweaks init behavior for embedded use (e.g. bare `cmdproxy` before serve).
type initOptions struct {
	skipNextSteps bool
}

// needsInit reports whether first-time shim setup should run (no shim dir or empty).
func needsInit() (bool, error) {
	shimDir, err := config.ShimDir()
	if err != nil {
		return false, err
	}
	fi, err := os.Stat(shimDir)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	if !fi.IsDir() {
		return true, nil
	}
	entries, err := os.ReadDir(shimDir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func runInit(names []string, assumeYes bool, stderr io.Writer, opts *initOptions) error {
	skipNext := false
	if opts != nil {
		skipNext = opts.skipNextSteps
	}
	cfgDir, err := config.Dir()
	if err != nil {
		return err
	}
	shimDir, err := config.ShimDir()
	if err != nil {
		return err
	}
	sock, err := config.SocketPath()
	if err != nil {
		return err
	}
	rules, err := config.RulesPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		return err
	}
	if err := os.MkdirAll(shimDir, 0o755); err != nil {
		return err
	}

	fmt.Fprintf(stderr, "cmdproxy layout (defaults; override with CMDPROXY_* only if needed):\n")
	fmt.Fprintf(stderr, "  configuration & rules: %s\n", cfgDir)
	fmt.Fprintf(stderr, "  rules file:            %s\n", rules)
	fmt.Fprintf(stderr, "  Unix socket:           %s\n", sock)
	fmt.Fprintf(stderr, "  shims directory:       %s\n", shimDir)
	fmt.Fprintln(stderr)

	if len(names) == 0 {
		return fmt.Errorf("no tool names to shim")
	}

	if !assumeYes {
		fmt.Fprintf(stderr, "Install symlinks for [%s] → %s ? [Y/n] ", strings.Join(names, ", "), shimDir)
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "" && line != "y" && line != "yes" {
			fmt.Fprintln(stderr, "Skipped install-shims.")
			printPATHSnippet(shimDir, stderr)
			if !skipNext {
				printNextSteps(stderr)
			}
			return nil
		}
	}

	if err := runInstallShims(shimDir, names, stderr); err != nil {
		return err
	}
	if !skipNext {
		printNextSteps(stderr)
	}
	return nil
}

func printNextSteps(stderr io.Writer) {
	fmt.Fprintln(stderr, "Next:")
	fmt.Fprintln(stderr, "  1. Ensure the PATH line above is in ~/.zshrc (or ~/.bashrc), then open a new terminal.")
	fmt.Fprintln(stderr, "  2. Start the daemon in a terminal (put secrets only in that environment):")
	fmt.Fprintln(stderr, "       export GITHUB_TOKEN=...   # example")
	fmt.Fprintln(stderr, "       cmdproxy serve")
	fmt.Fprintln(stderr, "  3. Run your agent from a shell without those secrets (allowed commands get the serve process env merged in).")
	fmt.Fprintln(stderr, "  (Or run `cmdproxy` alone: it auto-inits if needed and starts the daemon.)")
}
