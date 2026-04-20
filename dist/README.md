# Pre-built macOS binaries

These are cross-compiled **darwin** binaries (pure Go, `CGO_ENABLED=0`). Copy the one that matches your Mac into a directory on your `PATH` (for example `/usr/local/bin`).

| Artifact | Use on |
|----------|--------|
| `darwin-arm64/cmdproxy` | Apple Silicon (M1 / M2 / M3 / …) |
| `darwin-amd64/cmdproxy` | Intel Macs |

After installing:

```bash
chmod +x /path/to/cmdproxy
sudo mv /path/to/cmdproxy /usr/local/bin/cmdproxy   # optional
cmdproxy   # auto-inits if needed, then runs the daemon (see project README)
```

Rebuild locally (any OS with Go 1.22+):

```bash
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o cmdproxy .
```

Regenerate both artifacts with `make dist-darwin` from the repository root (uses Docker if you do not have Go installed).
