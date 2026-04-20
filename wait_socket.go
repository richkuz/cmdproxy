package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"cmdproxy/internal/config"
)

func effectiveSocketPath() (string, error) {
	if p := os.Getenv("CMDPROXY_SOCKET"); p != "" {
		return p, nil
	}
	return config.SocketPath()
}

func runWaitSocket(stderr io.Writer, timeout time.Duration, logOnFail string) error {
	sock, err := effectiveSocketPath()
	if err != nil {
		return err
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		fi, err := os.Stat(sock)
		if err == nil && fi.Mode().Type() == fs.ModeSocket {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	if logOnFail != "" {
		if data, rerr := os.ReadFile(logOnFail); rerr == nil && len(data) > 0 {
			fmt.Fprintf(stderr, "--- %s ---\n", logOnFail)
			_, _ = stderr.Write(data)
			if data[len(data)-1] != '\n' {
				fmt.Fprintln(stderr)
			}
		}
	}
	return fmt.Errorf("timed out after %v waiting for socket %s", timeout, sock)
}
