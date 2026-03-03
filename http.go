package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// HttpAPI is a very simple API inspired by HUMA with generc helpers
// for json serialization/deserialization
// See [Register] and [RegisterNoInputs]
type HttpAPI struct {
    mux *http.ServeMux
}

func (api *HttpAPI) Mux() *http.ServeMux {
    return api.mux
}

func NewHttpAPI() *HttpAPI {
    return &HttpAPI{
        mux: http.NewServeMux(),
    }
}

// Register registers a handler for a pattern and handles the json deserialization 
// from request body
func Register[Input any, Output any](
    api *HttpAPI, 
    pattern string,
    handler func(input *Input) (*Output, error),
) {
    api.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Accept", "application/json; charset=utf-8")
        if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
            w.WriteHeader(http.StatusUnsupportedMediaType)
            fmt.Fprintf(w, "use content-type: application/json")
            return
        }
        if r.Body == nil || r.ContentLength == 0 {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprintf(w, "body is required")
            return
        }
        var input Input
        err := json.NewDecoder(r.Body).Decode(&input)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprintf(w, "unable to parse json as %T: %s", input, err.Error())
            return
        }
        output, err := handler(&input)
        if err != nil {
            // todo: check error type
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprint(w, err.Error())
            return
        }
        buf := &bytes.Buffer{}
        err = json.NewEncoder(buf).Encode(output)
        if err != nil {
            // todo: check error type
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintf(w, "json encoding error: %s", err.Error())
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write(buf.Bytes())
    })
}

// RegisterNoInputs registers a handler for a pattern that does not expect
// any inputs and returns the Output serialized as json
func RegisterNoInputs[Output any](
    api *HttpAPI, 
    pattern string,
    handler func() (*Output, error),
) {
    api.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
        output, err := handler()
        if err != nil {
            // todo: check error type
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprint(w, err.Error())
            return
        }
        buf := &bytes.Buffer{}
        err = json.NewEncoder(buf).Encode(output)
        if err != nil {
            // todo: check error type
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintf(w, "json encoding error: %s", err.Error())
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write(buf.Bytes())
    })
}
