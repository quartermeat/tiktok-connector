package connector

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteAPIEventConsumesEvent(t *testing.T) {
	server := NewServer()
	event := server.hub.Publish(StreamEvent{Type: "comment", User: "viewer", Text: "!attract"})

	req := httptest.NewRequest(http.MethodDelete, "/api/events/1", nil)
	rec := httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if len(server.hub.Recent()) != 0 {
		t.Fatalf("recent len = %d, want 0", len(server.hub.Recent()))
	}
	if event.ID != 1 {
		t.Fatalf("event id = %d, want 1", event.ID)
	}
}

func TestDeleteAPIEventMissingIDReturnsNotFound(t *testing.T) {
	server := NewServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/events/404", nil)
	rec := httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
