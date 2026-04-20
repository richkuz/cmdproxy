# cmdproxy

`cmdproxy` is a **macOS-oriented helper** that sits in front of selected CLI tools (via `PATH` shims), asks you whether each invocation is allowed, and—when you allow it—runs the **real** binary with the **`cmdproxy serve` process environment merged in** (so the agent can keep a minimal env while you approve passing through secrets from the daemon).

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

## Pre-built macOS binary

See **`dist/`** — `darwin-arm64` (Apple Silicon) and `darwin-amd64` (Intel). Copy the matching binary onto your `PATH`, then run **`cmdproxy`** (or `cmdproxy init` if you prefer explicit setup).

Regenerate artifacts: `make dist-darwin` (uses Docker; see `Makefile`).

## First-time flow (recommended)

1. **Install the binary** on your `PATH` (e.g. `/usr/local/bin/cmdproxy`).

2. **Start everything** — with no arguments, `cmdproxy` creates the data directory and default shims if needed (`gh` and `git`), then runs the daemon:

   ```bash
   export GITHUB_TOKEN="..."   # example: only in this terminal

   cmdproxy
   ```

   That is equivalent to running `cmdproxy init -y` once (when the shims folder is missing or empty), then `cmdproxy serve`. You can still use **`cmdproxy init`** interactively or **`cmdproxy serve`** alone if you already initialized.

3. **Add shims to your PATH** — the first-time output prints a line like:

   ```bash
   export PATH="$HOME/Library/Application Support/cmdproxy/shims:$PATH"
   ```

   Put that in `~/.zshrc` or `~/.bashrc` (or export it in the terminal where you launch the agent).

4. **Run your agent** from a shell **without** those secrets in the environment.

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
