package template

import (
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) error {
	tpl := &Template{
		Title:          input.Title,
		Description:    input.Description,
		RecurrenceType: input.RecurrenceType,
		Interval:       input.Interval,
		DaysOfMonth:    input.DaysOfMonth,
		SpecificDates:  input.SpecificDates,
		Parity:         input.Parity,
		StartDate:      input.StartDate,
		EndDate:        input.EndDate,
	}

	return s.repo.Create(ctx, tpl)
}

func (s *Service) GetAll(ctx context.Context) ([]Template, error) {
	return s.repo.GetAll(ctx)
}
