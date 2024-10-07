package server

import (
	"log/slog"
	"net/http"
	"time"
)

func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		slog.InfoContext(r.Context(), "request",
			"method", r.Method,
			"uri", r.RequestURI,
			"request", name,
			"start", time.Since(start))

		inner.ServeHTTP(w, r)
	})
}
