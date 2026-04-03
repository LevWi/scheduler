package api

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"net/http/httptest"
	"net/url"
	tgauth "scheduler/appointment-service/internal/auth"
	"scheduler/appointment-service/internal/dbase/bots"
	"scheduler/appointment-service/internal/dbase/test"
	"testing"
	"time"
)

func TestAddSlotsAuthWebAppAuthorization(t *testing.T) {
	db := test.InitTmpDB(t)
	storage := &bots.BotsStorage{DB: db}

	const (
		clientID   = "777000"
		businessID = "biz-webapp"
	)
	if _, err := storage.AddBot(clientID, "secret-token", businessID); err != nil {
		t.Fatal(err)
	}

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Unix(1_700_000_000, 0)
	validator := tgauth.TelegramWebAppInitDataValidator{
		PublicKey: publicKey,
		MaxAge:    time.Hour,
		Now: func() time.Time {
			return now
		},
	}

	req := httptest.NewRequest("POST", "/slots/webapp", nil)
	req.Header.Set("X-Client-ID", clientID)
	req.Header.Set("Authorization", "tma "+signedAPITelegramInitData(t, privateKey, clientID, now, `{"id":555001}`))

	result, err := (AddSlotsAuthTgWebApp{
		BotsStorage: storage,
		Validator:   validator,
	}).Authorization(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Business != businessID {
		t.Fatalf("unexpected business id: %s", result.Business)
	}
	if result.Customer != "555001" {
		t.Fatalf("unexpected customer id: %s", result.Customer)
	}
}

func signedAPITelegramInitData(t *testing.T, privateKey ed25519.PrivateKey, telegramBotID string, authDate time.Time, user string) string {
	t.Helper()

	values := url.Values{}
	values.Set("auth_date", fmt.Sprintf("%d", authDate.Unix()))
	values.Set("query_id", "api-webapp-query")
	values.Set("user", user)
	payload := telegramBotID + ":WebAppData\nauth_date=" + values.Get("auth_date") + "\nquery_id=" + values.Get("query_id") + "\nuser=" + user
	values.Set("signature", base64.RawURLEncoding.EncodeToString(ed25519.Sign(privateKey, []byte(payload))))
	return values.Encode()
}
