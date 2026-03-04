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

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "embed"
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

func connectToDatabase(ctx context.Context) (*sql.DB, error) {
    driver := "pgx"
    connString := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
    db, err := sql.Open(driver, connString)
    if err != nil {
        return nil, fmt.Errorf("connect to %s db at %s: %w", driver, connString, err)
    }
    err = db.PingContext(ctx)
    if err != nil {
        return nil, fmt.Errorf("connect to %s db at %s: %w", driver, connString, err)
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
    insertIntoSourcesIfNotExist := `INSERT INTO sources (address, user_agent) VALUES ($1, $2) ON CONFLICT DO NOTHING`
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
    VALUES ($1, $2, $3, $4, $5, $6)`
    headers, err := json.Marshal(entry.Headers)
    if err != nil {
        return fmt.Errorf("add request log entry (%#v): %w", *entry, err)
    }
    _, err = tx.ExecContext(
        ctx,
        insertIntoRequests,
        entry.Source.Address,
        entry.Source.UserAgent,
        headers,
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

func newAPI(db *sql.DB) *HttpAPI {
    api := NewHttpAPI()
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
        slog.Info("Response", "data", *m)
        return m, nil
    })
    return api
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
    rootCtx, cancel := context.WithCancel(context.Background())
    defer cancel()
    db, err := connectToDatabase(rootCtx)
    if err != nil {
        log.Fatalln(err)
    }
    defer db.Close()
    err =  migrateDB(rootCtx, db) 
    if err != nil {
        log.Fatalln(err)
    }
    api := newAPI(db)
    port := 8080
    srv := http.Server{
        Addr: fmt.Sprintf(":%d", port),
        Handler: api.Mux(),
        BaseContext: func(_ net.Listener) context.Context { return rootCtx },
    }
    slog.Info("Starting api", "addr", fmt.Sprintf("http://0.0.0.0:%d", port))
    log.Fatalln(srv.ListenAndServe())
}
