package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// Dir returns the application support directory for cmdproxy (config + socket).
func Dir() (string, error) {
	if d := os.Getenv("CMDPROXY_CONFIG_DIR"); d != "" {
		return filepath.Clean(d), nil
	}
	if runtime.GOOS == "darwin" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "cmdproxy"), nil
	}
	// Non-macOS: XDG-style fallback for development/tests.
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg, "cmdproxy"), nil
}

// SocketPath returns the Unix domain socket path for the daemon.
func SocketPath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "cmdproxy.sock"), nil
}

// RulesPath returns the path to rules.json.
func RulesPath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "rules.json"), nil
}

// ShimDir returns the default directory for tool symlinks (…/cmdproxy/shims).
func ShimDir() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "shims"), nil
}

// PrintDataDirLocation writes the canonical config directory to w (e.g. stderr on serve).
func PrintDataDirLocation(w io.Writer) error {
	d, err := Dir()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "cmdproxy configuration directory: %s\n", d)
	return err
}
