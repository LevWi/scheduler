package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http/httptest"
	"net/url"
	common "scheduler/appointment-service/internal"
	"testing"
	"time"
)

func TestTelegramWebAppInitDataValidatorValidate(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Unix(1_700_000_000, 0)
	validator := TelegramWebAppInitDataValidator{
		PublicKey: publicKey,
		MaxAge:    time.Hour,
		Now: func() time.Time {
			return now
		},
	}

	initData := signedTelegramInitData(t, privateKey, "777000", now, `{"id":123456789,"first_name":"Test"}`)

	result, err := validator.Validate(initData, "777000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.User.ID != 123456789 {
		t.Fatalf("unexpected user id: %d", result.User.ID)
	}
}

func TestTelegramWebAppInitDataValidatorValidateExpired(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Unix(1_700_000_000, 0)
	validator := TelegramWebAppInitDataValidator{
		PublicKey: publicKey,
		MaxAge:    time.Minute,
		Now: func() time.Time {
			return now
		},
	}

	initData := signedTelegramInitData(t, privateKey, "777000", now.Add(-2*time.Minute), `{"id":123456789}`)

	_, err = validator.Validate(initData, "777000")
	if err == nil || !errors.Is(err, common.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestGetTelegramWebAppAuthToken(t *testing.T) {
	req := httptest.NewRequest("POST", "/slots/webapp", nil)
	req.Header.Set("Authorization", "tma init-data")

	token, err := TelegramWebAppAuthToken(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "init-data" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func signedTelegramInitData(t *testing.T, privateKey ed25519.PrivateKey, telegramBotID string, authDate time.Time, user string) string {
	t.Helper()

	values := url.Values{}
	values.Set("auth_date", fmt.Sprintf("%d", authDate.Unix()))
	values.Set("query_id", "AAHdF6IQAAAAAN0XohDhrOrc")
	values.Set("user", user)

	payload := buildTelegramWebAppDataCheckString(values, telegramBotID)
	signature := ed25519.Sign(privateKey, []byte(payload))
	values.Set("signature", base64.RawURLEncoding.EncodeToString(signature))
	values.Set("hash", "ignored-for-third-party-validation")

	return values.Encode()
}
