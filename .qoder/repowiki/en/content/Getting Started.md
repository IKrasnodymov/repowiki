# Getting Started

<cite>
Source files referenced:
- [README.md](file://README.md)
- [cmd/repowiki/enable.go](file://cmd/repowiki/enable.go)
- [cmd/repowiki/main.go](file://cmd/repowiki/main.go)
- [internal/config/config.go](file://internal/config/config.go)
</cite>

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Project Setup](#project-setup)
- [Initial Generation](#initial-generation)
- [Daily Usage](#daily-usage)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)

## Prerequisites

Before using repowiki, ensure you have:

1. **Git repository** with at least one commit
2. **Go 1.22+** installed (for building from source)
3. **One of the supported AI engines** installed and authenticated:
   - **Qoder CLI** (`qodercli`)
   - **Claude Code** (`claude`)
   - **OpenAI Codex CLI** (`codex`)

### Checking Requirements

```bash
# Verify Git
git --version

# Verify Go
go version

# Verify your chosen AI engine
qodercli --version
# or
claude --version
# or
codex --version
```

## Installation

### From Source

```bash
# Install directly
go install github.com/IKrasnodymov/repowiki/cmd/repowiki@latest

# Or clone and build
git clone https://github.com/IKrasnodymov/repowiki.git
cd repowiki
make install
```

Make sure `~/go/bin` is in your PATH:

```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH="$HOME/go/bin:$PATH"
```

### Build Locally

```bash
# Clone repository
git clone https://github.com/ikrasnodymov/repowiki.git
cd repowiki

# Build binary
make build

# Binary will be at: bin/repowiki
```

### Verify Installation

```bash
repowiki version
# Output: repowiki v0.1.0
```

## Project Setup

### Authenticate Your AI Engine (One Time)

Before enabling repowiki, authenticate with your chosen AI engine:

```bash
# Qoder
qodercli /login
# or: export QODER_PERSONAL_ACCESS_TOKEN=<token>

# Claude Code
claude  # follow the auth prompts

# Codex
codex  # follow the auth prompts
# or: export CODEX_API_KEY=<key>
```

### Enable repowiki in Your Project

Navigate to your project directory and run:

```bash
cd /path/to/your/project
repowiki enable
```

This command will **auto-detect** the first available AI engine in the following priority order:
1. `claude-code` (Claude CLI)
2. `qoder` (Qoder CLI)
3. `codex` (OpenAI Codex CLI)

The command also:
1. Creates `.repowiki/config.json` with default settings
2. Installs a post-commit hook in `.git/hooks/post-commit`
3. Creates a custom Qoder command at `.qoder/commands/update-wiki.md`

### Enable Options

```bash
# Enable with auto-detected engine (tries claude-code, qoder, codex in order)
repowiki enable

# Enable with specific engine
repowiki enable --engine claude-code

# Enable with Qoder explicitly
repowiki enable --engine qoder

# Enable with OpenAI Codex
repowiki enable --engine codex

# Enable with specific model
repowiki enable --engine claude-code --model sonnet

# Force reinstall if already enabled
repowiki enable --force

# Specify custom engine binary path
repowiki enable --engine-path /path/to/engine

# Disable auto-commit (generate wiki but don't commit automatically)
repowiki enable --no-auto-commit
```

### Enable Command Flags

| Flag | Description |
|------|-------------|
| `--engine` | AI engine: `qoder`, `claude-code`, or `codex` (auto-detected if not specified) |
| `--engine-path` | Custom path to the engine CLI binary |
| `--model` | Engine-specific model (e.g., `sonnet` for Claude, `performance` for Qoder) |
| `--force` | Reinstall hook even if already present |
| `--no-auto-commit` | Generate wiki but don't auto-commit changes |

### Engine Auto-Detection

When you run `repowiki enable` without specifying an engine, repowiki attempts to find an available AI engine automatically in this priority order:

1. `claude-code` (Claude Code by Anthropic) — most commonly available
2. `qoder` (Qoder CLI)
3. `codex` (OpenAI Codex CLI)

The first successfully detected engine will be used. If you explicitly specify an engine with `--engine` or `--engine-path`, auto-detection is skipped and the command will fail if that specific engine is not found.

Example output with auto-detection:
```
Auto-detected engine: claude-code (/usr/local/bin/claude)

repowiki enabled in /path/to/project

  Engine:  claude-code
  Binary:  /usr/local/bin/claude
  Config:  .repowiki/config.json
  Hook:    .git/hooks/post-commit

Every commit will now auto-update the repo wiki.
Run 'repowiki generate' for initial full wiki generation.
```

### What Happens During Enable

```go
// From cmd/repowiki/enable.go
func handleEnable(args []string) {
    // 1. Find git root
    gitRoot, err := git.FindRoot()

    // 2. Load or create config
    cfg, err := config.Load(gitRoot)
    if err != nil {
        cfg = config.Default()
    }

    // 3. Apply flag overrides
    engineExplicit := *engine != ""
    if engineExplicit {
        if !config.IsValidEngine(*engine) {
            fmt.Fprintf(os.Stderr, "Error: unknown engine %q\n", *engine)
            os.Exit(1)
        }
        cfg.Engine = *engine
    }
    if *enginePath != "" {
        cfg.EnginePath = *enginePath
    }
    cfg.Enabled = true

    // 4. Validate engine binary (with auto-detection fallback)
    binPath, findErr := wiki.FindEngineBinary(cfg)
    if findErr != nil {
        if engineExplicit || *enginePath != "" {
            // Explicit engine not found — fail hard
            fmt.Fprintf(os.Stderr, "Error: %v\n", findErr)
            os.Exit(1)
        }
        // Auto-detect first available engine
        for _, eng := range config.EngineDetectOrder {
            cfg.Engine = eng
            cfg.EnginePath = ""
            binPath, findErr = wiki.FindEngineBinary(cfg)
            if findErr == nil {
                break
            }
        }
    }

    // 5. Save config
    config.Save(gitRoot, cfg)

    // 6. Install git hook
    hook.Install(gitRoot, *force, selfPath)

    // 7. Create Qoder command
    createQoderCommand(gitRoot)
}
```

## Initial Generation

After enabling, generate the initial wiki:

```bash
repowiki generate
```

This will:
1. Analyze your entire codebase
2. Generate comprehensive documentation in `.qoder/repowiki/en/content/`
3. Create metadata file at `.qoder/repowiki/en/meta/repowiki-metadata.json`
4. Auto-commit changes (if enabled)

**Note**: Initial generation takes **3-5 minutes** depending on codebase size.

## Daily Usage

Once enabled, repowiki operates automatically:

### Automatic Mode

```bash
# Make changes and commit as usual
git add .
git commit -m "Add new feature"

# repowiki automatically updates in the background
```

### Manual Commands

```bash
# Check status
repowiki status

# Force full regeneration
repowiki generate

# Update for specific commit
repowiki update --commit abc123

# View latest generation log
repowiki logs

# Disable repowiki
repowiki disable
```

## Configuration

### Configuration File

Located at `.repowiki/config.json`:

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
  "full_generate_threshold": 20
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | `true` | Whether repowiki is active |
| `engine` | string | `"qoder"` | AI engine: `qoder`, `claude-code`, `codex` |
| `engine_path` | string | `""` | Override path to engine CLI binary (auto-detected if empty) |
| `model` | string | `""` | Engine-specific model (e.g., `sonnet` for Claude, `performance` for Qoder) |
| `max_turns` | integer | `50` | Maximum AI interaction turns |
| `language` | string | `"en"` | Wiki language code |
| `auto_commit` | boolean | `true` | Auto-commit wiki changes |
| `commit_prefix` | string | `"[repowiki]"` | Prefix for wiki commits |
| `excluded_paths` | array | `[...]` | Paths to ignore when detecting changes |
| `wiki_path` | string | `".qoder/repowiki"` | Output directory for wiki |
| `full_generate_threshold` | integer | `20` | File change threshold for full regeneration |

## Troubleshooting

### Common Issues

#### "not a git repository"

```
Error: not a git repository
```

**Solution**: Run `git init` or navigate to a git repository.

#### "repowiki not configured"

```
Error: repowiki not configured. Run 'repowiki enable' first.
```

**Solution**: Run `repowiki enable` to initialize configuration.

#### "engine binary not found"

```
Warning: qodercli not found
```

**Solution**: Install the AI engine or specify path:
```bash
repowiki enable --engine-path /Applications/Qoder.app/.../qodercli
```

#### "another repowiki process is running"

```
Error: another repowiki process is running
```

**Solution**: Wait for the other process to complete, or manually remove `.repowiki/.repowiki.lock` if the process crashed.

### Checking Status

```bash
repowiki status
```

Example output:
```
repowiki v0.1.0

  Status:       enabled
  Engine:       qoder
  Hook:         installed (.git/hooks/post-commit)
  Binary:       /usr/local/bin/qodercli
  Wiki path:    .qoder/repowiki/en/content/ (12 pages)
  Model:        auto
  Auto-commit:  true
  Max turns:    50
  Last run:     2026-02-19T15:30:00Z
  Last commit:  abc123def456
```

### Viewing Logs

```bash
# View the most recent log
repowiki logs

# View all logs in the directory
ls -la .repowiki/logs/
```

### Disabling

To stop automatic wiki generation:

```bash
repowiki disable
```

This removes the git hook but preserves all wiki files.
