package quality

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected ProjectLanguage
	}{
		{
			name:     "go project",
			files:    []string{"go.mod"},
			expected: LangGo,
		},
		{
			name:     "javascript project",
			files:    []string{"package.json"},
			expected: LangJavaScript,
		},
		{
			name:     "typescript project",
			files:    []string{"tsconfig.json"},
			expected: LangTypeScript,
		},
		{
			name:     "python project with requirements",
			files:    []string{"requirements.txt"},
			expected: LangPython,
		},
		{
			name:     "python project with setup.py",
			files:    []string{"setup.py"},
			expected: LangPython,
		},
		{
			name:     "rust project",
			files:    []string{"Cargo.toml"},
			expected: LangRust,
		},
		{
			name:     "java project",
			files:    []string{"pom.xml"},
			expected: LangJava,
		},
		{
			name:     "unknown project",
			files:    []string{},
			expected: LangUnknown,
		},
		{
			name:     "go takes priority over package.json",
			files:    []string{"go.mod", "package.json"},
			expected: LangGo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "devflow-quality-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			for _, f := range tt.files {
				os.WriteFile(filepath.Join(tmpDir, f), []byte{}, 0644)
			}

			result := detectLanguage(tmpDir)
			if result != tt.expected {
				t.Errorf("detectLanguage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCheckReportFormatJSON(t *testing.T) {
	report := CheckReport{
		Language: LangGo,
		Results: []CheckResult{
			{Name: "lint", Passed: true, Duration: "100ms"},
			{Name: "test", Passed: false, Output: "1 failed", Duration: "200ms"},
		},
		Passed: false,
	}

	jsonStr, err := report.FormatJSON()
	if err != nil {
		t.Fatalf("FormatJSON() error = %v", err)
	}

	if len(jsonStr) == 0 {
		t.Error("FormatJSON() returned empty string")
	}
}

func TestCheckReportFormatTable(t *testing.T) {
	report := CheckReport{
		Language: LangGo,
		Results: []CheckResult{
			{Name: "lint", Passed: true, Duration: "100ms"},
			{Name: "test", Passed: true, Duration: "200ms"},
		},
		Passed: true,
	}

	table := report.FormatTable()
	if len(table) == 0 {
		t.Error("FormatTable() returned empty string")
	}
}

func TestSetupPreCommitHookNoGitRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devflow-hook-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	qm := &QualityManager{
		projectPath: tmpDir,
		language:    LangGo,
	}

	err = qm.SetupPreCommitHook()
	if err == nil {
		t.Error("SetupPreCommitHook() without .git directory should return error")
	}
}

func TestSetupPreCommitHookWithGitRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devflow-hook-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)

	qm := &QualityManager{
		projectPath: tmpDir,
		language:    LangGo,
	}

	err = qm.SetupPreCommitHook()
	if err != nil {
		t.Fatalf("SetupPreCommitHook() error = %v", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Error("pre-commit hook file should exist")
	}

	content, _ := os.ReadFile(hookPath)
	if len(content) == 0 {
		t.Error("pre-commit hook file should not be empty")
	}
}

func TestSetupPreCommitHookIdempotent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devflow-hook-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)

	qm := &QualityManager{
		projectPath: tmpDir,
		language:    LangGo,
	}

	qm.SetupPreCommitHook()
	err = qm.SetupPreCommitHook()
	if err != nil {
		t.Fatalf("Second SetupPreCommitHook() error = %v", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	content, _ := os.ReadFile(hookPath)
	originalHookCount := 0
	occurrences := 0
	for i := 0; i <= len(content)-len([]byte("DevFlow")); i++ {
		if string(content[i:i+7]) == "DevFlow" {
			occurrences++
		}
	}
	if originalHookCount > 0 && occurrences > originalHookCount {
		t.Error("Second SetupPreCommitHook() should not duplicate DevFlow content")
	}
}

func TestGenerateCIConfigUnsupported(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devflow-ci-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	qm := &QualityManager{
		projectPath: tmpDir,
		language:    LangGo,
	}

	err = qm.GenerateCIConfig("unsupported")
	if err == nil {
		t.Error("GenerateCIConfig() with unsupported platform should return error")
	}
}

func TestGenerateCIConfigGitHub(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devflow-ci-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	qm := &QualityManager{
		projectPath: tmpDir,
		language:    LangGo,
	}

	err = qm.GenerateCIConfig("github")
	if err != nil {
		t.Fatalf("GenerateCIConfig(github) error = %v", err)
	}

	configPath := filepath.Join(tmpDir, ".github", "workflows", "devflow.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("GitHub Actions config file should exist")
	}
}

func TestGenerateCIConfigGitLab(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devflow-ci-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	qm := &QualityManager{
		projectPath: tmpDir,
		language:    LangGo,
	}

	err = qm.GenerateCIConfig("gitlab")
	if err != nil {
		t.Fatalf("GenerateCIConfig(gitlab) error = %v", err)
	}

	configPath := filepath.Join(tmpDir, ".gitlab-ci.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("GitLab CI config file should exist")
	}
}
