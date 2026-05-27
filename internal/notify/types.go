package notify

import "context"

const (
	SeverityAlert    = "alert"
	SeverityRecovery = "recovery"
)

// Message is the payload to be delivered through a Notifier.
type Message struct {
	Title    string
	Body     string
	Severity string
}

// Notifier delivers a Message via a specific channel implementation.
type Notifier interface {
	Type() string
	Send(ctx context.Context, msg Message) error
}

// Meta exposes channel-type metadata for the UI.
type Meta struct {
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Schema      map[string]any `json:"schema"`
}

// Factory builds a Notifier from a decrypted payload map.
type Factory func(payload map[string]any) (Notifier, error)
