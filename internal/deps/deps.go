package deps

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/devflow/devflow/internal/ui"
)

type PackageManager string

const (
	PMNpm    PackageManager = "npm"
	PMYarn   PackageManager = "yarn"
	PMPnpm   PackageManager = "pnpm"
	PMPip    PackageManager = "pip"
	PMPoetry PackageManager = "poetry"
	PMCargo  PackageManager = "cargo"
	PMGoMod  PackageManager = "go"
	PMMaven  PackageManager = "maven"
)

type Dependency struct {
	Name         string `json:"name"`
	Current      string `json:"current"`
	Wanted       string `json:"wanted,omitempty"`
	Latest       string `json:"latest,omitempty"`
	Outdated     bool   `json:"outdated"`
	Vulnerable   bool   `json:"vulnerable,omitempty"`
}

type DependencyReport struct {
	Manager     PackageManager `json:"manager"`
	Count       int            `json:"count"`
	Dependencies []Dependency   `json:"dependencies"`
}

type DependencyManager struct {
	projectPath string
	manager     PackageManager
}

func NewDependencyManager() (*DependencyManager, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("无法获取当前目录: %w", err)
	}

	manager := detectPackageManager(wd)

	return &DependencyManager{
		projectPath: wd,
		manager:     manager,
	}, nil
}

func detectPackageManager(path string) PackageManager {
	if _, err := os.Stat(filepath.Join(path, "package-lock.json")); err == nil {
		return PMNpm
	}
	if _, err := os.Stat(filepath.Join(path, "yarn.lock")); err == nil {
		return PMYarn
	}
	if _, err := os.Stat(filepath.Join(path, "pnpm-lock.yaml")); err == nil {
		return PMPnpm
	}
	if _, err := os.Stat(filepath.Join(path, "package.json")); err == nil {
		return PMNpm
	}
	if _, err := os.Stat(filepath.Join(path, "poetry.lock")); err == nil {
		return PMPoetry
	}
	if _, err := os.Stat(filepath.Join(path, "requirements.txt")); err == nil {
		return PMPip
	}
	if _, err := os.Stat(filepath.Join(path, "Cargo.toml")); err == nil {
		return PMCargo
	}
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		return PMGoMod
	}
	if _, err := os.Stat(filepath.Join(path, "pom.xml")); err == nil {
		return PMMaven
	}

	return PMNpm
}

func (d *DependencyManager) runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = d.projectPath
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

func (d *DependencyManager) List() DependencyReport {
	report := DependencyReport{
		Manager:      d.manager,
		Dependencies: []Dependency{},
	}

	switch d.manager {
	case PMNpm, PMYarn, PMPnpm:
		deps := d.listNpm()
		report.Dependencies = deps

	case PMPip, PMPoetry:
		deps := d.listPython()
		report.Dependencies = deps

	case PMCargo:
		deps := d.listRust()
		report.Dependencies = deps

	case PMGoMod:
		deps := d.listGo()
		report.Dependencies = deps

	case PMMaven:
		deps := d.listMaven()
		report.Dependencies = deps
	}

	report.Count = len(report.Dependencies)
	return report
}

func (d *DependencyManager) listNpm() []Dependency {
	var deps []Dependency

	cmdArgs := []string{"list", "--depth=0", "--json"}
	output, _ := d.runCommand("npm", cmdArgs...)

	var result struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}

	if err := json.Unmarshal([]byte(output), &result); err == nil {
		for name, info := range result.Dependencies {
			deps = append(deps, Dependency{
				Name:    name,
				Current: info.Version,
			})
		}
	}

	return deps
}

func (d *DependencyManager) listPython() []Dependency {
	var deps []Dependency

	reqPath := filepath.Join(d.projectPath, "requirements.txt")
	if _, err := os.Stat(reqPath); err == nil {
		file, err := os.Open(reqPath)
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				parts := strings.SplitN(line, "==", 2)
				if len(parts) == 2 {
					deps = append(deps, Dependency{
						Name:    strings.TrimSpace(parts[0]),
						Current: strings.TrimSpace(parts[1]),
					})
				} else {
					deps = append(deps, Dependency{
						Name:    line,
						Current: "latest",
					})
				}
			}
		}
	}

	return deps
}

func (d *DependencyManager) listRust() []Dependency {
	var deps []Dependency

	output, _ := d.runCommand("cargo", "tree", "--depth=1")
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, " v") {
			parts := strings.SplitN(line, " v", 2)
			if len(parts) == 2 {
				name := strings.TrimPrefix(strings.TrimPrefix(parts[0], "├── "), "└── ")
				version := strings.SplitN(parts[1], " ", 2)[0]
				deps = append(deps, Dependency{
					Name:    name,
					Current: version,
				})
			}
		}
	}

	return deps
}

func (d *DependencyManager) listGo() []Dependency {
	var deps []Dependency

	output, _ := d.runCommand("go", "list", "-m", "all")
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			deps = append(deps, Dependency{
				Name:    parts[0],
				Current: parts[1],
			})
		}
	}

	return deps
}

func (d *DependencyManager) listMaven() []Dependency {
	var deps []Dependency

	output, _ := d.runCommand("mvn", "dependency:list")
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, ":jar:") || strings.Contains(line, ":compile") {
			parts := strings.Split(line, ":")
			if len(parts) >= 4 {
				deps = append(deps, Dependency{
					Name:    parts[0] + ":" + parts[1],
					Current: parts[3],
				})
			}
		}
	}

	return deps
}

func (d *DependencyManager) Outdated() DependencyReport {
	report := d.List()

	switch d.manager {
	case PMNpm, PMYarn, PMPnpm:
		d.checkNpmOutdated(&report)

	case PMGoMod:
		d.checkGoOutdated(&report)

	case PMCargo:
		d.checkCargoOutdated(&report)

	default:
		ui.Warning(fmt.Sprintf("%s 暂不支持自动检测过时依赖", d.manager))
	}

	return report
}

