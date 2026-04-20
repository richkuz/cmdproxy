package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cmdproxy/internal/config"
	"cmdproxy/internal/rules"
)

func runConfig(in io.Reader, out io.Writer, errOut io.Writer) error {
	cfgDir, err := config.Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		return err
	}
	rulePath, err := config.RulesPath()
	if err != nil {
		return err
	}
	store, err := rules.NewStore(rulePath)
	if err != nil {
		return err
	}

	br := bufio.NewReader(in)
	for {
		fmt.Fprintln(out, "cmdproxy config — choose:")
		fmt.Fprintln(out, "  1) Shims")
		fmt.Fprintln(out, "  2) Allow / deny rules")
		fmt.Fprintln(out, "  3) Quit")
		fmt.Fprint(out, "> ")
		line, err := br.ReadString('\n')
		if err != nil {
			return err
		}
		switch strings.TrimSpace(line) {
		case "1":
			if err := configShimsMenu(br, out, errOut); err != nil {
				return err
			}
		case "2":
			if err := configRulesMenu(br, out, errOut, store, rulePath); err != nil {
				return err
			}
		case "3", "q", "quit":
			fmt.Fprintln(out, "Done.")
			return nil
		default:
			fmt.Fprintln(out, "Unknown option.")
		}
	}
}

func configShimsMenu(br *bufio.Reader, out, errOut io.Writer) error {
	for {
		shimDir, err := config.ShimDir()
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "\nShims directory: %s\n", shimDir)
		entries, err := os.ReadDir(shimDir)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if os.IsNotExist(err) {
			_ = os.MkdirAll(shimDir, 0o755)
			entries = nil
		}
		fmt.Fprintln(out, "Installed shims:")
		if len(entries) == 0 {
			fmt.Fprintln(out, "  (none)")
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			p := filepath.Join(shimDir, e.Name())
			fmt.Fprintf(out, "  %s\n", e.Name())
			if li, err := os.Lstat(p); err == nil && li.Mode()&os.ModeSymlink != 0 {
				if t, err := os.Readlink(p); err == nil {
					fmt.Fprintf(out, "    -> %s\n", t)
				}
			}
		}
		fmt.Fprintln(out, "  a) Add shims   r) Remove shim   b) Back")
		fmt.Fprint(out, "> ")
		line, err := br.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(strings.ToLower(line))
		switch line {
		case "b", "back":
			return nil
		case "a", "add":
			fmt.Fprint(out, "Comma-separated tool names (e.g. gh,git): ")
			tools, err := br.ReadString('\n')
			if err != nil {
				return err
			}
			names := splitToolNames(tools)
			if len(names) == 0 {
				fmt.Fprintln(out, "No names given.")
				continue
			}
			if err := runInstallShims("", names, errOut); err != nil {
				fmt.Fprintf(errOut, "error: %v\n", err)
			}
		case "r", "remove":
			fmt.Fprint(out, "Shim name to remove: ")
			name, err := br.ReadString('\n')
			if err != nil {
				return err
			}
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			target := filepath.Join(shimDir, name)
			if err := os.Remove(target); err != nil {
				fmt.Fprintf(errOut, "remove %s: %v\n", target, err)
			} else {
				fmt.Fprintf(out, "Removed %s\n", target)
			}
		default:
			fmt.Fprintln(out, "Unknown option.")
		}
	}
}

