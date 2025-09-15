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
	"scheduler/appointment-service/internal/storage"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	logger := common.NewLoggerWithCtxHandler(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Server started")

	dbPath := "tmp_test.db" //TODO remove

	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		slog.Error(err.Error())
	}
	defer db.Close()

	sessionStore := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	sessionStore.MaxAge(86400 * 5) // 5 days

	strg := &storage.Storage{DB: db}
	storage.CreateOIDCTable(strg)
	storage.CreateUsersTable(strg)
	storage.CreateBusinessTable(strg)
	storage.CreateTableAppointments(strg)
	storage.CreateTableUserBots(strg)

	userSessionStore := auth.NewUserSessionStore(sessionStore, auth.WithAuthStatusCheck(), auth.WithSessionLifeTime(time.Hour*24*5))

	router := server.NewRouter(strg, userSessionStore)

	log.Fatal(http.ListenAndServe(":8080", router)) //TODO
}
