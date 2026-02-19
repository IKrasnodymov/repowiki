# repowiki

Go CLI tool that auto-generates repo wiki on git commits. Works like [Entire CLI](https://entire.io) but for documentation instead of AI session capture. Supports multiple AI engines: Qoder CLI, Claude Code, OpenAI Codex CLI.

## Architecture

```
cmd/repowiki/          CLI commands (entry points)
  main.go              Subcommand dispatch: enable|disable|status|generate|update|hooks|logs
  enable.go            Install hook + config, --engine flag selects AI backend
  disable.go           Remove hook, set enabled=false
  status.go            Print config, hook status, engine detection, wiki page count
  generate.go          Full wiki generation
  update.go            Incremental update — detects changed files, filters excluded
  hooks.go             Post-commit callback — 3-layer loop prevention, spawns background process
  logs.go              Show latest generation log

internal/
  config/config.go     Config struct (JSON), Load/Save, engine constants (qoder|claude-code|codex)
  git/git.go           Git wrappers: FindRoot, HeadCommit, CommitMessage, ChangedFiles, Stage, Commit
  hook/hook.go         Hook install/uninstall with marker comments, coexists with other hooks
  lockfile/lockfile.go PID-based lock with stale detection (30min timeout)
  wiki/
    engine.go          Multi-engine abstraction: FindEngineBinary(), RunEngine() + per-engine impls
    wiki.go            Orchestrator: FullGenerate() and IncrementalUpdate()
    prompt.go          AI prompt construction for full and incremental modes
    detect.go          Changed files → affected wiki sections (metadata reverse index + heuristics)
    commit.go          Auto-commit wiki with [repowiki] prefix + sentinel file
```

## Engine Support

| Engine | CLI binary | Non-interactive invocation |
|--------|-----------|--------------------------|
| `qoder` | `qodercli` | `-p "prompt" -q -w <dir> --max-turns N --dangerously-skip-permissions` |
| `claude-code` | `claude` | `-p "prompt" --dangerously-skip-permissions --allowedTools ...` |
| `codex` | `codex` | `exec "prompt" --full-auto` |

Binaries auto-detected from: `engine_path` config → `$PATH` → known OS locations.

## Critical Design: Loop Prevention

Wiki auto-commits could trigger the post-commit hook infinitely. Three layers prevent this:

1. **Sentinel file** `.repowiki/.committing` — created before git commit, checked by hook
2. **Lock file** `.repowiki/.repowiki.lock` — prevents concurrent runs (PID + stale detection)
3. **Commit prefix** `[repowiki]` — hook skips commits starting with this prefix

## Key Paths

| Path | Purpose |
|------|---------|
| `.repowiki/config.json` | Project config (engine, model, thresholds) |
| `.repowiki/logs/` | Daily generation logs |
| `.repowiki/.repowiki.lock` | Process lock (gitignored) |
| `.repowiki/.committing` | Commit sentinel (gitignored) |
| `.qoder/repowiki/en/content/` | Generated wiki markdown |
| `.qoder/repowiki/en/meta/repowiki-metadata.json` | Code snippet index |
| `.git/hooks/post-commit` | Hook with absolute path to repowiki binary |

## Build & Run

```bash
go build -o bin/repowiki ./cmd/repowiki   # build
go install ./cmd/repowiki                  # install to ~/go/bin/
make build                                 # same via Makefile
```

## Commands Quick Reference

```bash
repowiki enable [--engine ENGINE] [--engine-path PATH] [--model MODEL] [--force] [--no-auto-commit]
repowiki disable
repowiki status
repowiki generate              # full wiki from scratch (~4min)
repowiki update [--commit HASH]  # incremental (~3min)
repowiki logs
repowiki version
```

## Config Format (.repowiki/config.json)

```json
{
  "enabled": true,
  "engine": "qoder",
  "engine_path": "",
  "model": "",
  "max_turns": 50,
  "language": "en",
  "auto_commit": true,
  "commit_prefix": "[repowiki]",
  "excluded_paths": [".qoder/repowiki/", ".repowiki/", "node_modules/", "vendor/", ".git/"],
  "wiki_path": ".qoder/repowiki",
  "full_generate_threshold": 20
}
```

## Dependencies

Zero external dependencies. Stdlib only: `os/exec`, `encoding/json`, `flag`, `path/filepath`, `syscall`.

## Development Notes

- Hook uses absolute binary path (resolved at `enable` time via `os.Executable()`) — git hooks run with minimal PATH
- Background process detached via `syscall.SysProcAttr{Setsid: true}` — won't block user's terminal
- `update --from-hook` flag suppresses stdout for background runs
- Incremental detection: parses `repowiki-metadata.json` for reverse index + heuristic path matching
- Old configs without `engine` field auto-migrate to `"qoder"` on load
