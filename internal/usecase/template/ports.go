package template

import (
	"context"
	"time"
)

type Template struct {
	ID             int64
	Title          string
	Description    string
	RecurrenceType string
	Interval       int
	DaysOfMonth    []int
	SpecificDates  []time.Time
	Parity         string
	StartDate      time.Time
	EndDate        *time.Time
}

type Repository interface {
	Create(ctx context.Context, tpl *Template) error
	GetAll(ctx context.Context) ([]Template, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) error
	GetAll(ctx context.Context) ([]Template, error)
}

type CreateInput struct {
	Title          string
	Description    string
	RecurrenceType string
	Interval       int
	DaysOfMonth    []int
	SpecificDates  []time.Time
	Parity         string
	StartDate      time.Time
	EndDate        *time.Time
}
