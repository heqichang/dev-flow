package cmd

import (
	"fmt"
	"os"

	"github.com/devflow/devflow/internal/quality"
	"github.com/devflow/devflow/internal/ui"
	"github.com/spf13/cobra"
)

func newQualityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quality",
		Short: "代码质量",
		Long: `代码质量检查工具。

可用命令:
  lint     - 运行代码检查
  test     - 运行测试
  check    - 综合检查
  hook     - 安装 pre-commit hook
  ci       - 生成 CI 配置`,
		Aliases: []string{"q"},
	}

	cmd.AddCommand(newLintCmd())
	cmd.AddCommand(newTestCmd())
	cmd.AddCommand(newCheckCmd())
	cmd.AddCommand(newHookCmd())
	cmd.AddCommand(newCICmd())

	return cmd
}

func newLintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "代码检查",
		Long: `运行项目的代码 linter。`,
		Run: runLint,
	}

	return cmd
}

func runLint(cmd *cobra.Command, args []string) {
	qm, err := quality.NewQualityManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Info(fmt.Sprintf("检测到项目语言: %s", qm.GetLanguage()))
	ui.Info("运行代码检查...")

	result := qm.Lint()

	if result.Passed {
		ui.Success(fmt.Sprintf("代码检查通过 (%s)", result.Duration))
	} else {
		ui.Error("代码检查发现问题:")
		if result.Output != "" {
			fmt.Println(result.Output)
		}
		os.Exit(1)
	}
}

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "运行测试",
		Long: `运行项目测试。`,
		Run: runTest,
	}

	cmd.Flags().BoolP("watch", "w", false, "watch 模式")

	return cmd
}

func runTest(cmd *cobra.Command, args []string) {
	watch, _ := cmd.Flags().GetBool("watch")

	qm, err := quality.NewQualityManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Info(fmt.Sprintf("检测到项目语言: %s", qm.GetLanguage()))
	ui.Info("运行测试...")

	result := qm.Test(watch)

	if result.Passed {
		ui.Success(fmt.Sprintf("测试通过 (%s)", result.Duration))
	} else {
		ui.Error("测试失败:")
		if result.Output != "" {
			fmt.Println(result.Output)
		}
		os.Exit(1)
	}
}

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "综合检查",
		Long: `运行综合检查 (lint + test + type check)。`,
		Run: runCheck,
	}

	cmd.Flags().BoolP("json", "j", false, "JSON 格式输出")

	return cmd
}

func runCheck(cmd *cobra.Command, args []string) {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	qm, err := quality.NewQualityManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Header("DevFlow 综合检查")
	report := qm.Check()

	if jsonOutput {
		jsonStr, err := report.FormatJSON()
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		fmt.Println(jsonStr)
	} else {
		fmt.Println(report.FormatTable())
		for _, result := range report.Results {
			if !result.Passed && result.Output != "" {
				ui.Info(fmt.Sprintf("\n%s 详细输出:", result.Name))
				fmt.Println(result.Output)
			}
		}
	}

	if !report.Passed {
		os.Exit(1)
	}
}

func newHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "安装 hook",
		Long: `安装 Git pre-commit hook。`,
		Run: runHook,
	}

	return cmd
}

func runHook(cmd *cobra.Command, args []string) {
	qm, err := quality.NewQualityManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := qm.SetupPreCommitHook(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

func newCICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ci [platform]",
		Short: "生成 CI 配置",
		Long: `生成 CI 配置文件。

支持的平台:
  github   - GitHub Actions
  gitlab   - GitLab CI`,
		Args: cobra.ExactArgs(1),
		Run: runCI,
	}

	return cmd
}

func runCI(cmd *cobra.Command, args []string) {
	platform := args[0]

	qm, err := quality.NewQualityManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := qm.GenerateCIConfig(platform); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}
