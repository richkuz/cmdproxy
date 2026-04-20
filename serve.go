package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"

	"cmdproxy/internal/config"
	"cmdproxy/internal/privenv"
	"cmdproxy/internal/protocol"
	"cmdproxy/internal/rules"
)

type serveState struct {
	mu       sync.Mutex
	priv     map[string]string
	rulePath string
	store    *rules.Store
}

func runServe(stderr io.Writer) error {
	if err := config.PrintDataDirLocation(stderr); err != nil {
		return err
	}
	if sh, err := config.ShimDir(); err == nil {
		fmt.Fprintf(stderr, "cmdproxy default shims directory: %s\n", sh)
	}
	cfgDir, err := config.Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		return err
	}
	rulePath, err := config.RulesPath()
	if err != nil {
		return err
	}
	store, err := rules.NewStore(rulePath)
	if err != nil {
		return err
	}
	priv := privenv.EnvironMap()
	fmt.Fprintf(stderr, "cmdproxy: serve env has %d vars; allow rules may restrict which keys are merged\n", len(priv))

	sockPath, err := config.SocketPath()
	if err != nil {
		return err
	}
	if p := os.Getenv("CMDPROXY_SOCKET"); p != "" {
		sockPath = p
	}
	_ = os.Remove(sockPath)

	st := &serveState{
		priv:     priv,
		rulePath: rulePath,
		store:    store,
	}

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return fmt.Errorf("listen %s: %w", sockPath, err)
	}
	if err := os.Chmod(sockPath, 0o600); err != nil {
		_ = ln.Close()
		return err
	}
	fmt.Fprintf(stderr, "cmdproxy listening on %s\n", sockPath)

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go st.handleConn(c)
	}
}

func (st *serveState) handleConn(c net.Conn) {
	defer c.Close()
	dec := json.NewDecoder(c)
	enc := json.NewEncoder(c)
	var req protocol.RunRequest
	if err := dec.Decode(&req); err != nil {
		_ = enc.Encode(protocol.RunResponse{Allow: false, Message: fmt.Sprintf("decode: %v", err)})
		return
	}
	resp := st.decide(&req)
	if err := enc.Encode(resp); err != nil {
		return
	}
}

func (st *serveState) decide(req *protocol.RunRequest) protocol.RunResponse {
	line := CommandLine(req.Argv)
	if err := st.store.Reload(); err != nil {
		fmt.Fprintf(os.Stderr, "cmdproxy: reload rules: %v\n", err)
	}
	if rule, ok := st.store.MatchRule(line); ok {
		if rule.Action == rules.ActionDeny {
			return protocol.RunResponse{Allow: false, Message: "denied by persisted rule"}
		}
		return protocol.RunResponse{Allow: true, MergeEnv: mergeEnvForAllowRule(st.priv, rule)}
	}

	st.mu.Lock()
	defer st.mu.Unlock()

	defaultPat := DefaultPattern(line)
	ch, editedPat, err := PromptCommand(line, defaultPat)
	if err != nil {
		if errors.Is(err, ErrCancelled) {
			return protocol.RunResponse{Allow: false, Message: "cancelled"}
		}
		return protocol.RunResponse{Allow: false, Message: err.Error()}
	}

	pattern := editedPat
	if strings.TrimSpace(pattern) == "" {
		pattern = defaultPat
	}
	pattern = strings.TrimSpace(pattern)
	if err := validatePattern(pattern); err != nil {
		return protocol.RunResponse{Allow: false, Message: fmt.Sprintf("invalid regex: %v", err)}
	}

	switch ch {
	case ChoiceAllowOnce:
		rule := rules.Rule{Pattern: pattern, Action: rules.ActionAllow, AllowEnvKeys: nil}
		return protocol.RunResponse{Allow: true, MergeEnv: mergeEnvForAllowRule(st.priv, rule)}
	case ChoiceDenyOnce:
		return protocol.RunResponse{Allow: false, Message: "denied once"}
	case ChoiceAllowAlways:
		keys, err := PromptAllowEnvKeys()
		if err != nil {
			if errors.Is(err, ErrCancelled) {
				return protocol.RunResponse{Allow: false, Message: "cancelled"}
			}
			return protocol.RunResponse{Allow: false, Message: err.Error()}
		}
		rule := rules.Rule{Pattern: pattern, Action: rules.ActionAllow, AllowEnvKeys: keys}
		if err := st.store.AppendRule(rule); err != nil {
			return protocol.RunResponse{Allow: false, Message: fmt.Sprintf("save rule: %v", err)}
		}
		fmt.Fprintf(os.Stderr, "cmdproxy: saved allow rule %q (env: %s) -> %s\n", pattern, formatAllowEnvKeys(keys), st.rulePath)
		return protocol.RunResponse{Allow: true, MergeEnv: mergeEnvForAllowRule(st.priv, rule)}
	case ChoiceDenyAlways:
		rule := rules.Rule{Pattern: pattern, Action: rules.ActionDeny}
		if err := st.store.AppendRule(rule); err != nil {
			return protocol.RunResponse{Allow: false, Message: fmt.Sprintf("save rule: %v", err)}
		}
		fmt.Fprintf(os.Stderr, "cmdproxy: saved deny rule %q -> %s\n", pattern, st.rulePath)
		return protocol.RunResponse{Allow: false, Message: "denied always"}
	default:
		return protocol.RunResponse{Allow: false, Message: "unexpected choice"}
	}
}

func validatePattern(p string) error {
	if _, err := regexp.Compile(p); err != nil {
		return err
	}
	return nil
}
