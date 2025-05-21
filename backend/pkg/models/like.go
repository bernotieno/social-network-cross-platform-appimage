package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Like represents a like on a post
type Like struct {
	ID        string    `json:"id"`
	PostID    string    `json:"postId"`
	UserID    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}

// LikeService handles like-related operations
type LikeService struct {
	DB *sql.DB
}

// NewLikeService creates a new LikeService
func NewLikeService(db *sql.DB) *LikeService {
	return &LikeService{DB: db}
}

// Create creates a new like
func (s *LikeService) Create(postID, userID string) error {
	// Check if the like already exists
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing like: %w", err)
	}

	if count > 0 {
		return errors.New("post already liked by user")
	}

	// Create the like
	like := &Like{
		ID:        uuid.New().String(),
		PostID:    postID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	_, err = s.DB.Exec(`
		INSERT INTO likes (id, post_id, user_id, created_at)
		VALUES (?, ?, ?, ?)
	`, like.ID, like.PostID, like.UserID, like.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create like: %w", err)
	}

	return nil
}

// Delete deletes a like
func (s *LikeService) Delete(postID, userID string) error {
	result, err := s.DB.Exec("DELETE FROM likes WHERE post_id = ? AND user_id = ?", postID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete like: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("like not found")
	}

	return nil
}

// GetLikesByPost retrieves all likes for a post
func (s *LikeService) GetLikesByPost(postID string, limit, offset int) ([]*Like, error) {
	rows, err := s.DB.Query(`
		SELECT id, post_id, user_id, created_at
		FROM likes
		WHERE post_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, postID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to get likes: %w", err)
	}
	defer rows.Close()

	var likes []*Like
	for rows.Next() {
		like := &Like{}
		err := rows.Scan(&like.ID, &like.PostID, &like.UserID, &like.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan like: %w", err)
		}
		likes = append(likes, like)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating likes: %w", err)
	}

	return likes, nil
}

// GetLikeCount returns the number of likes for a post
func (s *LikeService) GetLikeCount(postID string) (int, error) {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ?", postID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get like count: %w", err)
	}

	return count, nil
}

// HasUserLikedPost checks if a user has liked a post
func (s *LikeService) HasUserLikedPost(postID, userID string) (bool, error) {
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if user liked post: %w", err)
	}

	return count > 0, nil
}
