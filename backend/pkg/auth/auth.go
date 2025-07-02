package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bernaotieno/social-network/backend/pkg/models"
	"github.com/gorilla/sessions"
)

// Session cookie name
const (
	SessionCookieName = "social_network_session"
	SessionDuration   = 7 * 24 * time.Hour // 7 days
)

// Store is the session store
var Store *sessions.CookieStore

// Initialize initializes the auth package
func Initialize(secret []byte) {
	Store = sessions.NewCookieStore(secret)
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(SessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
}

// CreateSession creates a new session for a user
func CreateSession(ctx context.Context, db *sql.DB, userID string, w http.ResponseWriter, r *http.Request) (string, error) {
	// Delete all existing sessions for this user first
	sessionService := models.NewSessionService(db)
	if err := sessionService.DeleteAllForUser(userID); err != nil {
		return "", fmt.Errorf("failed to delete existing sessions: %w", err)
	}

	// Create new session
	session, err := sessionService.Create(userID, SessionDuration)
	if err != nil {
		return "", fmt.Errorf("failed to create session in database: %w", err)
	}

	// Create cookie session
	cookieSession, err := Store.Get(r, SessionCookieName)
	if err != nil {
		return "", fmt.Errorf("failed to get cookie session: %w", err)
	}

	// Set session values
	if cookieSession.Values == nil {
		cookieSession.Values = make(map[interface{}]interface{})
	}
	cookieSession.Values["session_id"] = session.ID

	// Save cookie
	if err := cookieSession.Save(r, w); err != nil {
		sessionService.Delete(session.ID)
		return "", fmt.Errorf("failed to save session cookie: %w", err)
	}

	log.Printf("Session created successfully: %s for user: %s", session.ID, userID)
	return session.ID, nil
}

// GetSessionCookie gets the session ID from the cookie
func GetSessionCookie(r *http.Request) (string, error) {
	session, err := Store.Get(r, SessionCookieName)
	if err != nil {
		return "", fmt.Errorf("failed to get session cookie: %w", err)
	}

	// Check if session values exist
	if session.Values == nil || len(session.Values) == 0 {
		return "", errors.New("session cookie is empty")
	}

	// Get session ID from cookie
	sessionIDValue, exists := session.Values["session_id"]
	if !exists {
		return "", errors.New("session_id key not found in cookie")
	}

	// Type assert to string
	sessionID, ok := sessionIDValue.(string)
	if !ok {
		return "", fmt.Errorf("session_id is not a string, got type %T", sessionIDValue)
	}

	// Check if session ID is empty
	if sessionID == "" {
		return "", errors.New("session_id is empty")
	}

	return sessionID, nil
}

// ValidateSession validates a session and returns the user ID
func ValidateSession(ctx context.Context, db *sql.DB, sessionID string) (string, error) {
	sessionService := models.NewSessionService(db)
	session, err := sessionService.GetByID(sessionID)
	if err != nil {
		return "", err
	}

	// Check if this is the most recent session for this user
	isValid, err := sessionService.IsLatestSession(session.UserID, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to validate session: %w", err)
	}

	if !isValid {
		// Delete invalid session
		sessionService.Delete(sessionID)
		return "", errors.New("session invalidated by newer login")
	}

	// Check expiration
	if session.ExpiresAt.Before(time.Now()) {
		sessionService.Delete(sessionID)
		return "", errors.New("session has expired")
	}

	return session.UserID, nil
}

// ClearSession clears the session cookie and removes the session from the database
func ClearSession(ctx context.Context, db *sql.DB, w http.ResponseWriter, r *http.Request) error {
	// Try to get session ID from cookie
	sessionID, err := GetSessionCookie(r)
	if err == nil {
		// Delete session from database if we found a session ID
		sessionService := models.NewSessionService(db)
		if err := sessionService.Delete(sessionID); err != nil {
			log.Printf("Failed to delete session from database: %v", err)
			// Continue to clear cookie even if database deletion fails
		}
	}

	// Always try to clear the cookie, even if we couldn't get the session ID
	session, err := Store.Get(r, SessionCookieName)
	if err != nil {
		// If we can't get the session, create a new one to clear it
		session = sessions.NewSession(Store, SessionCookieName)
	}

	// Clear all values and delete the cookie by setting MaxAge to -1
	session.Values = nil
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("failed to clear session cookie: %w", err)
	}

	log.Printf("Session cleared successfully")
	return nil
}
