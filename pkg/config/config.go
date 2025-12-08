package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the dnsbro runtime configuration.
type Config struct {
	Listen   string `yaml:"listen"`
	Upstream struct {
		DoHEndpoint string        `yaml:"doh_endpoint"`
		Timeout     time.Duration `yaml:"timeout"`
	} `yaml:"upstream"`
	Rules struct {
		Blocklist []string `yaml:"blocklist"`
		Allowlist []string `yaml:"allowlist"`
	} `yaml:"rules"`
	Log struct {
		File  string `yaml:"file"`
		Level string `yaml:"level"`
	} `yaml:"log"`
}

// Defaults returns a Config populated with sensible defaults.
func Defaults() Config {
	var cfg Config
	cfg.Listen = "127.0.0.1:53"
	cfg.Upstream.DoHEndpoint = "https://1.1.1.1/dns-query"
	cfg.Upstream.Timeout = 5 * time.Second
	cfg.Log.Level = "info"
	return cfg
}

// Load reads the configuration file and returns the parsed Config.
func Load(path string) (Config, error) {
	cfg := Defaults()

	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}

	if cfg.Listen == "" {
		return cfg, errors.New("listen address required")
	}
	if cfg.Upstream.DoHEndpoint == "" {
		return cfg, errors.New("upstream.doh_endpoint required")
	}
	if cfg.Upstream.Timeout == 0 {
		cfg.Upstream.Timeout = 5 * time.Second
	}
	return cfg, nil
}

// Write persists the config to the given path, creating parent directories when needed.
func Write(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0o644)
}
