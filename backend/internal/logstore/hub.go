package logstore

import (
	"context"
	"sync"

	"github.com/nugroho/siswa/backend/internal/model"
)

// hub is a tiny in-process fan-out used by the SSE endpoint. Slow subscribers
// are skipped rather than blocking the publisher.
type hub struct {
	mu          sync.RWMutex
	subscribers map[int]chan model.RequestLog
	nextID      int
	seen        map[string]struct{}
	seenOrder   []string
	closed      bool
}

const (
	subscriberBuffer = 64
	seenWindow       = 512
)

func newHub() *hub {
	return &hub{
		subscribers: make(map[int]chan model.RequestLog),
		seen:        make(map[string]struct{}),
	}
}

func (h *hub) subscribe(ctx context.Context) (<-chan model.RequestLog, func()) {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		ch := make(chan model.RequestLog)
		close(ch)
		return ch, func() {}
	}

	id := h.nextID
	h.nextID++
	ch := make(chan model.RequestLog, subscriberBuffer)
	h.subscribers[id] = ch
	h.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			h.mu.Lock()
			if sub, ok := h.subscribers[id]; ok {
				delete(h.subscribers, id)
				close(sub)
			}
			h.mu.Unlock()
		})
	}

	// Release the subscription when the HTTP request goes away.
	go func() {
		<-ctx.Done()
		cancel()
	}()

	return ch, cancel
}

func (h *hub) broadcast(entry model.RequestLog) {
	h.markSeen(entry.ID)
	h.publish(entry)
}

// broadcastOnce delivers an entry only if it has not been broadcast locally yet
// (used by the Redis bridge to avoid duplicating this replica's own entries).
func (h *hub) broadcastOnce(entry model.RequestLog) {
	h.mu.Lock()
	if _, dup := h.seen[entry.ID]; dup {
		h.mu.Unlock()
		return
	}
	h.rememberLocked(entry.ID)
	h.mu.Unlock()

	h.publish(entry)
}

func (h *hub) publish(entry model.RequestLog) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, sub := range h.subscribers {
		select {
		case sub <- entry:
		default: // subscriber is too slow, skip this entry for it
		}
	}
}

func (h *hub) markSeen(id string) {
	h.mu.Lock()
	h.rememberLocked(id)
	h.mu.Unlock()
}

func (h *hub) rememberLocked(id string) {
	if id == "" {
		return
	}
	if _, ok := h.seen[id]; ok {
		return
	}
	h.seen[id] = struct{}{}
	h.seenOrder = append(h.seenOrder, id)
	if len(h.seenOrder) > seenWindow {
		oldest := h.seenOrder[0]
		h.seenOrder = h.seenOrder[1:]
		delete(h.seen, oldest)
	}
}

func (h *hub) close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	for id, sub := range h.subscribers {
		delete(h.subscribers, id)
		close(sub)
	}
}
