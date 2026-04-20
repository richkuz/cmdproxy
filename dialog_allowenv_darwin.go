//go:build darwin

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// PromptAllowEnvKeys asks which serve-process env vars to merge for an allow rule.
func PromptAllowEnvKeys() (*[]string, error) {
	script := `
on run argv
	set def to item 1 of argv
	set d to display dialog "Env vars to pass from cmdproxy (comma-separated), or ALL for full environment." default answer def with title "cmdproxy allow rule" buttons {"Cancel", "OK"} default button "OK"
	if button returned of d is "Cancel" then error "cancelled"
	return text returned of d
end run
`
	cmd := exec.Command("osascript", "-", "ALL")
	cmd.Stdin = strings.NewReader(script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		s := string(out)
		if strings.Contains(s, "cancelled") || strings.Contains(s, "User canceled") {
			return nil, ErrCancelled
		}
		return nil, fmt.Errorf("osascript: %w: %s", err, strings.TrimSpace(s))
	}
	return parseAllowEnvLine(strings.TrimSpace(string(out))), nil
}
