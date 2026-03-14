package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"os"
	"runtime"

	_ "embed"
	_ "net/http/pprof"

	_ "modernc.org/sqlite"
)

type Request struct {
    Source Source 
    Headers map[string][]string
    URL string
    Host string
    Method string
}

type Source struct {
    Address string
    UserAgent string
}

type DBSerializer struct {
    db *sql.DB
    queue chan func(*sql.DB)
}

func NewDbSerializer(db *sql.DB, size int) *DBSerializer {
    return &DBSerializer{
        db: db,
        queue: make(chan func(*sql.DB), size),
    }
}

func (s *DBSerializer) Run(work func(*sql.DB)) {
    s.queue <- work
}

func (s *DBSerializer) Poll(ctx context.Context) error {
    for {
        select {
        case work := <- s.queue:
            work(s.db)
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func mustGetConnString() string {
    connString, ok := os.LookupEnv("DB_CONN_STRING")
    if !ok {
        log.Fatalf("DB_CONN_STRING not set")
    }
    return connString
}

func connectToDatabase(ctx context.Context) (*sql.DB, error) {
    driver := "sqlite"
    connString := mustGetConnString()
    db, err := sql.Open(driver, connString)
    if err != nil {
        return nil, fmt.Errorf("connect to db: %w", err)
    }

    pragmas := []string{
		"PRAGMA journal_mode = WAL",   // Write-Ahead Logging for better concurrency
		"PRAGMA synchronous = NORMAL", // Good balance of safety and performance
		"PRAGMA foreign_keys = ON",    // Enable foreign key constraints
		"PRAGMA busy_timeout = 5000",  // Wait up to 5s when database is locked
		"PRAGMA cache_size = -64000",  // 64MB cache (negative = KB, positive = pages)
		"PRAGMA temp_store = MEMORY",  // Store temp tables in memory
	}

    for _, pragma := range pragmas {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
            return nil, fmt.Errorf("failed to set pragma %s: %w", pragma, err)
		}
	}
        
    db.SetMaxOpenConns(1)

    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("connect to db: %w", err)
    }
    slog.Info("Connected to database", "driver", driver, "conn", connString)
    return db, nil
}

// source_address TEXT NOT NULL,
// source_user_agent TEXT NOT NULL,
// FOREIGN KEY (source_address, source_user_agent) REFERENCES sources (address, user_agent),
// 
// headers JSONB NOT NULL,
// method TEXT NOT NULL,
// url TEXT NOT NULL,
// host TEXT NOT NULL

func addRequestLogEntry(ctx context.Context, db *sql.DB, entry *Request) (err error) {
    tx, err := db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelDefault,
    })
    if err != nil {
        return fmt.Errorf("add request log entry (%#v): %w", *entry, err)
    }
    defer func(){
        if err != nil {
            slog.ErrorContext(ctx, "Rolling back", "cause", err.Error())
            err = errors.Join(tx.Rollback(), err)
        }
    }()
    insertIntoSourcesIfNotExist := `INSERT INTO sources (address, user_agent) VALUES (?, ?) ON CONFLICT DO NOTHING`
    _, err = tx.ExecContext(ctx, insertIntoSourcesIfNotExist, entry.Source.Address, entry.Source.UserAgent)
    if err != nil {
        return fmt.Errorf("add request log entry (%#v): %w", *entry, err)
    }
    insertIntoRequests := `INSERT INTO requests (
        source_address,
        source_user_agent,
        headers,
        method,
        url,
        host
    )
    VALUES (?, ?, ?, ?, ?, ?)`
    headers, err := json.Marshal(entry.Headers)
    if err != nil {
        return fmt.Errorf("add request log entry (%#v): %w", *entry, err)
    }
    _, err = tx.ExecContext(
        ctx,
        insertIntoRequests,
        entry.Source.Address,
        entry.Source.UserAgent,
        string(headers),
        entry.Method,
        entry.URL,
        entry.Host,
    )
    if err != nil {
        return fmt.Errorf("add request log entry (%#v): %w", *entry, err)
    }
    err = tx.Commit()
    return
}

func newAPIHandler(db *sql.DB) http.Handler {
    api := NewHttpAPI()
    instrumentPrometheus(api)
    RegisterNoInputs(api, "POST /", func(r *http.Request) (*Request, error) {
        m := &Request{
            Host: r.Host,
            Source: Source{
                Address: r.RemoteAddr,
                UserAgent: r.UserAgent(),
            },
            Headers: make(map[string][]string), 
            URL: r.URL.String(),
            Method: r.Method,
        }
        maps.Copy(m.Headers, r.Header)
        err := addRequestLogEntry(r.Context(), db, m)
        if err != nil {
            return nil, fmt.Errorf("post %s: %w", r.URL.String(), err)
        }
        // slog.Info("Response", "data", *m)
        return m, nil
    })
    globalMiddleware := NewChain(
		// LoggingMiddleware,
		NewOtelHTTPMiddleware(),
	)

    return globalMiddleware.Wrap(api.Mux())
}

//go:embed schema.sql
var schema string

func migrateDB(ctx context.Context, db *sql.DB) error {
    _, err := db.ExecContext(ctx, schema)
    if err != nil {
        return fmt.Errorf("migrateDB (schema=%s): %w", schema, err)
    }
    slog.Info("Migrated the database")
    return nil
}

func main(){
    go func(){
        // pprof
        runtime.SetBlockProfileRate(1)
        runtime.SetMutexProfileFraction(1)
        slog.Info("Starting pprof", "addr", "http://0.0.0.0:6060")
        log.Fatalln(http.ListenAndServe(":6060", nil))
    }()
    rootCtx, cancel := context.WithCancel(context.Background())
    defer cancel()
    shutdownOtel, err := SetupOtelSDK(rootCtx)
    if err != nil {
        log.Fatalln(err)
    }
    defer shutdownOtel(rootCtx)
    db, err := connectToDatabase(rootCtx)
    if err != nil {
        log.Fatalln(err)
    }
    defer db.Close()
    err =  migrateDB(rootCtx, db) 
    if err != nil {
        log.Fatalln(err)
    }
    api := newAPIHandler(db)
    port := 8080
    srv := http.Server{
        Addr: fmt.Sprintf(":%d", port),
        Handler: api,
        BaseContext: func(_ net.Listener) context.Context { return rootCtx },
    }
    slog.Info("Starting api", "addr", fmt.Sprintf("http://0.0.0.0:%d", port))
    log.Fatalln(srv.ListenAndServe())
}
