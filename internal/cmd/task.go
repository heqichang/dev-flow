package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/devflow/devflow/internal/task"
	"github.com/devflow/devflow/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "任务管理",
		Long: `项目任务管理，支持与 Git 分支关联。

可用命令:
  add     - 添加任务
  list    - 列出任务
  start   - 开始任务
  done    - 完成任务
  update  - 更新任务
  delete  - 删除任务
  show    - 查看任务详情`,
	}

	cmd.AddCommand(newTaskAddCmd())
	cmd.AddCommand(newTaskListCmd())
	cmd.AddCommand(newTaskStartCmd())
	cmd.AddCommand(newTaskDoneCmd())
	cmd.AddCommand(newTaskUpdateCmd())
	cmd.AddCommand(newTaskDeleteCmd())
	cmd.AddCommand(newTaskShowCmd())

	return cmd
}

func newTaskAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "添加任务",
		Long: `添加新任务。`,
		Args: cobra.MinimumNArgs(1),
		Run: runTaskAdd,
	}

	cmd.Flags().StringP("description", "d", "", "任务描述")
	cmd.Flags().StringP("priority", "p", "medium", "优先级 (low, medium, high, urgent)")
	cmd.Flags().StringSlice("tags", []string{}, "标签")

	return cmd
}

func runTaskAdd(cmd *cobra.Command, args []string) {
	title := strings.Join(args, " ")
	description, _ := cmd.Flags().GetString("description")
	priorityStr, _ := cmd.Flags().GetString("priority")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	var priority task.TaskPriority
	switch strings.ToLower(priorityStr) {
	case "low":
		priority = task.PriorityLow
	case "medium":
		priority = task.PriorityMedium
	case "high":
		priority = task.PriorityHigh
	case "urgent":
		priority = task.PriorityUrgent
	default:
		ui.Error(fmt.Sprintf("无效的优先级: %s", priorityStr))
		os.Exit(1)
	}

	tm, err := task.NewTaskManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	t, err := tm.Add(title, description, priority, tags)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success(fmt.Sprintf("任务 #%d 已添加", t.ID))
	ui.Info(fmt.Sprintf("  标题: %s", t.Title))
	ui.Info(fmt.Sprintf("  优先级: %s", t.Priority))
}

func newTaskListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出任务",
		Long: `列出所有任务。`,
		Run: runTaskList,
	}

	cmd.Flags().StringP("status", "s", "", "过滤状态 (todo, in_progress, done)")

	return cmd
}

func runTaskList(cmd *cobra.Command, args []string) {
	statusStr, _ := cmd.Flags().GetString("status")

	var status task.TaskStatus
	switch strings.ToLower(statusStr) {
	case "todo":
		status = task.StatusTodo
	case "in_progress", "progress":
		status = task.StatusInProgress
	case "done", "completed":
		status = task.StatusDone
	}

	tm, err := task.NewTaskManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	tasks := tm.List(status)

	if len(tasks) == 0 {
		ui.Info("没有任务")
		return
	}

	ui.Header("任务列表")

	for _, t := range tasks {
		printTaskSummary(t)
	}
}

func printTaskSummary(t task.Task) {
	var statusIcon string
	var statusColor func(a ...interface{}) string

	switch t.Status {
	case task.StatusTodo:
		statusIcon = "○"
		statusColor = color.New(color.FgWhite).SprintFunc()
	case task.StatusInProgress:
		statusIcon = "●"
		statusColor = color.New(color.FgYellow).SprintFunc()
	case task.StatusDone:
		statusIcon = "✓"
		statusColor = color.New(color.FgGreen).SprintFunc()
	}

	var priorityIcon string
	switch t.Priority {
	case task.PriorityUrgent:
		priorityIcon = "🔴"
	case task.PriorityHigh:
		priorityIcon = "🟠"
	case task.PriorityMedium:
		priorityIcon = "🟡"
	case task.PriorityLow:
		priorityIcon = "🟢"
	}

	ui.Info(fmt.Sprintf("%s %s #%d %s %s",
		statusColor(statusIcon),
		priorityIcon,
		t.ID,
		t.Title,
		strings.Join(t.Tags, " "),
	))

	if t.Branch != "" {
		ui.Info(fmt.Sprintf("    分支: %s", t.Branch))
	}
}

func newTaskStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <id>",
		Short: "开始任务",
		Long: `开始任务并创建分支。`,
		Args: cobra.ExactArgs(1),
		Run: runTaskStart,
	}

	return cmd
}

func runTaskStart(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		ui.Error("无效的任务 ID")
		os.Exit(1)
	}

	tm, err := task.NewTaskManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := tm.Start(id); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

func newTaskDoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done <id>",
		Short: "完成任务",
		Long: `完成任务。`,
		Args: cobra.ExactArgs(1),
		Run: runTaskDone,
	}

	return cmd
}

func runTaskDone(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		ui.Error("无效的任务 ID")
		os.Exit(1)
	}

	tm, err := task.NewTaskManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := tm.Done(id); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

func newTaskUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "更新任务",
		Long: `更新任务信息。`,
		Args: cobra.ExactArgs(1),
		Run: runTaskUpdate,
	}

	cmd.Flags().StringP("title", "t", "", "新标题")
	cmd.Flags().StringP("description", "d", "", "新描述")
	cmd.Flags().StringP("priority", "p", "", "新优先级")
	cmd.Flags().StringSlice("tags", []string{}, "新标签")

	return cmd
}

func runTaskUpdate(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		ui.Error("无效的任务 ID")
		os.Exit(1)
	}

	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	priorityStr, _ := cmd.Flags().GetString("priority")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	var priority task.TaskPriority
	switch strings.ToLower(priorityStr) {
	case "low":
		priority = task.PriorityLow
	case "medium":
		priority = task.PriorityMedium
	case "high":
		priority = task.PriorityHigh
	case "urgent":
		priority = task.PriorityUrgent
	}

	tm, err := task.NewTaskManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := tm.Update(id, title, description, priority, tags); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success(fmt.Sprintf("任务 #%d 已更新", id))
}

func newTaskDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "删除任务",
		Long: `删除任务。`,
		Args: cobra.ExactArgs(1),
		Run: runTaskDelete,
	}

	cmd.Flags().BoolP("force", "f", false, "不确认直接删除")

	return cmd
}

func runTaskDelete(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		ui.Error("无效的任务 ID")
		os.Exit(1)
	}

	force, _ := cmd.Flags().GetBool("force")

	if !force {
		reader := bufio.NewReader(os.Stdin)
		ui.Prompt(fmt.Sprintf("确认删除任务 #%d? [y/N]", id))
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			ui.Info("已取消")
			return
		}
	}

	tm, err := task.NewTaskManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := tm.Delete(id); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success(fmt.Sprintf("任务 #%d 已删除", id))
}

func newTaskShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "查看任务",
		Long: `查看任务详细信息。`,
		Args: cobra.ExactArgs(1),
		Run: runTaskShow,
	}

	return cmd
}

func runTaskShow(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		ui.Error("无效的任务 ID")
		os.Exit(1)
	}

	tm, err := task.NewTaskManager()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	t, err := tm.Get(id)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Header(fmt.Sprintf("任务 #%d", t.ID))
	ui.Info(fmt.Sprintf("标题: %s", t.Title))
	ui.Info(fmt.Sprintf("状态: %s", t.Status))
	ui.Info(fmt.Sprintf("优先级: %s", t.Priority))
	if len(t.Tags) > 0 {
		ui.Info(fmt.Sprintf("标签: %s", strings.Join(t.Tags, ", ")))
	}
	if t.Description != "" {
		ui.Info(fmt.Sprintf("描述: %s", t.Description))
	}
	if t.Branch != "" {
		ui.Info(fmt.Sprintf("分支: %s", t.Branch))
	}
	if t.PRURL != "" {
		ui.Info(fmt.Sprintf("PR: %s", t.PRURL))
	}
	ui.Info(fmt.Sprintf("创建时间: %s", t.CreatedAt.Format(time.RFC3339)))
	ui.Info(fmt.Sprintf("更新时间: %s", t.UpdatedAt.Format(time.RFC3339)))
	if t.StartedAt != nil {
		ui.Info(fmt.Sprintf("开始时间: %s", t.StartedAt.Format(time.RFC3339)))
	}
	if t.CompletedAt != nil {
		ui.Info(fmt.Sprintf("完成时间: %s", t.CompletedAt.Format(time.RFC3339)))
	}
}
