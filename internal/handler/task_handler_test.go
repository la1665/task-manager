package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/la1665/task-manager/internal/dto"
	"github.com/la1665/task-manager/internal/utils"
)

type mockTaskService struct {
	createTaskFunc  func(ctx context.Context, req dto.CreateTaskRequest) (*dto.TaskResponse, error)
	getTaskByIDFunc func(ctx context.Context, id int64) (*dto.TaskResponse, error)
	listTasksFunc   func(ctx context.Context) ([]*dto.TaskResponse, error)
	updateTaskFunc  func(ctx context.Context, id int64, req dto.UpdateTaskRequest) (*dto.TaskResponse, error)
	deleteTaskFunc  func(ctx context.Context, id int64) error
}

func (m *mockTaskService) CreateTask(ctx context.Context, req dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	return m.createTaskFunc(ctx, req)
}

func (m *mockTaskService) GetTaskByID(ctx context.Context, id int64) (*dto.TaskResponse, error) {
	return m.getTaskByIDFunc(ctx, id)
}

func (m *mockTaskService) ListTasks(ctx context.Context) ([]*dto.TaskResponse, error) {
	return m.listTasksFunc(ctx)
}

func (m *mockTaskService) UpdateTask(ctx context.Context, id int64, req dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	return m.updateTaskFunc(ctx, id, req)
}

func (m *mockTaskService) DeleteTask(ctx context.Context, id int64) error {
	return m.deleteTaskFunc(ctx, id)
}

func setupRouter(h *TaskHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/tasks", h.CreateTask)
	r.GET("/tasks", h.ListTasks)
	r.GET("/tasks/:id", h.GetTask)
	r.PUT("/tasks/:id", h.UpdateTask)
	r.DELETE("/tasks/:id", h.DeleteTask)
	return r
}

func TestTaskHandler_CreateTask_Success(t *testing.T) {
	svc := &mockTaskService{
		createTaskFunc: func(ctx context.Context, req dto.CreateTaskRequest) (*dto.TaskResponse, error) {
			return &dto.TaskResponse{ID: 1, Title: req.Title, Status: "pending", Priority: "medium", Assignee: req.Assignee}, nil
		},
	}
	router := setupRouter(NewTaskHandler(svc))

	body, _ := json.Marshal(dto.CreateTaskRequest{Title: "Test", Assignee: "ali"})
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d (%s)", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestTaskHandler_CreateTask_MissingRequiredField(t *testing.T) {
	router := setupRouter(NewTaskHandler(&mockTaskService{}))

	body, _ := json.Marshal(map[string]string{"description": "no title, no assignee"})
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTaskHandler_CreateTask_ServiceValidationError(t *testing.T) {
	svc := &mockTaskService{
		createTaskFunc: func(ctx context.Context, req dto.CreateTaskRequest) (*dto.TaskResponse, error) {
			return nil, errors.New("invalid priority")
		},
	}
	router := setupRouter(NewTaskHandler(svc))

	body, _ := json.Marshal(dto.CreateTaskRequest{Title: "T", Assignee: "ali"})
	req, _ := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTaskHandler_GetTask_Success(t *testing.T) {
	svc := &mockTaskService{
		getTaskByIDFunc: func(ctx context.Context, id int64) (*dto.TaskResponse, error) {
			return &dto.TaskResponse{ID: id, Title: "Found"}, nil
		},
	}
	router := setupRouter(NewTaskHandler(svc))

	req, _ := http.NewRequest(http.MethodGet, "/tasks/5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTaskHandler_GetTask_NotFound(t *testing.T) {
	svc := &mockTaskService{
		getTaskByIDFunc: func(ctx context.Context, id int64) (*dto.TaskResponse, error) {
			return nil, utils.ErrTaskNotFound
		},
	}
	router := setupRouter(NewTaskHandler(svc))

	req, _ := http.NewRequest(http.MethodGet, "/tasks/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestTaskHandler_GetTask_InvalidID(t *testing.T) {
	router := setupRouter(NewTaskHandler(&mockTaskService{}))

	req, _ := http.NewRequest(http.MethodGet, "/tasks/not-a-number", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestTaskHandler_ListTasks_Success(t *testing.T) {
	svc := &mockTaskService{
		listTasksFunc: func(ctx context.Context) ([]*dto.TaskResponse, error) {
			return []*dto.TaskResponse{{ID: 1}, {ID: 2}}, nil
		},
	}
	router := setupRouter(NewTaskHandler(svc))

	req, _ := http.NewRequest(http.MethodGet, "/tasks", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTaskHandler_UpdateTask_Success(t *testing.T) {
	svc := &mockTaskService{
		updateTaskFunc: func(ctx context.Context, id int64, req dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
			return &dto.TaskResponse{ID: id, Title: *req.Title}, nil
		},
	}
	router := setupRouter(NewTaskHandler(svc))

	newTitle := "updated"
	body, _ := json.Marshal(dto.UpdateTaskRequest{Title: &newTitle})
	req, _ := http.NewRequest(http.MethodPut, "/tasks/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTaskHandler_UpdateTask_NotFound(t *testing.T) {
	svc := &mockTaskService{
		updateTaskFunc: func(ctx context.Context, id int64, req dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
			return nil, utils.ErrTaskNotFound
		},
	}
	router := setupRouter(NewTaskHandler(svc))

	body, _ := json.Marshal(dto.UpdateTaskRequest{})
	req, _ := http.NewRequest(http.MethodPut, "/tasks/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestTaskHandler_DeleteTask_Success(t *testing.T) {
	svc := &mockTaskService{
		deleteTaskFunc: func(ctx context.Context, id int64) error { return nil },
	}
	router := setupRouter(NewTaskHandler(svc))

	req, _ := http.NewRequest(http.MethodDelete, "/tasks/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestTaskHandler_DeleteTask_NotFound(t *testing.T) {
	svc := &mockTaskService{
		deleteTaskFunc: func(ctx context.Context, id int64) error { return utils.ErrTaskNotFound },
	}
	router := setupRouter(NewTaskHandler(svc))

	req, _ := http.NewRequest(http.MethodDelete, "/tasks/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, w.Code)
	}
}
