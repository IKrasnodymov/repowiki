# Configuration Management

<cite>
Source files referenced:
- [internal/config/config.go](file://internal/config/config.go)
- [.repowiki/config.json](file://.repowiki/config.json)
- [.gitignore](file://.gitignore)
</cite>

## Table of Contents

- [Configuration File](#configuration-file)
- [Configuration Options](#configuration-options)
- [Environment Variables](#environment-variables)
- [File Locations](#file-locations)
- [Git Ignore](#git-ignore)
- [Manual Configuration](#manual-configuration)

## Configuration File

repowiki stores its configuration in `.repowiki/config.json`. This file is created automatically when you run `repowiki enable`.

### Example Configuration

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
  "excluded_paths": [
    ".qoder/repowiki/",
    ".repowiki/",
    "node_modules/",
    "vendor/",
    ".git/"
  ],
  "wiki_path": ".qoder/repowiki",
  "full_generate_threshold": 20,
  "last_run": "2026-02-19T15:30:00Z",
  "last_commit_hash": "abc123def456"
}
```

## Configuration Options

### Core Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | `true` | Master switch for repowiki functionality |
| `engine` | string | `"qoder"` | AI engine: `qoder`, `claude-code`, or `codex` |
| `engine_path` | string | `""` | Override path to engine CLI binary |
| `model` | string | `""` | Engine-specific model for wiki generation |
| `max_turns` | integer | `50` | Maximum AI interaction turns per generation |
| `language` | string | `"en"` | Language code for generated wiki |

### Supported Engines

| Engine | CLI Binary | Description |
|--------|-----------|-------------|
| `qoder` | `qodercli` | Qoder CLI (default) |
| `claude-code` | `claude` | Claude Code by Anthropic |
| `codex` | `codex` | OpenAI Codex CLI |

### Engine Auto-Detection

When running `repowiki enable` without explicitly specifying an engine, repowiki attempts to auto-detect an available engine. The detection order is:

1. `claude-code` (Claude Code by Anthropic)
2. `qoder` (Qoder CLI)
3. `codex` (OpenAI Codex CLI)

If the default engine (`qoder`) is not found, repowiki iterates through the detection order and uses the first available engine. This behavior is controlled by the `EngineDetectOrder` slice in the config package.

If an engine is explicitly specified via `--engine` or `--engine-path` but not found, the command fails with an error rather than attempting auto-detection.

### Model Selection

Model values are engine-specific:

**Qoder CLI:**
- `""` (empty) - Use Qoder default
- `efficient` - Faster, less detailed generation
- `performance` - Balanced speed and quality
- `ultimate` - Highest quality, slower generation

**Claude Code:**
- `""` (empty) - Use Claude default
- `sonnet` - Claude Sonnet model
- `opus` - Claude Opus model
- `haiku` - Claude Haiku model

### Commit Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `auto_commit` | boolean | `true` | Automatically commit wiki changes |
| `commit_prefix` | string | `"[repowiki]"` | Prefix for wiki commit messages |

When `auto_commit` is enabled, wiki changes are committed automatically with messages like:
```
[repowiki] full wiki generation
[repowiki] update wiki for 5 changed files
```

To disable auto-commit during setup:
```bash
repowiki enable --no-auto-commit
```

When auto-commit is disabled, the wiki files are still generated but you must commit them manually.

### Path Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `wiki_path` | string | `".qoder/repowiki"` | Output directory for wiki files |
| `excluded_paths` | array | `[...]` | Paths to exclude from change detection |

### Default Excluded Paths

```json
[
  ".qoder/repowiki/",
  ".repowiki/",
  "node_modules/",
  "vendor/",
  ".git/"
]
```

These paths are excluded because:
- `.qoder/repowiki/` — Wiki output (prevents self-triggering)
- `.repowiki/` — Configuration and logs
- `node_modules/` — Node.js dependencies
- `vendor/` — Go vendor directory
- `.git/` — Git internals

### Threshold Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `full_generate_threshold` | integer | `20` | File change threshold for full regeneration |

When more than this many files change, a full regeneration is performed instead of an incremental update.

### Runtime Tracking

| Option | Type | Description |
|--------|------|-------------|
| `last_run` | string (ISO 8601) | Timestamp of last wiki generation |
| `last_commit_hash` | string | Git hash of last processed commit |

These fields are updated automatically after each wiki generation and should not be modified manually.

## Environment Variables

repowiki does not currently use environment variables for configuration. All settings are managed through the configuration file.

However, the following environment variables affect operation:

| Variable | Effect |
|----------|--------|
| `PATH` | Used to locate `git` and `qodercli` binaries |
| `HOME` | Used for resolving paths in some Git operations |

## File Locations

### Configuration Directory

```
.repowiki/
├── config.json          # Main configuration
├── .repowiki.lock       # Process lock file
├── .committing          # Sentinel file (loop prevention)
└── logs/
    ├── 2026-02-19.log   # Daily execution logs
    └── hook.log         # Hook execution log
```

### Wiki Output Directory

```
.qoder/repowiki/
└── en/
    ├── content/         # Wiki markdown files
    │   ├── System Overview.md
    │   ├── Technology Stack.md
    │   ├── Getting Started.md
    │   ├── Backend Architecture/
    │   ├── Frontend Architecture/
    │   ├── Core Features/
    │   ├── API Reference/
    │   └── Configuration Management.md
    └── meta/
        └── repowiki-metadata.json
```

### Path Resolution

```go
// internal/config/config.go
func Dir(gitRoot string) string {
    return filepath.Join(gitRoot, ConfigDir)  // .repowiki
}

func Path(gitRoot string) string {
    return filepath.Join(Dir(gitRoot), ConfigFile)  // .repowiki/config.json
}

func LogPath(gitRoot string) string {
    return filepath.Join(Dir(gitRoot), LogDir)  // .repowiki/logs
}
```

## Git Ignore

The `.gitignore` file should exclude repowiki runtime files:

```gitignore
# repowiki runtime files (keep config.json)
.repowiki/.repowiki.lock
.repowiki/.committing
.repowiki/logs/
```

**Note**: `config.json` should NOT be ignored as it contains project-specific settings that should be shared.

### Recommended .gitignore

```gitignore
# Binaries
bin/
*.exe

# repowiki runtime files
.repowiki/.repowiki.lock
.repowiki/.committing
.repowiki/logs/
```

## Manual Configuration

You can edit `.repowiki/config.json` directly to change settings:

### Change Model Level

```json
{
  "model": "performance"
}
```

### Disable Auto-Commit

```json
{
  "auto_commit": false
}
```

### Add Custom Excluded Paths

```json
{
  "excluded_paths": [
    ".qoder/repowiki/",
    ".repowiki/",
    "node_modules/",
    "vendor/",
    ".git/",
    "generated/",
    "*.pb.go"
  ]
}
```

### Specify Engine Path

```json
{
  "engine_path": "/Applications/Qoder.app/Contents/Resources/app/resources/bin/aarch64_darwin/qodercli"
}
```

Or for Claude Code:

```json
{
  "engine": "claude-code",
  "engine_path": "/usr/local/bin/claude"
}
```

### Configuration Validation

After manual edits, verify the configuration:

```bash
repowiki status
```

Invalid JSON will cause repowiki to fail with an error:
```
Error: failed to parse config: invalid character ',' looking for beginning of object key string
```

### Reset to Defaults

To reset configuration to defaults:

```bash
# Disable and re-enable
repowiki disable
repowiki enable
```

Or manually delete `.repowiki/config.json` and run `repowiki enable`.
