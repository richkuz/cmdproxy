package main

import (
	"regexp"
	"testing"
)

func TestCommandLine(t *testing.T) {
	if got := CommandLine([]string{"gh", "pr", "list"}); got != "gh pr list" {
		t.Fatalf("got %q", got)
	}
	if got := CommandLine([]string{"/x/y/gh", "status"}); got != "gh status" {
		t.Fatalf("got %q", got)
	}
}

func TestDefaultPattern(t *testing.T) {
	line := "gh pr list"
	p := DefaultPattern(line)
	re := regexp.MustCompile(p)
	if !re.MatchString(line) {
		t.Fatalf("pattern %q did not match %q", p, line)
	}
	quoted := `gh pr "x"`
	p2 := DefaultPattern(quoted)
	re2 := regexp.MustCompile(p2)
	if !re2.MatchString(quoted) {
		t.Fatalf("pattern %q did not match %q", p2, quoted)
	}
}
