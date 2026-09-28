-- +goose Up
-- +goose StatementBegin

ALTER TABLE outbox_event ADD COLUMN status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE outbox_event ADD COLUMN attempts INT NOT NULL DEFAULT 0;
ALTER TABLE outbox_event ADD COLUMN last_error TEXT;
ALTER TABLE outbox_event ADD COLUMN sent_at TIMESTAMPTZ;

CREATE INDEX idx_outbox_events_pending ON outbox_event (created_at) WHERE status = 'pending';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_outbox_events_pending;
ALTER TABLE outbox_event DROP COLUMN status;
ALTER TABLE outbox_event DROP COLUMN attempts;
ALTER TABLE outbox_event DROP COLUMN last_error;
ALTER TABLE outbox_event DROP COLUMN sent_at;

-- +goose StatementEnd
