// Package service provides the task service layer.
package service

import (
	"context"
	"fmt"

	"github.com/la1665/task-manager/internal/dto"
	"github.com/la1665/task-manager/internal/model"
	"github.com/la1665/task-manager/internal/repository"
)

type TaskService interface {
	CreateTask(ctx context.Context, req dto.CreateTaskRequest) (*dto.TaskResponse, error)
	GetTaskByID(ctx context.Context, id int64) (*dto.TaskResponse, error)
	ListTasks(ctx context.Context) ([]*dto.TaskResponse, error)
	UpdateTask(ctx context.Context, id int64, req dto.UpdateTaskRequest) (*dto.TaskResponse, error)
	DeleteTask(ctx context.Context, id int64) error
}

type taskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func resolvePriority(raw string) (model.TaskPriority, error) {
	if raw == "" {
		return model.TaskPriorityMedium, nil
	}
	p := model.TaskPriority(raw)
	if !p.IsValid() {
		return "", fmt.Errorf("invalid priority: %q", raw)
	}
	return p, nil
}

func resolveStatus(raw string) (model.TaskStatus, error) {
	if raw == "" {
		return model.TaskStatusPending, nil
	}
	s := model.TaskStatus(raw)
	if !s.IsValid() {
		return "", fmt.Errorf("invalid status: %q", raw)
	}
	return s, nil
}

func (s *taskService) CreateTask(ctx context.Context, req dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	priority, err := resolvePriority(req.Priority)
	if err != nil {
		return nil, err
	}

	status, err := resolveStatus(req.Status)
	if err != nil {
		return nil, err
	}

	task := &model.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		Assignee:    req.Assignee,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return toTaskResponse(task), nil
}

func (s *taskService) GetTaskByID(ctx context.Context, id int64) (*dto.TaskResponse, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toTaskResponse(task), nil
}

func (s *taskService) ListTasks(ctx context.Context) ([]*dto.TaskResponse, error) {
	tasks, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return toTaskResponseList(tasks), nil
}

func (s *taskService) UpdateTask(ctx context.Context, id int64, req dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Status != nil {
		status := model.TaskStatus(*req.Status)
		if !status.IsValid() {
			return nil, fmt.Errorf("invalid status: %q", *req.Status)
		}
		existing.Status = status
	}
	if req.Priority != nil {
		priority := model.TaskPriority(*req.Priority)
		if !priority.IsValid() {
			return nil, fmt.Errorf("invalid priority: %q", *req.Priority)
		}
		existing.Priority = priority
	}
	if req.Assignee != nil {
		existing.Assignee = *req.Assignee
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return toTaskResponse(existing), nil
}

func (s *taskService) DeleteTask(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}
