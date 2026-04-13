package command

import (
	"context"
	"strings"
	common "scheduler/appointment-service/internal"
	"scheduler/appointment-service/internal/bot"
	"scheduler/appointment-service/internal/bot/chat"
	"scheduler/appointment-service/internal/bot/i18n/messages"
	"testing"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type stubChat struct {
	showOptionsMessage string
	showOptions        []chat.ChatOption
}

func (s *stubChat) Print(_ *chat.ChatContext, _ string) error { return nil }

func (s *stubChat) ShowOptions(_ *chat.ChatContext, message string, m []chat.ChatOption) error {
	s.showOptionsMessage = message
	s.showOptions = m
	return nil
}

func (s *stubChat) ShowMenu(_ *chat.ChatContext, _ string, _ []string) error { return nil }

func (s *stubChat) HideMenu(_ *chat.ChatContext) error { return nil }

func testLocalization() *messages.Localization {
	bundle := i18n.NewBundle(language.English)
	bundle.AddMessages(language.English, messages.SelectRequestMessage, messages.DialogTimeZone)
	return messages.NewLocalization(bundle, "en")
}

func TestChatAdapter_ShowAsOptions_AppendsGMTWhenOffsetsAreSame(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}

	cha := &stubChat{}
	adapter := NewChatAdapter(cha, &bot.UserSettings{Loc: testLocalization(), TimeZone: loc})

	ops := []LabeledSlot{
		{ID: 1, Slot: common.Slot{Start: time.Date(2026, time.April, 6, 10, 0, 0, 0, time.UTC), Dur: 30 * time.Minute}},
		{ID: 2, Slot: common.Slot{Start: time.Date(2026, time.April, 13, 10, 0, 0, 0, time.UTC), Dur: 30 * time.Minute}},
		{ID: 3, Slot: common.Slot{Start: time.Date(2026, time.April, 20, 10, 0, 0, 0, time.UTC), Dur: 30 * time.Minute}},
	}

	err = adapter.ShowAsOptions(&chat.ChatContext{Ctx: context.Background(), ChatID: "c1"}, messages.SelectRequestMessage, ops)
	if err != nil {
		t.Fatalf("ShowAsOptions() error = %v", err)
	}

	if !strings.Contains(cha.showOptionsMessage, "Time zone: Europe/Berlin (GMT+02:00)") {
		t.Fatalf("expected GMT offset in message, got: %q", cha.showOptionsMessage)
	}

	if len(cha.showOptions) != 3 {
		t.Fatalf("expected 3 day options, got %d", len(cha.showOptions))
	}

	for _, option := range cha.showOptions {
		if !strings.HasPrefix(option.ID, DayMarker) {
			t.Fatalf("expected day marker option id, got: %q", option.ID)
		}
	}
}

func TestChatAdapter_ShowAsOptions_OmitsGMTWhenOffsetsDiffer(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}

	cha := &stubChat{}
	adapter := NewChatAdapter(cha, &bot.UserSettings{Loc: testLocalization(), TimeZone: loc})

	ops := []LabeledSlot{
		{ID: 1, Slot: common.Slot{Start: time.Date(2026, time.March, 23, 10, 0, 0, 0, time.UTC), Dur: 30 * time.Minute}},
		{ID: 2, Slot: common.Slot{Start: time.Date(2026, time.March, 30, 10, 0, 0, 0, time.UTC), Dur: 30 * time.Minute}},
		{ID: 3, Slot: common.Slot{Start: time.Date(2026, time.April, 6, 10, 0, 0, 0, time.UTC), Dur: 30 * time.Minute}},
	}

	err = adapter.ShowAsOptions(&chat.ChatContext{Ctx: context.Background(), ChatID: "c1"}, messages.SelectRequestMessage, ops)
	if err != nil {
		t.Fatalf("ShowAsOptions() error = %v", err)
	}

	if strings.Contains(cha.showOptionsMessage, "(GMT") {
		t.Fatalf("did not expect GMT offset in message, got: %q", cha.showOptionsMessage)
	}

	if !strings.Contains(cha.showOptionsMessage, "Time zone: Europe/Berlin") {
		t.Fatalf("expected time zone line in message, got: %q", cha.showOptionsMessage)
	}

	if len(cha.showOptions) != 3 {
		t.Fatalf("expected 3 day options, got %d", len(cha.showOptions))
	}
}
