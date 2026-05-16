package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitClient struct {
	RepoPath string
}

type BranchInfo struct {
	Name      string
	IsCurrent bool
}

type RepoStatus struct {
	Branch     string
	HasChanges bool
	IsAhead    bool
	IsBehind   bool
	IsDetached bool
	Upstream   string
}

func (r *RepoStatus) IsProtected(branch string) bool {
	protected := []string{"main", "master", "develop", "release", "production"}
	for _, b := range protected {
		if branch == b {
			return true
		}
	}
	return false
}

type CommitInfo struct {
	Hash    string
	Message string
	Author  string
	Date    string
}

func NewGitClient() *GitClient {
	return &GitClient{}
}

func (g *GitClient) runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if g.RepoPath != "" {
		cmd.Dir = g.RepoPath
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func (g *GitClient) IsGitRepo() bool {
	_, err := g.runGit("rev-parse", "--is-inside-work-tree")
	return err == nil
}

func (g *GitClient) Init(repoPath string) error {
	g.RepoPath = repoPath
	_, err := g.runGit("init")
	return err
}

func (g *GitClient) Clone(url, targetDir string) error {
	_, err := g.runGit("clone", url, targetDir)
	return err
}

func (g *GitClient) GetCurrentBranch() (string, error) {
	output, err := g.runGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (g *GitClient) ListBranches() ([]BranchInfo, error) {
	output, err := g.runGit("branch", "--all")
	if err != nil {
		return nil, err
	}

	var branches []BranchInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		isCurrent := strings.HasPrefix(line, "*")
		name := strings.TrimPrefix(line, "* ")
		name = strings.TrimPrefix(name, "remotes/origin/")
		name = strings.TrimSpace(name)

		branches = append(branches, BranchInfo{
			Name:      name,
			IsCurrent: isCurrent,
		})
	}

	return branches, nil
}

func (g *GitClient) CreateBranch(name string) error {
	_, err := g.runGit("checkout", "-b", name)
	return err
}

func (g *GitClient) CreateBranchFrom(source, target string) error {
	_, err := g.runGit("checkout", "-b", target, source)
	return err
}

func (g *GitClient) Checkout(branch string) error {
	_, err := g.runGit("checkout", branch)
	return err
}

func (g *GitClient) DeleteBranch(name string, force bool) error {
	args := []string{"branch", "-d", name}
	if force {
		args[1] = "-D"
	}
	_, err := g.runGit(args...)
	return err
}

func (g *GitClient) Merge(branch string) error {
	_, err := g.runGit("merge", branch)
	return err
}

func (g *GitClient) Pull() error {
	_, err := g.runGit("pull")
	return err
}

func (g *GitClient) Push() error {
	_, err := g.runGit("push")
	return err
}

func (g *GitClient) PushBranch(branch string) error {
	_, err := g.runGit("push", "-u", "origin", branch)
	return err
}

func (g *GitClient) Status() (RepoStatus, error) {
	status := RepoStatus{}

	branch, err := g.GetCurrentBranch()
	if err != nil {
		return status, err
	}
	status.Branch = branch

	if branch == "HEAD" {
		status.IsDetached = true
	}

	statusOutput, err := g.runGit("status", "--porcelain")
	if err != nil {
		return status, err
	}
	status.HasChanges = strings.TrimSpace(statusOutput) != ""

	upstream, _ := g.runGit("rev-parse", "--abbrev-ref", "@{u}")
	status.Upstream = strings.TrimSpace(upstream)

	if status.Upstream != "" {
		localHash, _ := g.runGit("rev-parse", "@")
		remoteHash, _ := g.runGit("rev-parse", "@{u}")
		baseHash, _ := g.runGit("merge-base", "@", "@{u}")

		localHash = strings.TrimSpace(localHash)
		remoteHash = strings.TrimSpace(remoteHash)
		baseHash = strings.TrimSpace(baseHash)

		status.IsAhead = localHash != baseHash
		status.IsBehind = remoteHash != baseHash
	}

	return status, nil
}

func (g *GitClient) Add(files ...string) error {
	args := append([]string{"add"}, files...)
	_, err := g.runGit(args...)
	return err
}

func (g *GitClient) Commit(message string) error {
	_, err := g.runGit("commit", "-m", message)
	return err
}

func (g *GitClient) GetRemoteURL(remote string) (string, error) {
	output, err := g.runGit("remote", "get-url", remote)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (g *GitClient) GetLog(n int) ([]CommitInfo, error) {
	return g.GetLogSince("", n)
}

func (g *GitClient) GetLogSince(sinceTag string, n int) ([]CommitInfo, error) {
	format := "%H|%s|%an|%ad"
	args := []string{"log", fmt.Sprintf("--format=%s", format), "--date=short"}
	if sinceTag != "" {
		args = append(args, fmt.Sprintf("%s..HEAD", sinceTag))
	}
	if n > 0 {
		args = append(args, fmt.Sprintf("-%d", n))
	}
	output, err := g.runGit(args...)
	if err != nil {
		return nil, err
	}

	var commits []CommitInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) == 4 {
			commits = append(commits, CommitInfo{
				Hash:    parts[0],
				Message: parts[1],
				Author:  parts[2],
				Date:    parts[3],
			})
		}
	}

	return commits, nil
}

func (g *GitClient) GetLastCommitMessage() (string, error) {
	output, err := g.runGit("log", "-1", "--format=%s")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (g *GitClient) IsProtectedBranch(branch string) bool {
	protected := []string{"main", "master", "develop", "release", "production"}
	for _, b := range protected {
		if branch == b {
			return true
		}
	}
	return false
}

func (g *GitClient) GetRootPath() (string, error) {
	output, err := g.runGit("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func FindGitRepoRoot(startPath string) (string, error) {
	current := startPath
	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("未找到 Git 仓库")
		}
		current = parent
	}
}
