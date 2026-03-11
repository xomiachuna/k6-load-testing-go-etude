package main

import (
	"log/slog"
	"net/http"
	"slices"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewOtelHTTPMiddleware() MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return otelhttp.NewHandler(
			h,
			"",
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				return r.RequestURI
			}),
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		)
	}
}

type Middleware interface {
	Wrap(handler http.Handler) http.Handler
}

// Chain represents a chain of middlewares being applied
// in sequential fashion.
type Chain struct {
	middlewares []Middleware
}

func NewChain(middlewares ...Middleware) *Chain {
	// the order of application needs to be reversed
	middlewares = slices.Clone(middlewares)
	slices.Reverse(middlewares)
	return &Chain{
		middlewares: middlewares,
	}
}

func (c *Chain) Wrap(handler http.Handler) http.Handler {
	wrapped := handler
	for _, middleware := range c.middlewares {
		wrapped = middleware.Wrap(wrapped)
	}
	return wrapped
}

// ensure that Chain conforms to Middleware interface.
var _ Middleware = (*Chain)(nil)

// MiddlewareFunc is an alias for a simple middleware function.
type MiddlewareFunc func(http.Handler) http.Handler

func (mf MiddlewareFunc) Wrap(handler http.Handler) http.Handler {
	return mf(handler)
}

// ensure that Func conforms to Middleware interface.
var _ Middleware = MiddlewareFunc(nil)

func WrapFunc(f func(http.ResponseWriter, *http.Request), middleware Middleware) http.Handler {
	return middleware.Wrap(http.HandlerFunc(f))
}

type loggingInterceptor struct {
	statusCode  int
	innerWriter http.ResponseWriter
}

var _ http.ResponseWriter = (*loggingInterceptor)(nil)

func (i *loggingInterceptor) Header() http.Header {
	return i.innerWriter.Header()
}

func (i *loggingInterceptor) Write(data []byte) (int, error) {
	return i.innerWriter.Write(data)
}

func (i *loggingInterceptor) WriteHeader(statusCode int) {
	i.statusCode = statusCode
	i.innerWriter.WriteHeader(statusCode)
}

var LoggingMiddleware = MiddlewareFunc(func(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		interceptor := &loggingInterceptor{innerWriter: w}
		h.ServeHTTP(interceptor, r)
		slog.InfoContext(r.Context(), "request",
			"method", r.Method,
			"url", r.URL,
			"statusCode", interceptor.statusCode,
			"status", http.StatusText(interceptor.statusCode),
		)
	})
})
