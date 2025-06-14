package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FollowStatus represents the status of a follow relationship
type FollowStatus string

const (
	FollowStatusPending  FollowStatus = "pending"
	FollowStatusAccepted FollowStatus = "accepted"
	FollowStatusRejected FollowStatus = "rejected"
)

// Follow represents a follow relationship between users
type Follow struct {
	ID          string       `json:"id"`
	FollowerID  string       `json:"followerId"`
	FollowingID string       `json:"followingId"`
	Status      FollowStatus `json:"status"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// FollowService handles follow-related operations
type FollowService struct {
	DB *sql.DB
}

// NewFollowService creates a new FollowService
func NewFollowService(db *sql.DB) *FollowService {
	return &FollowService{DB: db}
}

// Create creates a new follow relationship
func (s *FollowService) Create(followerID, followingID string, isPrivate bool) (*Follow, error) {
	// Check if the follow relationship already exists
	existing, err := s.GetByUserIDs(followerID, followingID)
	if err == nil {
		// Follow relationship already exists
		return existing, errors.New("follow relationship already exists")
	}

	// Determine initial status based on target user's privacy setting
	status := FollowStatusAccepted
	if isPrivate {
		status = FollowStatusPending
	}

	follow := &Follow{
		ID:          uuid.New().String(),
		FollowerID:  followerID,
		FollowingID: followingID,
		Status:      status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = s.DB.Exec(`
		INSERT INTO follows (id, follower_id, following_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, follow.ID, follow.FollowerID, follow.FollowingID, follow.Status, follow.CreatedAt, follow.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create follow: %w", err)
	}

	return follow, nil
}

// GetByID retrieves a follow relationship by ID
func (s *FollowService) GetByID(id string) (*Follow, error) {
	follow := &Follow{}
	err := s.DB.QueryRow(`
		SELECT id, follower_id, following_id, status, created_at, updated_at
		FROM follows
		WHERE id = ?
	`, id).Scan(&follow.ID, &follow.FollowerID, &follow.FollowingID, &follow.Status, &follow.CreatedAt, &follow.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("follow relationship not found")
		}
		return nil, fmt.Errorf("failed to get follow: %w", err)
	}

	return follow, nil
}

// GetByUserIDs retrieves a follow relationship by follower and following IDs
func (s *FollowService) GetByUserIDs(followerID, followingID string) (*Follow, error) {
	follow := &Follow{}
	err := s.DB.QueryRow(`
		SELECT id, follower_id, following_id, status, created_at, updated_at
		FROM follows
		WHERE follower_id = ? AND following_id = ?
	`, followerID, followingID).Scan(&follow.ID, &follow.FollowerID, &follow.FollowingID, &follow.Status, &follow.CreatedAt, &follow.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("follow relationship not found")
		}
		return nil, fmt.Errorf("failed to get follow: %w", err)
	}

	return follow, nil
}

// UpdateStatus updates the status of a follow relationship
func (s *FollowService) UpdateStatus(id string, status FollowStatus) error {
	_, err := s.DB.Exec(`
		UPDATE follows
		SET status = ?, updated_at = ?
		WHERE id = ?
	`, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update follow status: %w", err)
	}

	return nil
}

// Delete deletes a follow relationship
func (s *FollowService) Delete(followerID, followingID string) error {
	_, err := s.DB.Exec(`
		DELETE FROM follows
		WHERE follower_id = ? AND following_id = ?
	`, followerID, followingID)
	if err != nil {
		return fmt.Errorf("failed to delete follow: %w", err)
	}

	return nil
}

// GetFollowers retrieves all users who follow a user
func (s *FollowService) GetFollowers(userID string, limit, offset int) ([]*User, error) {
	rows, err := s.DB.Query(`
		SELECT u.id, u.username, u.email, u.password, u.full_name, u.bio, u.profile_picture, u.cover_photo, u.is_private, u.created_at, u.updated_at
		FROM users u
		JOIN follows f ON u.id = f.follower_id
		WHERE f.following_id = ? AND f.status = ?
		LIMIT ? OFFSET ?
	`, userID, FollowStatusAccepted, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}
	defer rows.Close()

	var followers []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.FullName, &user.Bio, &user.ProfilePicture, &user.CoverPhoto, &user.IsPrivate, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follower: %w", err)
		}
		followers = append(followers, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating followers: %w", err)
	}

	return followers, nil
}

