ALTER TABLE tasks
ADD COLUMN template_id BIGINT,
ADD COLUMN scheduled_for DATE;
