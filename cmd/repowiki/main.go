package main

import (
	"fmt"
	"os"
)

const Version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "enable":
		handleEnable(os.Args[2:])
	case "disable":
		handleDisable(os.Args[2:])
	case "status":
		handleStatus(os.Args[2:])
	case "generate":
		handleGenerate(os.Args[2:])
	case "update":
		handleUpdate(os.Args[2:])
	case "hooks":
		handleHooks(os.Args[2:])
	case "logs":
		handleLogs(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("repowiki v%s\n", Version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'repowiki help' for usage.\n", os.Args[1])
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`repowiki v%s — Auto-generate Qoder repo wiki on git commits

Usage:
  repowiki <command> [flags]

Commands:
  enable      Enable repowiki in current project (install git hook)
  disable     Disable repowiki (remove git hook)
  status      Show current status and configuration
  generate    Run full wiki generation
  update      Run incremental wiki update for recent changes
  version     Show version

Flags for 'enable':
  --force             Reinstall hook even if already present
  --qodercli-path     Path to qodercli binary
  --model             Qoder model level (auto, efficient, performance, ultimate)
  --no-auto-commit    Don't auto-commit wiki changes

Flags for 'update':
  --commit            Specific commit hash to process
  --from-hook         Internal: indicates hook-triggered run

Examples:
  repowiki enable                    # Enable in current project
  repowiki enable --force            # Reinstall hook
  repowiki generate                  # Full wiki generation
  repowiki update --commit abc123    # Update for specific commit
  repowiki disable                   # Remove hook
`, Version)
}