// GetFollowing retrieves all users a user follows
func (s *FollowService) GetFollowing(userID string, limit, offset int) ([]*User, error) {
	rows, err := s.DB.Query(`
		SELECT u.id, u.username, u.email, u.password, u.full_name, u.bio, u.profile_picture, u.cover_photo, u.is_private, u.created_at, u.updated_at
		FROM users u
		JOIN follows f ON u.id = f.following_id
		WHERE f.follower_id = ? AND f.status = ?
		LIMIT ? OFFSET ?
	`, userID, FollowStatusAccepted, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}
	defer rows.Close()

	var following []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.FullName, &user.Bio, &user.ProfilePicture, &user.CoverPhoto, &user.IsPrivate, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan following: %w", err)
		}
		following = append(following, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating following: %w", err)
	}

	return following, nil
}

// GetFollowRequests retrieves all pending follow requests for a user
func (s *FollowService) GetFollowRequests(userID string) ([]*Follow, error) {
	rows, err := s.DB.Query(`
		SELECT id, follower_id, following_id, status, created_at, updated_at
		FROM follows
		WHERE following_id = ? AND status = ?
	`, userID, FollowStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to get follow requests: %w", err)
	}
	defer rows.Close()

	var requests []*Follow
	for rows.Next() {
		follow := &Follow{}
		err := rows.Scan(&follow.ID, &follow.FollowerID, &follow.FollowingID, &follow.Status, &follow.CreatedAt, &follow.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follow request: %w", err)
		}
		requests = append(requests, follow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating follow requests: %w", err)
	}

	return requests, nil
}

// IsFollowing checks if a user is following another user
func (s *FollowService) IsFollowing(followerID, followingID string) (bool, error) {
	var count int
	err := s.DB.QueryRow(`
		SELECT COUNT(*)
		FROM follows
		WHERE follower_id = ? AND following_id = ? AND status = ?
	`, followerID, followingID, FollowStatusAccepted).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if following: %w", err)
	}

	return count > 0, nil
}

// GetFollowersCount returns the number of followers a user has
func (s *FollowService) GetFollowersCount(userID string) (int, error) {
	var count int
	err := s.DB.QueryRow(`
		SELECT COUNT(*)
		FROM follows
		WHERE following_id = ? AND status = ?
	`, userID, FollowStatusAccepted).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get followers count: %w", err)
	}

	return count, nil
}

// GetFollowingCount returns the number of users a user is following
func (s *FollowService) GetFollowingCount(userID string) (int, error) {
	var count int
	err := s.DB.QueryRow(`
		SELECT COUNT(*)
		FROM follows
		WHERE follower_id = ? AND status = ?
	`, userID, FollowStatusAccepted).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get following count: %w", err)
	}

	return count, nil
}

// GetFollowStatus returns the follow status between two users
func (s *FollowService) GetFollowStatus(followerID, followingID string) (FollowStatus, error) {
	var status FollowStatus
	err := s.DB.QueryRow(`
		SELECT status
		FROM follows
		WHERE follower_id = ? AND following_id = ?
	`, followerID, followingID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("follow relationship not found")
		}
		return "", fmt.Errorf("failed to get follow status: %w", err)
	}

	return status, nil
}

// HasPendingRequest checks if there's a pending follow request between two users
func (s *FollowService) HasPendingRequest(followerID, followingID string) (bool, error) {
	var count int
	err := s.DB.QueryRow(`
		SELECT COUNT(*)
		FROM follows
		WHERE follower_id = ? AND following_id = ? AND status = ?
	`, followerID, followingID, FollowStatusPending).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check pending request: %w", err)
	}

	return count > 0, nil
}
