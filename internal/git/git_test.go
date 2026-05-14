package git

import (
	"testing"
)

func TestRepoStatusIsProtected(t *testing.T) {
	status := RepoStatus{}

	tests := []struct {
		branch   string
		expected bool
	}{
		{"main", true},
		{"master", true},
		{"develop", true},
		{"release", true},
		{"production", true},
		{"feature/login", false},
		{"hotfix/fix-bug", false},
		{"my-branch", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			result := status.IsProtected(tt.branch)
			if result != tt.expected {
				t.Errorf("IsProtected(%q) = %v, want %v", tt.branch, result, tt.expected)
			}
		})
	}
}

func TestGitClientIsProtectedBranch(t *testing.T) {
	client := &GitClient{}

	tests := []struct {
		branch   string
		expected bool
	}{
		{"main", true},
		{"master", true},
		{"develop", true},
		{"release", true},
		{"production", true},
		{"feature/auth", false},
		{"bugfix/issue-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			result := client.IsProtectedBranch(tt.branch)
			if result != tt.expected {
				t.Errorf("IsProtectedBranch(%q) = %v, want %v", tt.branch, result, tt.expected)
			}
		})
	}
}
