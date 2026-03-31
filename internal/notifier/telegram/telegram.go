package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Config holds settings for the Telegram notifier.
// ChatID is the channel or chat ID (e.g. -100xxxxxxxxxx for a channel).
// The bot must be added to the channel with permission to post.
type Config struct {
	BotToken string `json:"bot_token"`
	ChatID   int64  `json:"chat_id"`
}

// Client calls the Telegram Bot API over HTTP and implements notifier.Notifier.
type Client struct {
	baseURL    string
	chatID     int64
	httpClient *http.Client
}

// New returns a Notifier that sends messages to the configured Telegram chat/channel.
func New(cfg Config) (*Client, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("telegram: bot_token is required")
	}
	baseURL := "https://api.telegram.org/bot" + cfg.BotToken
	return &Client{
		baseURL: baseURL,
		chatID:  cfg.ChatID,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}, nil
}

// sendMessageResponse is the JSON shape returned by the Telegram sendMessage API.
type sendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// Notify sends the message to the configured chat/channel.
func (c *Client) Notify(ctx context.Context, message string) error {
	reqBody := struct {
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
	}{c.chatID, message}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("telegram: encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/sendMessage", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: send: %w", err)
	}
	defer resp.Body.Close()
	var apiResp sendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("telegram: decode response: %w", err)
	}
	if !apiResp.OK {
		return fmt.Errorf("telegram: api error: %s", apiResp.Description)
	}
	return nil
}
