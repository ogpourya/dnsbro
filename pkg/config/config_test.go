package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteCreatesConfigAndParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.yaml")

	cfg := Defaults()
	cfg.Listen = "127.0.0.2:53"
	cfg.Upstream.Timeout = 3 * time.Second

	if err := Write(path, cfg); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file missing: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Listen != cfg.Listen {
		t.Fatalf("Listen mismatch: got %q want %q", loaded.Listen, cfg.Listen)
	}
	if loaded.Upstream.DoHEndpoint != cfg.Upstream.DoHEndpoint {
		t.Fatalf("DoH endpoint mismatch: got %q want %q", loaded.Upstream.DoHEndpoint, cfg.Upstream.DoHEndpoint)
	}
	if loaded.Upstream.Timeout != cfg.Upstream.Timeout {
		t.Fatalf("Timeout mismatch: got %v want %v", loaded.Upstream.Timeout, cfg.Upstream.Timeout)
	}
}
