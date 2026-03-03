package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/dbase/auth"
	"scheduler/appointment-service/internal/dbase/backend/slots"
	"scheduler/appointment-service/internal/dbase/bots"
	"scheduler/appointment-service/internal/dbase/test"

	"github.com/teambition/rrule-go"
)

const UnknownFn = "<fn unknown>"

func CallerLine(skip int) int {
	_, _, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return -1
	}
	return line
}

func PrintVar(name string, value any) {
	fmt.Printf("%-10s = %#v\n", name, value)
}

func Fatal(msg any) {
	_, _, line, ok := runtime.Caller(1)
	if !ok {
		line = -1
	}
	fmt.Println("L:", line)
	fmt.Println(msg)
	os.Exit(1)
}

func createRule() common.IntervalRRule {
	const rruleStr = "DTSTART=20260101T090000Z;FREQ=DAILY;COUNT=1000"
	var rule common.IntervalRRule
	{
		r, e := rrule.StrToRRule(rruleStr)
		if e != nil {
			Fatal(e)
		}
		rule.RRule = r
		rule.Len = 60 * 60 * 8
	}
	return rule
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

	slts := slots.TimeSlotsStorage{DB: db}

	rule := common.IntervalRRuleWithType{
		Rule: createRule(),
		Type: common.Inclusion,
	}

	strRule, err := json.Marshal(rule)
	if err != nil {
		Fatal(err)
	}

	err = slts.AddBusinessRule(userId, string(strRule))
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
	PrintVar("Rule", string(strRule))
}
