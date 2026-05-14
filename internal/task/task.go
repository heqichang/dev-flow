package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/devflow/devflow/internal/git"
	"github.com/devflow/devflow/internal/gitflow"
	"github.com/devflow/devflow/internal/ui"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

type TaskPriority string

const (
	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
	PriorityUrgent TaskPriority = "urgent"
)

type Task struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	Tags        []string     `json:"tags,omitempty"`
	Branch      string       `json:"branch,omitempty"`
	PRURL       string       `json:"prURL,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	StartedAt   *time.Time   `json:"startedAt,omitempty"`
	CompletedAt *time.Time   `json:"completedAt,omitempty"`
}

type TaskStore struct {
	Tasks     []Task    `json:"tasks"`
	NextID    int       `json:"nextID"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TaskManager struct {
	repoPath  string
	store     *TaskStore
	storePath string
}

const taskDir = ".devflow"
const taskFile = "tasks.json"

func NewTaskManager() (*TaskManager, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("无法获取当前目录: %w", err)
	}

	repoPath, err := git.FindGitRepoRoot(wd)
	if err != nil {
		repoPath = wd
	}

	tm := &TaskManager{
		repoPath:  repoPath,
		storePath: filepath.Join(repoPath, taskDir, taskFile),
	}

	if err := tm.load(); err != nil {
		return nil, err
	}

	return tm, nil
}

func (tm *TaskManager) load() error {
	if _, err := os.Stat(tm.storePath); err != nil {
		if os.IsNotExist(err) {
			tm.store = &TaskStore{
				Tasks:     []Task{},
				NextID:    1,
				UpdatedAt: time.Now(),
			}
			return nil
		}
		return err
	}

	data, err := os.ReadFile(tm.storePath)
	if err != nil {
		return fmt.Errorf("读取任务存储失败: %w", err)
	}

	var store TaskStore
	if err := json.Unmarshal(data, &store); err != nil {
		return fmt.Errorf("解析任务存储失败: %w", err)
	}

	tm.store = &store
	return nil
}

func (tm *TaskManager) save() error {
	dir := filepath.Dir(tm.storePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	tm.store.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(tm.store, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化任务存储失败: %w", err)
	}

	if err := os.WriteFile(tm.storePath, data, 0644); err != nil {
		return fmt.Errorf("写入任务存储失败: %w", err)
	}

	return nil
}

func (tm *TaskManager) Add(title, description string, priority TaskPriority, tags []string) (*Task, error) {
	if title == "" {
		return nil, fmt.Errorf("任务标题不能为空")
	}

	if priority == "" {
		priority = PriorityMedium
	}

	now := time.Now()
	task := Task{
		ID:          tm.store.NextID,
		Title:       title,
		Description: description,
		Status:      StatusTodo,
		Priority:    priority,
		Tags:        tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tm.store.Tasks = append(tm.store.Tasks, task)
	tm.store.NextID++

	if err := tm.save(); err != nil {
		return nil, err
	}

	return &task, nil
}

func (tm *TaskManager) Get(id int) (*Task, error) {
	for i := range tm.store.Tasks {
		if tm.store.Tasks[i].ID == id {
			return &tm.store.Tasks[i], nil
		}
	}
	return nil, fmt.Errorf("任务不存在: %d", id)
}

func (tm *TaskManager) List(status TaskStatus) []Task {
	var tasks []Task
	for _, t := range tm.store.Tasks {
		if status == "" || t.Status == status {
			tasks = append(tasks, t)
		}
	}

	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Priority != tasks[j].Priority {
			priorityOrder := map[TaskPriority]int{
				PriorityUrgent: 0,
				PriorityHigh:   1,
				PriorityMedium: 2,
				PriorityLow:    3,
			}
			return priorityOrder[tasks[i].Priority] < priorityOrder[tasks[j].Priority]
		}
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})

	return tasks
}

