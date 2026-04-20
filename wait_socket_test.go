package main

import (
	"bytes"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunWaitSocket_OK(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CMDPROXY_CONFIG_DIR", dir)
	t.Setenv("CMDPROXY_SOCKET", "")

	sockPath := filepath.Join(dir, "cmdproxy.sock")
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	if err := runWaitSocket(io.Discard, time.Second, ""); err != nil {
		t.Fatal(err)
	}
}

func TestRunWaitSocket_Timeout(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CMDPROXY_CONFIG_DIR", dir)
	err := runWaitSocket(io.Discard, 150*time.Millisecond, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunWaitSocket_LogOnFail(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CMDPROXY_CONFIG_DIR", dir)
	logf := filepath.Join(dir, "log.txt")
	if err := os.WriteFile(logf, []byte("daemon output\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := runWaitSocket(&buf, 50*time.Millisecond, logf)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(buf.String(), "daemon output") {
		t.Fatalf("expected log in stderr, got %q", buf.String())
	}
}
