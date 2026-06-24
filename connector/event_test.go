package connector

import "testing"

func TestParseCommand(t *testing.T) {
	command, args := ParseCommand("!repel left now")
	if command != "repel" {
		t.Fatalf("command = %q, want repel", command)
	}
	if len(args) != 2 || args[0] != "left" || args[1] != "now" {
		t.Fatalf("args = %#v", args)
	}
}

func TestNormalizeGiftCommand(t *testing.T) {
	event := NormalizeEvent(StreamEvent{Type: "gift", Gift: "Rose Power"})
	if event.Command != "gift:rose-power" {
		t.Fatalf("gift command = %q", event.Command)
	}
}