func configRulesMenu(br *bufio.Reader, out, errOut io.Writer, store *rules.Store, rulePath string) error {
	for {
		if err := store.Reload(); err != nil {
			fmt.Fprintf(errOut, "reload rules: %v\n", err)
		}
		rulesList := store.Snapshot()
		fmt.Fprintln(out, "\nRules ("+rulePath+"):")
		if len(rulesList) == 0 {
			fmt.Fprintln(out, "  (none)")
		}
		for i, r := range rulesList {
			env := ""
			if r.Action == rules.ActionAllow {
				env = " env=" + formatAllowEnvKeys(r.AllowEnvKeys)
			}
			fmt.Fprintf(out, "  [%d] %s pattern=%q%s\n", i, r.Action, r.Pattern, env)
		}
		fmt.Fprintln(out, "  e) Edit rule   d) Delete   aa) Add allow   ad) Add deny   b) Back")
		fmt.Fprint(out, "> ")
		line, err := br.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(strings.ToLower(line))
		switch line {
		case "b", "back":
			return nil
		case "d", "delete":
			fmt.Fprint(out, "Index to delete: ")
			idx, err := readIndex(br)
			if err != nil {
				fmt.Fprintln(out, "invalid index")
				continue
			}
			if err := store.DeleteRule(idx); err != nil {
				fmt.Fprintf(errOut, "%v\n", err)
			} else {
				fmt.Fprintln(out, "Deleted.")
			}
		case "e", "edit":
			fmt.Fprint(out, "Index to edit: ")
			idx, err := readIndex(br)
			if err != nil {
				fmt.Fprintln(out, "invalid index")
				continue
			}
			cur := store.Snapshot()
			if idx < 0 || idx >= len(cur) {
				fmt.Fprintln(out, "out of range")
				continue
			}
			r := cur[idx]
			fmt.Fprintf(out, "Pattern [%s]: ", r.Pattern)
			pat, err := br.ReadString('\n')
			if err != nil {
				return err
			}
			pat = strings.TrimSpace(pat)
			if pat != "" {
				r.Pattern = pat
			}
			if err := validatePattern(r.Pattern); err != nil {
				fmt.Fprintf(errOut, "invalid pattern: %v\n", err)
				continue
			}
			if r.Action == rules.ActionAllow {
				fmt.Fprintf(out, "Env keys (ALL or comma-separated) [%s]: ", formatAllowEnvKeys(r.AllowEnvKeys))
				envLine, err := br.ReadString('\n')
				if err != nil {
					return err
				}
				envTrim := strings.TrimSpace(envLine)
				if envTrim != "" {
					r.AllowEnvKeys = parseAllowEnvLine(envTrim)
				}
			}
			if err := store.SetRule(idx, r); err != nil {
				fmt.Fprintf(errOut, "%v\n", err)
			} else {
				fmt.Fprintln(out, "Updated.")
			}
		case "aa":
			fmt.Fprint(out, "Regex pattern: ")
			pat, err := br.ReadString('\n')
			if err != nil {
				return err
			}
			pat = strings.TrimSpace(pat)
			if err := validatePattern(pat); err != nil {
				fmt.Fprintf(errOut, "invalid pattern: %v\n", err)
				continue
			}
			fmt.Fprint(out, "Env keys (ALL or comma-separated): ")
			envLine, err := br.ReadString('\n')
			if err != nil {
				return err
			}
			keys := parseAllowEnvLine(strings.TrimSpace(envLine))
			r := rules.Rule{Pattern: pat, Action: rules.ActionAllow, AllowEnvKeys: keys}
			if err := store.AppendRule(r); err != nil {
				fmt.Fprintf(errOut, "%v\n", err)
			} else {
				fmt.Fprintln(out, "Appended allow rule.")
			}
		case "ad":
			fmt.Fprint(out, "Regex pattern: ")
			pat, err := br.ReadString('\n')
			if err != nil {
				return err
			}
			pat = strings.TrimSpace(pat)
			if err := validatePattern(pat); err != nil {
				fmt.Fprintf(errOut, "invalid pattern: %v\n", err)
				continue
			}
			r := rules.Rule{Pattern: pat, Action: rules.ActionDeny}
			if err := store.AppendRule(r); err != nil {
				fmt.Fprintf(errOut, "%v\n", err)
			} else {
				fmt.Fprintln(out, "Appended deny rule.")
			}
		default:
			fmt.Fprintln(out, "Unknown option.")
		}
	}
}

func readIndex(br *bufio.Reader) (int, error) {
	line, err := br.ReadString('\n')
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(line))
}
