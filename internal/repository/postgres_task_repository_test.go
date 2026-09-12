package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/la1665/task-manager/internal/model"
	"github.com/la1665/task-manager/internal/utils"

	"github.com/joho/godotenv"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	_ = godotenv.Load("../../.env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set, skipping integration test")
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	db.MustExec("TRUNCATE TABLE tasks RESTART IDENTITY")
	t.Cleanup(func() {
		db.MustExec("TRUNCATE TABLE tasks RESTART IDENTITY")
		db.Close()
	})

	return db
}

func TestPostgresTaskRepository_CreateAndGetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresTaskRepository(db)
	ctx := context.Background()

	task := &model.Task{
		Title:    "learning go",
		Status:   model.TaskStatusPending,
		Priority: model.TaskPriorityMedium,
		Assignee: "Amir",
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Title != task.Title {
		t.Errorf("Title = %q, want %q", got.Title, task.Title)
	}
	if got.Priority != task.Priority {
		t.Errorf("Priority = %q, want %q", got.Priority, task.Priority)
	}
}

func TestPostgresTaskRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresTaskRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 1)
	if !errors.Is(err, utils.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestPostgresTaskRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresTaskRepository(db)
	ctx := context.Background()

	tasks := []*model.Task{
		{ID: 1, Title: "Task A", Status: model.TaskStatusPending, Priority: model.TaskPriorityLow, Assignee: "John"},
		{ID: 2, Title: "Task B", Status: model.TaskStatusDone, Priority: model.TaskPriorityHigh, Assignee: "Doe"},
	}
	for _, task := range tasks {
		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	got, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(got))
	}
	for _, task := range got {
		if task.ID != 1 && task.ID != 2 {
			t.Fatalf("expected task ID to be 1 or 2, got %d", task.ID)
		}
	}
}

func TestPostgresTaskRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresTaskRepository(db)
	ctx := context.Background()

	task := &model.Task{
		Title:  "Original title",
		Status: model.TaskStatusPending, Priority: model.TaskPriorityMedium, Assignee: "John",
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	task.Title = "Updated title"
	task.Status = model.TaskStatusInProgress
	if err := repo.Update(ctx, task); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Title != "Updated title" || got.Status != model.TaskStatusInProgress {
		t.Errorf("got = %+v, want updated title/status", got)
	}
}

func TestPostgresTaskRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresTaskRepository(db)
	ctx := context.Background()

	task := &model.Task{ID: 1, Title: "ghost"}
	if err := repo.Update(ctx, task); !errors.Is(err, utils.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestPostgresTaskRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresTaskRepository(db)
	ctx := context.Background()

	task := &model.Task{
		Title:  "to delete",
		Status: model.TaskStatusPending, Priority: model.TaskPriorityLow, Assignee: "John",
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.GetByID(ctx, task.ID); !errors.Is(err, utils.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound after delete, got %v", err)
	}
}

func TestPostgresTaskRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresTaskRepository(db)
	ctx := context.Background()

	if err := repo.Delete(ctx, 1); !errors.Is(err, utils.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}
