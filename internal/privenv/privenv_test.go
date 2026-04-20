package privenv

import (
	"testing"
)

func TestEnvironMap(t *testing.T) {
	t.Setenv("CMDPROXY_TEST_VAR", "abc")
	m := EnvironMap()
	if m["CMDPROXY_TEST_VAR"] != "abc" {
		t.Fatalf("got %q", m["CMDPROXY_TEST_VAR"])
	}
}
