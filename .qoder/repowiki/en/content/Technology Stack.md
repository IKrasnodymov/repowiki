# Technology Stack

<cite>
Source files referenced:
- [go.mod](/to/go.mod)
- [Makefile](/to/Makefile)
- [cmd/repowiki/main.go](/to/cmd/repowiki/main.go)
- [internal/git/git.go](/to/internal/git/git.go)
- [internal/wiki/qoder.go](/to/internal/wiki/qoder.go)
</cite>

## Table of Contents

- [Programming Language](#programming-language)
- [Build System](#build-system)
- [Dependencies](#dependencies)
- [External Tools](#external-tools)
- [Project Structure](#project-structure)

## Programming Language

**Go 1.22+** — The entire project is written in Go, leveraging:
- Standard library for all core functionality
- No external dependencies (zero third-party imports)
- Cross-platform compatibility (macOS, Linux)
- Single binary deployment

```go
// go.mod
module github.com/IKrasnodymov/repowiki

go 1.22.0
```

## Build System

The project uses a simple **Makefile** for build automation:

```makefile
.PHONY: build install test clean

BINARY := repowiki
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/repowiki

install:
	go install ./cmd/repowiki

test:
	go test ./internal/... -v -race

clean:
	rm -rf $(BUILD_DIR)
```

### Build Commands

| Command | Description |
|---------|-------------|
| `make build` | Compile binary to `bin/repowiki` |
| `make install` | Install to `$GOPATH/bin` |
| `make test` | Run tests with race detection |
| `make clean` | Remove build artifacts |

## Dependencies

This project has **zero external dependencies**. All functionality is implemented using Go's standard library:

### Standard Library Packages Used

| Package | Purpose | Files Using |
|---------|---------|-------------|
| `os` | File operations, process management | All files |
| `os/exec` | External command execution | `git/git.go`, `wiki/qoder.go` |
| `path/filepath` | Cross-platform path manipulation | All files |
| `encoding/json` | Configuration serialization | `config/config.go` |
| `fmt` | Formatted I/O | All files |
| `strings` | String manipulation | Multiple files |
| `time` | Timestamp handling | `config/config.go`, `wiki/wiki.go` |
| `syscall` | Process detachment | `hooks.go` |
| `strconv` | String/number conversion | `wiki/commit.go`, `wiki/qoder.go` |
| `bytes` | Buffer management | `wiki/qoder.go` |

### Why Zero Dependencies?

The project intentionally avoids external dependencies for:
- **Simplicity**: No dependency management complexity
- **Reliability**: No risk of broken or compromised dependencies
- **Portability**: Single binary that just works
- **Security**: Minimal attack surface

## External Tools

While the binary itself has no dependencies, it integrates with external tools:

### Required

| Tool | Purpose | Detection |
|------|---------|-----------|
| `git` | Repository operations | Must be in PATH |
| `qodercli` | AI-powered wiki generation | Configurable path |

### Qoder CLI Detection

The tool searches for `qodercli` in multiple locations:

```go
func FindQoderCLI(cfg *config.Config) (string, error) {
    // 1. Use config override
    if cfg.QoderCLIPath != "" && cfg.QoderCLIPath != "qodercli" {
        if _, err := os.Stat(cfg.QoderCLIPath); err == nil {
            return cfg.QoderCLIPath, nil
        }
    }

    // 2. Check PATH
    if path, err := exec.LookPath("qodercli"); err == nil {
        return path, nil
    }

    // 3. Check known macOS locations
    if runtime.GOOS == "darwin" {
        knownPaths := []string{
            "/Applications/Qoder.app/Contents/Resources/app/resources/bin/aarch64_darwin/qodercli",
            "/Applications/Qoder.app/Contents/Resources/app/resources/bin/x86_64_darwin/qodercli",
        }
        // ...
    }

    // 4. Check known Linux locations
    if runtime.GOOS == "linux" {
        knownPaths := []string{
            "/usr/bin/qodercli",
            "/usr/local/bin/qodercli",
        }
        // ...
    }
}
```

## Project Structure

```
repowiki/
├── cmd/repowiki/          # CLI entry points
│   ├── main.go            # Command router
│   ├── enable.go          # Enable command
│   ├── disable.go         # Disable command
│   ├── status.go          # Status command
│   ├── generate.go        # Full generation
│   ├── update.go          # Incremental update
│   └── hooks.go           # Hook entry point
├── internal/              # Internal packages
│   ├── config/            # Configuration management
│   ├── git/               # Git operations
│   ├── hook/              # Git hook management
│   ├── lockfile/          # Process locking
│   └── wiki/              # Wiki generation
│       ├── wiki.go        # Core generation logic
│       ├── qoder.go       # Qoder CLI integration
│       ├── commit.go      # Auto-commit logic
│       ├── detect.go      # Change detection
│       └── prompt.go      # Prompt building
├── go.mod                 # Go module definition
├── Makefile               # Build automation
└── README.md              # User documentation
```

## Configuration Files

### Runtime Configuration

| File | Format | Purpose |
|------|--------|---------|
| `.repowiki/config.json` | JSON | Tool configuration |
| `.repowiki/.repowiki.lock` | Text | Process lock file |
| `.repowiki/.committing` | Text | Sentinel file (loop prevention) |
| `.repowiki/logs/*.log` | Text | Execution logs |

### Generated Output

| File/Directory | Format | Purpose |
|----------------|--------|---------|
| `.qoder/repowiki/en/content/*.md` | Markdown | Wiki documentation |
| `.qoder/repowiki/en/meta/repowiki-metadata.json` | JSON | Code snippet index |
