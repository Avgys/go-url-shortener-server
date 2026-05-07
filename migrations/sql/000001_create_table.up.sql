CREATE TABLE IF NOT EXISTS urls (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    short_url VARCHAR(255) NOT NULL,
    long_url VARCHAR(255) NOT NULL,
    user_id BIGINT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at_utc timestamp NULL
);

CREATE UNIQUE INDEX ux_urls_long_active
ON urls (long_url)
WHERE deleted_at_utc IS NULL;

CREATE UNIQUE INDEX ux_urls_short_active
ON urls (short_url)
WHERE deleted_at_utc IS NULL;