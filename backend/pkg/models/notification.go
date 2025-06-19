package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeFollowRequest     NotificationType = "follow_request"
	NotificationTypeFollowAccepted    NotificationType = "follow_accepted"
	NotificationTypeNewFollower       NotificationType = "new_follower"
	NotificationTypePostLike          NotificationType = "post_like"
	NotificationTypePostComment       NotificationType = "post_comment"
	NotificationTypeGroupInvite       NotificationType = "group_invite"
	NotificationTypeGroupJoinRequest  NotificationType = "group_join_request"
	NotificationTypeGroupJoinApproved NotificationType = "group_join_approved"
	NotificationTypeGroupJoinRejected NotificationType = "group_join_rejected"
	NotificationTypeEventInvite       NotificationType = "event_invite"
	NotificationTypeGroupEventCreated NotificationType = "group_event_created"
)

// Notification represents a notification
type Notification struct {
	ID        string           `json:"id"`
	UserID    string           `json:"userId"`
	SenderID  string           `json:"senderId"`
	Type      NotificationType `json:"type"`
	Content   string           `json:"content"`
	Data      string           `json:"data,omitempty"`
	ReadAt    *time.Time       `json:"readAt,omitempty"`
	CreatedAt time.Time        `json:"createdAt"`
	// Additional fields for API responses
	Sender *User `json:"sender,omitempty"`
}

