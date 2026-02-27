package main

import (
	"flag"
	"fmt"
	"os"
	"scheduler/appointment-service/internal/dbase/auth"
	"scheduler/appointment-service/internal/dbase/bots"
	"scheduler/appointment-service/internal/dbase/test"
)

func PrintVar(name string, value any) {
	fmt.Printf("%-10s = %#v\n", name, value)
}

func Fatal(msg any) {
	fmt.Println(msg)
	os.Exit(1)
}

func main() {
	var dbPath string
	flag.StringVar(&dbPath, "db", "", "sqlite database path")

	flag.Parse()

	if dbPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	db, err := test.InitSqliteDB(dbPath)
	if err != nil {
		Fatal(err)
	}

	user := "Bob"
	pass := "123"

	auth := auth.AuthStorage{DB: db}

	userId, err := auth.CreateUserPassword(user, pass)
	if err != nil {
		Fatal(err)
	}

	botStorage := bots.BotsStorage{DB: db}

	botID := "bobs_bot"
	botToken := "bot_token_123"

	botInfo, err := botStorage.AddBot(botID, botToken, userId)
	if err != nil {
		Fatal(err)
	}

	PrintVar("dbPath", dbPath)
	PrintVar("user", user)
	PrintVar("pass", pass)
	PrintVar("userId", userId)
	PrintVar("botID", botID)
	PrintVar("botToken", botToken)
	PrintVar("botInfo", botInfo)

	//slots.TimeSlotsStorage{test.InitTmpDB(t)}

}
