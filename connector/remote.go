package connector

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	DefaultRemoteRelayBase = "https://ntfy.sh"
	DefaultRemoteTopic     = "quartermeat-tiktok-connector"
)

type RemoteRelayMessage struct {
	ID      string `json:"id"`
	Event   string `json:"event"`
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

func RemotePublishURL(baseURL, topic string) string {
	return remoteTopicURL(baseURL, topic)
}

func RemoteSubscribeURL(baseURL, topic string) string {
	return remoteTopicURL(baseURL, topic) + "/json"
}

func ParseRemoteRelayLine(line []byte) (StreamEvent, bool, error) {
	line = []byte(strings.TrimSpace(string(line)))
	if len(line) == 0 {
		return StreamEvent{}, false, nil
	}

	var message RemoteRelayMessage
	if err := json.Unmarshal(line, &message); err != nil {
		return StreamEvent{}, false, err
	}
	if message.Event != "message" || strings.TrimSpace(message.Message) == "" {
		return StreamEvent{}, false, nil
	}

	var event StreamEvent
	if err := json.Unmarshal([]byte(message.Message), &event); err != nil {
		return StreamEvent{}, false, err
	}
	if event.Source == "" {
		event.Source = "remote"
	}
	if event.Raw == nil {
		event.Raw = map[string]string{}
	}
	if message.ID != "" {
		event.Raw["relay_id"] = message.ID
	}
	if message.Topic != "" {
		event.Raw["relay_topic"] = message.Topic
	}
	return event, true, nil
}

func StartRemoteRelay(ctx context.Context, hub *Hub, subscribeURL string) {
	if strings.TrimSpace(subscribeURL) == "" {
		return
	}
	go func() {
		backoff := time.Second
		for {
			if err := consumeRemoteRelay(ctx, hub, subscribeURL); err != nil && ctx.Err() == nil {
				log.Printf("remote connector relay disconnected: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			select {
			case <-time.After(backoff):
				if backoff < 30*time.Second {
					backoff *= 2
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func consumeRemoteRelay(ctx context.Context, hub *Hub, subscribeURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, subscribeURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/x-ndjson, application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("relay returned %s", resp.Status)
	}
	return readRemoteRelay(resp.Body, hub)
}

func readRemoteRelay(reader io.Reader, hub *Hub) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for scanner.Scan() {
		event, ok, err := ParseRemoteRelayLine(scanner.Bytes())
		if err != nil {
			log.Printf("remote connector relay ignored message: %v", err)
			continue
		}
		if ok {
			hub.Publish(event)
		}
	}
	return scanner.Err()
}

func remoteTopicURL(baseURL, topic string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	topic = strings.Trim(strings.TrimSpace(topic), "/")
	if baseURL == "" {
		baseURL = DefaultRemoteRelayBase
	}
	if topic == "" {
		topic = DefaultRemoteTopic
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return strings.TrimRight(DefaultRemoteRelayBase, "/") + "/" + url.PathEscape(topic)
	}
	parsed.Path = path.Join(parsed.Path, topic)
	return parsed.String()
}
