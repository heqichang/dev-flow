package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *ProjectConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &ProjectConfig{
				ProjectName: "test-project",
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			cfg: &ProjectConfig{
				Version: "1.0.0",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devflow-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &ProjectConfig{
		ProjectName: "test-project",
		Version:     "1.0.0",
		Language:    "go",
		Author:      "Test Author",
		License:     "MIT",
		Scripts: map[string]Script{
			"test": {
				Command: "go test ./...",
			},
		},
	}

	configPath := filepath.Join(tmpDir, ".devflow.yml")
	if err := SaveProjectConfig(cfg, configPath); err != nil {
		t.Fatalf("SaveProjectConfig() error = %v", err)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working dir: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	loaded, err := LoadConfig(Development)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if loaded.Project.ProjectName != cfg.ProjectName {
		t.Errorf("ProjectName = %v, want %v", loaded.Project.ProjectName, cfg.ProjectName)
	}

	if loaded.Project.Version != cfg.Version {
		t.Errorf("Version = %v, want %v", loaded.Project.Version, cfg.Version)
	}

	if loaded.Project.Language != cfg.Language {
		t.Errorf("Language = %v, want %v", loaded.Project.Language, cfg.Language)
	}
}

func TestMergeConfig(t *testing.T) {
	base := &ProjectConfig{
		ProjectName: "base",
		Version:     "1.0.0",
		Language:    "go",
		Scripts: map[string]Script{
			"test": {Command: "go test"},
		},
		Env: map[string]string{
			"KEY1": "value1",
		},
	}

	override := &ProjectConfig{
		Version:  "2.0.0",
		Language: "javascript",
		Scripts: map[string]Script{
			"build": {Command: "npm run build"},
		},
		Env: map[string]string{
			"KEY2": "value2",
		},
	}

	mergeConfig(base, override)

	if base.ProjectName != "base" {
		t.Errorf("ProjectName should not be overridden, got %v", base.ProjectName)
	}

	if base.Version != "2.0.0" {
		t.Errorf("Version should be overridden to 2.0.0, got %v", base.Version)
	}

	if base.Language != "javascript" {
		t.Errorf("Language should be overridden to javascript, got %v", base.Language)
	}

	if _, exists := base.Scripts["test"]; !exists {
		t.Error("test script should still exist")
	}

	if _, exists := base.Scripts["build"]; !exists {
		t.Error("build script should be added")
	}

	if base.Env["KEY1"] != "value1" {
		t.Error("KEY1 should still exist")
	}

	if base.Env["KEY2"] != "value2" {
		t.Error("KEY2 should be added")
	}
}

func TestDefaultProjectConfig(t *testing.T) {
	cfg := DefaultProjectConfig()

	if cfg.Version != "0.1.0" {
		t.Errorf("Default Version = %v, want 0.1.0", cfg.Version)
	}

	if cfg.License != "MIT" {
		t.Errorf("Default License = %v, want MIT", cfg.License)
	}

	if cfg.Scripts == nil {
		t.Error("Scripts should not be nil")
	}

	if cfg.Env == nil {
		t.Error("Env should not be nil")
	}

	if cfg.Environments == nil {
		t.Error("Environments should not be nil")
	}
}
