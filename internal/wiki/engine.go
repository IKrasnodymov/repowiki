package wiki

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/ikrasnodymov/repowiki/internal/config"
)

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
		return "", fmt.Errorf("unknown engine: %s (valid: qoder, claude-code, codex)", cfg.Engine)
	}
}

// RunEngine invokes the configured engine with the given prompt in non-interactive mode.
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

// --- Qoder CLI ---

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
	return "", fmt.Errorf("qodercli not found; install Qoder or set engine_path in config")
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

// --- Claude Code ---

func findClaudeCodeBinary(cfg *config.Config) (string, error) {
	if cfg.EnginePath != "" {
		if _, err := os.Stat(cfg.EnginePath); err == nil {
			return cfg.EnginePath, nil
		}
	}
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}
	// Common locations
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
	return "", fmt.Errorf("claude not found; install Claude Code or set engine_path in config")
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

// --- Codex CLI ---

func findCodexBinary(cfg *config.Config) (string, error) {
	if cfg.EnginePath != "" {
		if _, err := os.Stat(cfg.EnginePath); err == nil {
			return cfg.EnginePath, nil
		}
	}
	if path, err := exec.LookPath("codex"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("codex not found; install OpenAI Codex CLI or set engine_path in config")
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

// --- Common executor ---

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
