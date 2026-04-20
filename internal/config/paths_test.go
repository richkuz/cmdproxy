package config

import (
	"path/filepath"
	"testing"
)

func TestDirOverride(t *testing.T) {
	t.Setenv("CMDPROXY_CONFIG_DIR", "/tmp/cp-test")
	d, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if d != "/tmp/cp-test" {
		t.Fatalf("got %q", d)
	}
	sp, err := SocketPath()
	if err != nil {
		t.Fatal(err)
	}
	if sp != filepath.Join(d, "cmdproxy.sock") {
		t.Fatalf("socket %q", sp)
	}
	sh, err := ShimDir()
	if err != nil {
		t.Fatal(err)
	}
	if sh != filepath.Join(d, "shims") {
		t.Fatalf("shim dir %q", sh)
	}
}
