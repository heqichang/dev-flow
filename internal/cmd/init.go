package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devflow/devflow/internal/config"
	"github.com/devflow/devflow/internal/template"
	"github.com/devflow/devflow/internal/ui"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "初始化新项目",
		Long: `交互式初始化一个新项目。

该命令会引导您完成项目配置，选择模板，并生成标准的项目结构。`,
		Run: runInit,
	}

	cmd.Flags().StringP("name", "n", "", "项目名称")
	cmd.Flags().StringP("template", "t", "", "项目模板 (frontend, backend, fullstack, cli, library)")
	cmd.Flags().StringP("language", "l", "", "编程语言")
	cmd.Flags().StringP("author", "a", "", "作者")
	cmd.Flags().StringP("license", "L", "MIT", "许可证")
	cmd.Flags().BoolP("git", "g", true, "初始化 Git 仓库")
	cmd.Flags().BoolP("force", "f", false, "强制覆盖现有目录")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) {
	ui.Header("DevFlow 项目初始化")

	reader := bufio.NewReader(os.Stdin)

	projectName, _ := cmd.Flags().GetString("name")
	templateType, _ := cmd.Flags().GetString("template")
	language, _ := cmd.Flags().GetString("language")
	author, _ := cmd.Flags().GetString("author")
	license, _ := cmd.Flags().GetString("license")
	initGit, _ := cmd.Flags().GetBool("git")
	force, _ := cmd.Flags().GetBool("force")

	if projectName == "" {
		ui.Prompt("项目名称:")
		projectName, _ = reader.ReadString('\n')
		projectName = strings.TrimSpace(projectName)
		if projectName == "" {
			ui.Error("项目名称不能为空")
			os.Exit(1)
		}
	}

	if templateType == "" {
		ui.Info("可用模板:")
		ui.Info("  1. frontend  - 前端项目")
		ui.Info("  2. backend   - 后端项目")
		ui.Info("  3. fullstack - 全栈项目")
		ui.Info("  4. cli       - CLI 工具")
		ui.Info("  5. library   - 库项目")
		ui.Prompt("选择模板 (1-5):")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			templateType = "frontend"
		case "2":
			templateType = "backend"
		case "3":
			templateType = "fullstack"
		case "4":
			templateType = "cli"
		case "5":
			templateType = "library"
		default:
			ui.Error("无效的选择")
			os.Exit(1)
		}
	}

	if language == "" {
		switch templateType {
		case "frontend":
			language = "javascript"
		case "backend":
			language = "go"
		case "fullstack":
			language = "javascript"
		case "cli":
			language = "go"
		case "library":
			language = "javascript"
		}
		ui.Prompt(fmt.Sprintf("编程语言 [%s]:", language))
		input, _ := reader.ReadString('\n')
		if trimmed := strings.TrimSpace(input); trimmed != "" {
			language = trimmed
		}
	}

	if author == "" {
		ui.Prompt("作者:")
		author, _ = reader.ReadString('\n')
		author = strings.TrimSpace(author)
	}

	ui.Info("\n项目配置:")
	ui.Step(1, 5, fmt.Sprintf("项目名称: %s", projectName))
	ui.Step(2, 5, fmt.Sprintf("模板: %s", templateType))
	ui.Step(3, 5, fmt.Sprintf("语言: %s", language))
	ui.Step(4, 5, fmt.Sprintf("作者: %s", author))
	ui.Step(5, 5, fmt.Sprintf("许可证: %s", license))

	targetDir := filepath.Join(".", projectName)
	if _, err := os.Stat(targetDir); err == nil {
		if !force {
			ui.Error(fmt.Sprintf("目录 %s 已存在，请使用 --force 强制覆盖", projectName))
			os.Exit(1)
		}
	}

	spinner := ui.Spinner("正在创建项目结构...")
	spinner.Start()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		spinner.StopWithError(err)
		os.Exit(1)
	}

	cfg := config.DefaultProjectConfig()
	cfg.ProjectName = projectName
	cfg.Language = language
	cfg.Author = author
	cfg.License = license
	cfg.Description = fmt.Sprintf("%s - %s 项目", projectName, templateType)

	tmpl := template.NewTemplate(projectName, templateType, author, license, language)
	if err := tmpl.Generate(targetDir); err != nil {
		spinner.StopWithError(err)
		os.Exit(1)
	}

	configPath := filepath.Join(targetDir, ".devflow.yml")
	if err := config.SaveProjectConfig(cfg, configPath); err != nil {
		spinner.StopWithError(err)
		os.Exit(1)
	}

	spinner.Update("正在初始化 Git 仓库...")
	if initGit {
		if err := initGitRepo(targetDir, projectName); err != nil {
			ui.Warning(fmt.Sprintf("Git 初始化失败: %v", err))
		}
	}

	spinner.Stop()
	ui.Success(fmt.Sprintf("项目 %s 创建成功！", projectName))
	ui.Info(fmt.Sprintf("  目录: %s", targetDir))
	ui.Info(fmt.Sprintf("  配置: %s", configPath))
	ui.Info("")
	ui.Info("下一步:")
	ui.Info(fmt.Sprintf("  cd %s", projectName))
	ui.Info("  devflow help")
}

func initGitRepo(dir, projectName string) error {
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		return err
	}

	readmePath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		readmeContent := fmt.Sprintf("# %s\n\n项目描述", projectName)
		os.WriteFile(readmePath, []byte(readmeContent), 0644)
	}

	return nil
}
