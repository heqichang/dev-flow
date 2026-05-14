package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devflow/devflow/internal/config"
	"github.com/devflow/devflow/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "管理项目配置",
		Long: `管理 DevFlow 项目配置。

可用子命令：
  show     - 显示当前配置
  get      - 获取单个配置项
  set      - 设置单个配置项
  validate - 验证配置文件`,
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigValidateCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "显示当前配置",
		Run: func(cmd *cobra.Command, args []string) {
			envFlag, _ := cmd.Flags().GetString("env")
			env := parseEnvironment(envFlag)
			cfg, err := config.LoadConfig(env)
			if err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			ui.Header(fmt.Sprintf("项目配置 (%s)", env))
			ui.Info(fmt.Sprintf("项目名称: %s", cfg.Project.ProjectName))
			ui.Info(fmt.Sprintf("版本: %s", cfg.Project.Version))
			ui.Info(fmt.Sprintf("语言: %s", cfg.Project.Language))
			ui.Info(fmt.Sprintf("框架: %s", cfg.Project.Framework))
			ui.Info(fmt.Sprintf("作者: %s", cfg.Project.Author))
			ui.Info(fmt.Sprintf("许可证: %s", cfg.Project.License))

			if len(cfg.Project.Scripts) > 0 {
				ui.Info("\n可用脚本:")
				for name, script := range cfg.Project.Scripts {
					ui.Info(fmt.Sprintf("  %s: %s", name, script.Command))
				}
			}

			if len(cfg.Project.Env) > 0 {
				ui.Info("\n环境变量:")
				for key, value := range cfg.Project.Env {
					ui.Info(fmt.Sprintf("  %s=%s", key, value))
				}
			}
		},
	}
	cmd.Flags().StringP("env", "e", "development", "运行环境 (development/staging/production)")
	return cmd
}

func newConfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [key]",
		Short: "获取单个配置项",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			envFlag, _ := cmd.Flags().GetString("env")
			env := parseEnvironment(envFlag)
			cfg, err := config.LoadConfig(env)
			if err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			key := args[0]
			var value string

			switch key {
			case "projectName":
				value = cfg.Project.ProjectName
			case "version":
				value = cfg.Project.Version
			case "language":
				value = cfg.Project.Language
			case "framework":
				value = cfg.Project.Framework
			case "author":
				value = cfg.Project.Author
			case "license":
				value = cfg.Project.License
			default:
				ui.Error(fmt.Sprintf("未知的配置项: %s", key))
				os.Exit(1)
			}

			fmt.Println(value)
		},
	}
	cmd.Flags().StringP("env", "e", "development", "运行环境 (development/staging/production)")
	return cmd
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "设置单个配置项",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]

			wd, err := os.Getwd()
			if err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			configPath := filepath.Join(wd, ".devflow.yml")
			data, err := os.ReadFile(configPath)
			if err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			var cfg config.ProjectConfig
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				ui.Error(fmt.Sprintf("解析配置文件失败: %v", err))
				os.Exit(1)
			}

			switch key {
			case "projectName":
				cfg.ProjectName = value
			case "version":
				cfg.Version = value
			case "language":
				cfg.Language = value
			case "framework":
				cfg.Framework = value
			case "author":
				cfg.Author = value
			case "license":
				cfg.License = value
			default:
				ui.Error(fmt.Sprintf("未知的配置项: %s", key))
				os.Exit(1)
			}

			if err := cfg.Validate(); err != nil {
				ui.Error(fmt.Sprintf("配置验证失败: %v", err))
				os.Exit(1)
			}

			if err := config.SaveProjectConfig(&cfg, configPath); err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			ui.Success(fmt.Sprintf("已设置 %s = %s", key, value))
		},
	}
}

func newConfigValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "验证配置文件",
		Run: func(cmd *cobra.Command, args []string) {
			envFlag, _ := cmd.Flags().GetString("env")
			env := parseEnvironment(envFlag)
			_, err := config.LoadConfig(env)
			if err != nil {
				ui.Error(fmt.Sprintf("配置验证失败 (%s): %v", env, err))
				os.Exit(1)
			}

			ui.Success(fmt.Sprintf("配置文件验证通过 (%s)", env))
		},
	}
	cmd.Flags().StringP("env", "e", "development", "运行环境 (development/staging/production)")
	return cmd
}

func parseEnvironment(envFlag string) config.Environment {
	switch envFlag {
	case "development", "dev":
		return config.Development
	case "staging":
		return config.Staging
	case "production", "prod":
		return config.Production
	default:
		ui.Error(fmt.Sprintf("无效的环境: %s", envFlag))
		os.Exit(1)
		return config.Development
	}
}
