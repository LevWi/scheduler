package bots

import (
	"scheduler/appointment-service/internal/dbase/test"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAddAndGetBot(t *testing.T) {
	st := BotsStorage{test.InitTmpDB(t)}

	botId := "bot_123"
	tokenHash := "hash_abc"
	businessId := "biz_1"

	bot, err := st.AddBot(botId, tokenHash, businessId)
	assert.NoError(t, err)

	assert.Equal(t, botId, bot.BotId)
	err = bcrypt.CompareHashAndPassword([]byte(bot.BotTokenHash), []byte(tokenHash))
	assert.NoError(t, err)
	assert.Equal(t, businessId, bot.BusinessId)
	assert.True(t, bot.IsActive)
	assert.WithinDuration(t, time.Now(), bot.CreatedAt, time.Minute)
	assert.WithinDuration(t, time.Now(), bot.LastUsedAt, time.Minute)

	// Получаем список ботов по бизнесу
	bots, err := st.GetBots(businessId)
	assert.NoError(t, err)
	assert.Len(t, bots, 1)

	got := bots[0]
	assert.Equal(t, botId, got.BotId)
	err = bcrypt.CompareHashAndPassword([]byte(bot.BotTokenHash), []byte(tokenHash))
	assert.NoError(t, err)
	assert.Equal(t, businessId, got.BusinessId)
	assert.True(t, got.IsActive)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Minute)
	assert.WithinDuration(t, time.Now(), got.LastUsedAt, time.Minute)
}

func TestEditBotStatus(t *testing.T) {
	st := BotsStorage{test.InitTmpDB(t)}

	botId := "bot_456"
	businessId := "biz_2"

	_, err := st.AddBot(botId, "hash_xyz", businessId)
	assert.NoError(t, err)

	err = st.EditBotStatus(businessId, botId, false)
	assert.NoError(t, err)

	bots, err := st.GetBots(businessId)
	assert.NoError(t, err)
	assert.Len(t, bots, 1)

	got := bots[0]
	assert.Equal(t, botId, got.BotId)
	err = bcrypt.CompareHashAndPassword([]byte(got.BotTokenHash), []byte("hash_xyz"))
	assert.NoError(t, err)
	assert.Equal(t, businessId, got.BusinessId)
	assert.False(t, got.IsActive)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Minute)
	assert.WithinDuration(t, time.Now(), got.LastUsedAt, time.Minute)
}

func TestDeleteBot(t *testing.T) {
	st := BotsStorage{test.InitTmpDB(t)}

	botId := "bot_789"
	businessId := "biz_3"

	_, err := st.AddBot(botId, "hash_del", businessId)
	assert.NoError(t, err)

	err = st.DeleteBot(businessId, botId)
	assert.NoError(t, err)

	bots, err := st.GetBots(businessId)
	assert.NoError(t, err)
	assert.Empty(t, bots)
}

func TestBotIdUniqueness(t *testing.T) {
	st := BotsStorage{test.InitTmpDB(t)}

	botId := "bot_unique"
	businessId := "biz_uniq"

	_, err := st.AddBot(botId, "hash1", businessId)
	assert.NoError(t, err)

	_, err = st.AddBot(botId, "hash2", businessId)
	assert.Error(t, err)
}

func TestValidateBotToken(t *testing.T) {
	st := BotsStorage{test.InitTmpDB(t)}

	botId := "bot_validate"
	businessId := "biz_validate"
	token := "token_secret"

	_, err := st.AddBot(botId, token, businessId)
	assert.NoError(t, err)

	// Test case 1: Valid token
	validatedBusinessId, err := st.ValidateBotToken(botId, token)
	assert.NoError(t, err)
	assert.Equal(t, businessId, validatedBusinessId)

	// Test case 2: Invalid token
	_, err = st.ValidateBotToken(botId, "wrong_token")
	assert.ErrorIs(t, err, ErrTokenMismatch)

	// Test case 3: Non-existent bot
	_, err = st.ValidateBotToken("non_existent_bot", token)
	assert.Error(t, err)
}
