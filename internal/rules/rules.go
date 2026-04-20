package rules

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"sync"
)

// Action is persisted per rule.
type Action string

const (
	ActionAllow Action = "allow"
	ActionDeny  Action = "deny"
)

// Rule is a single regex match with a decision.
// For allow rules, AllowEnvKeys controls which serve-process env vars are merged:
//   - nil: merge all (same as "ALL" in the UI).
//   - non-nil: merge only these keys ([] means merge none).
//
// Deny rules ignore AllowEnvKeys.
type Rule struct {
	Pattern      string    `json:"pattern"`
	Action       Action    `json:"action"`
	AllowEnvKeys *[]string `json:"allow_env_keys,omitempty"`
}

// Doc is the on-disk rules.json shape.
type Doc struct {
	Rules []Rule `json:"rules"`
}

// Store loads and saves rules with safe concurrent access.
type Store struct {
	mu    sync.RWMutex
	path  string
	rules []Rule
}

// NewStore loads rules from path or starts empty if missing.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.reloadLocked(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) reloadLocked() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.rules = nil
			return nil
		}
		return err
	}
	var doc Doc
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}
	s.rules = doc.Rules
	return nil
}

// Reload re-reads rules from disk.
func (s *Store) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reloadLocked()
}

// MatchRule returns the first matching rule and true, or false if none.
func (s *Store) MatchRule(commandLine string) (Rule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			continue
		}
		if re.MatchString(commandLine) {
			return r, true
		}
	}
	return Rule{}, false
}

// AppendRule adds a rule and persists.
func (s *Store) AppendRule(r Rule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = append(s.rules, r)
	return s.saveLocked()
}

// SetRule replaces the rule at index i.
func (s *Store) SetRule(i int, r Rule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i < 0 || i >= len(s.rules) {
		return errors.New("invalid rule index")
	}
	s.rules[i] = r
	return s.saveLocked()
}

// DeleteRule removes the rule at index i.
func (s *Store) DeleteRule(i int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i < 0 || i >= len(s.rules) {
		return errors.New("invalid rule index")
	}
	s.rules = append(s.rules[:i], s.rules[i+1:]...)
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	doc := Doc{Rules: s.rules}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Snapshot returns a copy for tests.
func (s *Store) Snapshot() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Rule, len(s.rules))
	copy(out, s.rules)
	return out
}

// Len returns the number of rules.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.rules)
}
