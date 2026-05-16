package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{rooms: map[string]map[*websocket.Conn]struct{}{}}
}

func (h *Hub) Join(chatID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room, ok := h.rooms[chatID]
	if !ok {
		room = map[*websocket.Conn]struct{}{}
		h.rooms[chatID] = room
	}
	room[conn] = struct{}{}
}

func (h *Hub) Leave(chatID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, ok := h.rooms[chatID]; ok {
		delete(room, conn)
		if len(room) == 0 {
			delete(h.rooms, chatID)
		}
	}
}

func (h *Hub) Broadcast(chatID string, payload interface{}) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, 4)
	if room, ok := h.rooms[chatID]; ok {
		for c := range room {
			conns = append(conns, c)
		}
	}
	h.mu.RUnlock()
	for _, c := range conns {
		_ = c.WriteJSON(payload)
	}
}
