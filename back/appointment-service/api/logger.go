package api

import (
	"log/slog"
	"net/http"
	"time"
)

// TODO Expand this to middleware. Remove "name" ? Add function for logging for each handler
func Logger(inner http.Handler, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		slog.InfoContext(r.Context(), "request",
			"method", r.Method,
			"uri", r.RequestURI,
			"request", name,
			"start", time.Since(start))

		inner.ServeHTTP(w, r)
	}
}
