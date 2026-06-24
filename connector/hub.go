package connector

import (
	"sync"
	"time"
)

const replayLimit = 100

type Hub struct {
	mu      sync.Mutex
	nextID  int64
	recent  []StreamEvent
	clients map[chan StreamEvent]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: map[chan StreamEvent]struct{}{}}
}

func (h *Hub) Publish(event StreamEvent) StreamEvent {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.nextID++
	event = NormalizeEvent(event)
	event.ID = h.nextID
	event.ReceivedAt = time.Now().UTC()
	h.recent = append(h.recent, event)
	if len(h.recent) > replayLimit {
		h.recent = h.recent[len(h.recent)-replayLimit:]
	}
	for ch := range h.clients {
		select {
		case ch <- event:
		default:
		}
	}
	return event
}

func (h *Hub) Subscribe() (<-chan StreamEvent, func()) {
	ch := make(chan StreamEvent, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	cancel := func() {
		h.mu.Lock()
		if _, ok := h.clients[ch]; ok {
			delete(h.clients, ch)
			close(ch)
		}
		h.mu.Unlock()
	}
	return ch, cancel
}

func (h *Hub) Recent() []StreamEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]StreamEvent, len(h.recent))
	copy(out, h.recent)
	return out
}

func (h *Hub) Consume(id int64) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, event := range h.recent {
		if event.ID == id {
			h.recent = append(h.recent[:i], h.recent[i+1:]...)
			return true
		}
	}
	return false
}
