# cmdproxy

NOTE: I've abandoned this. It's too cumbersome to add shims for `gh` such that they are reliably invoked by `cursor-agent` and `claude`.

Just stick with the normal allowlist/denylist in AI harnesses. Or use properly scoped read-only GitHub tokens.

# OLD

`cmdproxy` is a **macOS** tool to let allow-list privileged ENV vars for commands that Claude Code and other AI agents run.

## Quick Start

```bash
# Install (Apple Silicon)
curl -fsSL -o cmdproxy "https://github.com/richkuz/cmdproxy/releases/latest/download/cmdproxy-darwin-arm64"
chmod +x cmdproxy
sudo mv cmdproxy /usr/local/bin/cmdproxy

# One-time initialization
cmdproxy init -y

# Usage:
export GITHUB_TOKEN=... # Or any privileged ENV vars
cmdproxy &

unset GITHUB_TOKEN
claude
```

Ask claude to read a PR or push a PR using `gh`.

`cmdproxy` will intercept the call and allow you to run the command on your behalf with privileged ENV vars that the AI agent cannot see.



## Install

Install the **latest release** from [GitHub Releases](https://github.com/richkuz/cmdproxy/releases) with `curl`, then put the binary on your `PATH` (here: `/usr/local/bin`). Assets are named `cmdproxy-darwin-arm64` and `cmdproxy-darwin-amd64`.

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

If `curl` prints **404**, there is no published release yet—use **Build from source** below, or a maintainer can [publish a release](#publishing-a-release).

### Build from source (requires [Go](https://go.dev/dl/) 1.22+)

```bash
git clone https://github.com/richkuz/cmdproxy.git
cd cmdproxy
go build -o cmdproxy .
chmod +x cmdproxy
sudo mv cmdproxy /usr/local/bin/cmdproxy
```

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

**Claude Code (and similar)** often use a shell environment that **does not** keep your shims first on `PATH`, so `command -v gh` may still show `/opt/homebrew/bin/gh`. Run once (after `cmdproxy init`):

```bash
sudo cmdproxy link-shims
```

That symlinks shims into `/usr/local/bin`, which usually appears **before** `/opt/homebrew/bin` in the default macOS `PATH` order, so `gh` resolves to cmdproxy.

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

## Install from source (same repo, no clone)

If you already have the repository checked out:

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

### Publishing a release

Pushing a **version tag** whose name starts with `v` runs [`.github/workflows/release.yml`](.github/workflows/release.yml). That workflow cross-compiles macOS binaries and uploads **`cmdproxy-darwin-arm64`** and **`cmdproxy-darwin-amd64`** to a [GitHub Release](https://github.com/richkuz/cmdproxy/releases) for that tag. After the workflow succeeds, the `curl` install URLs under [Install](#install) resolve to the new build.

1. Merge your changes to `main` and update your local clone (`git pull`) so the commit you tag is the one you intend to ship.
2. Choose the next version name (e.g. `v0.1.0`, then later `v0.1.1`). Tags must be unique; you cannot reuse a version.
3. Create the tag at the current commit and **push the tag** (pushing `main` alone does not publish binaries):

```bash
git tag v0.1.1
git push origin v0.1.1
```

4. On GitHub, open the **Actions** tab and confirm the **Release** workflow completed successfully. The new release should list both darwin assets; **`releases/latest/download/...`** then points at the newest tag.

If you tagged the wrong commit, remove the remote tag with `git push origin :refs/tags/v0.1.1`, delete it locally with `git tag -d v0.1.1`, then tag again and push.

## License

Use at your own risk; this is a sharp tool around process execution and secrets.