func (d *DependencyManager) checkNpmOutdated(report *DependencyReport) {
	output, _ := d.runCommand("npm", "outdated", "--json")

	var outdated map[string]struct {
		Current string `json:"current"`
		Wanted  string `json:"wanted"`
		Latest  string `json:"latest"`
	}

	if err := json.Unmarshal([]byte(output), &outdated); err != nil {
		return
	}

	for i, dep := range report.Dependencies {
		if info, ok := outdated[dep.Name]; ok {
			report.Dependencies[i].Wanted = info.Wanted
			report.Dependencies[i].Latest = info.Latest
			report.Dependencies[i].Outdated = true
		}
	}
}

func (d *DependencyManager) checkGoOutdated(report *DependencyReport) {
	for i := range report.Dependencies {
		report.Dependencies[i].Outdated = false
	}
}

func (d *DependencyManager) checkCargoOutdated(report *DependencyReport) {
	output, _ := d.runCommand("cargo", "outdated", "--root-deps-only")

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "---") || strings.Contains(line, "Name") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 4 {
			name := parts[0]
			for i := range report.Dependencies {
				if report.Dependencies[i].Name == name {
					report.Dependencies[i].Latest = parts[3]
					report.Dependencies[i].Outdated = true
				}
			}
		}
	}
}

func (d *DependencyManager) Update(interactive bool) error {
	ui.Info(fmt.Sprintf("使用包管理器: %s", d.manager))

	if interactive {
		reader := bufio.NewReader(os.Stdin)
		ui.Prompt("更新所有依赖? [y/N]")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			ui.Info("已取消")
			return nil
		}
	}

	switch d.manager {
	case PMNpm:
		ui.Info("运行 npm update...")
		_, err := d.runCommand("npm", "update")
		if err != nil {
			return err
		}

	case PMYarn:
		ui.Info("运行 yarn upgrade...")
		_, err := d.runCommand("yarn", "upgrade")
		if err != nil {
			return err
		}

	case PMPnpm:
		ui.Info("运行 pnpm update...")
		_, err := d.runCommand("pnpm", "update")
		if err != nil {
			return err
		}

	case PMPip:
		ui.Info("运行 pip install --upgrade...")
		_, err := d.runCommand("pip", "install", "--upgrade", "-r", "requirements.txt")
		if err != nil {
			return err
		}

	case PMCargo:
		ui.Info("运行 cargo update...")
		_, err := d.runCommand("cargo", "update")
		if err != nil {
			return err
		}

	case PMGoMod:
		ui.Info("运行 go get -u...")
		_, err := d.runCommand("go", "get", "-u", "./...")
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("不支持的包管理器: %s", d.manager)
	}

	ui.Success("依赖更新完成")
	return nil
}

func (d *DependencyManager) Audit() error {
	ui.Info(fmt.Sprintf("使用包管理器: %s", d.manager))

	switch d.manager {
	case PMNpm, PMPnpm:
		ui.Info("运行 npm audit...")
		output, err := d.runCommand("npm", "audit")
		if output != "" {
			fmt.Println(output)
		}
		if err != nil {
			return fmt.Errorf("发现安全漏洞")
		}
		ui.Success("未发现安全漏洞")

	case PMYarn:
		ui.Info("运行 yarn audit...")
		output, err := d.runCommand("yarn", "audit")
		if output != "" {
			fmt.Println(output)
		}
		if err != nil {
			return fmt.Errorf("发现安全漏洞")
		}
		ui.Success("未发现安全漏洞")

	case PMGoMod:
		ui.Info("运行 go list -m all...")
		output, _ := d.runCommand("go", "list", "-m", "-json", "all")
		fmt.Println(output)
		ui.Info("建议使用 govulncheck 进行详细的安全检查")

	case PMCargo:
		if _, err := exec.LookPath("cargo-audit"); err == nil {
			ui.Info("运行 cargo audit...")
			output, err := d.runCommand("cargo", "audit")
			if output != "" {
				fmt.Println(output)
			}
			if err != nil {
				return fmt.Errorf("发现安全漏洞")
			}
			ui.Success("未发现安全漏洞")
		} else {
			ui.Warning("cargo-audit 未安装，建议安装: cargo install cargo-audit")
		}

	default:
		ui.Warning(fmt.Sprintf("%s 暂不支持安全审计", d.manager))
	}

	return nil
}

func (d *DependencyManager) GetManager() PackageManager {
	return d.manager
}

func (r DependencyReport) FormatTable() string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("包管理器: %s\n", r.Manager))
	output.WriteString(fmt.Sprintf("依赖数量: %d\n\n", r.Count))

	if r.Count == 0 {
		output.WriteString("未找到依赖\n")
		return output.String()
	}

	output.WriteString("名称\t\t当前版本\t最新版本\t状态\n")
	output.WriteString("================================================================\n")

	for _, dep := range r.Dependencies {
		status := "  最新"
		if dep.Outdated {
			status = "✗ 过时"
		}
		latest := dep.Latest
		if latest == "" {
			latest = "-"
		}
		output.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\n", dep.Name, dep.Current, latest, status))
	}

	outdatedCount := 0
	for _, dep := range r.Dependencies {
		if dep.Outdated {
			outdatedCount++
		}
	}

	output.WriteString("\n")
	if outdatedCount > 0 {
		output.WriteString(fmt.Sprintf("发现 %d 个过时依赖\n", outdatedCount))
	} else {
		output.WriteString("所有依赖都是最新的\n")
	}

	return output.String()
}

func (r DependencyReport) FormatJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
