package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTemplate(t *testing.T) {
	tmpl := NewTemplate("test-project", "backend", "Test Author", "MIT", "go")

	if tmpl.ProjectName != "test-project" {
		t.Errorf("ProjectName = %v, want test-project", tmpl.ProjectName)
	}

	if tmpl.TemplateType != "backend" {
		t.Errorf("TemplateType = %v, want backend", tmpl.TemplateType)
	}

	if tmpl.Author != "Test Author" {
		t.Errorf("Author = %v, want Test Author", tmpl.Author)
	}

	if tmpl.License != "MIT" {
		t.Errorf("License = %v, want MIT", tmpl.License)
	}

	if tmpl.Language != "go" {
		t.Errorf("Language = %v, want go", tmpl.Language)
	}
}

func TestProcessTemplate(t *testing.T) {
	tmpl := NewTemplate("my-app", "frontend", "John", "Apache", "javascript")

	content := `# {{PROJECT_NAME}}
Author: {{AUTHOR}}
License: {{LICENSE}}
Language: {{LANGUAGE}}`

	processed := tmpl.processTemplate(content)
	expected := `# my-app
Author: John
License: Apache
Language: javascript`

	if processed != expected {
		t.Errorf("Processed template = %v, want %v", processed, expected)
	}
}

func TestGenerate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devflow-template-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpl := NewTemplate("test-gen", "backend", "Author", "MIT", "go")

	if err := tmpl.Generate(tmpDir); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	expectedFiles := []string{
		"README.md",
		".gitignore",
		"main.go",
		"go.mod",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
		}
	}

	readmePath := filepath.Join(tmpDir, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	if !contains(string(data), "test-gen") {
		t.Error("README.md should contain project name")
	}
}

func TestFrontendTemplate(t *testing.T) {
	tmpl := NewTemplate("frontend-app", "frontend", "Dev", "MIT", "javascript")
	files := tmpl.getFrontendTemplate()

	requiredFiles := []string{
		"src/index.js",
		"src/App.js",
		"public/index.html",
		"package.json",
		"vite.config.js",
	}

	for _, file := range requiredFiles {
		if _, exists := files[file]; !exists {
			t.Errorf("Frontend template missing file: %s", file)
		}
	}
}

func TestBackendTemplate(t *testing.T) {
	tmpl := NewTemplate("backend-app", "backend", "Dev", "MIT", "go")
	files := tmpl.getBackendTemplate()

	requiredFiles := []string{
		"main.go",
		"go.mod",
		"internal/server/server.go",
		"Makefile",
	}

	for _, file := range requiredFiles {
		if _, exists := files[file]; !exists {
			t.Errorf("Backend template missing file: %s", file)
		}
	}
}

func TestFullstackTemplate(t *testing.T) {
	tmpl := NewTemplate("fullstack-app", "fullstack", "Dev", "MIT", "javascript")
	files := tmpl.getFullstackTemplate()

	requiredFiles := []string{
		"server/main.go",
		"server/go.mod",
		"client/package.json",
		"client/src/index.js",
	}

	for _, file := range requiredFiles {
		if _, exists := files[file]; !exists {
			t.Errorf("Fullstack template missing file: %s", file)
		}
	}
}

func TestCLITemplate(t *testing.T) {
	tmpl := NewTemplate("cli-app", "cli", "Dev", "MIT", "go")
	files := tmpl.getCLITemplate()

	requiredFiles := []string{
		"cmd/cli-app/main.go",
		"go.mod",
		"Makefile",
	}

	for _, file := range requiredFiles {
		if _, exists := files[file]; !exists {
			t.Errorf("CLI template missing file: %s", file)
		}
	}
}

func TestLibraryTemplate(t *testing.T) {
	tmpl := NewTemplate("lib-app", "library", "Dev", "MIT", "javascript")
	files := tmpl.getLibraryTemplate()

	requiredFiles := []string{
		"index.js",
		"package.json",
	}

	for _, file := range requiredFiles {
		if _, exists := files[file]; !exists {
			t.Errorf("Library template missing file: %s", file)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || 
		(len(s) > 0 && (s[0:len(substr)] == substr || 
			contains(s[1:], substr))))
}
