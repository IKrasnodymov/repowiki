# repowiki

Go CLI tool that auto-generates Qoder repo wiki on git commits. Works like [Entire CLI](https://entire.io) but for documentation instead of AI session capture.

## Architecture

```
cmd/repowiki/          CLI commands (entry points)
  main.go              Subcommand dispatch: enable|disable|status|generate|update|hooks|logs
  enable.go            Install hook + config + .qoder/commands/update-wiki.md
  disable.go           Remove hook, set enabled=false
  status.go            Print config, hook status, qodercli detection, wiki page count
  generate.go          Full wiki generation via qodercli
  update.go            Incremental update — detects changed files, filters excluded, runs qodercli
  hooks.go             Post-commit callback — 3-layer loop prevention, spawns background process
  logs.go              Show latest generation log

internal/
  config/config.go     Config struct (JSON), Load/Save, defaults — stored in .repowiki/config.json
  git/git.go           Git wrappers: FindRoot, HeadCommit, CommitMessage, ChangedFiles, Stage, Commit
  hook/hook.go         Hook install/uninstall with marker comments, coexists with other hooks
  lockfile/lockfile.go PID-based lock with stale detection (30min timeout)
  wiki/
    wiki.go            Orchestrator: FullGenerate() and IncrementalUpdate()
    qoder.go           qodercli binary detection + invocation (-p, -q, --dangerously-skip-permissions)
    prompt.go          AI prompt construction for full and incremental modes
    detect.go          Changed files → affected wiki sections (metadata reverse index + heuristics)
    commit.go          Auto-commit wiki with [repowiki] prefix + sentinel file
```

## Critical Design: Loop Prevention

Wiki auto-commits could trigger the post-commit hook infinitely. Three layers prevent this:

1. **Sentinel file** `.repowiki/.committing` — created before git commit, checked by hook
2. **Lock file** `.repowiki/.repowiki.lock` — prevents concurrent runs (PID + stale detection)
3. **Commit prefix** `[repowiki]` — hook skips commits starting with this prefix

Flow: hook fires → checks all 3 layers → spawns detached background `repowiki update --from-hook` → qodercli runs → wiki committed with prefix → hook fires again → prefix check → SKIP.

## Key Paths

| Path | Purpose |
|------|---------|
| `.repowiki/config.json` | Project config |
| `.repowiki/logs/` | Daily generation logs |
| `.repowiki/.repowiki.lock` | Process lock (gitignored) |
| `.repowiki/.committing` | Commit sentinel (gitignored) |
| `.qoder/repowiki/en/content/` | Generated wiki markdown |
| `.qoder/repowiki/en/meta/repowiki-metadata.json` | Code snippet index |
| `.qoder/commands/update-wiki.md` | Custom Qoder command for manual use |
| `.git/hooks/post-commit` | Hook with absolute path to repowiki binary |

## Build & Run

```bash
go build -o bin/repowiki ./cmd/repowiki   # build
go install ./cmd/repowiki                  # install to ~/go/bin/
make build                                 # same via Makefile
```

## Commands Quick Reference

```bash
repowiki enable [--force] [--qodercli-path PATH] [--model MODEL] [--no-auto-commit]
repowiki disable
repowiki status
repowiki generate              # full wiki from scratch (slow, ~4min)
repowiki update [--commit HASH]  # incremental (fast, ~3min)
repowiki logs                  # show latest log
repowiki version
```

## Config Format (.repowiki/config.json)

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

`full_generate_threshold` — if more files changed than this, run full generation instead of incremental.

## Dependencies

Zero external dependencies. Stdlib only: `os/exec`, `encoding/json`, `flag`, `path/filepath`, `syscall`.

## qodercli Integration

Binary auto-detected from: config → PATH → `/Applications/Qoder.app/.../qodercli` (macOS).
Invoked as: `qodercli -p "<prompt>" -q -w <dir> --max-turns 50 --dangerously-skip-permissions --allowed-tools Read,Write,Edit,Glob,Grep,Bash`.
Requires authentication: `qodercli /login` or `QODER_PERSONAL_ACCESS_TOKEN` env var.

## Development Notes

- Hook uses absolute binary path (resolved at `enable` time via `os.Executable()`) — git hooks run with minimal PATH
- Background process detached via `syscall.SysProcAttr{Setsid: true}` — won't block user's terminal
- `update --from-hook` flag suppresses stdout for background runs
- Incremental detection: parses `repowiki-metadata.json` for reverse index (source file → wiki pages) + heuristic path matching (e.g. `backend/` → "Backend Architecture")
- Logs written to `.repowiki/logs/YYYY-MM-DD.log` with RFC3339 timestamps
