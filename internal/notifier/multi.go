package notifier

import "context"

// Multi sends each notification to all wrapped Notifiers.
// Notify returns the first error from any backend and stops calling the rest.
type Multi struct {
	Notifiers []Notifier
}

// NewMulti returns a Notifier that fans out to the given backends.
func NewMulti(n ...Notifier) *Multi {
	return &Multi{Notifiers: n}
}

// Notify calls Notify on each backend in order; returns the first error, if any.
func (m *Multi) Notify(ctx context.Context, message string) error {
	for _, n := range m.Notifiers {
		if err := n.Notify(ctx, message); err != nil {
			return err
		}
	}
	return nil
}
