package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/devflow/devflow/internal/config"
	"github.com/devflow/devflow/internal/ui"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "管理环境变量",
		Long: `管理项目环境变量和 .env 文件。

可用子命令：
  list     - 列出所有环境变量
  get      - 获取单个环境变量
  set      - 设置环境变量
  switch   - 切换到不同的环境文件`,
	}

	cmd.AddCommand(newEnvListCmd())
	cmd.AddCommand(newEnvGetCmd())
	cmd.AddCommand(newEnvSetCmd())
	cmd.AddCommand(newEnvSwitchCmd())
	cmd.AddCommand(newEnvCheckCmd())

	return cmd
}

func newEnvListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有环境变量",
		Run: func(cmd *cobra.Command, args []string) {
			wd, err := os.Getwd()
			if err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			envPath := filepath.Join(wd, ".env")
			if _, err := os.Stat(envPath); os.IsNotExist(err) {
				ui.Warning("未找到 .env 文件")
				return
			}

			env, err := godotenv.Read(envPath)
			if err != nil {
				ui.Error(fmt.Sprintf("读取 .env 文件失败: %v", err))
				os.Exit(1)
			}

			ui.Header("环境变量")
			if len(env) == 0 {
				ui.Info(".env 文件为空")
				return
			}

			keys := make([]string, 0, len(env))
			for k := range env {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, key := range keys {
				ui.Info(fmt.Sprintf("  %s=%s", key, env[key]))
			}

			ui.Info(fmt.Sprintf("\n共 %d 个环境变量", len(env)))
		},
	}
}

func newEnvGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "获取单个环境变量",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			wd, err := os.Getwd()
			if err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			envPath := filepath.Join(wd, ".env")
			if _, err := os.Stat(envPath); os.IsNotExist(err) {
				ui.Error("未找到 .env 文件")
				os.Exit(1)
			}

			env, err := godotenv.Read(envPath)
			if err != nil {
				ui.Error(fmt.Sprintf("读取 .env 文件失败: %v", err))
				os.Exit(1)
			}

			key := args[0]
			value, exists := env[key]
			if !exists {
				ui.Error(fmt.Sprintf("环境变量 '%s' 不存在", key))
				os.Exit(1)
			}

			fmt.Println(value)
		},
	}
}

func newEnvSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "设置环境变量",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]

			wd, err := os.Getwd()
			if err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			envPath := filepath.Join(wd, ".env")
			var env map[string]string

			if _, err := os.Stat(envPath); os.IsNotExist(err) {
				env = make(map[string]string)
			} else {
				env, err = godotenv.Read(envPath)
				if err != nil {
					ui.Error(fmt.Sprintf("读取 .env 文件失败: %v", err))
					os.Exit(1)
				}
			}

			env[key] = value

			if err := writeEnvFile(envPath, env); err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			ui.Success(fmt.Sprintf("已设置 %s=%s", key, value))
		},
	}
}

func newEnvSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch [environment]",
		Short: "切换到不同的环境文件",
		Args:  cobra.ExactArgs(1),
		Long: `切换到指定的环境文件。

例如:
  devflow env switch development  - 使用 .env.development
  devflow env switch production   - 使用 .env.production`,
		Run: func(cmd *cobra.Command, args []string) {
			envName := args[0]
			wd, err := os.Getwd()
			if err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			sourcePath := filepath.Join(wd, fmt.Sprintf(".env.%s", envName))
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				ui.Error(fmt.Sprintf("环境文件 .env.%s 不存在", envName))
				os.Exit(1)
			}

			targetPath := filepath.Join(wd, ".env")
			data, err := os.ReadFile(sourcePath)
			if err != nil {
				ui.Error(fmt.Sprintf("读取环境文件失败: %v", err))
				os.Exit(1)
			}

			if err := os.WriteFile(targetPath, data, 0644); err != nil {
				ui.Error(fmt.Sprintf("写入 .env 文件失败: %v", err))
				os.Exit(1)
			}

			ui.Success(fmt.Sprintf("已切换到 %s 环境", envName))
		},
	}
}

func newEnvCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "检查必需的环境变量",
		Run: func(cmd *cobra.Command, args []string) {
			wd, err := os.Getwd()
			if err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			devflowConfigPath := filepath.Join(wd, ".devflow.yml")
			if _, err := os.Stat(devflowConfigPath); os.IsNotExist(err) {
				ui.Warning("未找到 .devflow.yml 文件，无法检查必需环境变量")
				return
			}

			data, err := os.ReadFile(devflowConfigPath)
			if err != nil {
				ui.Error(fmt.Sprintf("读取配置文件失败: %v", err))
				os.Exit(1)
			}

			var cfg config.ProjectConfig
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				ui.Error(fmt.Sprintf("解析配置文件失败: %v", err))
				os.Exit(1)
			}

			if len(cfg.RequiredEnv) == 0 {
				ui.Info("未配置必需的环境变量 (requiredEnv)")
				return
			}

			envPath := filepath.Join(wd, ".env")
			var env map[string]string
			if _, err := os.Stat(envPath); os.IsNotExist(err) {
				env = make(map[string]string)
			} else {
				env, err = godotenv.Read(envPath)
				if err != nil {
					ui.Error(fmt.Sprintf("读取 .env 文件失败: %v", err))
					os.Exit(1)
				}
			}

			ui.Info("检查必需环境变量...")
			missing := []string{}

			for _, requiredKey := range cfg.RequiredEnv {
				if _, exists := env[requiredKey]; exists {
					ui.Success(fmt.Sprintf("✓ %s 已设置", requiredKey))
				} else {
					ui.Error(fmt.Sprintf("✗ %s 未设置", requiredKey))
					missing = append(missing, requiredKey)
				}
			}

			if len(missing) == 0 {
				ui.Success("所有必需环境变量已设置")
			} else {
				ui.Error(fmt.Sprintf("缺少以下必需环境变量: %s", strings.Join(missing, ", ")))
				os.Exit(1)
			}
		},
	}
}

func writeEnvFile(path string, env map[string]string) error {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var content strings.Builder
	for _, key := range keys {
		value := env[key]
		if strings.Contains(value, " ") || strings.Contains(value, "\"") || strings.Contains(value, "'") {
			value = fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
		}
		content.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	return os.WriteFile(path, []byte(content.String()), 0644)
}

func readEnvFile(path string) (map[string]string, error) {
	env := make(map[string]string)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return env, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"'")
			env[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return env, nil
}
