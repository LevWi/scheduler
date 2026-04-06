package telegram

import (
	"context"
	"fmt"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot/chat"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Tg struct {
	*bot.Bot
}

//TODO add default update handler
//TODO move *bot.Bot to command.ChatContext ?

func (t *Tg) send(c context.Context, message *bot.SendMessageParams) error {
	_, err := t.Bot.SendMessage(c, message)
	if err != nil {
		method := common.CallerFuncOnly(1)
		return fmt.Errorf("telegram %s err: %w", method, err)
	}
	return nil
}

func (t *Tg) Print(c *chat.ChatContext, message string) error {
	return t.send(c.Ctx, &bot.SendMessageParams{
		ChatID: c.ChatID,
		Text:   message,
	})
}

func (t *Tg) ShowMenu(c *chat.ChatContext, message string, options []string) error {
	kb := make([]models.KeyboardButton, len(options))
	for i, o := range options {
		kb[i] = models.KeyboardButton{
			Text: o,
		}
	}

	km := &models.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard: [][]models.KeyboardButton{
			kb,
		},
	}
	return t.send(c.Ctx, &bot.SendMessageParams{
		ChatID:      c.ChatID,
		Text:        message,
		ReplyMarkup: km,
	})
}

func (t *Tg) ShowOptions(c *chat.ChatContext, message string, m []chat.ChatOption) error {
	km := &models.InlineKeyboardMarkup{
		InlineKeyboard: toInlineKeyboardButton(m),
	}
	return t.send(c.Ctx, &bot.SendMessageParams{
		ChatID:      c.ChatID,
		Text:        message,
		ReplyMarkup: km,
	})
}

func (t *Tg) HideMenu(c *chat.ChatContext) error {
	return t.send(c.Ctx, &bot.SendMessageParams{
		ChatID: c.ChatID,
		ReplyMarkup: &models.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		},
	})
}

func toInlineKeyboardButton(in []chat.ChatOption) [][]models.InlineKeyboardButton {
	columns := 1
	if len(in) > 7 {
		columns = 3
	}
	rows := len(in) / columns
	if len(in)%columns != 0 {
		rows += 1
	}
	out := make([][]models.InlineKeyboardButton, rows)
	for i, v := range in {
		row := out[i/columns]
		if row == nil {
			row = make([]models.InlineKeyboardButton, 0, columns)
		}
		out[i/columns] = append(row, models.InlineKeyboardButton{
			Text:         v.Text,
			CallbackData: fmt.Sprintf("slot_id_%v", v.ID)}, //TODO fix it
		)
	}
	return out
}
