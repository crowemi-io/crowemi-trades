package telegram

import (
	"context"
	"fmt"

	"github.com/crowemi-io/crowemi-trades/internal/notifier"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Config holds settings for the Telegram notifier.
// ChatID is the channel or chat ID (e.g. -100xxxxxxxxxx for a channel).
// The bot must be added to the channel with permission to post.
type Config struct {
	BotToken string `json:"bot_token"`
	ChatID   int64  `json:"chat_id"`
}

// client wraps the Telegram Bot API and implements notifier.Notifier.
type client struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

// New returns a Notifier that sends messages to the configured Telegram chat/channel.
// The bot is created and the token is validated at construction time.
func New(cfg Config) (notifier.Notifier, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("telegram: bot_token is required")
	}
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("telegram: create bot: %w", err)
	}
	return &client{bot: bot, chatID: cfg.ChatID}, nil
}

// Notify sends the message to the configured chat/channel.
func (c *client) Notify(ctx context.Context, message string) error {
	msg := tgbotapi.NewMessage(c.chatID, message)
	_, err := c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("telegram: send: %w", err)
	}
	return nil
}
