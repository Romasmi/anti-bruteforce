-- +goose Up
CREATE TABLE ip_whitelist (
    network inet PRIMARY KEY
);

CREATE TABLE ip_blacklist (
    network inet PRIMARY KEY
);

CREATE INDEX idx_ip_whitelist_network ON ip_whitelist USING gist (network inet_ops);
CREATE INDEX idx_ip_blacklist_network ON ip_blacklist USING gist (network inet_ops);

-- +goose Down
DROP TABLE ip_whitelist;
DROP TABLE ip_blacklist;
