package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	server "scheduler/appointment-service/api"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/auth"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" //TODO replace to Postgres
)

func main() {
	cfg, err := LoadServiceConfig()
	if err != nil {
		slog.Error("[LoadServiceConfig]", "err", err.Error())
		log.Fatal(err)
	}

	opts := &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}
	logger := common.NewLoggerWithCtxHandler(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	slog.Info("Server started")

	db, err := sqlx.Connect(cfg.DB.Driver, cfg.DB.Connection)
	if err != nil {
		slog.Error("Open db error", "err", err.Error())
		log.Fatal(err)
	}
	defer db.Close()

	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionsKey))
	sessionStore.MaxAge(86400 * 5) // 5 days

	//TODO move LifeTime to config?
	userSessionStore := auth.NewUserSessionStore(sessionStore, auth.WithAuthStatusCheck(), auth.WithSessionLifeTime(time.Hour*24*5))

	router := server.NewRouterBuilder(db, userSessionStore).
		AddTimeSlotsHandlers().
		AddBusinessRulesHandlers().
		AddUserAccountHandlers().
		AppendFileServerLogic(cfg.FrontPath).
		AddOIDCHandlers(cfg.Auth.OAuthGoogleConfig).
		Done()

	err = http.ListenAndServe(cfg.Addr, router)
	if err != nil {
		slog.Error("[http.ListenAndServe]", "err", err.Error())
		log.Fatal(err)
	}
}
