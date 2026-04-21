package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cmdproxy/internal/config"
)

func runUninstall(linkDir string, names []string, removeSystemLinks bool, assumeYes bool, stdin io.Reader, stderr io.Writer) error {
	cfgDir, err := config.Dir()
	if err != nil {
		return err
	}
	shimDir, err := config.ShimDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(cfgDir); os.IsNotExist(err) {
		fmt.Fprintf(stderr, "cmdproxy: nothing to remove (%s does not exist)\n", cfgDir)
		return nil
	} else if err != nil {
		return err
	}

	if !assumeYes {
		fmt.Fprintf(stderr, "This removes cmdproxy data under:\n  %s\n", cfgDir)
		if removeSystemLinks {
			fmt.Fprintf(stderr, "and symlinks under %s for [%s] (only if they match this install’s shims)\n", linkDir, strings.Join(names, ", "))
		}
		fmt.Fprintf(stderr, "Stop cmdproxy serve first if it is running. Continue? [y/N] ")
		line, err := bufio.NewReader(stdin).ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			fmt.Fprintln(stderr, "Aborted.")
			return nil
		}
	}

	if removeSystemLinks {
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			shimEntry := filepath.Join(shimDir, name)
			dst := filepath.Join(linkDir, name)
			ok, err := systemLinkMatchesShim(dst, shimEntry)
			if err != nil || !ok {
				if err != nil && !os.IsNotExist(err) {
					fmt.Fprintf(stderr, "cmdproxy: skip %s: %v\n", dst, err)
				}
				continue
			}
			if err := os.Remove(dst); err != nil {
				fmt.Fprintf(stderr, "cmdproxy: remove %s: %v (try sudo)\n", dst, err)
				continue
			}
			fmt.Fprintf(stderr, "removed %s\n", dst)
		}
	}

	if err := os.RemoveAll(cfgDir); err != nil {
		return fmt.Errorf("remove data directory: %w", err)
	}
	fmt.Fprintf(stderr, "removed %s\n", cfgDir)
	return nil
}

// systemLinkMatchesShim reports whether dst is a symlink whose resolved target matches
// the resolved target of shimEntry (same as link-shims: both resolve to the cmdproxy binary).
func systemLinkMatchesShim(dst, shimEntry string) (bool, error) {
	fi, err := os.Lstat(dst)
	if err != nil {
		return false, err
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}
	if _, err := os.Lstat(shimEntry); err != nil {
		return false, err
	}
	got, err := filepath.EvalSymlinks(dst)
	if err != nil {
		return false, err
	}
	want, err := filepath.EvalSymlinks(shimEntry)
	if err != nil {
		return false, err
	}
	return got == want, nil
}
