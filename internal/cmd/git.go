package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/devflow/devflow/internal/git"
	"github.com/devflow/devflow/internal/gitflow"
	"github.com/devflow/devflow/internal/ui"
	"github.com/spf13/cobra"
)

func newGitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "Git 工作流",
		Long: `Git 工作流自动化命令。

可用命令:
  flow     - Git Flow 操作
  commit   - 交互式提交
  check    - 检查提交规范
  changelog- 生成变更日志`,
	}

	cmd.AddCommand(newGitFlowCmd())
	cmd.AddCommand(newGitCommitCmd())
	cmd.AddCommand(newGitCheckCmd())
	cmd.AddCommand(newGitChangelogCmd())

	return cmd
}

func newGitFlowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flow",
		Short: "Git Flow 操作",
		Long: `Git Flow 分支管理。

可用命令:
  start   - 开始新的功能/修复/发布
  finish  - 完成并合并分支
  pr      - 创建 Pull Request`,
	}

	cmd.AddCommand(newGitFlowStartCmd())
	cmd.AddCommand(newGitFlowFinishCmd())
	cmd.AddCommand(newGitFlowPRCmd())

	return cmd
}

func newGitFlowStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <feature|hotfix|release> <name>",
		Short: "开始新分支",
		Long: `从基准分支创建新的功能分支。`,
		Args: cobra.ExactArgs(2),
		Run: runGitFlowStart,
	}

	return cmd
}

func runGitFlowStart(cmd *cobra.Command, args []string) {
	flowTypeStr := args[0]
	name := args[1]

	var flowType gitflow.FlowType
	switch flowTypeStr {
	case "feature":
		flowType = gitflow.FlowFeature
	case "hotfix":
		flowType = gitflow.FlowHotfix
	case "release":
		flowType = gitflow.FlowRelease
	default:
		ui.Error(fmt.Sprintf("无效的流类型: %s (可用: feature, hotfix, release", flowTypeStr))
		os.Exit(1)
	}

	gfm := gitflow.NewGitFlowManager()
	if _, err := gfm.StartFlow(flowType, name); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

func newGitFlowFinishCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finish",
		Short: "完成分支",
		Long: `合并当前功能分支并删除。`,
		Run: runGitFlowFinish,
	}

	cmd.Flags().BoolP("no-delete", "n", false, "不删除分支")

	return cmd
}

func runGitFlowFinish(cmd *cobra.Command, args []string) {
	noDelete, _ := cmd.Flags().GetBool("no-delete")

	gfm := gitflow.NewGitFlowManager()
	if err := gfm.FinishFlow("", !noDelete); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

func newGitFlowPRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "创建 PR",
		Long: `创建 Pull Request。`,
		Run: runGitFlowPR,
	}

	cmd.Flags().String("title", "", "PR 标题")
	cmd.Flags().String("body", "", "PR 描述")
	cmd.Flags().String("base", "", "目标分支")

	return cmd
}

func runGitFlowPR(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	body, _ := cmd.Flags().GetString("body")
	base, _ := cmd.Flags().GetString("base")

	gitClient := git.NewGitClient()

	if !gitClient.IsGitRepo() {
		ui.Error("当前目录不是 Git 仓库")
		os.Exit(1)
	}

	currentBranch, err := gitClient.GetCurrentBranch()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if title == "" {
		title = currentBranch
	}

	if base == "" {
		if strings.HasPrefix(currentBranch, "feature/") {
			base = "develop"
		} else {
			base = "main"
		}
	}

	remoteURL, err := gitClient.GetRemoteURL("origin")
	if err != nil {
		ui.Error("无法获取远程仓库地址")
		os.Exit(1)
	}

	ui.Info("Pull Request 信息:")
	ui.Info(fmt.Sprintf("  源分支: %s", currentBranch))
	ui.Info(fmt.Sprintf("  目标分支: %s", base))
	ui.Info(fmt.Sprintf("  标题: %s", title))
	ui.Info(fmt.Sprintf("  描述: %s", body))
	ui.Info(fmt.Sprintf("  仓库: %s", remoteURL))
	ui.Info("")
	ui.Warning("PR 创建需要配置 GitHub/GitLab 凭据")
	ui.Info("请手动创建 Pull Request 或配置 API 访问令牌")
	ui.Info(fmt.Sprintf("建议的 PR URL 格式: %s/compare/%s...%s", remoteURL, base, currentBranch))
}

func newGitCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "交互式提交",
		Long: `使用 Conventional Commits 规范交互式创建提交。`,
		Run: runGitCommit,
	}

	return cmd
}

func runGitCommit(cmd *cobra.Command, args []string) {
	gfm := gitflow.NewGitFlowManager()
	if err := gfm.InteractiveCommit(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

func newGitCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "检查提交",
		Long: `检查最后一次提交是否符合 Conventional Commits 规范。`,
		Run: runGitCheck,
	}

	return cmd
}

func runGitCheck(cmd *cobra.Command, args []string) {
	gitClient := git.NewGitClient()

	if !gitClient.IsGitRepo() {
		ui.Error("当前目录不是 Git 仓库")
		os.Exit(1)
	}

	message, err := gitClient.GetLastCommitMessage()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := gitflow.ValidateCommit(message); err != nil {
		ui.Error(fmt.Sprintf("提交不符合规范: %v", err))
		ui.Info(fmt.Sprintf("提交信息: %s", message))
		os.Exit(1)
	}

	ui.Success("提交符合 Conventional Commits 规范")
}

func newGitChangelogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changelog",
		Short: "生成变更日志",
		Long: `基于提交历史生成变更日志。`,
		Run: runGitChangelog,
	}

	cmd.Flags().String("since", "", "从哪个标签开始")
	cmd.Flags().StringP("output", "o", "", "输出文件")

	return cmd
}

func runGitChangelog(cmd *cobra.Command, args []string) {
	since, _ := cmd.Flags().GetString("since")
	output, _ := cmd.Flags().GetString("output")

	gfm := gitflow.NewGitFlowManager()
	changelog, err := gfm.GenerateChangelog(since)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if output != "" {
		if err := os.WriteFile(output, []byte(changelog), 0644); err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		ui.Success(fmt.Sprintf("变更日志已保存到: %s", output))
	} else {
		fmt.Println(changelog)
	}
}
