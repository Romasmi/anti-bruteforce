-- +goose Up
CREATE TABLE ip_whitelist (
    network TEXT PRIMARY KEY
);

CREATE TABLE ip_blacklist (
    network TEXT PRIMARY KEY
);

-- +goose Down
DROP TABLE ip_whitelist;
DROP TABLE ip_blacklist;
