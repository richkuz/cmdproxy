package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cmdproxy/internal/config"
)

func main() {
	base := filepath.Base(os.Args[0])
	if base != "cmdproxy" {
		os.Exit(runShim())
	}

	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	installCmd := flag.NewFlagSet("install-shims", flag.ExitOnError)
	installDir := installCmd.String("dir", "", "shim directory (default: ~/Library/Application Support/cmdproxy/shims on macOS)")
	installNames := installCmd.String("names", "gh,git", "comma-separated tool names to shim")

	linkCmd := flag.NewFlagSet("link-shims", flag.ExitOnError)
	linkTo := linkCmd.String("to", "/usr/local/bin", "directory for symlinks (should sort before Homebrew in PATH; often /usr/local/bin)")
	linkNames := linkCmd.String("names", "gh,git", "comma-separated tool names (must already exist under the shims directory)")

	uninstallCmd := flag.NewFlagSet("uninstall", flag.ExitOnError)
	uninstallYes := uninstallCmd.Bool("y", false, "non-interactive: do not prompt")
	uninstallKeepLinks := uninstallCmd.Bool("keep-system-links", false, "do not remove symlinks from -from")
	uninstallFrom := uninstallCmd.String("from", "/usr/local/bin", "directory where link-shims placed symlinks (must match link-shims -to)")
	uninstallNames := uninstallCmd.String("names", "gh,git", "comma-separated names to unlink from -from when removing system links")

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initNames := initCmd.String("names", "gh,git", "comma-separated tool names to shim")
	initYes := initCmd.Bool("y", false, "non-interactive: create directories and install shims without prompting")

	waitSockCmd := flag.NewFlagSet("wait-socket", flag.ExitOnError)
	waitTimeout := waitSockCmd.Duration("timeout", 5*time.Second, "max time to wait for the Unix socket")
	waitLog := waitSockCmd.String("log", "", "on failure, print this file to stderr (e.g. daemon log path)")

	if len(os.Args) < 2 {
		need, err := needsInit()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cmdproxy: %v\n", err)
			os.Exit(1)
		}
		if need {
			fmt.Fprintf(os.Stderr, "cmdproxy: first-time setup (same as `cmdproxy init -y`)…\n\n")
			names := splitToolNames("gh,git")
			if err := runInit(names, true, os.Stderr, &initOptions{skipNextSteps: true}); err != nil {
				fmt.Fprintf(os.Stderr, "cmdproxy: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintln(os.Stderr)
		}
		if err := runServe(os.Stderr); err != nil {
			fmt.Fprintf(os.Stderr, "cmdproxy serve: %v\n", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "serve":
		_ = serveCmd.Parse(os.Args[2:])
		if err := runServe(os.Stderr); err != nil {
			fmt.Fprintf(os.Stderr, "cmdproxy serve: %v\n", err)
			os.Exit(1)
		}
	case "install-shims":
		_ = installCmd.Parse(os.Args[2:])
		names := splitToolNames(*installNames)
		if err := runInstallShims(*installDir, names, os.Stderr); err != nil {
			fmt.Fprintf(os.Stderr, "cmdproxy install-shims: %v\n", err)
			os.Exit(1)
		}
	case "link-shims":
		_ = linkCmd.Parse(os.Args[2:])
		names := splitToolNames(*linkNames)
		if err := runLinkShims(*linkTo, names, os.Stderr); err != nil {
			fmt.Fprintf(os.Stderr, "cmdproxy link-shims: %v\n", err)
			os.Exit(1)
		}
	case "uninstall":
		_ = uninstallCmd.Parse(os.Args[2:])
		names := splitToolNames(*uninstallNames)
		removeLinks := !*uninstallKeepLinks
		if err := runUninstall(*uninstallFrom, names, removeLinks, *uninstallYes, os.Stdin, os.Stderr); err != nil {
			fmt.Fprintf(os.Stderr, "cmdproxy uninstall: %v\n", err)
			os.Exit(1)
		}
	case "init":
		_ = initCmd.Parse(os.Args[2:])
		names := splitToolNames(*initNames)
		if err := runInit(names, *initYes, os.Stderr, nil); err != nil {
			fmt.Fprintf(os.Stderr, "cmdproxy init: %v\n", err)
			os.Exit(1)
		}
	case "shim-path":
		if len(os.Args) > 2 {
			fmt.Fprintf(os.Stderr, "cmdproxy shim-path: unexpected arguments\n")
			os.Exit(2)
		}
		d, err := config.ShimDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cmdproxy shim-path: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(d)
	case "wait-socket":
		_ = waitSockCmd.Parse(os.Args[2:])
		if err := runWaitSocket(os.Stderr, *waitTimeout, *waitLog); err != nil {
			fmt.Fprintf(os.Stderr, "cmdproxy wait-socket: %v\n", err)
			os.Exit(1)
		}
	case "config":
		if len(os.Args) > 2 {
			fmt.Fprintf(os.Stderr, "cmdproxy config: unexpected arguments\n")
			os.Exit(2)
		}
		if err := runConfig(os.Stdin, os.Stdout, os.Stderr); err != nil {
			fmt.Fprintf(os.Stderr, "cmdproxy config: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func splitToolNames(s string) []string {
	var names []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			names = append(names, p)
		}
	}
	return names
}

func usage() {
	fmt.Fprintf(os.Stderr, `cmdproxy — gate subprocess execution and merge a privileged environment after approval.

  cmdproxy
    If shims are missing, runs first-time setup (like init -y), then starts the daemon.
    Same as: cmdproxy serve (after auto-init if needed).

  cmdproxy init [-names gh,git,...] [-y]
    Creates the data directory, prints default paths, and installs shims under
    ~/Library/Application Support/cmdproxy/shims (macOS). No env vars required for defaults.

  cmdproxy serve
    Run the daemon (osascript prompts on macOS). Prints configuration and default paths.
    On allow, merges env from this process into the child (all keys by default; allow rules can restrict keys).

  cmdproxy install-shims [-dir DIR] [-names gh,git,...]
    Create symlinks only (default shim dir matches init).

  cmdproxy link-shims [-to /usr/local/bin] [-names gh,git,...]
    Symlink each shim into -to (default /usr/local/bin). Use when an agent ignores PATH:
    macOS usually orders /usr/local/bin before /opt/homebrew/bin, so gh resolves to cmdproxy.
    May require: sudo cmdproxy link-shims

  cmdproxy uninstall [-y] [-keep-system-links] [-from /usr/local/bin] [-names gh,git,...]
    Remove the cmdproxy data directory (rules, shims, socket) and, by default, matching
    symlinks created by link-shims under -from. Stop cmdproxy serve first.

  cmdproxy shim-path
    Print the default shims directory (stdout only). For scripts:
      export PATH="$(cmdproxy shim-path):$PATH"

  cmdproxy wait-socket [-timeout 5s] [-log PATH]
    Block until the daemon's Unix socket exists (same path as serve / shims use).
    Optional -log: on timeout, print that file (e.g. /tmp/cmdproxy-serve.log) for debugging.

  cmdproxy config
    Interactive session: manage shims and allow/deny rules (including per-allow env keys).

When invoked via a shim name (e.g. "gh"), connects to the daemon and proxies execution.

Optional overrides (only if you change layout):
  CMDPROXY_CONFIG_DIR   Override the data directory (socket + rules + default shims path).
  CMDPROXY_SOCKET       Unix socket path.
  CMDPROXY_SHIM_DIR     Shim directory if not using defaults.

`)
}
