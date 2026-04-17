package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/taskservice/internal/usecase/template"
)

type TemplateRepository struct {
	pool *pgxpool.Pool
}

func NewTemplateRepo(pool *pgxpool.Pool) *TemplateRepository {
	return &TemplateRepository{pool: pool}
}

func (r *TemplateRepository) Create(ctx context.Context, tpl *template.Template) error {
	query := `
		INSERT INTO task_templates
		(title, description, recurrence_type, interval, days_of_month, specific_dates, parity, start_date, end_date)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`

	_, err := r.pool.Exec(ctx, query,
		tpl.Title,
		tpl.Description,
		tpl.RecurrenceType,
		tpl.Interval,
		tpl.DaysOfMonth,
		tpl.SpecificDates,
		tpl.Parity,
		tpl.StartDate,
		tpl.EndDate,
	)

	return err
}

func (r *TemplateRepository) GetAll(ctx context.Context) ([]template.Template, error) {
	query := `
		SELECT id, title, description, recurrence_type, interval, days_of_month, specific_dates, parity, start_date, end_date
		FROM task_templates
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []template.Template

	for rows.Next() {
		var t template.Template

		err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&t.RecurrenceType,
			&t.Interval,
			&t.DaysOfMonth,
			&t.SpecificDates,
			&t.Parity,
			&t.StartDate,
			&t.EndDate,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, t)
	}

	return result, nil
}
