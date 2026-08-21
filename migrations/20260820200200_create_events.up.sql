CREATE TABLE IF NOT EXISTS events (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    script_path VARCHAR(512) NOT NULL,
    agent_user  VARCHAR(255) NOT NULL,
    action      VARCHAR(16)  NOT NULL,
    event_time  TIMESTAMPTZ  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT chk_events_action CHECK (action IN ('open', 'execute'))
);

CREATE INDEX idx_events_script_path ON events (script_path);
