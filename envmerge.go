package main

import (
	"strings"

	"cmdproxy/internal/rules"
)

// mergeEnv overlays extra keys onto base env lines (KEY=VAL). Later entries in base win unless extra replaces same key.
func mergeEnv(base []string, extra map[string]string) []string {
	m := parseEnv(base)
	for k, v := range extra {
		m[k] = v
	}
	var out []string
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}

func parseEnv(lines []string) map[string]string {
	m := make(map[string]string)
	for _, line := range lines {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		m[k] = v
	}
	return m
}

// mergeEnvForAllowRule returns the env map to merge into the child for an allow rule.
// nil AllowEnvKeys on the rule means merge all keys from priv.
func mergeEnvForAllowRule(priv map[string]string, rule rules.Rule) map[string]string {
	if rule.AllowEnvKeys == nil {
		return cloneStringMap(priv)
	}
	out := make(map[string]string)
	for _, k := range *rule.AllowEnvKeys {
		if v, ok := priv[k]; ok {
			out[k] = v
		}
	}
	return out
}

func cloneStringMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
