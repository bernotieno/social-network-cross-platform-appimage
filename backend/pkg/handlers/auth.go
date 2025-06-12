package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bernaotieno/social-network/backend/pkg/auth"
	"github.com/bernaotieno/social-network/backend/pkg/middleware"
	"github.com/bernaotieno/social-network/backend/pkg/models"
	"github.com/bernaotieno/social-network/backend/pkg/utils"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	FullName    string `json:"fullName"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DateOfBirth string `json:"dateOfBirth"` // Will be parsed to time.Time
	Bio         string `json:"bio"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (for avatar upload support)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	// Extract form values
	req := RegisterRequest{
		Username:    r.FormValue("username"),
		Email:       r.FormValue("email"),
		Password:    r.FormValue("password"),
		FullName:    r.FormValue("fullName"),
		FirstName:   r.FormValue("firstName"),
		LastName:    r.FormValue("lastName"),
		DateOfBirth: r.FormValue("dateOfBirth"),
		Bio:         r.FormValue("bio"),
	}
	log.Println("register request:", req)

	// Validate request - only required fields
	if req.Username == "" || req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.DateOfBirth == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Username, email, password, first name, last name, and date of birth are required")
		return
	}

	// Validate email format
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	// Validate password length
	if len(req.Password) < 6 {
		utils.RespondWithError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	// Check if user already exists
	exists, err := h.UserService.UserExists(req.Email, req.Username)
	log.Println("user exist", exists)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check user existence")
		return
	}
	if exists {
		utils.RespondWithError(w, http.StatusConflict, "User with this email or username already exists")
		return
	}

	// Parse date of birth
	var dateOfBirth *time.Time
	if req.DateOfBirth != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.DateOfBirth); err == nil {
			dateOfBirth = &parsedDate
		} else {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid date of birth format. Use YYYY-MM-DD")
			return
		}
	}

	// Handle avatar upload if provided
	var avatarPath string
	file, header, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()

		// Save avatar image
		avatarPath, err = utils.SaveImage(file, header, "avatars")
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Failed to save avatar: "+err.Error())
			return
		}
	}

	// Generate fullName from firstName and lastName
	fullName := req.FullName
	if fullName == "" {
		fullName = strings.TrimSpace(req.FirstName + " " + req.LastName)
	}

	// Create user with all fields
	user := &models.User{
		Username:       req.Username,
		Email:          req.Email,
		Password:       req.Password,
		FullName:       fullName,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    dateOfBirth,
		Bio:            req.Bio,
		ProfilePicture: avatarPath,
	}
	if err := h.UserService.Create(user); err != nil {
		// If user creation fails and we uploaded an avatar, clean it up
		if avatarPath != "" {
			utils.DeleteImage(avatarPath)
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Create session
	sessionID, err := auth.CreateSession(r.Context(), h.DB, user.ID, w, r)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Return user data (without password)
	user.Password = ""
	utils.RespondWithSuccess(w, http.StatusCreated, "User registered successfully", map[string]interface{}{
		"user":  user,
		"token": sessionID,
	})
}

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	log.Println("login request:", req)

	// Validate request
	if req.Email == "" || req.Password == "" {
		log.Println("Login validation failed: missing email or password")
		utils.RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	log.Printf("Looking up user by email: %s", req.Email)
	// Get user by email
	user, err := h.UserService.GetByEmail(req.Email)
	if err != nil {
		log.Printf("User lookup failed for email %s: %v", req.Email, err)
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	log.Printf("User found: ID=%s, Email=%s", user.ID, user.Email)

	// Check password
	log.Println("Checking password...")
	if !h.UserService.CheckPassword(user, req.Password) {
		log.Printf("Password check failed for user %s", user.Email)
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	log.Printf("Password check successful for user %s", user.Email)

	// Create session
	log.Printf("Creating session for user ID: %s", user.ID)
	sessionID, err := auth.CreateSession(r.Context(), h.DB, user.ID, w, r)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}
	log.Printf("Session created successfully with ID: %s", sessionID)

	// Return user data (without password)
	user.Password = ""
	responseData := map[string]interface{}{
		"user":  user,
		"token": sessionID,
	}
	log.Printf("Sending login response with token: %s", sessionID)
	utils.RespondWithSuccess(w, http.StatusOK, "Login successful", responseData)
}

// Logout handles user logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Clear session
	if err := auth.ClearSession(r.Context(), h.DB, w, r); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	// Delete all sessions for the user
	if err := h.SessionService.DeleteAllForUser(userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete sessions")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Logout successful", nil)
}
