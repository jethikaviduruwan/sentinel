package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temp config file
	content := `
server_id: "test-server"
hq_address: "localhost:50051"
interval_seconds: 5
services:
  - nginx
  - bash
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ServerID != "test-server" {
		t.Errorf("expected server_id 'test-server', got '%s'", cfg.ServerID)
	}

	if cfg.HQAddress != "localhost:50051" {
		t.Errorf("expected hq_address 'localhost:50051', got '%s'", cfg.HQAddress)
	}

	if cfg.IntervalSeconds != 5 {
		t.Errorf("expected interval_seconds 5, got %d", cfg.IntervalSeconds)
	}

	if len(cfg.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(cfg.Services))
	}

	t.Logf("config loaded: server_id=%s services=%v", cfg.ServerID, cfg.Services)
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := Load("nonexistent-file.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}