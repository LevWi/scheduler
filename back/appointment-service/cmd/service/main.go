package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	server "scheduler/appointment-service/api"

	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/storage"

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

	storage := storage.Storage{DB: db}
	router := server.NewRouter(&storage)

	log.Fatal(http.ListenAndServe(":8080", router))
}
