package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/bernaotieno/social-network/backend/pkg/middleware"
	"github.com/bernaotieno/social-network/backend/pkg/models"
	"github.com/bernaotieno/social-network/backend/pkg/websocket"
	"github.com/gorilla/mux"
)

// MessageServiceAdapter adapts the models.MessageService to the websocket.MessageService interface
type MessageServiceAdapter struct {
	service *models.MessageService
}

// Create adapts the Create method to match the websocket interface
func (a *MessageServiceAdapter) Create(dbMessage *websocket.DBMessage) error {
	message := &models.Message{
		SenderID:   dbMessage.SenderID,
		ReceiverID: dbMessage.ReceiverID,
		Content:    dbMessage.Content,
	}
	return a.service.Create(message)
}

// Handler contains all the HTTP handlers for the API
type Handler struct {
	DB                   *sql.DB
	Hub                  *websocket.Hub
	UserService          *models.UserService
	SessionService       *models.SessionService
	FollowService        *models.FollowService
	PostService          *models.PostService
	PostViewerService    *models.PostViewerService
	CommentService       *models.CommentService
	LikeService          *models.LikeService
	GroupService         *models.GroupService
	GroupMemberService   *models.GroupMemberService
	GroupPostService     *models.GroupPostService
	EventService         *models.EventService
	EventResponseService *models.EventResponseService
	MessageService       *models.MessageService
	NotificationService  *models.NotificationService
	Upgrader             websocket.Upgrader
}

// NewHandler creates a new Handler
func NewHandler(db *sql.DB, hub *websocket.Hub) *Handler {
	handler := &Handler{
		DB:                   db,
		Hub:                  hub,
		UserService:          models.NewUserService(db),
		SessionService:       models.NewSessionService(db),
		FollowService:        models.NewFollowService(db),
		PostService:          models.NewPostService(db),
		PostViewerService:    models.NewPostViewerService(db),
		CommentService:       models.NewCommentService(db),
		LikeService:          models.NewLikeService(db),
		GroupService:         models.NewGroupService(db),
		GroupMemberService:   models.NewGroupMemberService(db),
		GroupPostService:     models.NewGroupPostService(db),
		EventService:         models.NewEventService(db),
		EventResponseService: models.NewEventResponseService(db),
		MessageService:       models.NewMessageService(db),
		NotificationService:  models.NewNotificationServiceWithHub(db, hub),
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins in development
				return true
			},
		},
	}

	// Set up notification broadcasting
	handler.setupNotificationBroadcasting()

	return handler
}

// setupNotificationBroadcasting sets up notification broadcasting functionality
func (h *Handler) setupNotificationBroadcasting() {
	// Override the notification service's broadcast method to use our hub
	if h.NotificationService != nil && h.Hub != nil {
		// We'll create a custom broadcast function
		h.NotificationService.SetBroadcastFunction(h.broadcastNotificationToUsers)
	}
}

// broadcastNotificationToUsers broadcasts a notification to the target user
func (h *Handler) broadcastNotificationToUsers(notification interface{}) {
	// Type assert to get the notification
	notif, ok := notification.(*models.Notification)
	if !ok {
		log.Printf("Invalid notification type for broadcasting")
		return
	}

	// Create broadcast message
	message := map[string]interface{}{
		"type":    "notification",
		"payload": notif,
	}

	// Serialize the message
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling notification for broadcast: %v", err)
		return
	}

	// Send notification only to the specific user who should receive it
	// Check if the target user is online
	if targetClient, exists := h.Hub.OnlineUsers[notif.UserID]; exists {
		select {
		case targetClient.Send <- data:
			log.Printf("Sent notification to user %s: %s", notif.UserID, notif.Content)
		default:
			// Client's send buffer is full, log but don't fail
			log.Printf("Failed to send notification to user %s: send buffer full", notif.UserID)
		}
	} else {
		log.Printf("User %s is not online, notification will be delivered when they connect", notif.UserID)
	}
}

// HandleWebSocket handles WebSocket connections
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleWebSocket: Starting WebSocket upgrade for user")

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	log.Printf("HandleWebSocket: WebSocket upgrade successful")

	// Get the user ID from the context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		log.Printf("HandleWebSocket: User ID not found in context: %v", err)
		conn.Close()
		return
	}
	log.Printf("HandleWebSocket: User ID found: %s", userID)

	// Create a new client with message service adapter
	messageAdapter := &MessageServiceAdapter{service: h.MessageService}
	client := websocket.NewClient(h.Hub, conn, userID, messageAdapter)
	log.Printf("HandleWebSocket: Created WebSocket client for user %s", userID)

	// Register the client with the hub
	registration := &websocket.Registration{
		Client: client,
		RoomID: "", // Default room
	}
	h.Hub.Register <- registration
	log.Printf("HandleWebSocket: Registered client with hub")

	// Start the client's read and write pumps
	go client.WritePump()
	go client.ReadPump()
	log.Printf("HandleWebSocket: Started read and write pumps for user %s", userID)
}

