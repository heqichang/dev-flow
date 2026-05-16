package gitflow

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/devflow/devflow/internal/git"
	"github.com/devflow/devflow/internal/ui"
)

type FlowType string

const (
	FlowFeature FlowType = "feature"
	FlowHotfix  FlowType = "hotfix"
	FlowRelease FlowType = "release"
)

type CommitType struct {
	Type        string
	Description string
}

type ConventionalCommit struct {
	Type    string
	Scope   string
	Subject string
	Body    string
}

var CommitTypes = []CommitType{
	{"feat", "新功能"},
	{"fix", "修复 bug"},
	{"docs", "文档更新"},
	{"style", "代码格式"},
	{"refactor", "重构"},
	{"perf", "性能优化"},
	{"test", "测试"},
	{"chore", "构建/工具"},
	{"build", "构建系统"},
	{"ci", "CI 配置"},
	{"revert", "回滚"},
}

var conventionalRegex = regexp.MustCompile(`^(\w+)(\(([^)]+)\))?!?:\s*(.+)$`)

type GitFlowManager struct {
	GitClient *git.GitClient
}

func NewGitFlowManager() *GitFlowManager {
	return &GitFlowManager{
		GitClient: git.NewGitClient(),
	}
}

func (g *GitFlowManager) StartFlow(flowType FlowType, name string) (string, error) {
	if !g.GitClient.IsGitRepo() {
		return "", fmt.Errorf("当前目录不是 Git 仓库")
	}

	currentBranch, err := g.GitClient.GetCurrentBranch()
	if err != nil {
		return "", err
	}

	if g.GitClient.IsProtectedBranch(currentBranch) {
		ui.Warning(fmt.Sprintf("当前在保护分支 %s，请确保已同步最新代码", currentBranch))
	}

	baseBranch := "develop"
	switch flowType {
	case FlowHotfix:
		baseBranch = "main"
	case FlowRelease:
		baseBranch = "main"
	}

	branchName := fmt.Sprintf("%s/%s", flowType, name)

	ui.Info(fmt.Sprintf("从 %s 创建分支 %s", baseBranch, branchName))

	if err := g.GitClient.CreateBranchFrom(baseBranch, branchName); err != nil {
		return "", err
	}

	ui.Success(fmt.Sprintf("分支 %s 创建成功", branchName))
	return branchName, nil
}

func (g *GitFlowManager) FinishFlow(branchName string, deleteBranch bool) error {
	if !g.GitClient.IsGitRepo() {
		return fmt.Errorf("当前目录不是 Git 仓库")
	}

	sourceBranch := branchName
	if sourceBranch == "" {
		currentBranch, err := g.GitClient.GetCurrentBranch()
		if err != nil {
			return err
		}
		sourceBranch = currentBranch
	}

	if !strings.HasPrefix(sourceBranch, "feature/") &&
		!strings.HasPrefix(sourceBranch, "hotfix/") &&
		!strings.HasPrefix(sourceBranch, "release/") {
		return fmt.Errorf("分支 %s 不是功能分支", sourceBranch)
	}

	var targetBranch string
	if strings.HasPrefix(sourceBranch, "feature/") {
		targetBranch = "develop"
	} else {
		targetBranch = "main"
	}

	ui.Info(fmt.Sprintf("合并 %s 到 %s", sourceBranch, targetBranch))

	if err := g.GitClient.Checkout(targetBranch); err != nil {
		return err
	}

	if err := g.GitClient.Merge(sourceBranch); err != nil {
		return err
	}

	if deleteBranch {
		ui.Info(fmt.Sprintf("删除分支 %s", sourceBranch))
		if err := g.GitClient.DeleteBranch(sourceBranch, true); err != nil {
			ui.Warning(fmt.Sprintf("删除本地分支失败: %v", err))
		}
	}

	ui.Success(fmt.Sprintf("合并完成"))
	return nil
}

func (g *GitFlowManager) CheckProtectedBranch() error {
	if !g.GitClient.IsGitRepo() {
		return nil
	}

	branch, err := g.GitClient.GetCurrentBranch()
	if err != nil {
		return err
	}

	if g.GitClient.IsProtectedBranch(branch) {
		ui.Warning(fmt.Sprintf("警告: 当前在保护分支 %s 上操作", branch))
		ui.Warning("建议在功能分支上进行开发")
	}

	return nil
}

