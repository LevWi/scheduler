package server

import (
	"log/slog"
	"net/http"

	common "scheduler/appointment-service/internal"

	"github.com/google/uuid"
)

func PassRequestIdToCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid := uuid.New().String()
		ctx := common.AppendCtx(r.Context(), slog.String("request_id", uuid))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
