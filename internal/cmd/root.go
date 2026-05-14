package cmd

import (
	"fmt"
	"os"

	"github.com/devflow/devflow/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "devflow",
		Short: "DevFlow - 开发者工作流管理工具",
		Long: `DevFlow 是一款面向开发者的命令行工作流工具，帮助您：
- 快速初始化项目
- 管理项目配置
- 运行和管理开发脚本
- 处理环境变量和密钥`,
	}
)

var (
	version string
	commit  string
	date    string
)

func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built at: %s)", v, c, d)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	
	rootCmd.PersistentFlags().BoolP("no-color", "n", false, "禁用彩色输出")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "显示详细输出")
	
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newEnvCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newCompletionCmd())
	
	rootCmd.SetVersionTemplate(`{{printf "DevFlow CLI v%s" .Version}}
`)
}

func initConfig() {
	noColor, _ := rootCmd.Flags().GetBool("no-color")
	if noColor {
		color.NoColor = true
	}
}