// SendMessage handles sending a message via HTTP API
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context using the middleware helper
	userID, err := middleware.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		ReceiverID string `json:"receiverId"`
		GroupID    string `json:"groupId"`
		Content    string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Validate that either receiverId or groupId is set, but not both
	if (req.ReceiverID == "" && req.GroupID == "") || (req.ReceiverID != "" && req.GroupID != "") {
		http.Error(w, "Either receiverId or groupId must be set, but not both", http.StatusBadRequest)
		return
	}

	var roomID string
	var message *models.Message

	if req.ReceiverID != "" {
		// Private message - validate follow relationship or public profile
		if err := h.validatePrivateMessagePermission(userID, req.ReceiverID); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		message = &models.Message{
			SenderID:   userID,
			ReceiverID: req.ReceiverID,
			Content:    req.Content,
		}
		roomID = generateRoomID(userID, req.ReceiverID)
	} else {
		// Group message - validate group membership
		isMember, err := h.GroupMemberService.IsGroupMember(req.GroupID, userID)
		if err != nil {
			http.Error(w, "Failed to check group membership", http.StatusInternalServerError)
			return
		}
		if !isMember {
			http.Error(w, "You must be a member of this group to send messages", http.StatusForbidden)
			return
		}

		message = &models.Message{
			SenderID: userID,
			GroupID:  req.GroupID,
			Content:  req.Content,
		}
		roomID = "group-" + req.GroupID
	}

	// Save to database
	if err := h.MessageService.Create(message); err != nil {
		log.Printf("Error creating message: %v", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// Also broadcast the message via WebSocket for real-time delivery
	responseMsg := map[string]interface{}{
		"roomId": roomID,
		"message": map[string]interface{}{
			"content":   req.Content,
			"sender":    userID,
			"timestamp": message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	// Serialize the response message
	data, err := json.Marshal(map[string]interface{}{
		"type":    "new_message",
		"payload": responseMsg,
	})
	if err != nil {
		log.Printf("Error marshaling WebSocket message: %v", err)
		// Continue even if WebSocket broadcast fails
	} else {
		// Broadcast to the room via WebSocket
		h.Hub.Broadcast <- &websocket.Broadcast{
			RoomID:  roomID,
			Message: data,
			Sender:  nil, // No specific sender client since this is from HTTP API
		}
		log.Printf("Message broadcasted via WebSocket to room %s", roomID)
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Message sent successfully",
	})
}

// GetMessages handles getting messages between two users
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context using the middleware helper
	userID, err := middleware.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get other user ID from URL
	vars := mux.Vars(r)
	otherUserID := vars["userId"]
	if otherUserID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get messages between the two users (limit to 50 messages, no offset)
	messages, err := h.MessageService.GetPrivateMessages(userID, otherUserID, 50, 0)
	if err != nil {
		log.Printf("Error getting messages: %v", err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	// Return messages
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"messages": messages,
	})
}

// GetOnlineUsers handles getting currently online users
func (h *Handler) GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context using the middleware helper
	userID, err := middleware.GetUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get online users from the WebSocket hub
	onlineUserIDs := h.getOnlineUserIDs()

	// Get user details for online users
	var onlineUsers []*models.User
	for _, id := range onlineUserIDs {
		if id != userID { // Don't include the current user
			user, err := h.UserService.GetByID(id)
			if err != nil {
				log.Printf("Error getting user %s: %v", id, err)
				continue
			}
			// Clear password for security
			user.Password = ""
			onlineUsers = append(onlineUsers, user)
		}
	}

	// Return online users
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"onlineUsers": onlineUsers,
	})
}

// getOnlineUserIDs returns a list of currently online user IDs from the WebSocket hub
func (h *Handler) getOnlineUserIDs() []string {
	if h.Hub == nil {
		return []string{}
	}

	var userIDs []string
	for userID := range h.Hub.OnlineUsers {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// validatePrivateMessagePermission checks if a user can send a private message to another user
func (h *Handler) validatePrivateMessagePermission(senderID, receiverID string) error {
	// Get receiver's profile
	receiver, err := h.UserService.GetByID(receiverID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// If receiver has a public profile, allow messaging
	if !receiver.IsPrivate {
		return nil
	}

	// For private profiles, check if there's a follow relationship
	// Either sender follows receiver OR receiver follows sender
	senderFollowsReceiver, err := h.FollowService.IsFollowing(senderID, receiverID)
	if err != nil {
		return fmt.Errorf("failed to check follow relationship")
	}

	receiverFollowsSender, err := h.FollowService.IsFollowing(receiverID, senderID)
	if err != nil {
		return fmt.Errorf("failed to check follow relationship")
	}

	if !senderFollowsReceiver && !receiverFollowsSender {
		return fmt.Errorf("you can only message users you follow or who follow you")
	}

	return nil
}

// generateRoomID creates a consistent room ID for two users
func generateRoomID(userID1, userID2 string) string {
	users := []string{userID1, userID2}
	sort.Strings(users)
	return strings.Join(users, "-")
}
