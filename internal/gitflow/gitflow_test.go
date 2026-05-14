package gitflow

import (
	"testing"
)

func TestParseConventionalCommit(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		wantType  string
		wantScope string
		wantSubj  string
		wantErr   bool
	}{
		{
			name:      "feat without scope",
			message:   "feat: add login page",
			wantType:  "feat",
			wantScope: "",
			wantSubj:  "add login page",
			wantErr:   false,
		},
		{
			name:      "feat with scope",
			message:   "feat(auth): add login page",
			wantType:  "feat",
			wantScope: "auth",
			wantSubj:  "add login page",
			wantErr:   false,
		},
		{
			name:      "fix with scope",
			message:   "fix(api): handle null response",
			wantType:  "fix",
			wantScope: "api",
			wantSubj:  "handle null response",
			wantErr:   false,
		},
		{
			name:      "breaking change with !",
			message:   "feat!: change API format",
			wantType:  "feat",
			wantScope: "",
			wantSubj:  "change API format",
			wantErr:   false,
		},
		{
			name:      "breaking change with scope and !",
			message:   "feat(core)!: rewrite engine",
			wantType:  "feat",
			wantScope: "core",
			wantSubj:  "rewrite engine",
			wantErr:   false,
		},
		{
			name:    "no type prefix",
			message: "random commit message",
			wantErr: true,
		},
		{
			name:    "missing colon",
			message: "feat add login page",
			wantErr: true,
		},
		{
			name:    "missing subject",
			message: "feat:",
			wantErr: true,
		},
		{
			name:    "empty message",
			message: "",
			wantErr: true,
		},
		{
			name:      "chore commit",
			message:   "chore: update dependencies",
			wantType:  "chore",
			wantScope: "",
			wantSubj:  "update dependencies",
			wantErr:   false,
		},
		{
			name:      "docs commit",
			message:   "docs(readme): update installation guide",
			wantType:  "docs",
			wantScope: "readme",
			wantSubj:  "update installation guide",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commit, err := ParseConventionalCommit(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConventionalCommit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if commit.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", commit.Type, tt.wantType)
			}
			if commit.Scope != tt.wantScope {
				t.Errorf("Scope = %q, want %q", commit.Scope, tt.wantScope)
			}
			if commit.Subject != tt.wantSubj {
				t.Errorf("Subject = %q, want %q", commit.Subject, tt.wantSubj)
			}
		})
	}
}

func TestValidateCommit(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{"valid feat", "feat: add feature", false},
		{"valid fix with scope", "fix(core): fix bug", false},
		{"invalid no type", "just a message", true},
		{"invalid empty", "", true},
		{"valid refactor", "refactor(utils): simplify logic", false},
		{"valid perf", "perf: optimize loop", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommit(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
