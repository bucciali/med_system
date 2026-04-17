package task

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/usecase/template"
)

type Service struct {
	repo         Repository
	templateRepo template.Repository
	now          func() time.Time
}

func NewService(repo Repository, templateRepo template.Repository) *Service {
	return &Service{
		repo:         repo,
		templateRepo: templateRepo,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

func normalizeDateUTC(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func sameDateUTC(a, b time.Time) bool {
	ay, am, ad := a.UTC().Date()
	by, bm, bd := b.UTC().Date()
	return ay == by && am == bm && ad == bd
}

func containsInt(arr []int, v int) bool {
	for _, x := range arr {
		if x == v {
			return true
		}
	}
	return false
}

func shouldCreateTask(tpl template.Template, date time.Time) bool {
	date = normalizeDateUTC(date)
	start := normalizeDateUTC(tpl.StartDate)

	if date.Before(start) {
		return false
	}
	if tpl.EndDate != nil {
		end := normalizeDateUTC(*tpl.EndDate)
		if date.After(end) {
			return false
		}
	}

	switch tpl.RecurrenceType {
	case "daily":
		interval := tpl.Interval
		if interval <= 0 {
			interval = 1
		}
		diffDays := int(date.Sub(start).Hours() / 24)
		return diffDays%interval == 0

	case "monthly":
		return containsInt(tpl.DaysOfMonth, date.Day())

	case "specific":
		for _, d := range tpl.SpecificDates {
			if sameDateUTC(d, date) {
				return true
			}
		}
		return false

	case "monthly_parity":
		if tpl.Parity == "odd" {
			return date.Day()%2 == 1
		}
		if tpl.Parity == "even" {
			return date.Day()%2 == 0
		}
		return false

	default:
		return false
	}
}

func (s *Service) ListByDate(ctx context.Context, date time.Time) ([]taskdomain.Task, error) {
	date = normalizeDateUTC(date)

	templates, err := s.templateRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, tpl := range templates {
		ok := shouldCreateTask(tpl, date)
		log.Printf("tpl=%d type=%q interval=%d start=%s date=%s ok=%v",
			tpl.ID,
			tpl.RecurrenceType,
			tpl.Interval,
			tpl.StartDate.Format("2006-01-02"),
			date.Format("2006-01-02"),
			ok,
		)
		if !ok {
			continue
		}
		if !shouldCreateTask(tpl, date) {
			continue
		}

		exists, err := s.repo.ExistsByTemplateAndDate(ctx, tpl.ID, date)
		if err != nil {
			return nil, err
		}
		if exists {
			continue
		}

		_, err = s.repo.Create(ctx, &taskdomain.Task{
			Title:        tpl.Title,
			Description:  tpl.Description,
			Status:       taskdomain.StatusNew,
			TemplateID:   &tpl.ID,
			ScheduledFor: date,
		})
		if err != nil {
			return nil, err
		}
	}

	return s.repo.GetByDate(ctx, date)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}
