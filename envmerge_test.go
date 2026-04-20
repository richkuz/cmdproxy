package main

import (
	"slices"
	"testing"

	"cmdproxy/internal/rules"
)

func TestMergeEnvForAllowRule(t *testing.T) {
	priv := map[string]string{"A": "1", "B": "2", "C": "3"}
	r := rules.Rule{Action: rules.ActionAllow, AllowEnvKeys: nil}
	got := mergeEnvForAllowRule(priv, r)
	if len(got) != 3 || got["A"] != "1" {
		t.Fatalf("%v", got)
	}
	keys := []string{"B"}
	r2 := rules.Rule{Action: rules.ActionAllow, AllowEnvKeys: &keys}
	got2 := mergeEnvForAllowRule(priv, r2)
	if len(got2) != 1 || got2["B"] != "2" {
		t.Fatalf("%v", got2)
	}
	empty := []string{}
	r3 := rules.Rule{Action: rules.ActionAllow, AllowEnvKeys: &empty}
	got3 := mergeEnvForAllowRule(priv, r3)
	if len(got3) != 0 {
		t.Fatalf("%v", got3)
	}
}

func TestMergeEnv(t *testing.T) {
	base := []string{"A=1", "B=2"}
	extra := map[string]string{"B": "3", "C": "4"}
	got := mergeEnv(base, extra)
	slices.Sort(got)
	want := []string{"A=1", "B=3", "C=4"}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
