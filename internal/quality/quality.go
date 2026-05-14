package quality

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/devflow/devflow/internal/ui"
)

type ProjectLanguage string

const (
	LangGo         ProjectLanguage = "go"
	LangJavaScript ProjectLanguage = "javascript"
	LangTypeScript ProjectLanguage = "typescript"
	LangPython     ProjectLanguage = "python"
	LangRust       ProjectLanguage = "rust"
	LangJava       ProjectLanguage = "java"
	LangUnknown    ProjectLanguage = "unknown"
)

type CheckResult struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Output   string `json:"output,omitempty"`
	Duration string `json:"duration,omitempty"`
}

type CheckReport struct {
	Language ProjectLanguage `json:"language"`
	Results  []CheckResult   `json:"results"`
	Passed   bool            `json:"passed"`
}

type QualityManager struct {
	projectPath string
	language    ProjectLanguage
}

func NewQualityManager() (*QualityManager, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("无法获取当前目录: %w", err)
	}

	lang := detectLanguage(wd)

	return &QualityManager{
		projectPath: wd,
		language:    lang,
	}, nil
}

func detectLanguage(path string) ProjectLanguage {
	markers := map[string]ProjectLanguage{
		"go.mod":       LangGo,
		"package.json": LangJavaScript,
		"requirements.txt": LangPython,
		"setup.py":     LangPython,
		"Cargo.toml":   LangRust,
		"pom.xml":      LangJava,
	}

	for file, lang := range markers {
		if _, err := os.Stat(filepath.Join(path, file)); err == nil {
			return lang
		}
	}

	if _, err := os.Stat(filepath.Join(path, "tsconfig.json")); err == nil {
		return LangTypeScript
	}

	return LangUnknown
}

func (q *QualityManager) runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = q.projectPath
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	output := stdout.String() + stderr.String()
	if err != nil {
		return output, err
	}
	return strings.TrimSpace(output), nil
}

func (q *QualityManager) Lint() CheckResult {
	start := time.Now()
	result := CheckResult{Name: "lint"}

	switch q.language {
	case LangGo:
		output, err := q.runCommand("go", "vet", "./...")
		if err != nil {
			result.Output = output
		} else {
			result.Output = output
			result.Passed = true
		}
		if _, err := exec.LookPath("golangci-lint"); err == nil {
			output, err := q.runCommand("golangci-lint", "run")
			if err != nil {
				result.Output = output
			} else {
				result.Output = output
				result.Passed = true
			}
		}

	case LangJavaScript, LangTypeScript:
		if _, err := os.Stat(filepath.Join(q.projectPath, "node_modules", ".bin", "eslint")); err == nil {
			output, err := q.runCommand("npx", "eslint", ".")
			if err != nil {
				result.Output = output
			} else {
				result.Output = output
				result.Passed = true
			}
		} else {
			result.Output = "eslint 未安装，跳过"
			result.Passed = true
		}

	case LangPython:
		if _, err := exec.LookPath("flake8"); err == nil {
			output, err := q.runCommand("flake8", ".")
			if err != nil {
				result.Output = output
			} else {
				result.Output = output
				result.Passed = true
			}
		} else {
			result.Output = "flake8 未安装，跳过"
			result.Passed = true
		}

	case LangRust:
		if _, err := exec.LookPath("cargo"); err == nil {
			output, err := q.runCommand("cargo", "clippy")
			if err != nil {
				result.Output = output
			} else {
				result.Output = output
				result.Passed = true
			}
		}

	default:
		result.Output = "未检测到项目语言"
		result.Passed = true
	}

	result.Duration = time.Since(start).String()
	return result
}

func (q *QualityManager) Test(watch bool) CheckResult {
	start := time.Now()
	result := CheckResult{Name: "test"}

	switch q.language {
	case LangGo:
		args := []string{"test", "./..."}
		if watch {
			ui.Warning("Go 测试不支持 watch 模式")
		}
		output, err := q.runCommand("go", args...)
		if err != nil {
			result.Output = output
		} else {
			result.Output = output
			result.Passed = true
		}

	case LangJavaScript, LangTypeScript:
		if _, err := os.Stat(filepath.Join(q.projectPath, "package.json")); err == nil {
			args := []string{"test"}
			if watch {
				args = append(args, "--", "--watch")
			}
			output, err := q.runCommand("npm", args...)
			if err != nil {
				result.Output = output
			} else {
				result.Output = output
				result.Passed = true
			}
		} else {
			result.Output = "未找到 package.json"
			result.Passed = true
		}

	case LangPython:
		if _, err := exec.LookPath("pytest"); err == nil {
			output, err := q.runCommand("pytest", "-v")
			if err != nil {
				result.Output = output
			} else {
				result.Output = output
				result.Passed = true
			}
		} else {
			result.Output = "pytest 未安装，跳过"
			result.Passed = true
		}

	case LangRust:
		output, err := q.runCommand("cargo", "test")
		if err != nil {
			result.Output = output
		} else {
			result.Output = output
			result.Passed = true
		}

	default:
		result.Output = "未检测到项目语言"
		result.Passed = true
	}

	result.Duration = time.Since(start).String()
	return result
}

