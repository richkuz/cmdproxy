//go:build darwin

package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrCancelled is returned when the user dismisses the dialog.
var ErrCancelled = errors.New("cancelled")

// Choice is the user's decision from the osascript UI.
type Choice string

const (
	ChoiceAllowAlways Choice = "Allow Always"
	ChoiceAllowOnce   Choice = "Allow Once"
	ChoiceDenyOnce    Choice = "Deny Once"
	ChoiceDenyAlways  Choice = "Deny Always"
)

// PromptCommand shows a modal dialog and optional regex editor. defaultPattern is used when saving Always rules.
func PromptCommand(displayLine string, defaultPattern string) (Choice, string, error) {
	prompt := displayLine
	if len(prompt) > 800 {
		prompt = prompt[:800] + "…"
	}
	script := `
on run argv
	set promptText to item 1 of argv
	set defPat to item 2 of argv
	set choiceList to choose from list {"Allow Always", "Allow Once", "Deny Once", "Deny Always"} with title "cmdproxy" with prompt promptText default items {"Allow Once"}
	if choiceList is false then error "cancelled"
	set c to item 1 of choiceList
	if c is "Allow Always" or c is "Deny Always" then
		set d to display dialog "Edit regex (RE2). Cancel to abort." default answer defPat with title "cmdproxy rule" buttons {"Cancel", "OK"} default button "OK"
		if button returned of d is "Cancel" then error "cancelled"
		set pat to text returned of d
		return c & (ASCII character 10) & pat
	else
		return c
	end if
end run
`
	cmd := exec.Command("osascript", "-", prompt, defaultPattern)
	cmd.Stdin = strings.NewReader(script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		s := string(out)
		if strings.Contains(s, "cancelled") || strings.Contains(s, "User canceled") {
			return "", "", ErrCancelled
		}
		return "", "", fmt.Errorf("osascript: %w: %s", err, strings.TrimSpace(s))
	}
	text := strings.TrimSpace(string(out))
	lines := strings.SplitN(text, "\n", 2)
	ch := Choice(lines[0])
	switch ch {
	case ChoiceAllowAlways, ChoiceAllowOnce, ChoiceDenyOnce, ChoiceDenyAlways:
	default:
		return "", "", fmt.Errorf("unexpected choice: %q", lines[0])
	}
	pat := ""
	if len(lines) == 2 {
		pat = strings.TrimRight(lines[1], "\r")
	}
	return ch, pat, nil
}
