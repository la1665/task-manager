package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/la1665/task-manager/internal/model"
	"github.com/la1665/task-manager/internal/utils"
)

type PostgresTaskRepository struct {
	db *sqlx.DB
}

func NewPostgresTaskRepository(db *sqlx.DB) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

func (r *PostgresTaskRepository) Create(ctx context.Context, task *model.Task) error {
	query := `
		INSERT INTO tasks (title, description, status, priority, assignee)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	row := r.db.QueryRowxContext(
		ctx, query,
		task.Title, task.Description, task.Status, task.Priority, task.Assignee,
	)
	return row.Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *PostgresTaskRepository) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	query := `SELECT * FROM tasks WHERE id = $1`
	var task model.Task
	err := r.db.GetContext(ctx, &task, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (r *PostgresTaskRepository) GetAll(ctx context.Context) ([]*model.Task, error) {
	query := `SELECT * FROM tasks ORDER BY created_at DESC`
	var tasks []*model.Task
	err := r.db.SelectContext(ctx, &tasks, query)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *PostgresTaskRepository) Update(ctx context.Context, task *model.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, priority = $4, assignee = $5, updated_at = now()
		WHERE id = $6
		RETURNING updated_at
	`
	row := r.db.QueryRowxContext(
		ctx, query,
		task.Title, task.Description, task.Status, task.Priority, task.Assignee, task.ID,
	)
	if err := row.Scan(&task.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.ErrTaskNotFound
		}
		return err
	}
	return nil
}

func (r *PostgresTaskRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM tasks WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return utils.ErrTaskNotFound
	}
	return nil
}
