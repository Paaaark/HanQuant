package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Paaaark/hanquant/internal/data"
)

type WebSocketService struct {
	kisClient *data.KISClient
	Hub       *data.Hub
}

func NewWebSocketService(kisClient *data.KISClient) *WebSocketService {
	return &WebSocketService{
		kisClient: kisClient,
		Hub:       data.NewHub(),
	}
}

func (s *WebSocketService) Start() {
	go s.Hub.Run()
	go s.periodicUpdates()
}

func (s *WebSocketService) HandleMessage(client *data.WSClient, msg []byte) {
	var wsMsg data.WSMessage
	if err := json.Unmarshal(msg, &wsMsg); err != nil {
		s.sendError(client, "invalid message format")
		return
	}
	switch wsMsg.Type {
	case "subscribe":
		client.Mu.Lock()
		for _, t := range wsMsg.Tickers {
			if len(client.Tickers) < 30 {
				client.Tickers[t] = true
			}
		}
		client.Mu.Unlock()
		// Send immediate snapshot
		s.sendSnapshot(client)
	case "unsubscribe":
		client.Mu.Lock()
		for _, t := range wsMsg.Tickers {
			delete(client.Tickers, t)
		}
		client.Mu.Unlock()
		// Send immediate snapshot
		s.sendSnapshot(client)
	default:
		s.sendError(client, "unknown message type")
	}
}

func (s *WebSocketService) sendError(client *data.WSClient, errMsg string) {
	resp, _ := json.Marshal(data.WSMessage{
		Type:  "error",
		Error: errMsg,
	})
	client.Send <- resp
}

func (s *WebSocketService) sendSnapshot(client *data.WSClient) {
	client.Mu.Lock()
	tickers := make([]string, 0, len(client.Tickers))
	for t := range client.Tickers {
		tickers = append(tickers, t)
	}
	client.Mu.Unlock()
	if len(tickers) == 0 {
		return
	}
	fmt.Println("Tickers received: ", tickers)
	snaps, err := s.kisClient.GetMultipleStockSnapshot(tickers)
	if err != nil {
		s.sendError(client, err.Error())
		return
	}

	if len(snaps) >= 1 {
		fmt.Println("Sent Message: ", snaps[0].Name, snaps[0].Price)
	}

	var buf bytes.Buffer
	buf.WriteString(`{"type":"snapshot","data":`)
	buf.Write(snaps.EncodeJSON())
	buf.WriteByte('}')
	client.Send <- buf.Bytes()
}

func (s *WebSocketService) periodicUpdates() {
	for {
		if isMarketOpen() {
			s.broadcastAll()
			time.Sleep(time.Second)
		} else {
			time.Sleep(2 * time.Second)
		}
	}
}

func (s *WebSocketService) broadcastAll() {
	s.Hub.Mu.Lock()
	clients := make([]*data.WSClient, 0, len(s.Hub.Clients))
	for c := range s.Hub.Clients {
		clients = append(clients, c)
	}
	s.Hub.Mu.Unlock()
	for _, client := range clients {
		client.Mu.Lock()
		tickers := make([]string, 0, len(client.Tickers))
		for t := range client.Tickers {
			tickers = append(tickers, t)
		}
		client.Mu.Unlock()
		if len(tickers) == 0 {
			continue
		}
		snaps, err := s.kisClient.GetMultipleStockSnapshot(tickers)
		if err != nil {
			s.sendError(client, err.Error())
			continue
		}

		if len(snaps) >= 1 {
			fmt.Println("Sent Message: ", snaps[0].Name, snaps[0].Price)
		}

		var buf bytes.Buffer
		buf.WriteString(`{"type":"snapshot","data":`)
		buf.Write(snaps.EncodeJSON())
		buf.WriteByte('}')
		client.Send <- buf.Bytes()
	}
}

// Placeholder for market open logic (Korea: 09:00-15:30 KST, Mon-Fri)
func isMarketOpen() bool {
	now := time.Now().In(time.FixedZone("KST", 9*60*60))
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	open := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	close := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, now.Location())
	return now.After(open) && now.Before(close)
} 