// NotificationService handles notification-related operations
type NotificationService struct {
	DB            *sql.DB
	Hub           interface{}       // WebSocket hub for broadcasting notifications
	BroadcastFunc func(interface{}) // Custom broadcast function
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(db *sql.DB) *NotificationService {
	return &NotificationService{DB: db, Hub: nil}
}

// NewNotificationServiceWithHub creates a new NotificationService with WebSocket hub
func NewNotificationServiceWithHub(db *sql.DB, hub interface{}) *NotificationService {
	return &NotificationService{DB: db, Hub: hub}
}

// SetBroadcastFunction sets a custom broadcast function
func (s *NotificationService) SetBroadcastFunction(fn func(interface{})) {
	s.BroadcastFunc = fn
}

// Create creates a new notification
func (s *NotificationService) Create(notification *Notification) error {
	notification.ID = uuid.New().String()
	notification.CreatedAt = time.Now()

	_, err := s.DB.Exec(`
		INSERT INTO notifications (id, user_id, sender_id, type, content, data, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, notification.ID, notification.UserID, notification.SenderID, notification.Type, notification.Content, notification.Data, notification.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Broadcast notification via WebSocket if hub is available
	s.broadcastNotification(notification)

	return nil
}

// CreateBatch creates multiple notifications in a single transaction
func (s *NotificationService) CreateBatch(notifications []*Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	// Start transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statement
	stmt, err := tx.Prepare(`
		INSERT INTO notifications (id, user_id, sender_id, type, content, data, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insert all notifications
	for _, notification := range notifications {
		notification.ID = uuid.New().String()
		notification.CreatedAt = time.Now()

		_, err := stmt.Exec(
			notification.ID,
			notification.UserID,
			notification.SenderID,
			notification.Type,
			notification.Content,
			notification.Data,
			notification.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a notification by ID
func (s *NotificationService) GetByID(id string) (*Notification, error) {
	notification := &Notification{Sender: &User{}}
	var readAt sql.NullTime

	err := s.DB.QueryRow(`
		SELECT n.id, n.user_id, n.sender_id, n.type, n.content, n.data, n.read_at, n.created_at,
			u.id, u.username, u.full_name, u.profile_picture
		FROM notifications n
		JOIN users u ON n.sender_id = u.id
		WHERE n.id = ?
	`, id).Scan(
		&notification.ID, &notification.UserID, &notification.SenderID, &notification.Type, &notification.Content, &notification.Data, &readAt, &notification.CreatedAt,
		&notification.Sender.ID, &notification.Sender.Username, &notification.Sender.FullName, &notification.Sender.ProfilePicture,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("notification not found")
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if readAt.Valid {
		notification.ReadAt = &readAt.Time
	}

	return notification, nil
}

// GetByUser retrieves notifications for a user
func (s *NotificationService) GetByUser(userID string, limit, offset int) ([]*Notification, error) {
	rows, err := s.DB.Query(`
		SELECT n.id, n.user_id, n.sender_id, n.type, n.content, n.data, n.read_at, n.created_at,
			u.id, u.username, u.full_name, u.profile_picture
		FROM notifications n
		JOIN users u ON n.sender_id = u.id
		WHERE n.user_id = ?
		ORDER BY n.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		notification := &Notification{Sender: &User{}}
		var readAt sql.NullTime

		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.SenderID, &notification.Type, &notification.Content, &notification.Data, &readAt, &notification.CreatedAt,
			&notification.Sender.ID, &notification.Sender.Username, &notification.Sender.FullName, &notification.Sender.ProfilePicture,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if readAt.Valid {
			notification.ReadAt = &readAt.Time
		}

		// Enhance notification with additional data based on type
		if err := s.enhanceNotificationData(notification); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to enhance notification data: %v\n", err)
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(id, userID string) error {
	// Check if notification belongs to user
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM notifications WHERE id = ? AND user_id = ?", id, userID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check notification ownership: %w", err)
	}

	if count == 0 {
		return errors.New("notification not found or not authorized")
	}

	// Mark as read
	now := time.Now()
	_, err = s.DB.Exec("UPDATE notifications SET read_at = ? WHERE id = ?", now, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllAsRead(userID string) error {
	now := time.Now()
	_, err := s.DB.Exec("UPDATE notifications SET read_at = ? WHERE user_id = ? AND read_at IS NULL", now, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// GetUnreadCount returns the number of unread notifications for a user
func (s *NotificationService) GetUnreadCount(userID string) (int, error) {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read_at IS NULL", userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// Delete deletes a specific notification
func (s *NotificationService) Delete(id, userID string) error {
	// Check if notification belongs to user
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM notifications WHERE id = ? AND user_id = ?", id, userID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check notification ownership: %w", err)
	}

	if count == 0 {
		return errors.New("notification not found or not authorized")
	}

	// Delete the notification
	_, err = s.DB.Exec("DELETE FROM notifications WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return nil
}

// DeleteAll deletes all notifications for a user
func (s *NotificationService) DeleteAll(userID string) error {
	_, err := s.DB.Exec("DELETE FROM notifications WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete all notifications: %w", err)
	}

	return nil
}

// enhanceNotificationData adds additional context to notifications based on their type
func (s *NotificationService) enhanceNotificationData(notification *Notification) error {
	switch notification.Type {
	case NotificationTypePostLike:
		return s.enhancePostLikeNotification(notification)
	case NotificationTypePostComment:
		return s.enhancePostCommentNotification(notification)
	case NotificationTypeGroupEventCreated:
		return s.enhanceGroupEventNotification(notification)
	case NotificationTypeGroupInvite, NotificationTypeGroupJoinRequest, NotificationTypeGroupJoinApproved, NotificationTypeGroupJoinRejected:
		return s.enhanceGroupNotification(notification)
	case NotificationTypeEventInvite:
		return s.enhanceEventInviteNotification(notification)
	default:
		return nil
	}
}

// enhancePostLikeNotification adds post information to like notifications
func (s *NotificationService) enhancePostLikeNotification(notification *Notification) error {
	if notification.Data == "" {
		return nil
	}

	// Parse the existing data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(notification.Data), &data); err != nil {
		return fmt.Errorf("failed to parse notification data: %w", err)
	}

	// Get post ID from data
	postID, ok := data["postId"].(string)
	if !ok {
		return nil
	}

	// Fetch post information - check both regular posts and group posts
	var postContent string
	var groupID sql.NullString

	// First try regular posts table
	err := s.DB.QueryRow("SELECT content FROM posts WHERE id = ?", postID).Scan(&postContent)
	if err != nil {
		if err == sql.ErrNoRows {
			// Try group posts table
			err = s.DB.QueryRow("SELECT content, group_id FROM group_posts WHERE id = ?", postID).Scan(&postContent, &groupID)
			if err != nil {
				if err == sql.ErrNoRows {
					// Post might have been deleted
					data["postContent"] = "a deleted post"
				} else {
					return fmt.Errorf("failed to fetch group post content: %w", err)
				}
			} else {
				// It's a group post, fetch group name
				if groupID.Valid && groupID.String != "" {
					var groupName string
					err := s.DB.QueryRow("SELECT name FROM groups WHERE id = ?", groupID.String).Scan(&groupName)
					if err == nil {
						data["groupId"] = groupID.String
						data["groupName"] = groupName
					}
				}
			}
		} else {
			return fmt.Errorf("failed to fetch post content: %w", err)
		}
	}

	// Truncate content if too long and we have content
	if postContent != "" {
		if len(postContent) > 50 {
			postContent = postContent[:50] + "..."
		}
		data["postContent"] = postContent
	}

	// Update notification data
	updatedData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal updated data: %w", err)
	}
	notification.Data = string(updatedData)

	return nil
}

// enhancePostCommentNotification ensures comment data is properly formatted
func (s *NotificationService) enhancePostCommentNotification(notification *Notification) error {
	if notification.Data == "" {
		return nil
	}

	// Parse the existing data to ensure it's valid JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(notification.Data), &data); err != nil {
		return fmt.Errorf("failed to parse notification data: %w", err)
	}

	// Get post ID and comment from data
	postID, _ := data["postId"].(string)
	comment, _ := data["comment"].(string)

	// If we have post ID, get post content and group info for context
	if postID != "" {
		var postContent string
		var groupID sql.NullString

		// First try regular posts table
		err := s.DB.QueryRow("SELECT content FROM posts WHERE id = ?", postID).Scan(&postContent)
		if err != nil {
			if err == sql.ErrNoRows {
				// Try group posts table
				err = s.DB.QueryRow("SELECT content, group_id FROM group_posts WHERE id = ?", postID).Scan(&postContent, &groupID)
				if err != nil {
					if err == sql.ErrNoRows {
						data["postContent"] = "a deleted post"
					}
				} else {
					// It's a group post, fetch group name
					if groupID.Valid && groupID.String != "" {
						var groupName string
						err := s.DB.QueryRow("SELECT name FROM groups WHERE id = ?", groupID.String).Scan(&groupName)
						if err == nil {
							data["groupId"] = groupID.String
							data["groupName"] = groupName
						}
					}
				}
			}
		}

		// Truncate content if too long and we have content
		if postContent != "" {
			if len(postContent) > 50 {
				postContent = postContent[:50] + "..."
			}
			data["postContent"] = postContent
		}
	}

	// Ensure comment is properly stored
	if comment != "" {
		data["comment"] = comment
	}

	// Update notification data
	updatedData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal updated data: %w", err)
	}
	notification.Data = string(updatedData)

	return nil
}

// enhanceGroupEventNotification adds event details to group event notifications
func (s *NotificationService) enhanceGroupEventNotification(notification *Notification) error {
	if notification.Data == "" {
		return nil
	}

	// Parse the existing data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(notification.Data), &data); err != nil {
		return fmt.Errorf("failed to parse notification data: %w", err)
	}

	// Get event ID from data
	eventID, ok := data["eventId"].(string)
	if !ok {
		return nil
	}

	// Fetch event information
	var eventTitle, eventLocation string
	var startTime, endTime time.Time
	err := s.DB.QueryRow(`
		SELECT title, location, start_time, end_time
		FROM events
		WHERE id = ?
	`, eventID).Scan(&eventTitle, &eventLocation, &startTime, &endTime)

	if err != nil {
		if err == sql.ErrNoRows {
			// Event might have been deleted
			data["eventTitle"] = "a deleted event"
		} else {
			return fmt.Errorf("failed to fetch event details: %w", err)
		}
	} else {
		// Add event details to data
		data["eventTitle"] = eventTitle
		if eventLocation != "" {
			data["eventLocation"] = eventLocation
		}
		data["eventStartTime"] = startTime.Format(time.RFC3339)
		data["eventEndTime"] = endTime.Format(time.RFC3339)
	}

	// Update notification data
	updatedData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal updated data: %w", err)
	}
	notification.Data = string(updatedData)

	return nil
}

// enhanceGroupNotification adds group details to group-related notifications
func (s *NotificationService) enhanceGroupNotification(notification *Notification) error {
	if notification.Data == "" {
		return nil
	}

	// Parse the existing data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(notification.Data), &data); err != nil {
		return fmt.Errorf("failed to parse notification data: %w", err)
	}

	// Get group ID from data
	groupID, ok := data["groupId"].(string)
	if !ok {
		return nil
	}

	// Fetch group information if not already present
	if _, hasGroupName := data["groupName"]; !hasGroupName {
		var groupName string
		err := s.DB.QueryRow("SELECT name FROM groups WHERE id = ?", groupID).Scan(&groupName)
		if err != nil {
			if err == sql.ErrNoRows {
				data["groupName"] = "a deleted group"
			} else {
				return fmt.Errorf("failed to fetch group name: %w", err)
			}
		} else {
			data["groupName"] = groupName
		}

		// Update notification data
		updatedData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal updated data: %w", err)
		}
		notification.Data = string(updatedData)
	}

	return nil
}

// enhanceEventInviteNotification adds event details to event invite notifications
func (s *NotificationService) enhanceEventInviteNotification(notification *Notification) error {
	if notification.Data == "" {
		return nil
	}

	// Parse the existing data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(notification.Data), &data); err != nil {
		return fmt.Errorf("failed to parse notification data: %w", err)
	}

	// Get event ID from data
	eventID, ok := data["eventId"].(string)
	if !ok {
		return nil
	}

	// Fetch event information
	var eventTitle, eventLocation string
	var startTime, endTime time.Time
	err := s.DB.QueryRow(`
		SELECT title, location, start_time, end_time
		FROM events
		WHERE id = ?
	`, eventID).Scan(&eventTitle, &eventLocation, &startTime, &endTime)

	if err != nil {
		if err == sql.ErrNoRows {
			// Event might have been deleted
			data["eventTitle"] = "a deleted event"
		} else {
			return fmt.Errorf("failed to fetch event details: %w", err)
		}
	} else {
		// Add event details to data
		data["eventTitle"] = eventTitle
		if eventLocation != "" {
			data["eventLocation"] = eventLocation
		}
		data["eventStartTime"] = startTime.Format(time.RFC3339)
		data["eventEndTime"] = endTime.Format(time.RFC3339)
	}

	// Update notification data
	updatedData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal updated data: %w", err)
	}
	notification.Data = string(updatedData)

	return nil
}

// BroadcastNotification broadcasts a notification via WebSocket
func (s *NotificationService) BroadcastNotification(notification *Notification) {
	if s.Hub == nil {
		return
	}

	// Enhance notification data before broadcasting
	if err := s.enhanceNotificationData(notification); err != nil {
		// Log error but continue with broadcast
		fmt.Printf("Warning: failed to enhance notification data for broadcast: %v\n", err)
	}

	// Use the custom broadcast function if available
	if s.BroadcastFunc != nil {
		s.BroadcastFunc(notification)
		fmt.Printf("Notification broadcast sent for user %s: %s\n", notification.UserID, notification.Type)
	} else {
		fmt.Printf("Warning: No broadcast function available for notifications\n")
	}
}

// sendNotificationBroadcast sends the notification broadcast
// This will be overridden by the handler to use the actual hub
func (s *NotificationService) sendNotificationBroadcast(data []byte) {
	// Default implementation - just log
	fmt.Printf("Notification broadcast ready: %s\n", string(data))
}

// broadcastNotification is the internal method called during Create
func (s *NotificationService) broadcastNotification(notification *Notification) {
	// Enhance notification with sender information before broadcasting
	if notification.SenderID != "" {
		sender := &User{}
		err := s.DB.QueryRow(`
			SELECT id, username, full_name, profile_picture
			FROM users
			WHERE id = ?
		`, notification.SenderID).Scan(
			&sender.ID, &sender.Username, &sender.FullName, &sender.ProfilePicture,
		)
		if err == nil {
			notification.Sender = sender
		}
	}

	if s.BroadcastFunc != nil {
		s.BroadcastFunc(notification)
	} else {
		// Fallback to the public method
		s.BroadcastNotification(notification)
	}
}
