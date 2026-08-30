-- +goose Up
CREATE TABLE playbooks (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created     TIMESTAMP,
    modified    TIMESTAMP,
    valid_from  TIMESTAMP,
    valid_until TIMESTAMP,
    labels      TEXT NOT NULL DEFAULT '[]',
    doc         TEXT NOT NULL
);

CREATE TABLE fins (
    fin_id           TEXT PRIMARY KEY,
    fin_token_hash   TEXT NOT NULL UNIQUE,
    display_name     TEXT NOT NULL DEFAULT '',
    protocol_version TEXT NOT NULL DEFAULT '',
    capabilities     TEXT NOT NULL DEFAULT '[]',
    registered_at    TIMESTAMP NOT NULL,
    last_seen        TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE fins;
DROP TABLE playbooks;
