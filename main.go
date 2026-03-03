package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
)



func newAPI() *HttpAPI {
    api := NewHttpAPI()
    result := 42
    RegisterNoInputs(api, "GET /", func() (*int, error) {
        slog.Info("GET /")
        return &result, nil
    })
    return api
}

func main(){
    api := newAPI()
    port := 8080
    srv := http.Server{
        Addr: fmt.Sprintf(":%d", port),
        Handler: api.Mux(),
    }
    slog.Info("Starting api", "addr", fmt.Sprintf("http://0.0.0.0:%d", port))
    log.Fatalln(srv.ListenAndServe())
}
