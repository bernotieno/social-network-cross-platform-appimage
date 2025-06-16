package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PostViewer represents a user who can view a custom visibility post
type PostViewer struct {
	ID        string    `json:"id"`
	PostID    string    `json:"postId"`
	UserID    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	// Additional fields for API responses
	User *User `json:"user,omitempty"`
}

// PostViewerService handles post viewer-related operations
type PostViewerService struct {
	DB *sql.DB
}

// NewPostViewerService creates a new PostViewerService
func NewPostViewerService(db *sql.DB) *PostViewerService {
	return &PostViewerService{DB: db}
}

// AddViewers adds multiple viewers to a post
func (s *PostViewerService) AddViewers(postID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	// Start transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing viewers for this post
	_, err = tx.Exec("DELETE FROM post_viewers WHERE post_id = ?", postID)
	if err != nil {
		return fmt.Errorf("failed to clear existing viewers: %w", err)
	}

	// Add new viewers
	stmt, err := tx.Prepare("INSERT INTO post_viewers (id, post_id, user_id, created_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, userID := range userIDs {
		id := uuid.New().String()
		_, err = stmt.Exec(id, postID, userID, now)
		if err != nil {
			return fmt.Errorf("failed to add viewer %s: %w", userID, err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetViewers retrieves all viewers for a post
func (s *PostViewerService) GetViewers(postID string) ([]*PostViewer, error) {
	rows, err := s.DB.Query(`
		SELECT pv.id, pv.post_id, pv.user_id, pv.created_at,
			u.id, u.username, u.full_name, u.profile_picture
		FROM post_viewers pv
		JOIN users u ON pv.user_id = u.id
		WHERE pv.post_id = ?
		ORDER BY pv.created_at ASC
	`, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get viewers: %w", err)
	}
	defer rows.Close()

	var viewers []*PostViewer
	for rows.Next() {
		viewer := &PostViewer{User: &User{}}
		err := rows.Scan(
			&viewer.ID, &viewer.PostID, &viewer.UserID, &viewer.CreatedAt,
			&viewer.User.ID, &viewer.User.Username, &viewer.User.FullName, &viewer.User.ProfilePicture,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan viewer: %w", err)
		}
		viewers = append(viewers, viewer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating viewers: %w", err)
	}

	return viewers, nil
}

// CanUserViewPost checks if a user can view a custom visibility post
func (s *PostViewerService) CanUserViewPost(postID, userID string) (bool, error) {
	var count int
	err := s.DB.QueryRow(`
		SELECT COUNT(*)
		FROM post_viewers
		WHERE post_id = ? AND user_id = ?
	`, postID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check viewer permission: %w", err)
	}

	return count > 0, nil
}

// RemoveViewer removes a specific viewer from a post
func (s *PostViewerService) RemoveViewer(postID, userID string) error {
	result, err := s.DB.Exec(`
		DELETE FROM post_viewers
		WHERE post_id = ? AND user_id = ?
	`, postID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove viewer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("viewer not found")
	}

	return nil
}

// GetViewerUserIDs returns just the user IDs of viewers for a post
func (s *PostViewerService) GetViewerUserIDs(postID string) ([]string, error) {
	rows, err := s.DB.Query(`
		SELECT user_id
		FROM post_viewers
		WHERE post_id = ?
		ORDER BY created_at ASC
	`, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get viewer user IDs: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user IDs: %w", err)
	}

	return userIDs, nil
}
