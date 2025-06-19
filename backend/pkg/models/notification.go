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
	DB *sql.DB
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(db *sql.DB) *NotificationService {
	return &NotificationService{DB: db}
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

	// Fetch post information
	var postContent string
	err := s.DB.QueryRow("SELECT content FROM posts WHERE id = ?", postID).Scan(&postContent)
	if err != nil {
		if err == sql.ErrNoRows {
			// Post might have been deleted
			data["postContent"] = "a deleted post"
		} else {
			return fmt.Errorf("failed to fetch post content: %w", err)
		}
	} else {
		// Truncate content if too long
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

	// If we have post ID, get post content for context
	if postID != "" {
		var postContent string
		err := s.DB.QueryRow("SELECT content FROM posts WHERE id = ?", postID).Scan(&postContent)
		if err != nil {
			if err == sql.ErrNoRows {
				data["postContent"] = "a deleted post"
			}
		} else {
			// Truncate content if too long
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
