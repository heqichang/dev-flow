package deps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPackageManager(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected PackageManager
	}{
		{
			name:     "npm with lock",
			files:    []string{"package-lock.json"},
			expected: PMNpm,
		},
		{
			name:     "yarn",
			files:    []string{"yarn.lock"},
			expected: PMYarn,
		},
		{
			name:     "pnpm",
			files:    []string{"pnpm-lock.yaml"},
			expected: PMPnpm,
		},
		{
			name:     "npm without lock",
			files:    []string{"package.json"},
			expected: PMNpm,
		},
		{
			name:     "poetry",
			files:    []string{"poetry.lock"},
			expected: PMPoetry,
		},
		{
			name:     "pip",
			files:    []string{"requirements.txt"},
			expected: PMPip,
		},
		{
			name:     "cargo",
			files:    []string{"Cargo.toml"},
			expected: PMCargo,
		},
		{
			name:     "go mod",
			files:    []string{"go.mod"},
			expected: PMGoMod,
		},
		{
			name:     "maven",
			files:    []string{"pom.xml"},
			expected: PMMaven,
		},
		{
			name:     "unknown defaults to npm",
			files:    []string{},
			expected: PMNpm,
		},
		{
			name:     "npm lock takes priority over package.json",
			files:    []string{"package-lock.json", "package.json"},
			expected: PMNpm,
		},
		{
			name:     "yarn lock takes priority over package.json",
			files:    []string{"yarn.lock", "package.json"},
			expected: PMYarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "devflow-deps-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			for _, f := range tt.files {
				os.WriteFile(filepath.Join(tmpDir, f), []byte{}, 0644)
			}

			result := detectPackageManager(tmpDir)
			if result != tt.expected {
				t.Errorf("detectPackageManager() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDependencyReportFormatJSON(t *testing.T) {
	report := DependencyReport{
		Manager: PMGoMod,
		Count:   2,
		Dependencies: []Dependency{
			{Name: "github.com/spf13/cobra", Current: "v1.8.0", Outdated: false},
			{Name: "github.com/fatih/color", Current: "v1.16.0", Latest: "v1.17.0", Outdated: true},
		},
	}

	jsonStr, err := report.FormatJSON()
	if err != nil {
		t.Fatalf("FormatJSON() error = %v", err)
	}
	if len(jsonStr) == 0 {
		t.Error("FormatJSON() returned empty string")
	}
}

func TestDependencyReportFormatTable(t *testing.T) {
	report := DependencyReport{
		Manager: PMNpm,
		Count:   1,
		Dependencies: []Dependency{
			{Name: "react", Current: "18.0.0", Latest: "18.2.0", Outdated: true},
		},
	}

	table := report.FormatTable()
	if len(table) == 0 {
		t.Error("FormatTable() returned empty string")
	}
}

func TestDependencyReportFormatTableEmpty(t *testing.T) {
	report := DependencyReport{
		Manager:      PMGoMod,
		Count:        0,
		Dependencies: []Dependency{},
	}

	table := report.FormatTable()
	if len(table) == 0 {
		t.Error("FormatTable() should return non-empty string even with no deps")
	}
}
