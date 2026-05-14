package task

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestTaskManager(t *testing.T) (*TaskManager, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "devflow-task-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	tm := &TaskManager{
		repoPath:  tmpDir,
		storePath: filepath.Join(tmpDir, taskDir, taskFile),
	}

	tm.store = &TaskStore{
		Tasks:     []Task{},
		NextID:    1,
		UpdatedAt: time.Now(),
	}

	return tm, tmpDir
}

func teardownTestTaskManager(t *testing.T, tmpDir string) {
	t.Helper()
	os.RemoveAll(tmpDir)
}

func TestTaskAdd(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	task, err := tm.Add("Test Task", "A description", PriorityHigh, []string{"bug"})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if task.ID != 1 {
		t.Errorf("Task ID = %d, want 1", task.ID)
	}
	if task.Title != "Test Task" {
		t.Errorf("Task Title = %q, want %q", task.Title, "Test Task")
	}
	if task.Status != StatusTodo {
		t.Errorf("Task Status = %q, want %q", task.Status, StatusTodo)
	}
	if task.Priority != PriorityHigh {
		t.Errorf("Task Priority = %q, want %q", task.Priority, PriorityHigh)
	}
	if len(task.Tags) != 1 || task.Tags[0] != "bug" {
		t.Errorf("Task Tags = %v, want [bug]", task.Tags)
	}
}

func TestTaskAddEmptyTitle(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	_, err := tm.Add("", "", PriorityMedium, nil)
	if err == nil {
		t.Error("Add() with empty title should return error")
	}
}

func TestTaskAddDefaultPriority(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	task, err := tm.Add("Test", "", "", nil)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if task.Priority != PriorityMedium {
		t.Errorf("Default Priority = %q, want %q", task.Priority, PriorityMedium)
	}
}

func TestTaskGet(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	tm.Add("Task 1", "", PriorityLow, nil)
	tm.Add("Task 2", "", PriorityHigh, nil)

	task, err := tm.Get(2)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if task.Title != "Task 2" {
		t.Errorf("Get() Title = %q, want %q", task.Title, "Task 2")
	}
}

func TestTaskGetNotFound(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	_, err := tm.Get(999)
	if err == nil {
		t.Error("Get() with non-existent ID should return error")
	}
}

func TestTaskList(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	tm.Add("Low Task", "", PriorityLow, nil)
	tm.Add("Urgent Task", "", PriorityUrgent, nil)
	tm.Add("Medium Task", "", PriorityMedium, nil)

	tasks := tm.List("")
	if len(tasks) != 3 {
		t.Fatalf("List() count = %d, want 3", len(tasks))
	}

	if tasks[0].Priority != PriorityUrgent {
		t.Errorf("First task Priority = %q, want %q (sorted by priority)", tasks[0].Priority, PriorityUrgent)
	}
}

func TestTaskListByStatus(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	tm.Add("Task 1", "", PriorityMedium, nil)
	tm.Add("Task 2", "", PriorityMedium, nil)

	task2, _ := tm.Get(2)
	task2.Status = StatusDone
	task2.Branch = "task/2-task-2"
	tm.save()

	tasks := tm.List(StatusDone)
	if len(tasks) != 1 {
		t.Fatalf("List(done) count = %d, want 1", len(tasks))
	}
	if tasks[0].Title != "Task 2" {
		t.Errorf("List(done) Title = %q, want %q", tasks[0].Title, "Task 2")
	}
}

