package privenv

import (
	"os"
	"strings"
)

// EnvironMap returns the current process environment as a map (from os.Environ).
// When a shimmed command is allowed, these key/value pairs are merged into the
// child (serve’s values win on conflict).
func EnvironMap() map[string]string {
	out := make(map[string]string)
	for _, line := range os.Environ() {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[k] = v
	}
	return out
}
