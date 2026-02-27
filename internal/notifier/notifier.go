package notifier

type Notifier interface {
	Notify(message string) error
}

type SlackNotifier struct {
	WebhookURL string
}

func (n *SlackNotifier) Notify(message string) error {
	return nil
}
