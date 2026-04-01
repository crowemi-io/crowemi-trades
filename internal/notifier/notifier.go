package notifier

import (
	"context"
	"fmt"

	"github.com/crowemi-io/crowemi-trades/internal/notifier/telegram"
)

// Notifier is a Telegram-only notification sender.
type Notifier struct {
	tn *telegram.Client
}

// Config holds settings for the Telegram notifier.
type Config struct {
	BotToken string `json:"bot_token"`
	ChatID   int64  `json:"chat_id"`
}

// New returns a Notifier that sends messages to Telegram.
func New(cfg Config) (*Notifier, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("telegram: bot_token is required")
	}
	if cfg.ChatID == 0 {
		return nil, fmt.Errorf("telegram: chat_id is required")
	}

	tn, err := telegram.New(telegram.Config{
		BotToken: cfg.BotToken,
		ChatID:   cfg.ChatID,
	})
	if err != nil {
		return nil, fmt.Errorf("telegram: %w", err)
	}

	return &Notifier{tn: tn}, nil
}

// Notify sends a message to the configured Telegram chat/channel.
func (n *Notifier) Notify(ctx context.Context, message string) error {
	return n.tn.Notify(ctx, message)
}
