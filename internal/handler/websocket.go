package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Paaaark/hanquant/internal/data"
	"github.com/Paaaark/hanquant/internal/service"
	"golang.org/x/net/websocket"
)

type WebSocketHandler struct {
	svc *service.WebSocketService
}

func NewWebSocketHandler(svc *service.WebSocketService) *WebSocketHandler {
	return &WebSocketHandler{svc: svc}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws := websocket.Server{Handler: func(conn *websocket.Conn) {
		client := &data.WSClient{
			Conn:    conn,
			Tickers: make(map[string]bool),
			Send:    make(chan []byte, 256),
			Hub:     h.svc.Hub,
		}
		h.svc.Hub.Register <- client
		fmt.Println("New WebSocket connection establihsed")
		defer func() {
			h.svc.Hub.Unregister <- client
			conn.Close()
		}()

		go h.writePump(client)
		h.readPump(client)
	}}
	ws.ServeHTTP(w, r)
}

func (h *WebSocketHandler) readPump(client *data.WSClient) {
	for {
		var msg []byte
		err := websocket.Message.Receive(client.Conn, &msg)
		if err != nil {
			log.Println("read error:", err)
			break
		}
		h.svc.HandleMessage(client, msg)
	}
}

func (h *WebSocketHandler) writePump(client *data.WSClient) {
	for msg := range client.Send {
		err := websocket.Message.Send(client.Conn, msg)
		if err != nil {
			log.Println("write error:", err)
			break
		}
	}
} 