package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchRuleAllow(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.json")
	s, err := NewStore(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.MatchRule("gh pr list"); ok {
		t.Fatal("expected no match")
	}
	if err := s.AppendRule(Rule{Pattern: `^gh pr list$`, Action: ActionAllow}); err != nil {
		t.Fatal(err)
	}
	r, ok := s.MatchRule("gh pr list")
	if !ok || r.Action != ActionAllow || r.AllowEnvKeys != nil {
		t.Fatalf("got %+v %v", r, ok)
	}
}

func TestMatchRuleDeny(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(p, []byte(`{"rules":[{"pattern":"^x$","action":"deny"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := NewStore(p)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := s.MatchRule("x")
	if !ok || r.Action != ActionDeny {
		t.Fatalf("got %+v %v", r, ok)
	}
}

func TestAllowEnvKeysJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.json")
	keys := []string{"A", "B"}
	if err := os.WriteFile(p, []byte(`{"rules":[{"pattern":"^t$","action":"allow","allow_env_keys":["A","B"]}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := NewStore(p)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := s.MatchRule("t")
	if !ok || r.AllowEnvKeys == nil || len(*r.AllowEnvKeys) != 2 {
		t.Fatalf("got %+v", r)
	}
	if (*r.AllowEnvKeys)[0] != keys[0] {
		t.Fatal()
	}
}

func TestDeleteRule(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.json")
	s, err := NewStore(p)
	if err != nil {
		t.Fatal(err)
	}
	_ = s.AppendRule(Rule{Pattern: "a", Action: ActionAllow})
	_ = s.AppendRule(Rule{Pattern: "b", Action: ActionDeny})
	if err := s.DeleteRule(0); err != nil {
		t.Fatal(err)
	}
	if s.Len() != 1 {
		t.Fatal(s.Len())
	}
}

func TestReload(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.json")
	s, err := NewStore(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(`{"rules":[{"pattern":"^z$","action":"allow"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.Reload(); err != nil {
		t.Fatal(err)
	}
	if s.Len() != 1 {
		t.Fatal()
	}
}
