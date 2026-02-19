package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ikrasnodymov/repowiki/internal/config"
	"github.com/ikrasnodymov/repowiki/internal/git"
)

func handleLogs(args []string) {
	gitRoot, err := git.FindRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not a git repository\n")
		os.Exit(1)
	}

	logDir := config.LogPath(gitRoot)
	entries, err := os.ReadDir(logDir)
	if err != nil {
		fmt.Println("No logs yet.")
		return
	}

	// Sort by name descending (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})

	// Show latest log
	if len(entries) == 0 {
		fmt.Println("No logs yet.")
		return
	}

	latest := entries[0]
	data, err := os.ReadFile(filepath.Join(logDir, latest.Name()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading log: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== %s ===\n%s", latest.Name(), string(data))
}
