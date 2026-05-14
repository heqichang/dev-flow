package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/devflow/devflow/internal/git"
	"github.com/devflow/devflow/internal/ui"
	"github.com/devflow/devflow/internal/workspace"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "工作区管理",
		Long: `管理多项目工作区，支持批量操作和项目依赖管理。

可用命令:
  init     - 初始化工作区
  add      - 添加项目到工作区
  status   - 查看所有项目状态
  sync     - 批量同步所有项目
  clone    - 克隆缺失的项目
  list     - 列出工作区项目列表`,
	}

	cmd.AddCommand(newWorkspaceInitCmd())
	cmd.AddCommand(newWorkspaceAddCmd())
	cmd.AddCommand(newWorkspaceStatusCmd())
	cmd.AddCommand(newWorkspaceSyncCmd())
	cmd.AddCommand(newWorkspaceCloneCmd())
	cmd.AddCommand(newWorkspaceListCmd())

	return cmd
}

func newWorkspaceInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "初始化工作区",
		Long: `在当前目录初始化一个新的工作区配置文件。`,
		Run: runWorkspaceInit,
	}

	return cmd
}

func runWorkspaceInit(cmd *cobra.Command, args []string) {
	ui.Header("初始化工作区")

	var name string
	if len(args) > 0 {
		name = args[0]
	}

	wm := workspace.NewWorkspaceManager()
	if err := wm.Init(name); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success(fmt.Sprintf("工作区初始化成功！"))
	ui.Info(fmt.Sprintf("工作区目录: %s", wm.Path))
	ui.Info(fmt.Sprintf("配置文件: %s/.devflow.workspace.yml", wm.Path))
	ui.Info("")
	ui.Info("下一步:")
	ui.Info("  devflow workspace add <name> <path> --url <repo-url>")
}

func newWorkspaceAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <path>",
		Short: "添加项目",
		Long: `添加项目到工作区配置。`,
		Args: cobra.ExactArgs(2),
		Run: runWorkspaceAdd,
	}

	cmd.Flags().String("url", "", "Git 仓库 URL（用于克隆）")
	cmd.Flags().StringSlice("depends-on", []string{}, "依赖的项目名称")

	return cmd
}

func runWorkspaceAdd(cmd *cobra.Command, args []string) {
	name := args[0]
	path := args[1]
	url, _ := cmd.Flags().GetString("url")
	dependsOn, _ := cmd.Flags().GetStringSlice("depends-on")

	wm := workspace.NewWorkspaceManager()
	if err := wm.Load(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := wm.AddProject(name, path, url, dependsOn); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success(fmt.Sprintf("项目 %s 已添加到工作区", name))
}

func newWorkspaceStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "查看状态",
		Long: `查看工作区中所有项目的 Git 状态。`,
		Run: runWorkspaceStatus,
	}

	return cmd
}

func runWorkspaceStatus(cmd *cobra.Command, args []string) {
	wm := workspace.NewWorkspaceManager()
	if err := wm.Load(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Header(fmt.Sprintf("工作区: %s", wm.Config.WorkspaceName))
	ui.Info("项目状态:")
	ui.Info("")

	statuses := wm.GetStatus()

	for _, ps := range statuses {
		if ps.HasError {
			ui.Error(fmt.Sprintf("%s: %v", ps.Project.Name, ps.Error))
			continue
		}

		status := ps.Status
		statusIcon := color.GreenString("✓")
		if status.HasChanges {
			statusIcon = color.YellowString("!")
		}

		branchColor := color.CyanString(status.Branch)
		if status.IsProtected(status.Branch) {
			branchColor = color.RedString(status.Branch)
		}

		ui.Info(fmt.Sprintf("%s %s [%s] %s", statusIcon, ps.Project.Name, branchColor, getStatusIndicator(status)))
	}
}

func getStatusIndicator(status git.RepoStatus) string {
	var parts []string
	if status.HasChanges {
		parts = append(parts, "有未提交更改")
	}
	if status.IsAhead {
		parts = append(parts, "本地领先")
	}
	if status.IsBehind {
		parts = append(parts, "本地落后")
	}
	if len(parts) == 0 {
		parts = append(parts, "干净")
	}
	return strings.Join(parts, ", ")
}

func newWorkspaceSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "同步项目",
		Long: `批量拉取所有项目的最新代码。`,
		Run: runWorkspaceSync,
	}

	return cmd
}

func runWorkspaceSync(cmd *cobra.Command, args []string) {
	ui.Header("同步工作区项目")

	wm := workspace.NewWorkspaceManager()
	if err := wm.Load(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := wm.Sync(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success("所有项目同步完成！")
}

func newWorkspaceCloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "克隆项目",
		Long: `克隆工作区中缺失的项目。`,
		Run: runWorkspaceClone,
	}

	cmd.Flags().BoolP("yes", "y", false, "自动确认所有克隆")

	return cmd
}

func runWorkspaceClone(cmd *cobra.Command, args []string) {
	autoConfirm, _ := cmd.Flags().GetBool("yes")

	wm := workspace.NewWorkspaceManager()
	if err := wm.Load(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := wm.CloneMissingProjects(!autoConfirm); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success("项目克隆完成！")
}

func newWorkspaceListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出项目",
		Long: `列出工作区中的所有项目。`,
		Run: runWorkspaceList,
	}

	return cmd
}

func runWorkspaceList(cmd *cobra.Command, args []string) {
	wm := workspace.NewWorkspaceManager()
	if err := wm.Load(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Header(fmt.Sprintf("工作区: %s", wm.Config.WorkspaceName))
	ui.Info("项目列表:")
	ui.Info("")

	for i, project := range wm.Config.Projects {
		ui.Step(i+1, len(wm.Config.Projects), fmt.Sprintf("%s", project.Name))
		ui.Info(fmt.Sprintf("  路径: %s", project.Path))
		if project.URL != "" {
			ui.Info(fmt.Sprintf("  仓库: %s", project.URL))
		}
		if len(project.DependsOn) > 0 {
			ui.Info(fmt.Sprintf("  依赖: %s", strings.Join(project.DependsOn, ", ")))
		}
	}
}
