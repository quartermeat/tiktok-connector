package connector

import "testing"

func TestRemoteRelayURLs(t *testing.T) {
	if got := RemotePublishURL("https://ntfy.sh", "quartermeat-test"); got != "https://ntfy.sh/quartermeat-test" {
		t.Fatalf("publish url = %q", got)
	}
	if got := RemoteSubscribeURL("https://ntfy.sh/", "/quartermeat-test/"); got != "https://ntfy.sh/quartermeat-test/json" {
		t.Fatalf("subscribe url = %q", got)
	}
}

func TestParseRemoteRelayLine(t *testing.T) {
	line := []byte(`{"id":"abc","event":"message","topic":"quartermeat-test","message":"{\"source\":\"remote\",\"type\":\"comment\",\"user\":\"viewer\",\"text\":\"!burst\",\"value\":2}"}`)

	event, ok, err := ParseRemoteRelayLine(line)
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	if !ok {
		t.Fatalf("parse returned ok=false")
	}
	if event.User != "viewer" || event.Text != "!burst" || event.Value != 2 {
		t.Fatalf("event = %#v", event)
	}
	if event.Raw["relay_id"] != "abc" {
		t.Fatalf("relay id = %q, want abc", event.Raw["relay_id"])
	}
}

func TestParseRemoteRelayLineIgnoresOpenEvent(t *testing.T) {
	_, ok, err := ParseRemoteRelayLine([]byte(`{"event":"open","topic":"quartermeat-test"}`))
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	if ok {
		t.Fatalf("open event should be ignored")
	}
}
