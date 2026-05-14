package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestWorkspace(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "devflow-workspace-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	teardown := func() {
		os.Chdir(originalWd)
		os.RemoveAll(tmpDir)
	}
	return tmpDir, teardown
}

func TestWorkspaceInit(t *testing.T) {
	tmpDir, teardown := setupTestWorkspace(t)
	defer teardown()

	wm := NewWorkspaceManager()
	if err := wm.Init("test-workspace"); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if wm.Config.WorkspaceName != "test-workspace" {
		t.Errorf("WorkspaceName = %q, want %q", wm.Config.WorkspaceName, "test-workspace")
	}

	configPath := filepath.Join(tmpDir, workspaceConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should exist after Init()")
	}
}

func TestWorkspaceInitDuplicate(t *testing.T) {
	_, teardown := setupTestWorkspace(t)
	defer teardown()

	wm := NewWorkspaceManager()
	wm.Init("test-workspace")

	err := wm.Init("another-workspace")
	if err == nil {
		t.Error("Init() on existing workspace should return error")
	}
}

func TestWorkspaceInitDefaultName(t *testing.T) {
	tmpDir, teardown := setupTestWorkspace(t)
	defer teardown()

	wm := NewWorkspaceManager()
	if err := wm.Init(""); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	expectedName := filepath.Base(tmpDir)
	if wm.Config.WorkspaceName != expectedName {
		t.Errorf("Default WorkspaceName = %q, want %q", wm.Config.WorkspaceName, expectedName)
	}
}

func TestWorkspaceAddProject(t *testing.T) {
	_, teardown := setupTestWorkspace(t)
	defer teardown()

	wm := NewWorkspaceManager()
	wm.Init("test-workspace")

	projectDir := filepath.Join(wm.Path, "project-a")
	os.MkdirAll(projectDir, 0755)

	err := wm.AddProject("project-a", "project-a", "https://github.com/test/a.git", []string{})
	if err != nil {
		t.Fatalf("AddProject() error = %v", err)
	}

	if len(wm.Config.Projects) != 1 {
		t.Fatalf("Projects count = %d, want 1", len(wm.Config.Projects))
	}

	p := wm.Config.Projects[0]
	if p.Name != "project-a" {
		t.Errorf("Project Name = %q, want %q", p.Name, "project-a")
	}
	if p.URL != "https://github.com/test/a.git" {
		t.Errorf("Project URL = %q, want %q", p.URL, "https://github.com/test/a.git")
	}
}

func TestWorkspaceAddProjectNoPathNoURL(t *testing.T) {
	_, teardown := setupTestWorkspace(t)
	defer teardown()

	wm := NewWorkspaceManager()
	wm.Init("test-workspace")

	err := wm.AddProject("missing", "nonexistent-path", "", []string{})
	if err == nil {
		t.Error("AddProject() with nonexistent path and no URL should return error")
	}
}

func TestWorkspaceLoad(t *testing.T) {
	_, teardown := setupTestWorkspace(t)
	defer teardown()

	wm := NewWorkspaceManager()
	wm.Init("test-workspace")

	projectDir := filepath.Join(wm.Path, "project-a")
	os.MkdirAll(projectDir, 0755)
	wm.AddProject("project-a", "project-a", "", []string{})

	wm2 := NewWorkspaceManager()
	if err := wm2.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if wm2.Config.WorkspaceName != "test-workspace" {
		t.Errorf("Loaded WorkspaceName = %q, want %q", wm2.Config.WorkspaceName, "test-workspace")
	}
	if len(wm2.Config.Projects) != 1 {
		t.Errorf("Loaded Projects count = %d, want 1", len(wm2.Config.Projects))
	}
}

func TestWorkspaceLoadNotFound(t *testing.T) {
	tmpDir, teardown := setupTestWorkspace(t)
	defer teardown()

	emptyDir := filepath.Join(tmpDir, "empty")
	os.MkdirAll(emptyDir, 0755)
	os.Chdir(emptyDir)

	wm := NewWorkspaceManager()
	err := wm.Load()
	if err == nil {
		t.Error("Load() without config file should return error")
	}
}

func TestWorkspaceSave(t *testing.T) {
	_, teardown := setupTestWorkspace(t)
	defer teardown()

	wm := NewWorkspaceManager()
	wm.Init("test-workspace")

	projectDir := filepath.Join(wm.Path, "project-a")
	os.MkdirAll(projectDir, 0755)
	wm.AddProject("project-a", "project-a", "", []string{})

	wm2 := NewWorkspaceManager()
	wm2.Load()

	if len(wm2.Config.Projects) != 1 {
		t.Errorf("After save/load, Projects count = %d, want 1", len(wm2.Config.Projects))
	}
}

func TestFindWorkspaceConfig(t *testing.T) {
	tmpDir, teardown := setupTestWorkspace(t)
	defer teardown()

	wm := NewWorkspaceManager()
	wm.Init("test-workspace")

	configPath, err := findWorkspaceConfig(tmpDir)
	if err != nil {
		t.Fatalf("findWorkspaceConfig() error = %v", err)
	}

	expected := filepath.Join(tmpDir, workspaceConfigFile)
	if configPath != expected {
		t.Errorf("Config path = %q, want %q", configPath, expected)
	}
}

func TestFindWorkspaceConfigParentDir(t *testing.T) {
	tmpDir, teardown := setupTestWorkspace(t)
	defer teardown()

	wm := NewWorkspaceManager()
	wm.Init("test-workspace")

	subDir := filepath.Join(tmpDir, "sub", "deep")
	os.MkdirAll(subDir, 0755)

	configPath, err := findWorkspaceConfig(subDir)
	if err != nil {
		t.Fatalf("findWorkspaceConfig() from subdirectory error = %v", err)
	}

	expected := filepath.Join(tmpDir, workspaceConfigFile)
	if configPath != expected {
		t.Errorf("Config path = %q, want %q", configPath, expected)
	}
}

func TestFindWorkspaceConfigNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "devflow-no-workspace")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = findWorkspaceConfig(tmpDir)
	if err == nil {
		t.Error("findWorkspaceConfig() without config should return error")
	}
}
