package main

import (
	"strings"
)

// parseAllowEnvLine interprets UI / config input for allow-rule env keys.
// Empty string or "ALL" (any case) means merge all serve env (returns nil).
// Otherwise comma-separated keys; empty list after parsing means merge no keys from serve.
func parseAllowEnvLine(s string) *[]string {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "ALL") {
		return nil
	}
	var keys []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			keys = append(keys, p)
		}
	}
	return &keys
}

func formatAllowEnvKeys(keys *[]string) string {
	if keys == nil {
		return "ALL"
	}
	if len(*keys) == 0 {
		return "(none)"
	}
	return strings.Join(*keys, ",")
}
