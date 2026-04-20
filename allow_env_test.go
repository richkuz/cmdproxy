package main

import "testing"

func TestParseAllowEnvLine(t *testing.T) {
	if parseAllowEnvLine("  ") != nil {
		t.Fatal("empty -> ALL")
	}
	if parseAllowEnvLine("ALL") != nil {
		t.Fatal()
	}
	k := parseAllowEnvLine(" FOO , BAR ")
	if k == nil || len(*k) != 2 {
		t.Fatal()
	}
	k2 := parseAllowEnvLine("x")
	if k2 == nil || len(*k2) != 1 || (*k2)[0] != "x" {
		t.Fatal()
	}
}

func TestFormatAllowEnvKeys(t *testing.T) {
	if formatAllowEnvKeys(nil) != "ALL" {
		t.Fatal()
	}
	empty := []string{}
	if formatAllowEnvKeys(&empty) != "(none)" {
		t.Fatal()
	}
}
