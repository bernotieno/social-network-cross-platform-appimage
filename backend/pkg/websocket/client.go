package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 10000
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	RoomID  string      `json:"roomId,omitempty"`
	Content interface{} `json:"content"`
	SenderID string     `json:"senderId"`
}

// Client represents a connected WebSocket client
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   string
	Rooms    map[string]bool
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userID,
		Rooms:    make(map[string]bool),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Parse the message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error parsing message: %v", err)
			continue
		}

		// Set the sender ID
		msg.SenderID = c.UserID

		// Handle different message types
		switch msg.Type {
		case "join_room":
			// Join a chat room
			if roomID, ok := msg.Content.(string); ok {
				c.JoinRoom(roomID)
			}
		case "leave_room":
			// Leave a chat room
			if roomID, ok := msg.Content.(string); ok {
				c.LeaveRoom(roomID)
			}
		case "chat_message":
			// Broadcast the message to the room
			if msg.RoomID != "" {
				// Serialize the message
				data, err := json.Marshal(msg)
				if err != nil {
					log.Printf("error marshaling message: %v", err)
					continue
				}

				// Broadcast to the room
				c.Hub.Broadcast <- &Broadcast{
					RoomID:  msg.RoomID,
					Message: data,
					Sender:  c,
				}
			}
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// JoinRoom adds the client to a room
func (c *Client) JoinRoom(roomID string) {
	c.Hub.Register <- &Registration{
		Client: c,
		RoomID: roomID,
	}
	c.Rooms[roomID] = true
}

// LeaveRoom removes the client from a room
func (c *Client) LeaveRoom(roomID string) {
	c.Hub.Unregister <- c
	delete(c.Rooms, roomID)
}
