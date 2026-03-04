BEGIN;
    CREATE TABLE IF NOT EXISTS sources (
        -- no primary key so far in order to look at the difference in perf before/after
        address TEXT NOT NULL,
        user_agent TEXT NOT NULL,
        CONSTRAINT unique_address_user_agent UNIQUE (address, user_agent)
    );

    CREATE TABLE IF NOT EXISTS requests (
        -- no primary key so far in order to look at the difference in perf before/after
        source_address TEXT NOT NULL,
        source_user_agent TEXT NOT NULL,
        FOREIGN KEY (source_address, source_user_agent) REFERENCES sources (address, user_agent),

        headers JSONB NOT NULL,
        method TEXT NOT NULL,
        url TEXT NOT NULL,
        host TEXT NOT NULL,

        timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
COMMIT;
