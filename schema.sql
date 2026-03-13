BEGIN;
    CREATE TABLE IF NOT EXISTS sources (
        -- no primary key so far in order to look at the difference in perf before/after
        address TEXT NOT NULL,
        user_agent TEXT NOT NULL,
        CONSTRAINT unique_address_user_agent UNIQUE (address, user_agent)
    ) STRICT;

    CREATE TABLE IF NOT EXISTS requests (
        source_address TEXT NOT NULL,
        source_user_agent TEXT NOT NULL,
        headers TEXT NOT NULL,
        method TEXT NOT NULL,
        url TEXT NOT NULL,
        host TEXT NOT NULL,
        -- this is UTC, with precision of a second (no subsecond precision)
        timestamp TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (source_address, source_user_agent) REFERENCES sources (address, user_agent)
    ) STRICT;
COMMIT;
