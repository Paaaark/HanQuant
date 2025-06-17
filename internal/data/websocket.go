package data

import (
	"encoding/json"
	"sync"

	"golang.org/x/net/websocket"
)

type WSMessage struct {
	Type    string      `json:"type"`
	Tickers []string    `json:"tickers,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type WSClient struct {
	Conn    *websocket.Conn
	Tickers map[string]bool
	Send    chan []byte
	Hub     *Hub
	Mu      sync.Mutex
}

type Hub struct {
	Clients    map[*WSClient]bool
	Register   chan *WSClient
	Unregister chan *WSClient
	Broadcast  chan *broadcastMsg
	Mu         sync.Mutex
}

type broadcastMsg struct {
	Tickers []string
	Data    []StockSnapshot
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*WSClient]bool),
		Register:   make(chan *WSClient),
		Unregister: make(chan *WSClient),
		Broadcast:  make(chan *broadcastMsg),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			h.Clients[client] = true
			h.Mu.Unlock()
		case client := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.Mu.Unlock()
		case msg := <-h.Broadcast:
			h.Mu.Lock()
			for client := range h.Clients {
				client.Mu.Lock()
				var interested []StockSnapshot
				for _, snap := range msg.Data {
					if client.Tickers[snap.Code] {
						interested = append(interested, snap)
					}
				}
				if len(interested) > 0 {
					resp, _ := json.Marshal(WSMessage{
						Type: "snapshot",
						Data: interested,
					})
					select {
					case client.Send <- resp:
					default:
						close(client.Send)
						delete(h.Clients, client)
					}
				}
				client.Mu.Unlock()
			}
			h.Mu.Unlock()
		}
	}
} 