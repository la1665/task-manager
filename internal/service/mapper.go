package service

import (
	"github.com/la1665/task-manager/internal/dto"
	"github.com/la1665/task-manager/internal/model"
)

func toTaskResponse(task *model.Task) *dto.TaskResponse {
	return &dto.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		Priority:    string(task.Priority),
		Assignee:    task.Assignee,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func toTaskResponseList(tasks []*model.Task) []*dto.TaskResponse {
	result := make([]*dto.TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, toTaskResponse(t))
	}
	return result
}
