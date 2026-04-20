//go:build !darwin

package main

import (
	"errors"
	"fmt"
)

// ErrCancelled is returned when the user dismisses the dialog.
var ErrCancelled = errors.New("cancelled")

type Choice string

const (
	ChoiceAllowAlways Choice = "Allow Always"
	ChoiceAllowOnce   Choice = "Allow Once"
	ChoiceDenyOnce    Choice = "Deny Once"
	ChoiceDenyAlways  Choice = "Deny Always"
)

// PromptCommand is unavailable off macOS.
func PromptCommand(displayLine string, defaultPattern string) (Choice, string, error) {
	_ = displayLine
	_ = defaultPattern
	return "", "", fmt.Errorf("cmdproxy UI is only supported on macOS (osascript)")
}
