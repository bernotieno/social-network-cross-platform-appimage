package websocket

import (
	"encoding/json"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients by room
	Rooms map[string]map[*Client]bool

	// Online users by user ID
	OnlineUsers map[string]*Client

	// Inbound messages to broadcast
	Broadcast chan *Broadcast

	// Register requests from clients
	Register chan *Registration

	// Unregister requests from clients
	Unregister chan *Client
}

// Broadcast represents a message to be broadcast to a room
type Broadcast struct {
	RoomID  string
	Message []byte
	Sender  *Client
}

// Registration represents a client registration to a room
type Registration struct {
	Client *Client
	RoomID string
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		Rooms:       make(map[string]map[*Client]bool),
		OnlineUsers: make(map[string]*Client),
		Broadcast:   make(chan *Broadcast),
		Register:    make(chan *Registration),
		Unregister:  make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case registration := <-h.Register:
			// Create the room if it doesn't exist
			if _, ok := h.Rooms[registration.RoomID]; !ok {
				h.Rooms[registration.RoomID] = make(map[*Client]bool)
			}
			// Add the client to the room
			h.Rooms[registration.RoomID][registration.Client] = true

			// Track user as online
			h.OnlineUsers[registration.Client.UserID] = registration.Client
			h.broadcastUserPresence(registration.Client.UserID, "online")

		case client := <-h.Unregister:
			// Remove user from online users
			if _, ok := h.OnlineUsers[client.UserID]; ok {
				delete(h.OnlineUsers, client.UserID)
				h.broadcastUserPresence(client.UserID, "offline")
			}

			// Remove the client from all rooms
			for roomID, room := range h.Rooms {
				if _, ok := room[client]; ok {
					delete(room, client)
					// Delete the room if it's empty
					if len(room) == 0 {
						delete(h.Rooms, roomID)
					}
				}
			}
			// Close the client's send channel safely
			select {
			case <-client.Send:
				// Channel is already closed
			default:
				close(client.Send)
			}

		case broadcast := <-h.Broadcast:
			// Get the room
			room, ok := h.Rooms[broadcast.RoomID]
			if !ok {
				continue
			}

			// Broadcast the message to all clients in the room
			for client := range room {
				// Don't send the message back to the sender (unless sender is nil, meaning it's from HTTP API)
				if broadcast.Sender != nil && client == broadcast.Sender {
					continue
				}

				select {
				case client.Send <- broadcast.Message:
				default:
					// Client's send buffer is full, remove them
					select {
					case <-client.Send:
						// Channel is already closed
					default:
						close(client.Send)
					}
					delete(room, client)
					// Delete the room if it's empty
					if len(room) == 0 {
						delete(h.Rooms, broadcast.RoomID)
					}
				}
			}
		}
	}
}

// broadcastUserPresence broadcasts user online/offline status to all connected clients
func (h *Hub) broadcastUserPresence(userID, status string) {
	presenceData := map[string]interface{}{
		"userId": userID,
		"status": status,
	}

	message := map[string]interface{}{
		"type":    "user_presence",
		"payload": presenceData,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling user presence message: %v", err)
		return
	}

	// Broadcast to all users in the default room (all connected users)
	if room, ok := h.Rooms[""]; ok {
		for client := range room {
			select {
			case client.Send <- data:
			default:
				// Client's send buffer is full, remove them
				select {
				case <-client.Send:
					// Channel is already closed
				default:
					close(client.Send)
				}
				delete(room, client)
			}
		}
	}
}

// broadcastTypingStatus broadcasts typing status to users in a specific room
func (h *Hub) broadcastTypingStatus(roomID, userID string, isTyping bool) {
	typingData := map[string]interface{}{
		"roomId":   roomID,
		"userId":   userID,
		"isTyping": isTyping,
	}

	message := map[string]interface{}{
		"type":    "typing_status",
		"payload": typingData,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling typing status message: %v", err)
		return
	}

	// Broadcast to users in the specific room
	if room, ok := h.Rooms[roomID]; ok {
		for client := range room {
			// Don't send typing status back to the user who is typing
			if client.UserID != userID {
				select {
				case client.Send <- data:
				default:
					// Client's send buffer is full, remove them
					select {
					case <-client.Send:
						// Channel is already closed
					default:
						close(client.Send)
					}
					delete(room, client)
				}
			}
		}
	}
}

// BroadcastSessionInvalidation notifies a user that their session has been invalidated
func (h *Hub) BroadcastSessionInvalidation(userID string) {
	message := map[string]interface{}{
		"type": "session_invalidated",
		"payload": map[string]interface{}{
			"message": "Your session has been invalidated due to a new login from another device",
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling session invalidation message: %v", err)
		return
	}

	// Send to the specific user if they're online
	if client, ok := h.OnlineUsers[userID]; ok {
		select {
		case client.Send <- data:
			log.Printf("Session invalidation message sent to user %s", userID)
		default:
			log.Printf("Failed to send session invalidation message to user %s - channel full", userID)
		}
	} else {
		log.Printf("User %s not online, cannot send session invalidation message", userID)
	}
}
