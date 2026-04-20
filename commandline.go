package main

import (
	"path/filepath"
	"regexp"
	"strings"
)

// CommandLine builds the canonical match string: basename argv0 + space + args...
func CommandLine(argv []string) string {
	if len(argv) == 0 {
		return ""
	}
	base := filepath.Base(argv[0])
	rest := argv[1:]
	var b strings.Builder
	b.WriteString(base)
	for _, a := range rest {
		b.WriteByte(' ')
		b.WriteString(a)
	}
	return b.String()
}

// DefaultPattern returns a regex that matches only this exact command line (metacharacters escaped).
func DefaultPattern(commandLine string) string {
	return "^" + regexp.QuoteMeta(commandLine) + "$"
}
