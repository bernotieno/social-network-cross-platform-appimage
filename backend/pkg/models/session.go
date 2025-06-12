package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// SessionService handles session-related operations
type SessionService struct {
	DB *sql.DB
}

// NewSessionService creates a new SessionService
func NewSessionService(db *sql.DB) *SessionService {
	return &SessionService{DB: db}
}

// Create creates a new session for a user
func (s *SessionService) Create(userID string, duration time.Duration) (*Session, error) {
	session := &Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(duration),
		CreatedAt: time.Now(),
	}

	log.Printf("Creating session with ID: %s for user: %s", session.ID, userID)

	_, err := s.DB.Exec(`
		INSERT INTO sessions (id, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`, session.ID, session.UserID, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		log.Printf("Failed to insert session into database: %v", err)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	log.Printf("Session created successfully in database with ID: %s", session.ID)
	return session, nil
}

// GetByID retrieves a session by ID
func (s *SessionService) GetByID(id string) (*Session, error) {
	session := &Session{}
	err := s.DB.QueryRow(`
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE id = ?
	`, id).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session has expired
	if session.ExpiresAt.Before(time.Now()) {
		// Delete expired session
		_, err := s.DB.Exec("DELETE FROM sessions WHERE id = ?", id)
		if err != nil {
			return nil, fmt.Errorf("failed to delete expired session: %w", err)
		}
		return nil, errors.New("session has expired")
	}

	return session, nil
}

// Delete deletes a session
func (s *SessionService) Delete(id string) error {
	_, err := s.DB.Exec("DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteAllForUser deletes all sessions for a user
func (s *SessionService) DeleteAllForUser(userID string) error {
	_, err := s.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	return nil
}

// CleanupExpiredSessions removes all expired sessions from the database
func (s *SessionService) CleanupExpiredSessions() error {
	_, err := s.DB.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}
