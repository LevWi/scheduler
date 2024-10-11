package server

import (
	"context"
	"log/slog"
	"net/http"

	common "scheduler/appointment-service/internal"

	"github.com/google/uuid"
)

type RequestIdKey struct{}

func PassRequestIdToCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIdKey{}, uuid)
		ctx = common.AppendSlogCtx(ctx, slog.String("request_id", uuid))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
