package connector

import (
	"strings"
	"time"
)

type StreamEvent struct {
	ID         int64             `json:"id"`
	Source     string            `json:"source"`
	Type       string            `json:"type"`
	User       string            `json:"user"`
	Text       string            `json:"text,omitempty"`
	Gift       string            `json:"gift,omitempty"`
	Value      int               `json:"value,omitempty"`
	Command    string            `json:"command,omitempty"`
	Args       []string          `json:"args,omitempty"`
	Raw        map[string]string `json:"raw,omitempty"`
	ReceivedAt time.Time         `json:"receivedAt"`
}

func NormalizeEvent(event StreamEvent) StreamEvent {
	if event.Source == "" {
		event.Source = "manual"
	}
	if event.Type == "" {
		event.Type = "comment"
	}
	event.Type = strings.ToLower(strings.TrimSpace(event.Type))
	event.User = strings.TrimSpace(event.User)
	event.Text = strings.TrimSpace(event.Text)
	event.Gift = strings.TrimSpace(event.Gift)
	if event.Value == 0 {
		event.Value = 1
	}
	if event.Type == "comment" && event.Command == "" {
		event.Command, event.Args = ParseCommand(event.Text)
	}
	if event.Type == "gift" && event.Command == "" && event.Gift != "" {
		event.Command = "gift:" + strings.ToLower(strings.ReplaceAll(event.Gift, " ", "-"))
	}
	return event
}

func ParseCommand(text string) (string, []string) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "!") {
		return "", nil
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "", nil
	}
	command := strings.TrimPrefix(strings.ToLower(fields[0]), "!")
	if command == "" {
		return "", nil
	}
	return command, fields[1:]
}
