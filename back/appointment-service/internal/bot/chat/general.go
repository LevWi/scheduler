package chat

import (
	"context"
)

type ChatID any

type ChatOption struct {
	ID       string
	Text     string
	TextLong string
}

type ChatContext struct {
	Ctx    context.Context
	ChatID ChatID
}

type Chat interface {
	Print(c *ChatContext, message string) error
	ShowOptions(c *ChatContext, message string, m []ChatOption) error
	ShowMenu(c *ChatContext, message string, options []string) error
	HideMenu(c *ChatContext) error
}
