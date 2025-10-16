package bots

import (
	"errors"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/dbase"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type BotsStorage struct {
	*sqlx.DB
}

type DbBot struct {
	BotId        string    `db:"bot_id"`
	BotTokenHash string    `db:"bot_token_hash"`
	BusinessId   string    `db:"business_id"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	LastUsedAt   time.Time `db:"last_used_at"`
}

func (s *BotsStorage) AddBot(botId, botToken, businessId string) (DbBot, error) {
	botTokenHash, err := bcrypt.GenerateFromPassword([]byte(botToken), bcrypt.DefaultCost)
	if err != nil {
		return DbBot{}, err
	}

	query := `INSERT INTO "user_bots" ("bot_id", "bot_token_hash", "business_id") VALUES ($1, $2, $3) RETURNING *;`
	var newBot DbBot
	err = s.Get(&newBot, query, botId, string(botTokenHash), businessId)
	return newBot, dbase.DbError(err)
}

var ErrTokenMismatch = errors.New("token mismatch")

func (s *BotsStorage) ValidateBotToken(botID common.ID, token string) (common.ID, error) {
	bot, err := s.GetBotByBotId(botID)
	if err != nil {
		return "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(bot.BotTokenHash), []byte(token))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", ErrTokenMismatch
		}
		return "", errors.Join(common.ErrInternal, err)
	}
	return bot.BusinessId, nil
}

func (s *BotsStorage) GetBotByBotId(botId string) (DbBot, error) {
	var bot DbBot
	err := s.Get(&bot, `SELECT * FROM "user_bots" WHERE "bot_id" = $1;`, botId)
	return bot, dbase.DbError(err)
}

// No errors if no bots found
func (s *BotsStorage) GetBots(businessId string) ([]DbBot, error) {
	query := `SELECT * FROM "user_bots" WHERE "business_id" = $1;`
	var bots []DbBot
	err := s.Select(&bots, query, businessId)
	return bots, dbase.DbError(err)
}

func (s *BotsStorage) EditBotStatus(businessId string, botId string, active bool) error {
	query := `UPDATE "user_bots" SET "is_active" = $1 WHERE "business_id" = $2 AND "bot_id" = $3;`
	_, err := s.Exec(query, active, businessId, botId)
	return dbase.DbError(err)
}

func (s *BotsStorage) DeleteBot(businessId string, botId string) error {
	query := `DELETE FROM "user_bots" WHERE "business_id" = $1 AND "bot_id" = $2;`
	_, err := s.Exec(query, businessId, botId)
	return dbase.DbError(err)
}
