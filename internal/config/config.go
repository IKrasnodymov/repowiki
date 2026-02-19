package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	ConfigDir  = ".repowiki"
	ConfigFile = "config.json"
	LogDir     = "logs"

	EngineQoder     = "qoder"
	EngineClaudeCode = "claude-code"
	EngineCodex     = "codex"
)

type Config struct {
	Enabled               bool     `json:"enabled"`
	Engine                string   `json:"engine"`
	EnginePath            string   `json:"engine_path,omitempty"`
	Model                 string   `json:"model"`
	MaxTurns              int      `json:"max_turns"`
	Language              string   `json:"language"`
	AutoCommit            bool     `json:"auto_commit"`
	CommitPrefix          string   `json:"commit_prefix"`
	ExcludedPaths         []string `json:"excluded_paths"`
	WikiPath              string   `json:"wiki_path"`
	FullGenerateThreshold int      `json:"full_generate_threshold"`
	LastRun               string   `json:"last_run,omitempty"`
	LastCommitHash        string   `json:"last_commit_hash,omitempty"`
}

func Default() *Config {
	return &Config{
		Enabled:      true,
		Engine:       EngineQoder,
		EnginePath:   "",
		Model:        "",
		MaxTurns:     50,
		Language:     "en",
		AutoCommit:   true,
		CommitPrefix: "[repowiki]",
		ExcludedPaths: []string{
			".qoder/repowiki/",
			".repowiki/",
			"node_modules/",
			"vendor/",
			".git/",
		},
		WikiPath:              ".qoder/repowiki",
		FullGenerateThreshold: 20,
	}
}

var ValidEngines = []string{EngineQoder, EngineClaudeCode, EngineCodex}

func IsValidEngine(engine string) bool {
	for _, e := range ValidEngines {
		if e == engine {
			return true
		}
	}
	return false
}

func Dir(gitRoot string) string {
	return filepath.Join(gitRoot, ConfigDir)
}

func Path(gitRoot string) string {
	return filepath.Join(Dir(gitRoot), ConfigFile)
}

func LogPath(gitRoot string) string {
	return filepath.Join(Dir(gitRoot), LogDir)
}

func Load(gitRoot string) (*Config, error) {
	data, err := os.ReadFile(Path(gitRoot))
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	// Migration: old configs without engine field default to qoder
	if cfg.Engine == "" {
		cfg.Engine = EngineQoder
	}
	return &cfg, nil
}

func Save(gitRoot string, cfg *Config) error {
	if err := os.MkdirAll(Dir(gitRoot), 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(Path(gitRoot), data, 0644)
}

func UpdateLastRun(gitRoot string, commitHash string) error {
	cfg, err := Load(gitRoot)
	if err != nil {
		return err
	}
	cfg.LastRun = time.Now().UTC().Format(time.RFC3339)
	cfg.LastCommitHash = commitHash
	return Save(gitRoot, cfg)
}