func ParseConventionalCommit(message string) (*ConventionalCommit, error) {
	matches := conventionalRegex.FindStringSubmatch(message)
	if matches == nil {
		return nil, fmt.Errorf("不符合 Conventional Commits 规范")
	}

	commit := &ConventionalCommit{
		Type:    matches[1],
		Scope:   matches[3],
		Subject: matches[4],
	}

	return commit, nil
}

func ValidateCommit(message string) error {
	_, err := ParseConventionalCommit(message)
	return err
}

func (g *GitFlowManager) InteractiveCommit() error {
	reader := bufio.NewReader(os.Stdin)

	ui.Info("Conventional Commits 引导")
	ui.Info("可用的提交类型:")
	for i, ct := range CommitTypes {
		ui.Info(fmt.Sprintf("  %d. %s - %s", i+1, ct.Type, ct.Description))
	}

	ui.Promptf("选择提交类型 (1-%d):", len(CommitTypes))
	typeInput, _ := reader.ReadString('\n')
	typeInput = strings.TrimSpace(typeInput)

	var commitType string
	for i, ct := range CommitTypes {
		if typeInput == ct.Type || typeInput == fmt.Sprintf("%d", i+1) {
			commitType = ct.Type
			break
		}
	}

	if commitType == "" {
		return fmt.Errorf("无效的提交类型")
	}

	ui.Prompt("作用域 (可选，直接回车跳过):")
	scope, _ := reader.ReadString('\n')
	scope = strings.TrimSpace(scope)

	ui.Prompt("简短描述:")
	subject, _ := reader.ReadString('\n')
	subject = strings.TrimSpace(subject)

	if subject == "" {
		return fmt.Errorf("描述不能为空")
	}

	ui.Prompt("详细描述 (可选，直接回车跳过):")
	body, _ := reader.ReadString('\n')
	body = strings.TrimSpace(body)

	var message string
	if scope != "" {
		message = fmt.Sprintf("%s(%s): %s", commitType, scope, subject)
	} else {
		message = fmt.Sprintf("%s: %s", commitType, subject)
	}

	ui.Info(fmt.Sprintf("\n提交信息: %s", message))

	if err := g.GitClient.Add("."); err != nil {
		return err
	}

	if err := g.GitClient.Commit(message); err != nil {
		return err
	}

	ui.Success("提交成功！")
	return nil
}

func (g *GitFlowManager) GenerateChangelog(sinceTag string) (string, error) {
	if !g.GitClient.IsGitRepo() {
		return "", fmt.Errorf("当前目录不是 Git 仓库")
	}

	commits, err := g.GitClient.GetLogSince(sinceTag, 0)
	if err != nil {
		return "", err
	}

	changelog := make(map[string][]git.CommitInfo)
	for _, commit := range commits {
		parsed, err := ParseConventionalCommit(commit.Message)
		if err != nil {
			continue
		}
		changelog[parsed.Type] = append(changelog[parsed.Type], commit)
	}

	var output strings.Builder
	header := "Changelog"
	if sinceTag != "" {
		header = fmt.Sprintf("Changelog (since %s)", sinceTag)
	}
	output.WriteString(fmt.Sprintf("# %s - %s\n\n", header, time.Now().Format("2006-01-02")))

	typeOrder := []string{"feat", "fix", "docs", "refactor", "perf", "test", "chore"}
	typeNames := map[string]string{
		"feat":     "新功能",
		"fix":      "修复",
		"docs":     "文档",
		"refactor": "重构",
		"perf":     "性能",
		"test":     "测试",
		"chore":    "其他",
	}

	for _, t := range typeOrder {
		if commits, ok := changelog[t]; ok {
			output.WriteString(fmt.Sprintf("\n## %s\n\n", typeNames[t]))
			for _, c := range commits {
				output.WriteString(fmt.Sprintf("- %s (%s)\n", c.Message, c.Hash[:7]))
			}
		}
	}

	return output.String(), nil
}
