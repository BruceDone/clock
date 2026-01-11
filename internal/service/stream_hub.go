package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// StreamEvent is a structured event for task execution streaming.
// It is sent to clients via SSE as JSON.
type StreamEvent struct {
	ID         int64  `json:"id"`
	TS         int64  `json:"ts"`
	Kind       string `json:"kind"` // task_start | task_end | stdout | stderr | meta
	RunID      string `json:"runId,omitempty"`
	Tid        int    `json:"tid,omitempty"`
	Cid        int    `json:"cid,omitempty"`
	TaskName   string `json:"taskName,omitempty"`
	Status     string `json:"status,omitempty"` // only for task_end
	DurationMs int64  `json:"durationMs,omitempty"`
	Msg        string `json:"msg,omitempty"`
}

// StreamHub broadcasts StreamEvents to all current subscribers.
// Each subscriber gets its own buffered channel.
//
// Backpressure policy: if a subscriber channel is full, that subscriber is disconnected.
// This protects the system from slow clients.
type StreamHub struct {
	subBuffer int

	nextID atomic.Int64

	mu   sync.RWMutex
	subs map[chan StreamEvent]struct{}

	slowDisconnects atomic.Int64
}

func NewStreamHub(subBuffer int) *StreamHub {
	if subBuffer <= 0 {
		subBuffer = 1000
	}
	return &StreamHub{
		subBuffer: subBuffer,
		subs:      make(map[chan StreamEvent]struct{}),
	}
}

func (h *StreamHub) Subscribe(ctx context.Context) <-chan StreamEvent {
	ch := make(chan StreamEvent, h.subBuffer)

	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	go func() {
		<-ctx.Done()
		h.removeSub(ch, false)
	}()

	return ch
}

func (h *StreamHub) Publish(ev StreamEvent) {
	if ev.TS == 0 {
		ev.TS = time.Now().UnixMilli()
	}
	if ev.ID == 0 {
		ev.ID = h.nextID.Add(1)
	}

	// Copy subscriber set to avoid holding lock while sending.
	h.mu.RLock()
	subs := make([]chan StreamEvent, 0, len(h.subs))
	for ch := range h.subs {
		subs = append(subs, ch)
	}
	h.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
			h.removeSub(ch, true)
		}
	}
}

func (h *StreamHub) removeSub(ch chan StreamEvent, slow bool) {
	h.mu.Lock()
	if _, ok := h.subs[ch]; !ok {
		h.mu.Unlock()
		return
	}
	delete(h.subs, ch)
	close(ch)
	h.mu.Unlock()

	if slow {
		h.slowDisconnects.Add(1)
	}
}

func (h *StreamHub) SubscriberCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs)
}

func (h *StreamHub) SlowDisconnects() int64 {
	return h.slowDisconnects.Load()
}
