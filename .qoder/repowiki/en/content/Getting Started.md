# Getting Started

<cite>
Source files referenced:
- [README.md](/to/README.md)
- [cmd/repowiki/enable.go](/to/cmd/repowiki/enable.go)
- [cmd/repowiki/main.go](/to/cmd/repowiki/main.go)
- [internal/config/config.go](/to/internal/config/config.go)
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
3. **Qoder CLI** (`qodercli`) installed and accessible
4. **Qoder account** — free or Pro (needed for `qodercli` authentication)

### Checking Requirements

```bash
# Verify Git
git --version

# Verify Go
go version

# Verify Qoder CLI
qodercli --version
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

### Enable repowiki in Your Project

Navigate to your project directory and run:

```bash
cd /path/to/your/project
repowiki enable
```

This command:
1. Creates `.repowiki/config.json` with default settings
2. Installs a post-commit hook in `.git/hooks/post-commit`
3. Creates a custom Qoder command at `.qoder/commands/update-wiki.md`

### Enable Options

```bash
# Force reinstall if already enabled
repowiki enable --force

# Specify Qoder CLI path
repowiki enable --qodercli-path /path/to/qodercli

# Set model level
repowiki enable --model performance

# Disable auto-commit
repowiki enable --no-auto-commit
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
    if *qoderPath != "" {
        cfg.QoderCLIPath = *qoderPath
    }
    cfg.Enabled = true
    
    // 4. Validate qodercli
    _, findErr := wiki.FindQoderCLI(&testCfg)
    
    // 5. Save config
    config.Save(gitRoot, cfg)
    
    // 6. Install git hook
    hook.Install(gitRoot, *force)
    
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

**Note**: Initial generation may take several minutes depending on codebase size.

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
  "qodercli_path": "qodercli",
  "model": "auto",
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
| `qodercli_path` | string | `"qodercli"` | Path to Qoder CLI binary |
| `model` | string | `"auto"` | AI model level (auto, efficient, performance, ultimate) |
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

#### "qodercli not found"

```
Warning: qodercli not found
```

**Solution**: Install Qoder or specify path:
```bash
repowiki enable --qodercli-path /Applications/Qoder.app/.../qodercli
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
  Hook:         installed (.git/hooks/post-commit)
  Qoder CLI:    /Applications/Qoder.app/.../qodercli
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
