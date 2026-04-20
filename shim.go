package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"cmdproxy/internal/config"
	"cmdproxy/internal/protocol"
)

func runShim() int {
	bin := filepath.Base(os.Args[0])
	shimDir, err := shimDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cmdproxy shim: %v\n", err)
		return 126
	}
	realPath, err := lookupRealBinary(bin, shimDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cmdproxy shim: resolve %q: %v\n", bin, err)
		return 126
	}

	sock, err := socketPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cmdproxy shim: %v\n", err)
		return 126
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cmdproxy shim: %v\n", err)
		return 126
	}

	req := protocol.RunRequest{
		Argv: append([]string{bin}, os.Args[1:]...),
		Cwd:  cwd,
		Env:  os.Environ(),
	}

	conn, err := net.Dial("unix", sock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cmdproxy shim: connect %s: %v (is cmdproxy serve running?)\n", sock, err)
		return 126
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(&req); err != nil {
		fmt.Fprintf(os.Stderr, "cmdproxy shim: send: %v\n", err)
		return 126
	}
	var resp protocol.RunResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		fmt.Fprintf(os.Stderr, "cmdproxy shim: read response: %v\n", err)
		return 126
	}
	if !resp.Allow {
		if resp.Message != "" {
			fmt.Fprintf(os.Stderr, "cmdproxy: %s\n", resp.Message)
		}
		return 1
	}

	merged := mergeEnv(os.Environ(), resp.MergeEnv)
	argv := append([]string{realPath}, os.Args[1:]...)
	if err := syscall.Exec(realPath, argv, merged); err != nil {
		fmt.Fprintf(os.Stderr, "cmdproxy shim: exec: %v\n", err)
		return 126
	}
	return 0
}

func shimDir() (string, error) {
	if d := os.Getenv("CMDPROXY_SHIM_DIR"); d != "" {
		return filepath.Clean(d), nil
	}
	// If invoked as /path/to/.../shims/gh, use that directory (do not resolve to the real cmdproxy binary).
	if len(os.Args) > 0 {
		a0 := os.Args[0]
		if filepath.IsAbs(a0) {
			return filepath.Clean(filepath.Dir(a0)), nil
		}
	}
	// e.g. argv0 is "gh" — use the same default directory as `cmdproxy init`.
	return config.ShimDir()
}

func lookupRealBinary(bin, shimDir string) (string, error) {
	oldPath := os.Getenv("PATH")
	cleaned := removePathEntry(oldPath, shimDir)
	if err := os.Setenv("PATH", cleaned); err != nil {
		return "", err
	}
	defer func() { _ = os.Setenv("PATH", oldPath) }()
	return exec.LookPath(bin)
}

func removePathEntry(pathEnv, dir string) string {
	dir = filepath.Clean(dir)
	parts := strings.Split(pathEnv, string(os.PathListSeparator))
	var kept []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		if filepath.Clean(p) == dir {
			continue
		}
		kept = append(kept, p)
	}
	return strings.Join(kept, string(os.PathListSeparator))
}

func socketPath() (string, error) {
	if p := os.Getenv("CMDPROXY_SOCKET"); p != "" {
		return p, nil
	}
	return config.SocketPath()
}