func (tm *TaskManager) Start(id int) error {
	task, err := tm.Get(id)
	if err != nil {
		return err
	}

	if task.Status == StatusInProgress {
		return fmt.Errorf("任务已在进行中")
	}

	gitClient := git.NewGitClient()
	gitClient.RepoPath = tm.repoPath

	if !gitClient.IsGitRepo() {
		return fmt.Errorf("当前目录不是 Git 仓库")
	}

	status, _ := gitClient.Status()
	if status.HasChanges {
		ui.Warning("当前分支有未提交的更改，请先提交或 stash")
	}

	branchName := fmt.Sprintf("task/%d-%s", task.ID, sanitizeBranchName(task.Title))

	ui.Info(fmt.Sprintf("创建分支: %s", branchName))
	if err := gitClient.CreateBranch(branchName); err != nil {
		return err
	}

	now := time.Now()
	task.Status = StatusInProgress
	task.Branch = branchName
	task.StartedAt = &now
	task.UpdatedAt = now

	if err := tm.save(); err != nil {
		return err
	}

	ui.Success(fmt.Sprintf("任务 #%d 已开始", task.ID))
	ui.Info(fmt.Sprintf("分支: %s", branchName))
	return nil
}

func (tm *TaskManager) Done(id int) error {
	task, err := tm.Get(id)
	if err != nil {
		return err
	}

	if task.Status == StatusDone {
		return fmt.Errorf("任务已完成")
	}

	if task.Branch == "" {
		return fmt.Errorf("任务未关联分支")
	}

	gitClient := git.NewGitClient()
	gitClient.RepoPath = tm.repoPath

	if gitClient.IsGitRepo() {
		currentBranch, _ := gitClient.GetCurrentBranch()
		if currentBranch != task.Branch {
			ui.Warning(fmt.Sprintf("当前不在任务分支 %s 上", task.Branch))
		} else {
			remoteURL, err := gitClient.GetRemoteURL("origin")
			if err == nil {
				ui.Info("请创建 Pull Request:")
				ui.Info(fmt.Sprintf("  分支: %s", task.Branch))
				ui.Info(fmt.Sprintf("  仓库: %s", remoteURL))
			}
		}
	}

	now := time.Now()
	task.Status = StatusDone
	task.CompletedAt = &now
	task.UpdatedAt = now

	if err := tm.save(); err != nil {
		return err
	}

	ui.Success(fmt.Sprintf("任务 #%d 已完成", task.ID))
	return nil
}

func (tm *TaskManager) Update(id int, title, description string, priority TaskPriority, tags []string) error {
	task, err := tm.Get(id)
	if err != nil {
		return err
	}

	if title != "" {
		task.Title = title
	}
	if description != "" {
		task.Description = description
	}
	if priority != "" {
		task.Priority = priority
	}
	if tags != nil {
		task.Tags = tags
	}
	task.UpdatedAt = time.Now()

	return tm.save()
}

func (tm *TaskManager) Delete(id int) error {
	for i, t := range tm.store.Tasks {
		if t.ID == id {
			tm.store.Tasks = append(tm.store.Tasks[:i], tm.store.Tasks[i+1:]...)
			return tm.save()
		}
	}
	return fmt.Errorf("任务不存在: %d", id)
}

func sanitizeBranchName(name string) string {
	replacer := strings.NewReplacer(
		" ", "-",
		"_", "-",
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
		".", "",
	)

	cleaned := replacer.Replace(strings.ToLower(name))

	for strings.Contains(cleaned, "--") {
		cleaned = strings.ReplaceAll(cleaned, "--", "-")
	}

	cleaned = strings.Trim(cleaned, "-")

	if len(cleaned) > 50 {
		cleaned = cleaned[:50]
	}

	return cleaned
}

func (tm *TaskManager) IntegrateWithGitFlow(id int) error {
	task, err := tm.Get(id)
	if err != nil {
		return err
	}

	if task.Status != StatusTodo {
		return fmt.Errorf("只能从未开始状态集成")
	}

	gfm := gitflow.NewGitFlowManager()

	ui.Info("选择 Git Flow 类型:")
	ui.Info("  1. feature - 新功能")
	ui.Info("  2. hotfix  - 紧急修复")

	flowType := gitflow.FlowFeature

	branchName, err := gfm.StartFlow(flowType, sanitizeBranchName(task.Title))
	if err != nil {
		return err
	}

	now := time.Now()
	task.Status = StatusInProgress
	task.Branch = branchName
	task.StartedAt = &now
	task.UpdatedAt = now

	return tm.save()
}
