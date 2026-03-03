package main

import (
	"fmt"
	"log"
	"log/slog"
	"maps"
	"net/http"
)



func newAPI() *HttpAPI {
    api := NewHttpAPI()
    RegisterNoInputs(api, "GET /", func(r *http.Request) (*map[string][]string, error) {
        slog.Info("GET /")
        m := make(map[string][]string)
        maps.Copy(m, r.Header)
        m["Host"] = []string{r.Host}
        m["Source"] = []string{r.RemoteAddr}
        m["URL"] = []string{r.URL.String()}
        slog.Info("Response", "data", m)
        return &m, nil
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
