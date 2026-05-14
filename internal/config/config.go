package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

type Script struct {
	Command   string            `yaml:"command" validate:"required"`
	DependsOn []string          `yaml:"dependsOn,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	Timeout   int               `yaml:"timeout,omitempty"`
}

type ProjectConfig struct {
	ProjectName string                 `yaml:"projectName" validate:"required"`
	Version     string                 `yaml:"version,omitempty"`
	Language    string                 `yaml:"language,omitempty"`
	Framework   string                 `yaml:"framework,omitempty"`
	Author      string                 `yaml:"author,omitempty"`
	License     string                 `yaml:"license,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Scripts     map[string]Script      `yaml:"scripts,omitempty"`
	Env         map[string]string      `yaml:"env,omitempty"`
	RequiredEnv []string               `yaml:"requiredEnv,omitempty"`
	Environments map[string]ProjectConfig `yaml:"environments,omitempty"`
}

type GlobalConfig struct {
	DefaultEnv Environment `yaml:"defaultEnv,omitempty"`
}

type Config struct {
	Project ProjectConfig
	Global  GlobalConfig
	Env     Environment
}

var validate = validator.New()

func (c *ProjectConfig) Validate() error {
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}
	return nil
}

func LoadConfig(env Environment) (*Config, error) {
	cfg := &Config{
		Env: env,
	}

	homeDir, _ := os.UserHomeDir()
	globalConfigPath := filepath.Join(homeDir, ".config", "devflow", "config.yml")
	if data, err := os.ReadFile(globalConfigPath); err == nil {
		yaml.Unmarshal(data, &cfg.Global)
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("无法获取当前目录: %w", err)
	}

	projectConfigPath := filepath.Join(wd, ".devflow.yml")
	data, err := os.ReadFile(projectConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("未找到 .devflow.yml 配置文件，请先运行 'devflow init'")
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg.Project); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if envConfig, ok := cfg.Project.Environments[string(env)]; ok {
		mergeConfig(&cfg.Project, &envConfig)
	}

	if err := cfg.Project.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func SaveProjectConfig(cfg *ProjectConfig, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

func mergeConfig(base, override *ProjectConfig) {
	if override.ProjectName != "" {
		base.ProjectName = override.ProjectName
	}
	if override.Version != "" {
		base.Version = override.Version
	}
	if override.Language != "" {
		base.Language = override.Language
	}
	if override.Framework != "" {
		base.Framework = override.Framework
	}
	if override.Author != "" {
		base.Author = override.Author
	}
	if override.License != "" {
		base.License = override.License
	}
	if override.Description != "" {
		base.Description = override.Description
	}

	if len(override.Scripts) > 0 {
		if base.Scripts == nil {
			base.Scripts = make(map[string]Script)
		}
		for name, script := range override.Scripts {
			base.Scripts[name] = script
		}
	}

	if len(override.Env) > 0 {
		if base.Env == nil {
			base.Env = make(map[string]string)
		}
		for key, value := range override.Env {
			base.Env[key] = value
		}
	}

	if len(override.RequiredEnv) > 0 {
		for _, env := range override.RequiredEnv {
			found := false
			for _, e := range base.RequiredEnv {
				if e == env {
					found = true
					break
				}
			}
			if !found {
				base.RequiredEnv = append(base.RequiredEnv, env)
			}
		}
	}
}

func DefaultProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		Version:  "0.1.0",
		License:  "MIT",
		Scripts:  map[string]Script{},
		Env:      map[string]string{},
		Environments: map[string]ProjectConfig{},
	}
}
