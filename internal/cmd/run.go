package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/devflow/devflow/internal/config"
	"github.com/devflow/devflow/internal/ui"
	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [script]",
		Short: "运行项目脚本",
		Long: `运行预定义的项目脚本。

如果不指定脚本名称，将列出所有可用的脚本。

内置脚本:
  dev    - 开发模式运行
  build  - 构建项目
  test   - 运行测试
  lint   - 代码检查
  clean  - 清理临时文件`,
		Args: cobra.MaximumNArgs(1),
		Run:  runScript,
	}

	cmd.Flags().BoolP("parallel", "p", false, "并行执行所有依赖脚本")
	cmd.Flags().StringP("env", "e", "development", "运行环境 (development/staging/production)")
	cmd.Flags().IntP("timeout", "t", 0, "超时时间（秒），0 表示不限制")

	return cmd
}

func runScript(cmd *cobra.Command, args []string) {
	envFlag, _ := cmd.Flags().GetString("env")
	parallel, _ := cmd.Flags().GetBool("parallel")
	timeout, _ := cmd.Flags().GetInt("timeout")

	var env config.Environment
	switch envFlag {
	case "development", "dev":
		env = config.Development
	case "staging":
		env = config.Staging
	case "production", "prod":
		env = config.Production
	default:
		ui.Error(fmt.Sprintf("无效的环境: %s", envFlag))
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(env)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if len(args) == 0 {
		listScripts(cfg)
		return
	}

	scriptName := args[0]
	script, exists := cfg.Project.Scripts[scriptName]
	if !exists {
		ui.Error(fmt.Sprintf("脚本 '%s' 不存在", scriptName))
		ui.Info("可用脚本:")
		listScripts(cfg)
		os.Exit(1)
	}

	if timeout == 0 && script.Timeout > 0 {
		timeout = script.Timeout
	}

	ui.Header(fmt.Sprintf("运行脚本: %s", scriptName))

	if len(script.DependsOn) > 0 {
		ui.Info(fmt.Sprintf("依赖脚本: %s", strings.Join(script.DependsOn, ", ")))
		
		if parallel {
			if err := runScriptsParallel(cfg, script.DependsOn, timeout); err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}
		} else {
			if err := runScriptsSequential(cfg, script.DependsOn, timeout); err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}
		}
	}

	if err := executeScript(scriptName, script, timeout, cfg); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success(fmt.Sprintf("脚本 '%s' 执行完成", scriptName))
}

func listScripts(cfg *config.Config) {
	ui.Header("可用脚本")
	if len(cfg.Project.Scripts) == 0 {
		ui.Info("当前项目没有定义任何脚本")
		return
	}

	for name, script := range cfg.Project.Scripts {
		ui.Info(fmt.Sprintf("  %s: %s", name, script.Command))
		if len(script.DependsOn) > 0 {
			ui.Info(fmt.Sprintf("    依赖: %s", strings.Join(script.DependsOn, ", ")))
		}
	}
}

func executeScript(name string, script config.Script, timeout int, cfg *config.Config) error {
	ui.Step(0, 1, fmt.Sprintf("执行: %s", script.Command))

	var shell string
	var shellArg string

	if runtime.GOOS == "windows" {
		shell = "cmd"
		shellArg = "/C"
	} else {
		shell = "sh"
		shellArg = "-c"
	}

	var cmd *exec.Cmd
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()
		cmd = exec.CommandContext(ctx, shell, shellArg, script.Command)
	} else {
		cmd = exec.Command(shell, shellArg, script.Command)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	for key, value := range cfg.Project.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	for key, value := range script.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("脚本 '%s' 执行失败，退出码: %d", name, exitErr.ExitCode())
		}
		return fmt.Errorf("脚本 '%s' 执行失败: %v", name, err)
	}

	return nil
}

func runScriptsSequential(cfg *config.Config, scripts []string, timeout int) error {
	for _, scriptName := range scripts {
		script, exists := cfg.Project.Scripts[scriptName]
		if !exists {
			return fmt.Errorf("依赖脚本 '%s' 不存在", scriptName)
		}

		if err := executeScript(scriptName, script, timeout, cfg); err != nil {
			return err
		}
	}
	return nil
}

func runScriptsParallel(cfg *config.Config, scripts []string, timeout int) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(scripts))

	for _, scriptName := range scripts {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			script, exists := cfg.Project.Scripts[name]
			if !exists {
				errChan <- fmt.Errorf("依赖脚本 '%s' 不存在", name)
				return
			}

			if err := executeScript(name, script, timeout, cfg); err != nil {
				errChan <- err
			}
		}(scriptName)
	}

	wg.Wait()
	close(errChan)

	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("脚本执行失败:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}
