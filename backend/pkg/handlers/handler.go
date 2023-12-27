package handlers

import (
	"database/sql"
	"net/http"

	"github.com/bernaotieno/social-network/backend/pkg/models"
	"github.com/bernaotieno/social-network/backend/pkg/websocket"
)

// Handler contains all the HTTP handlers for the API
type Handler struct {
	DB                   *sql.DB
	Hub                  *websocket.Hub
	UserService          *models.UserService
	SessionService       *models.SessionService
	FollowService        *models.FollowService
	PostService          *models.PostService
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
	return &Handler{
		DB:                   db,
		Hub:                  hub,
		UserService:          models.NewUserService(db),
		SessionService:       models.NewSessionService(db),
		FollowService:        models.NewFollowService(db),
		PostService:          models.NewPostService(db),
		CommentService:       models.NewCommentService(db),
		LikeService:          models.NewLikeService(db),
		GroupService:         models.NewGroupService(db),
		GroupMemberService:   models.NewGroupMemberService(db),
		GroupPostService:     models.NewGroupPostService(db),
		EventService:         models.NewEventService(db),
		EventResponseService: models.NewEventResponseService(db),
		MessageService:       models.NewMessageService(db),
		NotificationService:  models.NewNotificationService(db),
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins in development
				return true
			},
		},
	}
}

// HandleWebSocket handles WebSocket connections
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	// Get the user ID from the context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		conn.Close()
		return
	}

	// Create a new client
	client := websocket.NewClient(h.Hub, conn, userID)

	// Register the client with the hub
	registration := &websocket.Registration{
		Client: client,
		RoomID: "", // Default room
	}
	h.Hub.Register <- registration

	// Start the client's read and write pumps
	go client.WritePump()
	go client.ReadPump()
}
