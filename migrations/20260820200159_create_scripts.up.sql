CREATE TABLE IF NOT EXISTS scripts (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    host         VARCHAR(255) NOT NULL,
    ssh_user     VARCHAR(255) NOT NULL,
    password     VARCHAR(255) NOT NULL,
    template     VARCHAR(64)  NOT NULL,
    path         VARCHAR(512) NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),

CONSTRAINT uq_scripts_host_path UNIQUE (host, path)
);

CREATE INDEX idx_scripts_host ON scripts (host);
