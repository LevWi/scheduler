package storage

import (
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func newTestStorage(t *testing.T) *Storage {
	st := initDB(t)
	err := CreateTableUserBots(&st)
	assert.NoError(t, err)

	return &st
}

func TestAddAndGetBot(t *testing.T) {
	st := newTestStorage(t)

	botId := "bot_123"
	tokenHash := "hash_abc"
	businessId := "biz_1"

	// Добавляем бота
	bot, err := st.AddBot(botId, tokenHash, businessId)
	assert.NoError(t, err)

	assert.Equal(t, botId, bot.BotId)
	assert.Equal(t, tokenHash, bot.BotTokenHash)
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
	assert.Equal(t, tokenHash, got.BotTokenHash)
	assert.Equal(t, businessId, got.BusinessId)
	assert.True(t, got.IsActive)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Minute)
	assert.WithinDuration(t, time.Now(), got.LastUsedAt, time.Minute)
}

func TestEditBotStatus(t *testing.T) {
	st := newTestStorage(t)

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
	assert.Equal(t, "hash_xyz", got.BotTokenHash)
	assert.Equal(t, businessId, got.BusinessId)
	assert.False(t, got.IsActive)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Minute)
	assert.WithinDuration(t, time.Now(), got.LastUsedAt, time.Minute)
}

func TestDeleteBot(t *testing.T) {
	st := newTestStorage(t)

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
	st := newTestStorage(t)

	botId := "bot_unique"
	businessId := "biz_uniq"

	_, err := st.AddBot(botId, "hash1", businessId)
	assert.NoError(t, err)

	// Пытаемся добавить второго бота с тем же bot_id
	_, err = st.AddBot(botId, "hash2", businessId)
	assert.Error(t, err)
}
