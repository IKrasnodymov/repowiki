# Wiki Engine

<cite>
Source files referenced:
- [internal/wiki/wiki.go](file://internal/wiki/wiki.go)
- [internal/wiki/engine.go](file://internal/wiki/engine.go)
- [internal/wiki/commit.go](file://internal/wiki/commit.go)
- [internal/wiki/detect.go](file://internal/wiki/detect.go)
- [internal/wiki/prompt.go](file://internal/wiki/prompt.go)
</cite>

## Table of Contents

- [Overview](#overview)
- [Core Generation](#core-generation)
- [Multi-Engine Support](#multi-engine-support)
- [Auto-Commit System](#auto-commit-system)
- [Change Detection](#change-detection)
- [Prompt Building](#prompt-building)

## Overview

The wiki engine (`internal/wiki/`) is the core component responsible for generating and updating repository documentation. It orchestrates the interaction with the configured AI engine (Qoder CLI, Claude Code, or OpenAI Codex) and manages the wiki lifecycle.

```mermaid
flowchart TB
    subgraph WikiEngine["Wiki Engine"]
        Wiki["wiki.go"]
        Engine["engine.go"]
        Commit["commit.go"]
        Detect["detect.go"]
        Prompt["prompt.go"]
    end

    subgraph External["External"]
        Qoder["Qoder CLI"]
        Claude["Claude Code"]
        Codex["OpenAI Codex"]
        Git["Git"]
    end

    Wiki --> Prompt
    Wiki --> Engine
    Wiki --> Commit
    Wiki --> Detect
    Engine --> Qoder
    Engine --> Claude
    Engine --> Codex
    Commit --> Git
    Detect --> Git
```

## Core Generation

**File**: `internal/wiki/wiki.go`

### Full Generation

Performs complete wiki generation from scratch:

```go
func FullGenerate(gitRoot string, cfg *config.Config, commitHash string) error {
    // Acquire process lock
    if err := lockfile.Acquire(gitRoot); err != nil {
        return fmt.Errorf("cannot acquire lock: %w", err)
    }
    defer lockfile.Release(gitRoot)

    logf(gitRoot, "starting full wiki generation")

    // Build prompt for full generation
    prompt := BuildFullGeneratePrompt(cfg)

    // Execute AI engine
    output, err := RunEngine(cfg, gitRoot, prompt)
    if err != nil {
        logf(gitRoot, "engine failed: %v", err)
        return fmt.Errorf("wiki generation failed: %w", err)
    }

    logf(gitRoot, "engine completed, output length: %d", len(output))

    // Auto-commit if enabled
    if cfg.AutoCommit {
        config.UpdateLastRun(gitRoot, commitHash)
        if err := CommitChanges(gitRoot, cfg, "full wiki generation"); err != nil {
            logf(gitRoot, "auto-commit failed: %v", err)
            return err
        }
        logf(gitRoot, "wiki changes committed")
    }

    return nil
}
```

The `commitHash` parameter is used to update the `last_commit_hash` tracking field in the configuration after successful generation, enabling incremental updates on subsequent runs.

### Incremental Update

Updates wiki for specific changed files:

```go
func IncrementalUpdate(gitRoot string, cfg *config.Config, changedFiles []string, commitHash string) error {
    if err := lockfile.Acquire(gitRoot); err != nil {
        return fmt.Errorf("cannot acquire lock: %w", err)
    }
    defer lockfile.Release(gitRoot)

    logf(gitRoot, "starting incremental update for %d files", len(changedFiles))

    // Determine affected wiki sections
    affectedSections := AffectedSections(gitRoot, cfg, changedFiles)
    logf(gitRoot, "affected sections: %v", affectedSections)

    // Build incremental prompt
    prompt := BuildIncrementalPrompt(cfg, changedFiles, affectedSections)

    // Execute AI engine
    output, err := RunEngine(cfg, gitRoot, prompt)
    if err != nil {
        logf(gitRoot, "engine failed: %v", err)
        return fmt.Errorf("wiki update failed: %w", err)
    }

    logf(gitRoot, "engine completed, output length: %d", len(output))

    // Auto-commit if enabled
    if cfg.AutoCommit {
        config.UpdateLastRun(gitRoot, commitHash)
        desc := fmt.Sprintf("update wiki for %d changed files", len(changedFiles))
        if err := CommitChanges(gitRoot, cfg, desc); err != nil {
            logf(gitRoot, "auto-commit failed: %v", err)
            return err
        }
        logf(gitRoot, "wiki changes committed")
    }

    return nil
}
```

### Wiki Existence Check

```go
func Exists(gitRoot string, cfg *config.Config) bool {
    contentPath := filepath.Join(gitRoot, cfg.WikiPath, cfg.Language, "content")
    entries, err := os.ReadDir(contentPath)
    return err == nil && len(entries) > 0
}
```

### Logging

```go
func logf(gitRoot string, format string, args ...any) {
    logDir := config.LogPath(gitRoot)
    os.MkdirAll(logDir, 0755)

    now := time.Now().UTC()
    logFile := filepath.Join(logDir, now.Format("2006-01-02")+".log")

    f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return
    }
    defer f.Close()

    msg := fmt.Sprintf("[%s] %s\n", now.Format(time.RFC3339), fmt.Sprintf(format, args...))
    f.WriteString(msg)
}
```

## Multi-Engine Support

**File**: `internal/wiki/engine.go`

The wiki engine supports multiple AI backends through a unified interface.

### Supported Engines

| Engine | CLI Binary | Non-Interactive Invocation |
|--------|-----------|--------------------------|
| `qoder` | `qodercli` | `-p "prompt" -q -w <dir> --max-turns N --dangerously-skip-permissions` |
| `claude-code` | `claude` | `-p "prompt" --dangerously-skip-permissions --allowedTools ...` |
| `codex` | `codex` | `exec "prompt" --full-auto` |

Binaries are auto-detected from: `engine_path` config → `$PATH` → known OS locations.

### Engine Interface

```go
// FindEngineBinary locates the CLI binary for the configured engine.
func FindEngineBinary(cfg *config.Config) (string, error) {
    switch cfg.Engine {
    case config.EngineQoder:
        return findQoderBinary(cfg)
    case config.EngineClaudeCode:
        return findClaudeCodeBinary(cfg)
    case config.EngineCodex:
        return findCodexBinary(cfg)
    default:
        return "", fmt.Errorf("unknown engine: %s", cfg.Engine)
    }
}

// RunEngine invokes the configured engine with the given prompt.
func RunEngine(cfg *config.Config, gitRoot string, prompt string) (string, error) {
    switch cfg.Engine {
    case config.EngineQoder:
        return runQoder(cfg, gitRoot, prompt)
    case config.EngineClaudeCode:
        return runClaudeCode(cfg, gitRoot, prompt)
    case config.EngineCodex:
        return runCodex(cfg, gitRoot, prompt)
    default:
        return "", fmt.Errorf("unknown engine: %s", cfg.Engine)
    }
}
```

### Qoder CLI Implementation

```go
func findQoderBinary(cfg *config.Config) (string, error) {
    if cfg.EnginePath != "" {
        if _, err := os.Stat(cfg.EnginePath); err == nil {
            return cfg.EnginePath, nil
        }
    }
    if path, err := exec.LookPath("qodercli"); err == nil {
        return path, nil
    }
    if runtime.GOOS == "darwin" {
        for _, p := range []string{
            "/Applications/Qoder.app/Contents/Resources/app/resources/bin/aarch64_darwin/qodercli",
            "/Applications/Qoder.app/Contents/Resources/app/resources/bin/x86_64_darwin/qodercli",
        } {
            if _, err := os.Stat(p); err == nil {
                return p, nil
            }
        }
    }
    return "", fmt.Errorf("qodercli not found")
}

func runQoder(cfg *config.Config, gitRoot string, prompt string) (string, error) {
    bin, err := findQoderBinary(cfg)
    if err != nil {
        return "", err
    }
    args := []string{
        "-p", prompt,
        "-q",
        "-w", gitRoot,
        "--max-turns", strconv.Itoa(cfg.MaxTurns),
        "--dangerously-skip-permissions",
        "--allowed-tools", "Read,Write,Edit,Glob,Grep,Bash",
    }
    if cfg.Model != "" {
        args = append(args, "--model", cfg.Model)
    }
    return execCLI(bin, gitRoot, args)
}
```

### Claude Code Implementation

```go
func findClaudeCodeBinary(cfg *config.Config) (string, error) {
    if cfg.EnginePath != "" {
        if _, err := os.Stat(cfg.EnginePath); err == nil {
            return cfg.EnginePath, nil
        }
    }
    if path, err := exec.LookPath("claude"); err == nil {
        return path, nil
    }
    home, _ := os.UserHomeDir()
    for _, p := range []string{
        home + "/.local/bin/claude",
        home + "/.claude/bin/claude",
        "/usr/local/bin/claude",
    } {
        if _, err := os.Stat(p); err == nil {
            return p, nil
        }
    }
    return "", fmt.Errorf("claude not found")
}

func runClaudeCode(cfg *config.Config, gitRoot string, prompt string) (string, error) {
    bin, err := findClaudeCodeBinary(cfg)
    if err != nil {
        return "", err
    }
    args := []string{
        "-p", prompt,
        "--dangerously-skip-permissions",
        "--allowedTools", "Read,Write,Edit,Glob,Grep,Bash",
    }
    if cfg.Model != "" {
        args = append(args, "--model", cfg.Model)
    }
    return execCLI(bin, gitRoot, args)
}
```

### OpenAI Codex Implementation

```go
func findCodexBinary(cfg *config.Config) (string, error) {
    if cfg.EnginePath != "" {
        if _, err := os.Stat(cfg.EnginePath); err == nil {
            return cfg.EnginePath, nil
        }
    }
    if path, err := exec.LookPath("codex"); err == nil {
        return path, nil
    }
    return "", fmt.Errorf("codex not found")
}

func runCodex(cfg *config.Config, gitRoot string, prompt string) (string, error) {
    bin, err := findCodexBinary(cfg)
    if err != nil {
        return "", err
    }
    args := []string{
        "exec", prompt,
        "--full-auto",
    }
    return execCLI(bin, gitRoot, args)
}
```

### Common Executor

```go
func execCLI(bin string, dir string, args []string) (string, error) {
    cmd := exec.Command(bin, args...)
    cmd.Dir = dir

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("%s error: %w\nstderr: %s", bin, err, stderr.String())
    }
    return stdout.String(), nil
}
```

## Auto-Commit System

**File**: `internal/wiki/commit.go`

### Sentinel File (Loop Prevention)

```go
const sentinelFile = ".committing"

func sentinelPath(gitRoot string) string {
    return filepath.Join(config.Dir(gitRoot), sentinelFile)
}

func IsSentinelPresent(gitRoot string) bool {
    _, err := os.Stat(sentinelPath(gitRoot))
    return err == nil
}
```

### Commit Changes

```go
func CommitChanges(gitRoot string, cfg *config.Config, description string) error {
    wikiDir := filepath.Join(gitRoot, cfg.WikiPath)

    // Check if there are any changes to commit
    hasChanges, err := git.HasChanges(gitRoot, wikiDir)
    if err != nil || !hasChanges {
        return nil
    }

    // Write sentinel file (loop prevention layer 1)
    sp := sentinelPath(gitRoot)
    if err := os.WriteFile(sp, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
        return fmt.Errorf("failed to write sentinel: %w", err)
    }
    defer os.Remove(sp)

    // Stage wiki files
    if err := git.StageFiles(gitRoot, []string{wikiDir}); err != nil {
        return fmt.Errorf("failed to stage wiki files: %w", err)
    }

    // Also stage config (updated last_run, last_commit_hash)
    configPath := config.Path(gitRoot)
    if _, err := os.Stat(configPath); err == nil {
        git.StageFiles(gitRoot, []string{configPath})
    }

    // Commit with recognizable prefix
    message := fmt.Sprintf("%s %s", cfg.CommitPrefix, description)
    if err := git.Commit(gitRoot, message); err != nil {
        return fmt.Errorf("failed to commit wiki: %w", err)
    }

    return nil
}
```

## Change Detection

**File**: `internal/wiki/detect.go`

### Affected Sections Detection

Determines which wiki sections need updating based on changed files:

```go
type codeSnippet struct {
    ID        string `json:"id"`
    Path      string `json:"path"`
    LineRange string `json:"line_range"`
}

type metadata struct {
    CodeSnippets []codeSnippet `json:"code_snippets"`
}

func AffectedSections(gitRoot string, cfg *config.Config, changedFiles []string) []string {
    affected := map[string]bool{}

    // 1. Build reverse index from metadata
    reverseIdx := buildReverseIndex(gitRoot, cfg)
    for _, f := range changedFiles {
        if pages, ok := reverseIdx[f]; ok {
            for _, p := range pages {
                affected[p] = true
            }
        }
    }

    // 2. Heuristic path matching
    for _, f := range changedFiles {
        for _, section := range heuristicMatch(f) {
            affected[section] = true
        }
    }

    result := make([]string, 0, len(affected))
    for s := range affected {
        result = append(result, s)
    }
    return result
}
```

### Reverse Index Building

```go
func buildReverseIndex(gitRoot string, cfg *config.Config) map[string][]string {
    idx := map[string][]string{}

    metaPath := filepath.Join(gitRoot, cfg.WikiPath, cfg.Language, "meta", "repowiki-metadata.json")
    data, err := os.ReadFile(metaPath)
    if err != nil {
        return idx
    }

    var meta metadata
    if err := json.Unmarshal(data, &meta); err != nil {
        return idx
    }

    sourceFiles := map[string]bool{}
    for _, s := range meta.CodeSnippets {
        sourceFiles[s.Path] = true
    }

    contentDir := filepath.Join(gitRoot, cfg.WikiPath, cfg.Language, "content")
    scanWikiContent(contentDir, "", sourceFiles, idx)

    return idx
}
```

### Wiki Content Scanning

```go
func scanWikiContent(dir string, relDir string, sourceFiles map[string]bool, idx map[string][]string) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return
    }

    for _, e := range entries {
        if e.IsDir() {
            subRel := filepath.Join(relDir, e.Name())
            scanWikiContent(filepath.Join(dir, e.Name()), subRel, sourceFiles, idx)
            continue
        }

        if !strings.HasSuffix(e.Name(), ".md") {
            continue
        }

        wikiPage := filepath.Join(relDir, e.Name())
        data, err := os.ReadFile(filepath.Join(dir, e.Name()))
        if err != nil {
            continue
        }

        content := string(data)
        for srcFile := range sourceFiles {
            if strings.Contains(content, "file://"+srcFile) ||
               strings.Contains(content, srcFile) {
                idx[srcFile] = append(idx[srcFile], wikiPage)
            }
        }
    }
}
```

### Heuristic Matching

```go
func heuristicMatch(filePath string) []string {
    var sections []string
    lower := strings.ToLower(filePath)

    switch {
    case strings.Contains(lower, "backend/") ||
         strings.Contains(lower, "server/") ||
         strings.Contains(lower, "src/api/"):
        sections = append(sections, "Backend Architecture")
    case strings.Contains(lower, "frontend/") ||
         strings.Contains(lower, "src/components/") ||
         strings.Contains(lower, "src/app/"):
        sections = append(sections, "Frontend Architecture")
    }

    if strings.Contains(lower, "api/") ||
       strings.Contains(lower, "routes/") ||
       strings.Contains(lower, "endpoints/") {
        sections = append(sections, "API Reference")
    }

    if strings.Contains(lower, "config") ||
       strings.Contains(lower, ".env") ||
       strings.Contains(lower, "settings") {
        sections = append(sections, "Configuration Management")
    }

    if strings.HasSuffix(lower, "readme.md") ||
       strings.HasSuffix(lower, "package.json") ||
       strings.HasSuffix(lower, "pyproject.toml") {
        sections = append(sections, "System Overview")
    }

    if strings.Contains(lower, "auth") ||
       strings.Contains(lower, "security") {
        sections = append(sections, "Authentication and Security")
    }

    if strings.Contains(lower, "database/") ||
       strings.Contains(lower, "models/") ||
       strings.Contains(lower, "migrations/") {
        sections = append(sections, "Backend Architecture")
    }

    return sections
}
```

## Prompt Building

**File**: `internal/wiki/prompt.go`

### Full Generation Prompt

```go
func BuildFullGeneratePrompt(cfg *config.Config) string {
    return fmt.Sprintf(`You are a technical documentation specialist. Generate a comprehensive repository wiki for this project.

OUTPUT REQUIREMENTS:
- Create documentation files in %s/%s/content/ directory
- Create a metadata file at %s/%s/meta/repowiki-metadata.json
- Each markdown file must follow this structure:
  1. Title as H1 heading
  2. <cite> block listing referenced source files
  3. Table of Contents with anchor links
  4. Detailed content with code examples
  5. Mermaid diagrams for architecture

WIKI STRUCTURE:
- System Overview.md
- Technology Stack.md
- Getting Started.md
- Backend Architecture/
- Frontend Architecture/
- Core Features/
- API Reference/
- Configuration Management.md

METADATA FORMAT:
{
  "code_snippets": [
    {
      "id": "<md5 hash>",
      "path": "relative/path/to/file",
      "line_range": "1-100",
      "gmt_create": "<ISO 8601 timestamp>",
      "gmt_modified": "<ISO 8601 timestamp>"
    }
  ]
}

Analyze ALL source files. Be thorough. Include actual code references.
Do NOT modify any source code. Only create/modify files within %s/.`,
        cfg.WikiPath, cfg.Language,
        cfg.WikiPath, cfg.Language,
        cfg.WikiPath)
}
```

### Incremental Update Prompt

```go
func BuildIncrementalPrompt(cfg *config.Config, changedFiles []string, affectedSections []string) string {
    fileList := "  - " + strings.Join(changedFiles, "\n  - ")

    sectionHint := ""
    if len(affectedSections) > 0 {
        sectionHint = fmt.Sprintf(`
POTENTIALLY AFFECTED WIKI SECTIONS:
  - %s
`, strings.Join(affectedSections, "\n  - "))
    }

    return fmt.Sprintf(`You are a technical documentation specialist. Update the repository wiki.

CHANGED SOURCE FILES:
%s
%s
INSTRUCTIONS:
1. Read each changed source file
2. Read existing wiki pages
3. Update affected wiki sections
4. Create new pages if needed
5. Update metadata with new references
6. Preserve existing formatting
7. Do NOT modify source code`,
        fileList, sectionHint)
}
```
