// Package repository provides the interface for the task repository.
package repository

import (
	"context"

	"github.com/la1665/task-manager/internal/model"
)

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id int64) (*model.Task, error)
	GetAll(ctx context.Context) ([]*model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id int64) error
}
