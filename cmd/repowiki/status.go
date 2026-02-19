package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ikrasnodymov/repowiki/internal/config"
	"github.com/ikrasnodymov/repowiki/internal/git"
	"github.com/ikrasnodymov/repowiki/internal/hook"
	"github.com/ikrasnodymov/repowiki/internal/wiki"
)

func handleStatus(args []string) {
	gitRoot, err := git.FindRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not a git repository\n")
		os.Exit(1)
	}

	fmt.Printf("repowiki v%s\n\n", Version)

	// Config
	cfg, cfgErr := config.Load(gitRoot)
	if cfgErr != nil {
		fmt.Printf("  Status:       not configured\n")
		fmt.Printf("  Run 'repowiki enable' to get started.\n")
		return
	}

	if cfg.Enabled {
		fmt.Printf("  Status:       enabled\n")
	} else {
		fmt.Printf("  Status:       disabled\n")
	}

	// Engine
	fmt.Printf("  Engine:       %s\n", cfg.Engine)

	// Hook
	if hook.IsInstalled(gitRoot) {
		fmt.Printf("  Hook:         installed (.git/hooks/post-commit)\n")
	} else {
		fmt.Printf("  Hook:         not installed\n")
	}

	// Engine binary
	binPath, engineErr := wiki.FindEngineBinary(cfg)
	if engineErr == nil {
		fmt.Printf("  Binary:       %s\n", binPath)
	} else {
		fmt.Printf("  Binary:       not found (%s)\n", cfg.Engine)
	}

	// Wiki
	contentDir := filepath.Join(gitRoot, cfg.WikiPath, cfg.Language, "content")
	if entries, err := os.ReadDir(contentDir); err == nil {
		count := countMdFiles(contentDir)
		fmt.Printf("  Wiki path:    %s/%s/content/ (%d pages)\n", cfg.WikiPath, cfg.Language, count)
	} else {
		fmt.Printf("  Wiki path:    %s (not generated yet)\n", cfg.WikiPath)
		_ = entries
	}

	// Config details
	if cfg.Model != "" {
		fmt.Printf("  Model:        %s\n", cfg.Model)
	}
	fmt.Printf("  Auto-commit:  %v\n", cfg.AutoCommit)
	fmt.Printf("  Max turns:    %d\n", cfg.MaxTurns)

	if cfg.LastRun != "" {
		fmt.Printf("  Last run:     %s\n", cfg.LastRun)
	}
	if cfg.LastCommitHash != "" {
		fmt.Printf("  Last commit:  %s\n", cfg.LastCommitHash)
	}
}

func countMdFiles(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			count++
		}
		return nil
	})
	return count
}
