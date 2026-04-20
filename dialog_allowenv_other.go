//go:build !darwin

package main

import "fmt"

// PromptAllowEnvKeys is unavailable off macOS.
func PromptAllowEnvKeys() (*[]string, error) {
	return nil, fmt.Errorf("cmdproxy env prompt is only supported on macOS (osascript)")
}
