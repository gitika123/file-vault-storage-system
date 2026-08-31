package files

import "sync"

type DownloadEvent struct {
	OwnerID       string `json:"-"`
	FileID        string `json:"fileId"`
	DownloadCount int64  `json:"downloadCount"`
}

type DownloadEventHub struct {
	mu    sync.RWMutex
	next  int
	items map[int]chan DownloadEvent
}

func NewDownloadEventHub() *DownloadEventHub {
	return &DownloadEventHub{items: make(map[int]chan DownloadEvent)}
}

func (h *DownloadEventHub) Subscribe() (<-chan DownloadEvent, func()) {
	h.mu.Lock()
	h.next++
	id := h.next
	ch := make(chan DownloadEvent, 8)
	h.items[id] = ch
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		if current, ok := h.items[id]; ok {
			delete(h.items, id)
			close(current)
		}
		h.mu.Unlock()
	}
}

func (h *DownloadEventHub) Publish(event DownloadEvent) {
	if h == nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.items {
		select {
		case ch <- event:
		default:
		}
	}
}
