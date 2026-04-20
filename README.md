# cmdproxy

`cmdproxy` is a **macOS-oriented helper** that sits in front of selected CLI tools (via `PATH` shims), asks you whether each invocation is allowed, and—when you allow it—runs the **real** binary with the **`cmdproxy serve` process environment merged in** (so the agent can keep a minimal env while you approve passing through secrets from the daemon).

## Install

Download a pre-built binary from [GitHub Releases](https://github.com/richkuz/cmdproxy/releases), then put it on your `PATH` (example: `/usr/local/bin`).

**Apple Silicon (M1 / M2 / M3 / …):**

```bash
curl -fsSL -o cmdproxy "https://github.com/richkuz/cmdproxy/releases/latest/download/cmdproxy-darwin-arm64"
chmod +x cmdproxy
sudo mv cmdproxy /usr/local/bin/cmdproxy
```

**Intel Mac:**

```bash
curl -fsSL -o cmdproxy "https://github.com/richkuz/cmdproxy/releases/latest/download/cmdproxy-darwin-amd64"
chmod +x cmdproxy
sudo mv cmdproxy /usr/local/bin/cmdproxy
```

If `releases/latest/download/...` returns 404, there may not be a published release yet—use **Install from source** below, or copy a matching binary from `dist/` after running `make dist-darwin` locally.

## Usage

```bash
# One-time interactive initialization and configuration
cmdproxy

# Usage:
export GITHUB_TOKEN=... # Or any privileged ENV vars
cmdproxy &

unset GITHUB_TOKEN
claude
```

The first `cmdproxy` run performs first-time setup and keeps the daemon in the **foreground**; press Ctrl+C when you are finished, then use the `cmdproxy &` pattern for regular agent sessions. For menus to edit shims and rules without starting the daemon, run **`cmdproxy config`**.

When Claude Code (or any tool) runs `gh` or another **shimmed** command on your `PATH`, `cmdproxy` shows a **macOS dialog** (Allow / Deny, once or always). Use **`cmdproxy config`** (or edit `rules.json`) to choose which commands are shimmed and which environment variable names are merged for each **allow** rule—so privileged env vars apply only to commands you allow.

After the first run, add the shims directory to `PATH` (the tool prints the exact line), typically:

```bash
export PATH="$HOME/Library/Application Support/cmdproxy/shims:$PATH"
```

Put that in `~/.zshrc` / `~/.bashrc`, or export it in the terminal where you launch the agent.

## Why this exists

Tools like Cursor CLI Agent or other automation often need to run commands (`gh`, `git`, `curl`, …). If you put secrets in that process’s environment, the agent (and anything it runs) can read them. `cmdproxy` splits the world in two:

1. **Agent process**: minimal environment (no GitHub token, etc.).
2. **`cmdproxy serve` process**: environment you control; when you **allow** a shimmed command, matching keys from this process are merged into the child. **Allow** rules can list specific env var names, or default to **all** keys from the serve process (`allow_env_keys` in `rules.json`).

When the agent runs a **shimmed** command name, the shim talks to `cmdproxy serve`, shows a **modal `osascript` dialog** (Allow / Deny, once or always), and only then merges the serve environment into the child process before `exec`.

This does **not** replace full OS sandboxing; it is a deliberate, human-in-the-loop gate plus an environment merge.

## Limitations (read this)

- **Not a global interceptor.** Only commands resolved through **shim names on `PATH`** are gated. If something runs `/opt/homebrew/bin/gh` by **absolute path**, your `gh` shim is **not** used and `cmdproxy` never sees the call.
- **Trust model**: anyone who can execute your shims and reach the Unix socket can request merges while `serve` is running. Protect your machine user session and socket permissions.
- **UI**: `osascript` dialogs are blocking and utilitarian; they may steal focus.
- **Non-macOS**: the daemon starts, but interactive prompts require macOS (`osascript`).

## Default layout (no env vars required)

On macOS, everything lives under **one directory**:

`~/Library/Application Support/cmdproxy/`

| Item | Path |
|------|------|
| Rules | `…/cmdproxy/rules.json` |
| Unix socket | `…/cmdproxy/cmdproxy.sock` |
| Shims (after `init`) | `…/cmdproxy/shims/` (`gh`, `git`, … symlinks) |

Running **`cmdproxy` with no arguments** auto-runs first-time setup (if the shims directory is missing or empty—equivalent to `cmdproxy init -y` with default tools `gh,git`), then starts the daemon—same end state as `cmdproxy serve`.

`cmdproxy serve` and `cmdproxy init` print these paths. Shims and the daemon use the **same** defaults, so you normally do **not** set `CMDPROXY_SOCKET` or `CMDPROXY_SHIM_DIR`.

Optional overrides: `CMDPROXY_CONFIG_DIR` (moves data + default socket + default shims root), or individual `CMDPROXY_SOCKET` / `CMDPROXY_SHIM_DIR` if you use a custom layout.

## Install from source

```bash
go build -o cmdproxy .
sudo mv cmdproxy /usr/local/bin/   # optional
cmdproxy   # auto-inits if needed, then serves
```

Pre-built artifacts for maintainers live under **`dist/`** — `darwin-arm64` and `darwin-amd64`. Regenerate: `make dist-darwin` (uses Docker; see `Makefile`).

## Rules (“Always”) and env keys

Persisted rules live in `rules.json`. The first matching regex wins. **Deny** rules block the command. **Allow** rules include:

- `pattern`: regex (Go `regexp` / RE2).
- `allow_env_keys`: optional. **Omitted or `null`** means merge **all** env vars from the serve process. A **JSON array** (possibly empty) means merge **only** those names (empty array = merge nothing from serve).

Use **`cmdproxy config`** for an interactive session to edit shims and rules (including env keys) without hand-editing JSON.

## Example launcher

`scripts/agent-start.example.sh` uses **`env -i`** so the daemon only sees what you pass (e.g. `GITHUB_TOKEN`); the agent runs without it. On allow, the merged env follows the matching **allow** rule (default: all keys from serve).

## Environment reference

| Variable | When |
|----------|------|
| `CMDPROXY_CONFIG_DIR` | Optional — override the data directory (socket, rules, default shims path). |
| `CMDPROXY_SOCKET` | Optional — override socket path (advanced). |
| `CMDPROXY_SHIM_DIR` | Optional — override shim directory (advanced). |

**Print default shims path for scripts:** `cmdproxy shim-path` (stdout only). Example: `export PATH="$(cmdproxy shim-path):$PATH"`

**Wait for the daemon after starting it in the background:** `cmdproxy wait-socket [-timeout 5s] [-log /path/to/log]` (exits 0 when the socket exists; on timeout, optional `-log` dumps that file to stderr).

**Interactive config:** `cmdproxy config` — manage shims and allow/deny rules (stdin menus). Prefer stopping **`cmdproxy serve`** while editing rules to avoid concurrent writes.

## Development

- `make` — test and build (requires local Go).
- `make docker-test` — format, vet, test in Linux container.
- `make docker-build` — Linux binary in repo root (for CI; **not** for running on macOS).
- `make dist-darwin` — cross-compile macOS binaries into `dist/`.

## License

Use at your own risk; this is a sharp tool around process execution and secrets.
