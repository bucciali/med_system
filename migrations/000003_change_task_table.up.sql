ALTER TABLE tasks
ADD COLUMN template_id BIGINT,
ADD COLUMN scheduled_for DATE;

CREATE UNIQUE INDEX IF NOT EXISTS ux_tasks_template_date
ON tasks(template_id, scheduled_for);