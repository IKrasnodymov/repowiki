package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ikrasnodymov/repowiki/internal/config"
	"github.com/ikrasnodymov/repowiki/internal/git"
	"github.com/ikrasnodymov/repowiki/internal/hook"
	"github.com/ikrasnodymov/repowiki/internal/wiki"
)

func handleEnable(args []string) {
	fs := flag.NewFlagSet("enable", flag.ExitOnError)
	force := fs.Bool("force", false, "reinstall hook even if present")
	engine := fs.String("engine", "", "AI engine: qoder, claude-code, codex")
	enginePath := fs.String("engine-path", "", "path to engine CLI binary")
	model := fs.String("model", "", "model level (engine-specific)")
	noAutoCommit := fs.Bool("no-auto-commit", false, "don't auto-commit wiki changes")
	fs.Parse(args)

	gitRoot, err := git.FindRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not a git repository\n")
		os.Exit(1)
	}

	// Load existing config or create default
	cfg, err := config.Load(gitRoot)
	if err != nil {
		cfg = config.Default()
	}

	// Apply flag overrides
	if *engine != "" {
		if !config.IsValidEngine(*engine) {
			fmt.Fprintf(os.Stderr, "Error: unknown engine %q (valid: %s)\n", *engine, strings.Join(config.ValidEngines, ", "))
			os.Exit(1)
		}
		cfg.Engine = *engine
	}
	if *enginePath != "" {
		cfg.EnginePath = *enginePath
	}
	if *model != "" {
		cfg.Model = *model
	}
	if *noAutoCommit {
		cfg.AutoCommit = false
	}
	cfg.Enabled = true

	// Validate engine binary is reachable
	binPath, findErr := wiki.FindEngineBinary(cfg)
	if findErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", findErr)
		fmt.Fprintf(os.Stderr, "Set the path with: repowiki enable --engine-path /path/to/binary\n\n")
	}

	// Save config
	if err := config.Save(gitRoot, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	// Determine absolute path to this binary for the hook
	selfPath, _ := os.Executable()

	// Install git hook
	if err := hook.Install(gitRoot, *force, selfPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error installing hook: %v\n", err)
		os.Exit(1)
	}

	// Create custom Qoder command (useful even with other engines)
	createQoderCommand(gitRoot)

	fmt.Printf("repowiki enabled in %s\n\n", gitRoot)
	fmt.Printf("  Engine:  %s\n", cfg.Engine)
	if findErr == nil {
		fmt.Printf("  Binary:  %s\n", binPath)
	}
	fmt.Printf("  Config:  %s\n", config.Path(gitRoot))
	fmt.Printf("  Hook:    .git/hooks/post-commit\n")
	fmt.Printf("\nEvery commit will now auto-update the repo wiki.\n")
	fmt.Printf("Run 'repowiki generate' for initial full wiki generation.\n")
}

func createQoderCommand(gitRoot string) {
	cmdDir := filepath.Join(gitRoot, ".qoder", "commands")
	os.MkdirAll(cmdDir, 0755)

	cmdPath := filepath.Join(cmdDir, "update-wiki.md")
	if _, err := os.Stat(cmdPath); err == nil {
		return // Already exists
	}

	content := `---
description: Update the repository wiki documentation based on recent code changes
---

You are a technical documentation specialist. Update the repository wiki in ` + "`" + `.qoder/repowiki/` + "`" + ` to reflect the current state of the codebase.

## Instructions

1. Run ` + "`" + `git diff --name-only HEAD~5 HEAD` + "`" + ` to see recently changed files
2. Read the changed source files to understand what was modified
3. Read the existing wiki pages in ` + "`" + `.qoder/repowiki/en/content/` + "`" + `
4. Update any wiki pages that reference or document the changed code
5. If new modules/features were added without wiki coverage, create new pages
6. Update ` + "`" + `.qoder/repowiki/en/meta/repowiki-metadata.json` + "`" + ` with new code snippet references

## Formatting Rules

- Each wiki page starts with an H1 title
- Include a ` + "`" + `<cite>` + "`" + ` block listing all source files referenced
- Include a Table of Contents after the cite block
- Use mermaid diagrams for architecture documentation
- Reference code with ` + "`" + `file://path/to/file` + "`" + ` format in cite blocks

## Constraints

- Do NOT modify any source code files
- Only create/modify files within ` + "`" + `.qoder/repowiki/` + "`" + `
`
	os.WriteFile(cmdPath, []byte(content), 0644)
}
