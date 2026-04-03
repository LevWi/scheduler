package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	common "scheduler/appointment-service/internal"
	"slices"
	"strconv"
	"strings"
	"time"
)

const telegramMiniAppAuthScheme = "tma"

var telegramMiniAppProductionPublicKey = ed25519.PublicKey(mustDecodeTelegramHexKey("e7bf03a2fa4602af4580703d88dda5bb59f32ed8b02a56c187fe7d34caed242d"))

func mustDecodeTelegramHexKey(value string) []byte {
	key, err := hex.DecodeString(value)
	if err != nil {
		panic(err)
	}
	return key
}

type TelegramWebAppInitDataValidator struct {
	PublicKey ed25519.PublicKey
	MaxAge    time.Duration
	Now       func() time.Time
}

type TelegramWebAppUser struct {
	ID int64 `json:"id"`
}

type TelegramWebAppInitData struct {
	User TelegramWebAppUser
}

func NewTelegramWebAppInitDataValidator() TelegramWebAppInitDataValidator {
	return TelegramWebAppInitDataValidator{
		PublicKey: telegramMiniAppProductionPublicKey,
		MaxAge:    15 * time.Minute,
		Now:       time.Now,
	}
}

func getClientIDFromHeader(r *http.Request) (common.ID, error) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		return "", fmt.Errorf("%w: 'X-Client-ID' header", common.ErrNotFound)
	}
	return common.ID(clientID), nil
}

func TelegramWebAppAuthToken(r *http.Request) (string, error) {
	return getAuthorizationTokenByScheme(r, telegramMiniAppAuthScheme)
}

func (v TelegramWebAppInitDataValidator) Validate(initDataRaw string, telegramBotID string) (TelegramWebAppInitData, error) {
	var result TelegramWebAppInitData
	if telegramBotID == "" {
		return result, fmt.Errorf("telegram_bot_id: %w", common.ErrNotFound)
	}
	if len(v.PublicKey) != ed25519.PublicKeySize {
		return result, fmt.Errorf("telegram public key: %w", common.ErrInvalidArgument)
	}

	initData, err := url.ParseQuery(initDataRaw)
	if err != nil {
		return result, fmt.Errorf("parse initData: %w", err)
	}

	signature := initData.Get("signature")
	if signature == "" {
		return result, fmt.Errorf("signature: %w", common.ErrNotFound)
	}

	authDateStr := initData.Get("auth_date")
	if authDateStr == "" {
		return result, fmt.Errorf("auth_date: %w", common.ErrNotFound)
	}

	authDateUnix, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return result, fmt.Errorf("auth_date: %w", common.ErrInvalidArgument)
	}

	now := time.Now
	if v.Now != nil {
		now = v.Now
	}
	authDate := time.Unix(authDateUnix, 0)
	if authDate.After(now().Add(time.Minute)) {
		return result, fmt.Errorf("auth_date: %w", common.ErrInvalidArgument)
	}
	if v.MaxAge > 0 && authDate.Before(now().Add(-v.MaxAge)) {
		return result, fmt.Errorf("auth_date expired: %w", common.ErrUnauthorized)
	}

	dataCheckString := buildTelegramWebAppDataCheckString(initData, telegramBotID)

	signatureBytes, err := decodeTelegramSignature(signature)
	if err != nil {
		return result, fmt.Errorf("signature decode: %w", err)
	}
	if !ed25519.Verify(v.PublicKey, []byte(dataCheckString), signatureBytes) {
		return result, fmt.Errorf("signature verify: %w", common.ErrUnauthorized)
	}

	userStr := initData.Get("user")
	if userStr == "" {
		return result, fmt.Errorf("user: %w", common.ErrNotFound)
	}
	if err := json.Unmarshal([]byte(userStr), &result.User); err != nil {
		return result, fmt.Errorf("user: %w", err)
	}
	if result.User.ID == 0 {
		return result, fmt.Errorf("user.id: %w", common.ErrInvalidArgument)
	}

	return result, nil
}

func decodeTelegramSignature(value string) ([]byte, error) {
	signature, err := base64.RawURLEncoding.DecodeString(value)
	if err == nil {
		return signature, nil
	}
	return base64.URLEncoding.DecodeString(value)
}

func buildTelegramWebAppDataCheckString(values url.Values, telegramBotID string) string {
	lines := make([]string, 0, len(values))
	for key, vals := range values {
		if key == "hash" || key == "signature" || len(vals) == 0 {
			continue
		}
		lines = append(lines, key+"="+vals[0])
	}
	slices.Sort(lines)
	return telegramBotID + ":WebAppData\n" + strings.Join(lines, "\n")
}