func TestTaskUpdate(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	tm.Add("Original", "Original desc", PriorityLow, []string{"old"})

	err := tm.Update(1, "Updated", "New desc", PriorityHigh, []string{"new"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	task, _ := tm.Get(1)
	if task.Title != "Updated" {
		t.Errorf("Title = %q, want %q", task.Title, "Updated")
	}
	if task.Description != "New desc" {
		t.Errorf("Description = %q, want %q", task.Description, "New desc")
	}
	if task.Priority != PriorityHigh {
		t.Errorf("Priority = %q, want %q", task.Priority, PriorityHigh)
	}
}

func TestTaskDelete(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	tm.Add("Task 1", "", PriorityMedium, nil)
	tm.Add("Task 2", "", PriorityMedium, nil)

	err := tm.Delete(1)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	tasks := tm.List("")
	if len(tasks) != 1 {
		t.Fatalf("List() count after delete = %d, want 1", len(tasks))
	}
	if tasks[0].ID != 2 {
		t.Errorf("Remaining task ID = %d, want 2", tasks[0].ID)
	}
}

func TestTaskDeleteNotFound(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	err := tm.Delete(999)
	if err == nil {
		t.Error("Delete() with non-existent ID should return error")
	}
}

func TestTaskAddReturnValueNotConnectedToStore(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	addedTask, err := tm.Add("Original Title", "", PriorityMedium, nil)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	addedTask.Title = "Modified Title"

	storedTask, _ := tm.Get(1)
	if storedTask.Title == "Modified Title" {
		t.Error("Add() returns a pointer to the store element; modifying it affects the store. " +
			"This is a bug: Add() returns &task (local variable), not &tm.store.Tasks[i]. " +
			"Modifying the returned *Task should NOT affect the store, but currently it might if " +
			"the local variable happens to share memory. The correct fix is to return &tm.store.Tasks[len-1].")
	}
}

func TestTaskDoneAlreadyDone(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	task, _ := tm.Add("Done Task", "", PriorityMedium, nil)
	task.Status = StatusDone
	task.Branch = "task/1-done-task"
	tm.save()

	err := tm.Done(1)
	if err == nil {
		t.Error("Done() on already done task should return error")
	}
}

func TestTaskDoneNoBranch(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	tm.Add("No Branch Task", "", PriorityMedium, nil)

	err := tm.Done(1)
	if err == nil {
		t.Error("Done() on task without branch should return error")
	}
}

func TestTaskIncrementID(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	t1, _ := tm.Add("Task 1", "", PriorityMedium, nil)
	t2, _ := tm.Add("Task 2", "", PriorityMedium, nil)
	t3, _ := tm.Add("Task 3", "", PriorityMedium, nil)

	if t1.ID != 1 || t2.ID != 2 || t3.ID != 3 {
		t.Errorf("IDs = %d, %d, %d; want 1, 2, 3", t1.ID, t2.ID, t3.ID)
	}
}

func TestTaskPersistAndReload(t *testing.T) {
	tm, tmpDir := setupTestTaskManager(t)
	defer teardownTestTaskManager(t, tmpDir)

	tm.Add("Persisted Task", "Should survive reload", PriorityHigh, []string{"test"})

	tm2 := &TaskManager{
		repoPath:  tmpDir,
		storePath: filepath.Join(tmpDir, taskDir, taskFile),
	}

	if err := tm2.load(); err != nil {
		t.Fatalf("load() error = %v", err)
	}

	task, err := tm2.Get(1)
	if err != nil {
		t.Fatalf("Get() after reload error = %v", err)
	}
	if task.Title != "Persisted Task" {
		t.Errorf("Title after reload = %q, want %q", task.Title, "Persisted Task")
	}
}

func TestSanitizeBranchName(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"Hello World", "hello-world"},
		{"fix: critical bug", "fix-critical-bug"},
		{"task/with/slashes", "task-with-slashes"},
		{"  spaces  ", "spaces"},
		{"multiple---dashes", "multiple-dashes"},
		{"special:chars*here", "special-chars-here"},
		{"UPPERCASE", "uppercase"},
		{"a.b.c", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeBranchName(tt.input)
			if result != tt.expect {
				t.Errorf("sanitizeBranchName(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestSanitizeBranchNameLength(t *testing.T) {
	longName := ""
	for i := 0; i < 100; i++ {
		longName += "a"
	}
	result := sanitizeBranchName(longName)
	if len(result) > 50 {
		t.Errorf("sanitizeBranchName() result length = %d, want <= 50", len(result))
	}
}
