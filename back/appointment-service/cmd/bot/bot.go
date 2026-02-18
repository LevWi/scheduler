package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/chat/telegram"
	"scheduler/appointment-service/internal/bot/command"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"golang.org/x/text/language"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := LoadBotConfig()
	if err != nil {
		slog.Error("[LoadBotConfig]", "err", err.Error())
		log.Fatal(err)
	}

	opts := []bot.Option{bot.WithDebug()} //TODO

	b, err := bot.New(cfg.BotAPIConnection, opts...)
	if err != nil {
		slog.Error("[bot.New]", "err", err.Error())
		log.Fatal(err)
	}

	bundle := i18n.NewBundle(language.Russian)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	const langFilePrefix = "bot.dict"
	bundle.LoadMessageFile(langFilePrefix + ".ru.toml")
	bundle.LoadMessageFile(langFilePrefix + ".en.toml")

	localization := messages.NewLocalization(bundle, "ru")
	cha := &telegram.Tg{Bot: b}
	dialogStorage, err := command.NewDialogStorage(cha, localization, &cfg.SchedulerAPI)
	if err != nil {
		slog.Error("[NewDialogStorage]", "err", err.Error())
		log.Fatal(err)
	}

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData,
		"slot_id_", //TODO move it to bot package
		bot.MatchTypePrefix,
		makeOptionsCallbackHandler(dialogStorage))
	b.RegisterHandlerMatchFunc(messageMatchFunc, makeHandler(dialogStorage))
	b.Start(ctx)
}

func messageMatchFunc(update *models.Update) bool {
	return update.Message != nil
}

func makeHandler(ds *command.DialogsStorage) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			slog.Error("[bot.HandlerFunc]", "err", "message is nil")
			return
		}

		customer := command.Customer(fmt.Sprint(update.Message.From.ID))
		menu := ds.GetOrCreateMenu(customer, update.Message.Chat.ID)
		if menu == nil {
			slog.Error("[bot.HandlerFunc]", "err", "unexpected: menu == nil")
			return
		}

		chctx := &chat.ChatContext{
			Ctx:    ctx,
			ChatID: update.Message.Chat.ID,
		}
		r := &command.Request{
			ChatContext: chctx,
			//According manual it is Unix time
			Time:     time.Unix(int64(update.Message.Date), 0),
			Text:     update.Message.Text,
			Customer: customer,
			//Choices: ,
		}

		err := menu.Process(r)
		if err != nil {
			slog.Error("[Handler:menu.Process]", "err", err.Error())
		}
	}
}

func makeOptionsCallbackHandler(ds *command.DialogsStorage) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		// answering callback query first to let Telegram know that we received the callback query,
		// and we're handling it. Otherwise, Telegram might retry sending the update repetitively
		// as it thinks the callback query doesn't reach to our application. learn more by
		// reading the footnote of the https://core.telegram.org/bots/api#callbackquery type.
		//TODO show some message to user if error?
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			ShowAlert:       false,
		})

		slog.Debug("[OptionsCallbackHandler]", "choice", update.CallbackQuery.Data,
			"user", update.CallbackQuery.From.ID)

		customer := command.Customer(fmt.Sprint(update.Message.From.ID))
		dialog := ds.GetDialog(customer)
		if dialog == nil {
			slog.Error("[OptionsCallbackHandler]", "err", "dialog not found", "customer", customer)
			return
		}

		chctx := &chat.ChatContext{
			Ctx:    ctx,
			ChatID: dialog.ChatID,
		}

		var t time.Time
		if update.CallbackQuery.Message.Message != nil &&
			update.CallbackQuery.Message.Type == models.MaybeInaccessibleMessageTypeMessage {
			t = time.Unix(int64(update.CallbackQuery.Message.Message.Date), 0)
		} else {
			t = time.Now().UTC()
		}

		r := &command.Request{
			ChatContext: chctx,
			Time:        t,
			//Text:   ,
			Customer: customer,
			Choices: []command.ChoiceID{
				update.CallbackQuery.Data[len("slot_id_"):], //TODO move slot_id_ to bot package
			},
		}

		err := dialog.Menu.Process(r)
		if err != nil {
			slog.Error("[OptionsCallbackHandler:menu.Process]", "err", err.Error())
		}
	}
}
