package cmd

import (
	"fmt"
	"os"

	"github.com/devflow/devflow/internal/deps"
	"github.com/devflow/devflow/internal/ui"
	"github.com/spf13/cobra"
)

func newDepsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deps",
		Short: "依赖管理",
		Long: `项目依赖管理工具。

可用命令:
  list      - 列出依赖
  outdated  - 检查过时依赖
  update    - 更新依赖
  audit     - 安全漏洞检查`,
		Aliases: []string{"dependencies"},
	}

	cmd.AddCommand(newDepsListCmd())
	cmd.AddCommand(newDepsOutdatedCmd())
	cmd.AddCommand(newDepsUpdateCmd())
	cmd.AddCommand(newDepsAuditCmd())

	return cmd
}

func newDepsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出依赖",
		Long: `列出项目的所有依赖。`,
		Run: runDepsList,
	}

	cmd.Flags().BoolP("json", "j", false, "JSON 格式输出")

	return cmd
}

func runDepsList(cmd *cobra.Command, args []string) {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	dm, err := deps.NewDependencyManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Info(fmt.Sprintf("检测到包管理器: %s", dm.GetManager()))

	report := dm.List()

	if jsonOutput {
		jsonStr, err := report.FormatJSON()
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		fmt.Println(jsonStr)
	} else {
		fmt.Println(report.FormatTable())
	}
}

func newDepsOutdatedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "检查过时依赖",
		Long: `检查项目的过时依赖。`,
		Run: runDepsOutdated,
	}

	cmd.Flags().BoolP("json", "j", false, "JSON 格式输出")

	return cmd
}

func runDepsOutdated(cmd *cobra.Command, args []string) {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	dm, err := deps.NewDependencyManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Info(fmt.Sprintf("检测到包管理器: %s", dm.GetManager()))
	ui.Info("检查过时依赖...")

	report := dm.Outdated()

	if jsonOutput {
		jsonStr, err := report.FormatJSON()
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		fmt.Println(jsonStr)
	} else {
		fmt.Println(report.FormatTable())
	}

	outdatedCount := 0
	for _, dep := range report.Dependencies {
		if dep.Outdated {
			outdatedCount++
		}
	}

	if outdatedCount > 0 {
		ui.Warning(fmt.Sprintf("发现 %d 个过时依赖", outdatedCount))
		ui.Info("运行 'devflow deps update' 更新依赖")
	} else {
		ui.Success("所有依赖都是最新的！")
	}
}

func newDepsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "更新依赖",
		Long: `更新项目的依赖。`,
		Run: runDepsUpdate,
	}

	cmd.Flags().BoolP("yes", "y", false, "跳过确认")

	return cmd
}

func runDepsUpdate(cmd *cobra.Command, args []string) {
	autoConfirm, _ := cmd.Flags().GetBool("yes")

	dm, err := deps.NewDependencyManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := dm.Update(!autoConfirm); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

func newDepsAuditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "安全检查",
		Long: `检查依赖的安全漏洞。`,
		Run: runDepsAudit,
	}

	return cmd
}

func runDepsAudit(cmd *cobra.Command, args []string) {
	dm, err := deps.NewDependencyManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := dm.Audit(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}