func (q *QualityManager) TypeCheck() CheckResult {
	start := time.Now()
	result := CheckResult{Name: "typecheck"}

	switch q.language {
	case LangGo:
		output, err := q.runCommand("go", "build", "./...")
		if err != nil {
			result.Output = output
		} else {
			result.Output = "编译通过"
			result.Passed = true
		}

	case LangTypeScript:
		if _, err := exec.LookPath("tsc"); err == nil {
			output, err := q.runCommand("tsc", "--noEmit")
			if err != nil {
				result.Output = output
			} else {
				result.Output = "类型检查通过"
				result.Passed = true
			}
		} else {
			result.Output = "tsc 未安装，跳过"
			result.Passed = true
		}

	case LangRust:
		output, err := q.runCommand("cargo", "check")
		if err != nil {
			result.Output = output
		} else {
			result.Output = "类型检查通过"
			result.Passed = true
		}

	default:
		result.Output = "不支持类型检查或无需检查"
		result.Passed = true
	}

	result.Duration = time.Since(start).String()
	return result
}

func (q *QualityManager) Check() CheckReport {
	report := CheckReport{
		Language: q.language,
		Results:  []CheckResult{},
		Passed:   true,
	}

	ui.Info(fmt.Sprintf("检测到项目语言: %s", q.language))

	lintResult := q.Lint()
	report.Results = append(report.Results, lintResult)
	if !lintResult.Passed {
		report.Passed = false
	}

	testResult := q.Test(false)
	report.Results = append(report.Results, testResult)
	if !testResult.Passed {
		report.Passed = false
	}

	typeResult := q.TypeCheck()
	report.Results = append(report.Results, typeResult)
	if !typeResult.Passed {
		report.Passed = false
	}

	return report
}

func (r CheckReport) FormatJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r CheckReport) FormatTable() string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("语言: %s\n\n", r.Language))
	output.WriteString("检查项\t\t状态\t耗时\n")
	output.WriteString("========================================\n")

	for _, result := range r.Results {
		status := "✓ 通过"
		if !result.Passed {
			status = "✗ 失败"
		}
		output.WriteString(fmt.Sprintf("%s\t\t%s\t%s\n", result.Name, status, result.Duration))
	}

	output.WriteString("\n")
	if r.Passed {
		output.WriteString("所有检查通过！\n")
	} else {
		output.WriteString("部分检查失败，请查看详细输出\n")
	}

	return output.String()
}

func (q *QualityManager) SetupPreCommitHook() error {
	hooksDir := filepath.Join(q.projectPath, ".git", "hooks")
	hookPath := filepath.Join(hooksDir, "pre-commit")

	if _, err := os.Stat(hooksDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("当前目录不是 Git 仓库")
		}
		return err
	}

	hookContent := `#!/bin/sh
# DevFlow pre-commit hook

echo "Running DevFlow checks..."
devflow check

if [ $? -ne 0 ]; then
    echo "DevFlow checks failed. Commit aborted."
    exit 1
fi

exit 0
`

	if _, err := os.Stat(hookPath); err == nil {
		existingContent, err := os.ReadFile(hookPath)
		if err != nil {
			return err
		}
		if strings.Contains(string(existingContent), "DevFlow") {
			ui.Info("pre-commit hook 已存在")
			return nil
		}
		hookContent = string(existingContent) + "\n" + hookContent
	}

	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return err
	}

	ui.Success("pre-commit hook 已安装")
	return nil
}

func (q *QualityManager) GenerateCIConfig(platform string) error {
	var config string

	switch platform {
	case "github":
		config = q.generateGitHubActions()
	case "gitlab":
		config = q.generateGitLabCI()
	default:
		return fmt.Errorf("不支持的 CI 平台: %s (支持: github, gitlab)", platform)
	}

	outputDir := filepath.Join(q.projectPath, ".github", "workflows")
	if platform == "gitlab" {
		outputDir = q.projectPath
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	var outputPath string
	switch platform {
	case "github":
		outputPath = filepath.Join(outputDir, "devflow.yml")
	case "gitlab":
		outputPath = filepath.Join(outputDir, ".gitlab-ci.yml")
	}

	if err := os.WriteFile(outputPath, []byte(config), 0644); err != nil {
		return err
	}

	ui.Success(fmt.Sprintf("CI 配置已生成: %s", outputPath))
	return nil
}

func (q *QualityManager) generateGitHubActions() string {
	var steps string

	switch q.language {
	case LangGo:
		steps = `
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Install dependencies
        run: go mod download

      - name: Run linter
        run: go vet ./...

      - name: Run tests
        run: go test -v ./...`

	case LangJavaScript, LangTypeScript:
		steps = `
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        run: npm ci

      - name: Run linter
        run: npm run lint

      - name: Run tests
        run: npm test`

	case LangPython:
		steps = `
      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'

      - name: Install dependencies
        run: pip install -r requirements.txt

      - name: Run linter
        run: flake8 .

      - name: Run tests
        run: pytest -v`

	case LangRust:
		steps = `
      - name: Setup Rust
        uses: actions-rs/toolchain@v1
        with:
          toolchain: stable

      - name: Run linter
        run: cargo clippy

      - name: Run tests
        run: cargo test`

	default:
		steps = `
      - name: Run DevFlow check
        run: devflow check`
	}

	return fmt.Sprintf(`name: DevFlow CI

on:
  push:
    branches: [main, master, develop]
  pull_request:
    branches: [main, master, develop]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4%s
`, steps)
}

func (q *QualityManager) generateGitLabCI() string {
	return `stages:
  - lint
  - test
  - build

lint:
  stage: lint
  script:
    - echo "Running lint checks..."
    - devflow lint

test:
  stage: test
  script:
    - echo "Running tests..."
    - devflow test

build:
  stage: build
  script:
    - echo "Running type check and build..."
    - devflow check
`
}

func (q *QualityManager) GetLanguage() ProjectLanguage {
	return q.language
}
