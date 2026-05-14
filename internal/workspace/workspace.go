package workspace

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/devflow/devflow/internal/git"
	"github.com/devflow/devflow/internal/ui"
	"gopkg.in/yaml.v3"
)

type Project struct {
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path"`
	URL      string   `yaml:"url,omitempty"`
	Branch   string   `yaml:"branch,omitempty"`
	DependsOn []string `yaml:"dependsOn,omitempty"`
}

type WorkspaceConfig struct {
	WorkspaceName string    `yaml:"workspaceName"`
	Projects      []Project `yaml:"projects"`
}

type ProjectStatus struct {
	Project    Project
	Status     git.RepoStatus
	HasError   bool
	Error      error
}

type WorkspaceManager struct {
	Config *WorkspaceConfig
	Path   string
}

const workspaceConfigFile = ".devflow.workspace.yml"

func NewWorkspaceManager() *WorkspaceManager {
	return &WorkspaceManager{}
}

func (wm *WorkspaceManager) Load() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("无法获取当前目录: %w", err)
	}

	configPath, err := findWorkspaceConfig(wd)
	if err != nil {
		return fmt.Errorf("未找到工作区配置，请先运行 'devflow workspace init'")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config WorkspaceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	wm.Config = &config
	wm.Path = filepath.Dir(configPath)
	return nil
}

func findWorkspaceConfig(startDir string) (string, error) {
	current := startDir
	for {
		configPath := filepath.Join(current, workspaceConfigFile)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("未找到工作区配置")
		}
		current = parent
	}
}

func (wm *WorkspaceManager) Init(name string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("无法获取当前目录: %w", err)
	}

	configPath := filepath.Join(wd, workspaceConfigFile)
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("工作区配置已存在: %s", configPath)
	}

	if name == "" {
		name = filepath.Base(wd)
	}

	config := WorkspaceConfig{
		WorkspaceName: name,
		Projects:      []Project{},
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	wm.Config = &config
	wm.Path = wd
	return nil
}

func (wm *WorkspaceManager) AddProject(name, path, url string, dependsOn []string) error {
	if wm.Config == nil {
		return fmt.Errorf("工作区未加载")
	}

	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(wm.Path, path)
	}

	if _, err := os.Stat(absPath); err != nil {
		if url == "" {
			return fmt.Errorf("路径不存在且未提供仓库 URL: %s", path)
		}
	}

	project := Project{
		Name:      name,
		Path:      path,
		URL:       url,
		DependsOn: dependsOn,
	}

	wm.Config.Projects = append(wm.Config.Projects, project)
	return wm.Save()
}

func (wm *WorkspaceManager) Save() error {
	if wm.Config == nil {
		return fmt.Errorf("工作区未加载")
	}

	configPath := filepath.Join(wm.Path, workspaceConfigFile)
	data, err := yaml.Marshal(wm.Config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

func (wm *WorkspaceManager) CloneMissingProjects(interactive bool) error {
	if wm.Config == nil {
		return fmt.Errorf("工作区未加载")
	}

	for _, project := range wm.Config.Projects {
		absPath := project.Path
		if !filepath.IsAbs(absPath) {
			absPath = filepath.Join(wm.Path, absPath)
		}

		if _, err := os.Stat(absPath); err == nil {
			continue
		}

		if project.URL == "" {
			ui.Warning(fmt.Sprintf("项目 %s 不存在且无克隆 URL，跳过", project.Name))
			continue
		}

		if interactive {
			reader := bufio.NewReader(os.Stdin)
			ui.Prompt(fmt.Sprintf("克隆项目 %s (%s) 到 %s? [Y/n]", project.Name, project.URL, absPath))
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response == "n" {
				continue
			}
		}

		ui.Info(fmt.Sprintf("正在克隆项目: %s", project.Name))
		gitClient := git.NewGitClient()
		if err := gitClient.Clone(project.URL, absPath); err != nil {
			return fmt.Errorf("克隆项目 %s 失败: %w", project.Name, err)
		}
		ui.Success(fmt.Sprintf("项目 %s 克隆完成", project.Name))
	}

	return nil
}

func (wm *WorkspaceManager) GetStatus() []ProjectStatus {
	if wm.Config == nil {
		return nil
	}

	var wg sync.WaitGroup
	statuses := make([]ProjectStatus, len(wm.Config.Projects))
	sem := make(chan struct{}, 5)

	for i, project := range wm.Config.Projects {
		wg.Add(1)
		go func(index int, p Project) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			absPath := p.Path
			if !filepath.IsAbs(absPath) {
				absPath = filepath.Join(wm.Path, absPath)
			}

			gitClient := git.NewGitClient()
			gitClient.RepoPath = absPath

			if !gitClient.IsGitRepo() {
				statuses[index] = ProjectStatus{
					Project:  p,
					HasError: true,
					Error:    fmt.Errorf("不是 Git 仓库"),
				}
				return
			}

			status, err := gitClient.Status()
			statuses[index] = ProjectStatus{
				Project:  p,
				Status:   status,
				HasError: err != nil,
				Error:    err,
			}
		}(i, project)
	}

	wg.Wait()
	return statuses
}

func (wm *WorkspaceManager) Sync() error {
	if wm.Config == nil {
		return fmt.Errorf("工作区未加载")
	}

	var wg sync.WaitGroup
	errors := make(chan error, len(wm.Config.Projects))
	sem := make(chan struct{}, 3)

	for _, project := range wm.Config.Projects {
		wg.Add(1)
		go func(p Project) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			absPath := p.Path
			if !filepath.IsAbs(absPath) {
				absPath = filepath.Join(wm.Path, absPath)
			}

			gitClient := git.NewGitClient()
			gitClient.RepoPath = absPath

			if !gitClient.IsGitRepo() {
				errors <- fmt.Errorf("项目 %s: 不是 Git 仓库", p.Name)
				return
			}

			ui.Info(fmt.Sprintf("正在同步项目: %s", p.Name))
			if err := gitClient.Pull(); err != nil {
				errors <- fmt.Errorf("项目 %s 同步失败: %w", p.Name, err)
				return
			}
			ui.Success(fmt.Sprintf("项目 %s 同步完成", p.Name))
		}(project)
	}

	wg.Wait()
	close(errors)

	var errorList []string
	for err := range errors {
		errorList = append(errorList, err.Error())
	}

	if len(errorList) > 0 {
		return fmt.Errorf("部分项目同步失败:\n%s", strings.Join(errorList, "\n"))
	}

	return nil
}
