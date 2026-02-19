package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ikrasnodymov/repowiki/internal/config"
	"github.com/ikrasnodymov/repowiki/internal/lockfile"
)

// FullGenerate performs a complete wiki generation from scratch.
func FullGenerate(gitRoot string, cfg *config.Config) error {
	if err := lockfile.Acquire(gitRoot); err != nil {
		return fmt.Errorf("cannot acquire lock: %w", err)
	}
	defer lockfile.Release(gitRoot)

	logf(gitRoot, "starting full wiki generation")

	prompt := BuildFullGeneratePrompt(cfg)

	output, err := RunEngine(cfg, gitRoot, prompt)
	if err != nil {
		logf(gitRoot, "engine failed: %v", err)
		return fmt.Errorf("wiki generation failed: %w", err)
	}

	logf(gitRoot, "engine completed, output length: %d", len(output))

	if cfg.AutoCommit {
		if err := CommitChanges(gitRoot, cfg, "full wiki generation"); err != nil {
			logf(gitRoot, "auto-commit failed: %v", err)
			return err
		}
		logf(gitRoot, "wiki changes committed")
	}

	return nil
}

// IncrementalUpdate updates wiki for specific changed files.
func IncrementalUpdate(gitRoot string, cfg *config.Config, changedFiles []string) error {
	if err := lockfile.Acquire(gitRoot); err != nil {
		return fmt.Errorf("cannot acquire lock: %w", err)
	}
	defer lockfile.Release(gitRoot)

	logf(gitRoot, "starting incremental update for %d files", len(changedFiles))

	affectedSections := AffectedSections(gitRoot, cfg, changedFiles)
	logf(gitRoot, "affected sections: %v", affectedSections)

	prompt := BuildIncrementalPrompt(cfg, changedFiles, affectedSections)

	output, err := RunEngine(cfg, gitRoot, prompt)
	if err != nil {
		logf(gitRoot, "engine failed: %v", err)
		return fmt.Errorf("wiki update failed: %w", err)
	}

	logf(gitRoot, "engine completed, output length: %d", len(output))

	if cfg.AutoCommit {
		desc := fmt.Sprintf("update wiki for %d changed files", len(changedFiles))
		if err := CommitChanges(gitRoot, cfg, desc); err != nil {
			logf(gitRoot, "auto-commit failed: %v", err)
			return err
		}
		logf(gitRoot, "wiki changes committed")
	}

	return nil
}

// Exists checks if the wiki directory has content.
func Exists(gitRoot string, cfg *config.Config) bool {
	contentPath := filepath.Join(gitRoot, cfg.WikiPath, cfg.Language, "content")
	entries, err := os.ReadDir(contentPath)
	return err == nil && len(entries) > 0
}

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
