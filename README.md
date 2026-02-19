# repowiki

Auto-generate [Qoder](https://qoder.com) repo wiki on every git commit.

Inspired by [Entire CLI](https://entire.io) — but instead of capturing AI sessions, repowiki keeps your project documentation in sync with your code automatically.

## What It Does

Every time you `git commit`, repowiki:

1. Detects which files changed
2. Determines which wiki sections are affected
3. Runs [Qoder CLI](https://qoder.com/cli) in the background to update documentation
4. Auto-commits the updated wiki as a separate `[repowiki]` commit

You write code. Documentation writes itself.

```
$ git log --oneline
d46096d docs: add CLAUDE.md                          # your commit
35540c8 [repowiki] update wiki for 5 changed files   # auto-generated
76ab77b fix: use absolute binary path                 # your commit
aeaf37f [repowiki] full wiki generation               # auto-generated
0ed5fe3 initial: repowiki v0.1.0 CLI tool             # your commit
```

## Requirements

- **Go 1.22+** (for building from source)
- **Git** repository with at least one commit
- **Qoder** IDE or CLI installed — [qoder.com](https://qoder.com)
- **Qoder account** — free or Pro (needed for `qodercli` authentication)

## Installation

### From source (recommended)

```bash
go install github.com/IKrasnodymov/repowiki/cmd/repowiki@latest
```

Make sure `~/go/bin` is in your PATH:

```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH="$HOME/go/bin:$PATH"
```

### Build locally

```bash
git clone https://github.com/IKrasnodymov/repowiki.git
cd repowiki
make install
```

## Quick Start

### 1. Authenticate Qoder CLI (one time)

```bash
qodercli /login
```

If `qodercli` is not on PATH (common on macOS), use the full path:

```bash
/Applications/Qoder.app/Contents/Resources/app/resources/bin/aarch64_darwin/qodercli /login
```

Or set a token:

```bash
export QODER_PERSONAL_ACCESS_TOKEN=<your-token>
```

Get your token at [qoder.com/account/integrations](https://qoder.com/account/integrations).

### 2. Enable in your project

```bash
cd /path/to/your/project
repowiki enable
```

This creates:
- `.repowiki/config.json` — configuration
- `.git/hooks/post-commit` — git hook (appended, won't break existing hooks)
- `.qoder/commands/update-wiki.md` — custom Qoder command for manual use

### 3. Generate wiki for the first time

```bash
repowiki generate
```

This takes **3-5 minutes** depending on project size. Qoder CLI analyzes the entire codebase and generates structured documentation in `.qoder/repowiki/`.

### 4. Done. Just keep coding.

Every commit now auto-updates the wiki in the background. The generation runs as a detached process — your terminal is never blocked.

## Commands

```bash
repowiki enable      # Install hook and configure (run once per project)
repowiki disable     # Remove hook (wiki files preserved)
repowiki status      # Show current config, hook status, wiki stats
repowiki generate    # Full wiki generation from scratch
repowiki update      # Incremental update for recent changes
repowiki logs        # View latest generation log
repowiki version     # Show version
```

### Flags

```bash
# enable
repowiki enable --force                    # Reinstall hook
repowiki enable --qodercli-path /path/to   # Custom qodercli path
repowiki enable --model performance        # Qoder model level
repowiki enable --no-auto-commit           # Generate but don't auto-commit

# update
repowiki update --commit abc123            # Update for specific commit
```

## Generated Wiki Structure

```
.qoder/repowiki/
  en/
    content/
      System Overview.md
      Technology Stack.md
      Getting Started.md
      Backend Architecture/
        Backend Architecture.md
        ...
      Frontend Architecture/
        ...
      Core Features/
        ...
      API Reference/
        ...
    meta/
      repowiki-metadata.json    # code snippet index
```

Each wiki page includes:
- Referenced source files with links
- Table of contents
- Mermaid architecture diagrams
- Code examples from actual source

Team members access the wiki via `git pull` — no extra setup needed.

## Configuration

Config is stored in `.repowiki/config.json` (auto-created by `enable`):

```json
{
  "enabled": true,
  "qodercli_path": "qodercli",
  "model": "auto",
  "max_turns": 50,
  "language": "en",
  "auto_commit": true,
  "commit_prefix": "[repowiki]",
  "excluded_paths": [".qoder/repowiki/", ".repowiki/", "node_modules/", "vendor/", ".git/"],
  "wiki_path": ".qoder/repowiki",
  "full_generate_threshold": 20
}
```

| Option | Default | Description |
|--------|---------|-------------|
| `qodercli_path` | `"qodercli"` | Path to qodercli binary. Auto-detected on macOS. |
| `model` | `"auto"` | Qoder model: `auto`, `efficient`, `performance`, `ultimate` |
| `max_turns` | `50` | Max agent iterations per generation |
| `language` | `"en"` | Wiki language (`en`, `zh`) |
| `auto_commit` | `true` | Auto-commit wiki changes after generation |
| `commit_prefix` | `"[repowiki]"` | Prefix for wiki commits (also used for loop prevention) |
| `excluded_paths` | `[...]` | Paths ignored during change detection |
| `full_generate_threshold` | `20` | If more than N files changed, run full generation instead of incremental |

## How It Works Internally

### Incremental vs Full Generation

- **< 20 files changed** (configurable) → incremental: only affected wiki sections are updated
- **> 20 files changed** or **no wiki exists yet** → full generation from scratch

### Change Detection

1. Parse `repowiki-metadata.json` to build a reverse index: source file → wiki pages that reference it
2. Heuristic path matching (e.g., files in `backend/` → "Backend Architecture" section)
3. Combine both to determine which wiki sections need updating

### Loop Prevention

Wiki auto-commits trigger the post-commit hook again. Three layers prevent infinite loops:

1. **Sentinel file** — `.repowiki/.committing` is created before the wiki commit and checked first by the hook
2. **Lock file** — `.repowiki/.repowiki.lock` with PID prevents concurrent runs (stale after 30 min)
3. **Commit prefix** — commits starting with `[repowiki]` are skipped by the hook

### Hook Coexistence

The hook is injected between marker comments and appended to existing `post-commit` file — it won't break hooks from Entire, Husky, or other tools:

```sh
#!/bin/sh
# ... existing hooks preserved ...

# repowiki hook start
REPOWIKI_BIN="/Users/you/go/bin/repowiki"
if [ -x "$REPOWIKI_BIN" ]; then
  "$REPOWIKI_BIN" hooks post-commit &
fi
# repowiki hook end
```

## Troubleshooting

### qodercli not found

repowiki auto-detects qodercli from: config path → `$PATH` → known macOS locations. If detection fails:

```bash
repowiki enable --qodercli-path /path/to/qodercli
```

### Wiki not updating after commits

1. Check `repowiki status` — is it enabled?
2. Check `repowiki logs` — any errors?
3. Verify qodercli auth: `qodercli status`
4. Check if `.git/hooks/post-commit` contains the repowiki block

### Stuck lock file

If a previous generation crashed, the lock file may persist:

```bash
rm .repowiki/.repowiki.lock
```

The lock auto-clears after 30 minutes if the owning process is no longer running.
