package notifier

import "context"

// Notifier sends a text message to one or more notification backends
// (e.g. Telegram, Discord, WhatsApp). Implementations may be backend-specific.
type Notifier interface {
	Notify(ctx context.Context, message string) error
}
