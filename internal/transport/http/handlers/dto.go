package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type createTemplateDTO struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	RecurrenceType string   `json:"recurrence_type"`
	Interval       int      `json:"interval,omitempty"`
	DaysOfMonth    []int    `json:"days_of_month,omitempty"`
	SpecificDates  []string `json:"specific_dates,omitempty"`
	Parity         string   `json:"parity,omitempty"`
	StartDate      string   `json:"start_date"`
	EndDate        *string  `json:"end_date,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
