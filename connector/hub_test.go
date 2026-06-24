package connector

import "testing"

func TestHubPublishAssignsIDAndStoresRecent(t *testing.T) {
	hub := NewHub()
	event := hub.Publish(StreamEvent{Type: "comment", User: "viewer", Text: "!attract"})
	if event.ID != 1 {
		t.Fatalf("id = %d, want 1", event.ID)
	}
	if event.Command != "attract" {
		t.Fatalf("command = %q, want attract", event.Command)
	}
	recent := hub.Recent()
	if len(recent) != 1 {
		t.Fatalf("recent len = %d, want 1", len(recent))
	}
}

func TestHubConsumeRemovesRecentEvent(t *testing.T) {
	hub := NewHub()
	first := hub.Publish(StreamEvent{Type: "comment", User: "viewer", Text: "!attract"})
	second := hub.Publish(StreamEvent{Type: "comment", User: "viewer", Text: "!repel"})

	if !hub.Consume(first.ID) {
		t.Fatalf("consume returned false for existing id")
	}
	if hub.Consume(first.ID) {
		t.Fatalf("consume returned true for already consumed id")
	}

	recent := hub.Recent()
	if len(recent) != 1 {
		t.Fatalf("recent len = %d, want 1", len(recent))
	}
	if recent[0].ID != second.ID {
		t.Fatalf("remaining event id = %d, want %d", recent[0].ID, second.ID)
	}
}
