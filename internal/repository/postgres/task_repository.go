package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func normalizeDateUTC(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (r *Repository) CreateIfNotExists(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const q = `
WITH ins AS (
	INSERT INTO tasks (title, description, status, template_id, scheduled_for, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	ON CONFLICT (template_id, scheduled_for) DO NOTHING
	RETURNING id, title, description, status, template_id, scheduled_for, created_at, updated_at
)
SELECT id, title, description, status, template_id, scheduled_for, created_at, updated_at
FROM ins
UNION ALL
SELECT id, title, description, status, template_id, scheduled_for, created_at, updated_at
FROM tasks
WHERE template_id = $4 AND scheduled_for = $5
LIMIT 1;
`

	var out taskdomain.Task
	var templateID int64 // если у тебя в домене TemplateID *int64

	err := r.pool.QueryRow(
		ctx, q,
		task.Title,
		task.Description,
		task.Status,
		*task.TemplateID,
		task.ScheduledFor,
	).Scan(
		&out.ID,
		&out.Title,
		&out.Description,
		&out.Status,
		&templateID,
		&out.ScheduledFor,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	out.TemplateID = &templateID
	return &out, nil
}

func (r *Repository) GetByDate(ctx context.Context, date time.Time) ([]taskdomain.Task, error) {
	date = normalizeDateUTC(date)

	query := `
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE scheduled_for = $1::date
		ORDER BY id
	`

	rows, err := r.pool.Query(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskdomain.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, rows.Err()
}

func (r *Repository) ExistsByTemplateAndDate(ctx context.Context, templateID int64, date time.Time) (bool, error) {
	date = normalizeDateUTC(date)

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM tasks
			WHERE template_id = $1
			  AND scheduled_for = $2::date
		)
	`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, templateID, date).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, template_id, scheduled_for)
		VALUES ($1,$2,$3,$4,$5::date)
		RETURNING id, title, description, status, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.TemplateID,
		normalizeDateUTC(task.ScheduledFor),
	)

	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
		RETURNING id, title, description, status, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}
