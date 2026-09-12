package service

import (
	"context"
	"errors"
	"testing"

	"github.com/la1665/task-manager/internal/dto"
	"github.com/la1665/task-manager/internal/model"
	"github.com/la1665/task-manager/internal/utils"
)

type mockTaskRepository struct {
	createFunc  func(ctx context.Context, task *model.Task) error
	getByIDFunc func(ctx context.Context, id int64) (*model.Task, error)
	getAllFunc  func(ctx context.Context) ([]*model.Task, error)
	updateFunc  func(ctx context.Context, task *model.Task) error
	deleteFunc  func(ctx context.Context, id int64) error
}

func (m *mockTaskRepository) Create(ctx context.Context, task *model.Task) error {
	return m.createFunc(ctx, task)
}

func (m *mockTaskRepository) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockTaskRepository) GetAll(ctx context.Context) ([]*model.Task, error) {
	return m.getAllFunc(ctx)
}

func (m *mockTaskRepository) Update(ctx context.Context, task *model.Task) error {
	return m.updateFunc(ctx, task)
}

func (m *mockTaskRepository) Delete(ctx context.Context, id int64) error {
	return m.deleteFunc(ctx, id)
}

func TestTaskService_CreateTask_DefaultsPriorityAndStatus(t *testing.T) {
	var saved *model.Task
	repo := &mockTaskRepository{
		createFunc: func(ctx context.Context, task *model.Task) error {
			task.ID = 42
			saved = task
			return nil
		},
	}
	svc := NewTaskService(repo)

	req := dto.CreateTaskRequest{
		Title:    "Write service tests",
		Assignee: "amir",
	}

	result, err := svc.CreateTask(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if result.ID != 42 {
		t.Errorf("ID = %d, want 42", result.ID)
	}
	if saved.Priority != model.TaskPriorityMedium {
		t.Errorf("Priority = %q, want default %q", saved.Priority, model.TaskPriorityMedium)
	}
	if saved.Status != model.TaskStatusPending {
		t.Errorf("Status = %q, want %q", saved.Status, model.TaskStatusPending)
	}
}

func TestTaskService_CreateTask_InvalidPriority(t *testing.T) {
	svc := NewTaskService(&mockTaskRepository{})

	req := dto.CreateTaskRequest{Title: "bad", Priority: "urgent", Assignee: "amir"}
	if _, err := svc.CreateTask(context.Background(), req); err == nil {
		t.Fatal("expected error for invalid priority, got nil")
	}
}

func TestTaskService_CreateTask_ExplicitStatus(t *testing.T) {
	var saved *model.Task
	repo := &mockTaskRepository{
		createFunc: func(ctx context.Context, task *model.Task) error {
			task.ID = 7
			saved = task
			return nil
		},
	}
	svc := NewTaskService(repo)

	req := dto.CreateTaskRequest{
		Title:    "Already started",
		Assignee: "amir",
		Status:   "in_progress",
	}

	if _, err := svc.CreateTask(context.Background(), req); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if saved.Status != model.TaskStatusInProgress {
		t.Errorf("Status = %q, want %q", saved.Status, model.TaskStatusInProgress)
	}
}

func TestTaskService_CreateTask_InvalidStatus(t *testing.T) {
	svc := NewTaskService(&mockTaskRepository{})

	req := dto.CreateTaskRequest{Title: "bad", Assignee: "amir", Status: "archived"}
	if _, err := svc.CreateTask(context.Background(), req); err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
}

func TestTaskService_GetTaskByID_NotFound(t *testing.T) {
	repo := &mockTaskRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*model.Task, error) {
			return nil, utils.ErrTaskNotFound
		},
	}
	svc := NewTaskService(repo)

	_, err := svc.GetTaskByID(context.Background(), 999)
	if !errors.Is(err, utils.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskService_ListTasks_Success(t *testing.T) {
	repo := &mockTaskRepository{
		getAllFunc: func(ctx context.Context) ([]*model.Task, error) {
			return []*model.Task{
				{ID: 1, Title: "A", Status: model.TaskStatusPending, Priority: model.TaskPriorityLow, Assignee: "amir"},
				{ID: 2, Title: "B", Status: model.TaskStatusDone, Priority: model.TaskPriorityHigh, Assignee: "john"},
			}, nil
		},
	}
	svc := NewTaskService(repo)

	got, err := svc.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(got))
	}
	if got[0].ID != 1 || got[1].ID != 2 {
		t.Errorf("unexpected task ids: %+v", got)
	}
}

func TestTaskService_ListTasks_Error(t *testing.T) {
	repo := &mockTaskRepository{
		getAllFunc: func(ctx context.Context) ([]*model.Task, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewTaskService(repo)

	if _, err := svc.ListTasks(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTaskService_UpdateTask_PartialUpdateKeepsOtherFields(t *testing.T) {
	existing := &model.Task{
		ID: 1, Title: "old title", Description: "old desc",
		Status: model.TaskStatusPending, Priority: model.TaskPriorityLow, Assignee: "amir",
	}

	var updated *model.Task
	repo := &mockTaskRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*model.Task, error) { return existing, nil },
		updateFunc: func(ctx context.Context, task *model.Task) error {
			updated = task
			return nil
		},
	}
	svc := NewTaskService(repo)

	newTitle := "new title"
	got, err := svc.UpdateTask(context.Background(), 1, dto.UpdateTaskRequest{Title: &newTitle})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if got.Title != "new title" {
		t.Errorf("Title = %q, want %q", got.Title, "new title")
	}
	if updated.Description != "old desc" {
		t.Errorf("Description changed unexpectedly to %q", updated.Description)
	}
	if updated.Priority != model.TaskPriorityLow {
		t.Errorf("Priority changed unexpectedly to %q", updated.Priority)
	}
}

func TestTaskService_DeleteTask_NotFound(t *testing.T) {
	repo := &mockTaskRepository{
		deleteFunc: func(ctx context.Context, id int64) error { return utils.ErrTaskNotFound },
	}
	svc := NewTaskService(repo)

	if err := svc.DeleteTask(context.Background(), 999); !errors.Is(err, utils.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}
