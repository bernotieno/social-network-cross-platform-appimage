package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/bernaotieno/social-network/backend/pkg/middleware"
	"github.com/bernaotieno/social-network/backend/pkg/models"
	"github.com/bernaotieno/social-network/backend/pkg/utils"
	"github.com/bernaotieno/social-network/backend/pkg/websocket"
	"github.com/gorilla/mux"
)

// CreateGroupRequest represents a request to create a group
type CreateGroupRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Privacy     models.GroupPrivacy `json:"privacy"`
}

// UpdateGroupRequest represents a request to update a group
type UpdateGroupRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Privacy     models.GroupPrivacy `json:"privacy"`
}

// CreateGroupPostRequest represents a request to create a group post
type CreateGroupPostRequest struct {
	Content string `json:"content"`
}

// CreateGroupEventRequest represents a request to create a group event
type CreateGroupEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
}

// EventResponseRequest represents a request to respond to an event
type EventResponseRequest struct {
	Response models.EventResponseType `json:"response"`
}

// GetGroups handles retrieving a list of groups
func (h *Handler) GetGroups(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set default values
	limit := 20
	offset := 0

	// Parse limit and offset
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get groups
	groups, err := h.GroupService.GetGroups(query, userID, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get groups")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Groups retrieved successfully", map[string]interface{}{
		"groups": groups,
	})
}

// CreateGroup handles creating a new group
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	// Get form values
	name := r.FormValue("name")
	description := r.FormValue("description")
	privacy := r.FormValue("privacy")

	// Validate name
	if name == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Validate privacy
	if privacy != string(models.GroupPrivacyPublic) && privacy != string(models.GroupPrivacyPrivate) {
		privacy = string(models.GroupPrivacyPublic)
	}

	// Create group
	group := &models.Group{
		Name:        name,
		Description: description,
		CreatorID:   userID,
		Privacy:     models.GroupPrivacy(privacy),
	}

	// Check if cover photo was uploaded
	file, header, err := r.FormFile("coverPhoto")
	if err == nil {
		defer file.Close()

		// Save image
		imagePath, err := utils.SaveImage(file, header, "groups")
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		group.CoverPhoto = imagePath
	}

	// Save group
	if err := h.GroupService.Create(group); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	// Add creator as admin
	member := &models.GroupMember{
		GroupID: group.ID,
		UserID:  userID,
		Role:    models.GroupMemberRoleAdmin,
		Status:  models.GroupMemberStatusAccepted,
	}

	if err := h.GroupMemberService.Create(member); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to add creator as admin")
		return
	}

	// Get user for response
	user, err := h.UserService.GetByID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}
	user.Password = ""

	// Add creator to group for response
	group.Creator = user
	group.MembersCount = 1
	group.IsJoined = true

	utils.RespondWithSuccess(w, http.StatusCreated, "Group created successfully", map[string]interface{}{
		"group": group,
	})
}

// GetGroup handles retrieving a group by ID
func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get current user ID from context
	currentUserID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group
	group, err := h.GroupService.GetByID(groupID, currentUserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Group not found")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Group retrieved successfully", map[string]interface{}{
		"group": group,
	})
}

// UpdateGroup handles updating a group
func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	// Get form values
	name := r.FormValue("name")
	description := r.FormValue("description")
	privacy := r.FormValue("privacy")

	// Validate name
	if name == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Validate privacy
	if privacy != string(models.GroupPrivacyPublic) && privacy != string(models.GroupPrivacyPrivate) {
		privacy = string(models.GroupPrivacyPublic)
	}

	// Get group
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Group not found")
		return
	}

	// Check if user is the group creator
	if group.CreatorID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Not authorized to update this group")
		return
	}

	// Update group fields
	group.Name = name
	group.Description = description
	group.Privacy = models.GroupPrivacy(privacy)

	// Check if cover photo was uploaded
	file, header, err := r.FormFile("coverPhoto")
	if err == nil {
		defer file.Close()

		// Save image
		imagePath, err := utils.SaveImage(file, header, "groups")
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Delete old cover photo if exists
		if group.CoverPhoto != "" {
			utils.DeleteImage(group.CoverPhoto)
		}

		group.CoverPhoto = imagePath
	}

	// Save changes
	if err := h.GroupService.Update(group); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update group")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Group updated successfully", map[string]interface{}{
		"group": group,
	})
}

// DeleteGroup handles deleting a group
func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Delete group
	if err := h.GroupService.Delete(groupID, userID); err != nil {
		if err.Error() == "not authorized to delete this group" {
			utils.RespondWithError(w, http.StatusForbidden, "Not authorized to delete this group")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete group")
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Group deleted successfully", nil)
}

// JoinGroup handles joining a group
func (h *Handler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get group
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Group not found")
		return
	}

	// Check if user is already a member
	if group.IsJoined {
		utils.RespondWithError(w, http.StatusConflict, "Already a member of this group")
		return
	}

	// Determine status based on group privacy
	status := models.GroupMemberStatusAccepted
	if group.Privacy == models.GroupPrivacyPrivate {
		status = models.GroupMemberStatusPending
	}

	// Create group member
	member := &models.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    models.GroupMemberRoleMember,
		Status:  status,
	}

	if err := h.GroupMemberService.Create(member); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to join group")
		return
	}

	// Create notification for group creator if it's a private group
	if group.Privacy == models.GroupPrivacyPrivate {
		notification := &models.Notification{
			UserID:   group.CreatorID,
			SenderID: userID,
			Type:     models.NotificationTypeGroupJoinRequest,
			Content:  "requested to join your group",
			Data:     `{"groupId":"` + groupID + `","groupName":"` + group.Name + `"}`,
		}

		if err := h.NotificationService.Create(notification); err != nil {
			// Log error but don't fail the request
			// TODO: Add proper logging
		}
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Group join request sent", map[string]interface{}{
		"status": status,
	})
}

// LeaveGroup handles leaving a group
func (h *Handler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get group
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Group not found")
		return
	}

	// Check if user is a member
	if !group.IsJoined {
		utils.RespondWithError(w, http.StatusBadRequest, "Not a member of this group")
		return
	}

	// Check if user is the creator
	if group.CreatorID == userID {
		utils.RespondWithError(w, http.StatusBadRequest, "Group creator cannot leave the group. Delete the group instead.")
		return
	}

	// Leave group
	if err := h.GroupMemberService.Delete(groupID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to leave group")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Left group successfully", nil)
}

// GetGroupPosts handles retrieving posts for a group
func (h *Handler) GetGroupPosts(w http.ResponseWriter, r *http.Request) {
	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get current user ID from context
	currentUserID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set default values
	limit := 20
	offset := 0

	// Parse limit and offset
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get posts
	posts, err := h.GroupPostService.GetByGroup(groupID, currentUserID, limit, offset)
	if err != nil {
		if err.Error() == "not authorized to view posts in this group" {
			utils.RespondWithError(w, http.StatusForbidden, "Not authorized to view posts in this group")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get group posts")
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Group posts retrieved successfully", map[string]interface{}{
		"posts": posts,
	})
}

// CreateGroupPost handles creating a post in a group
func (h *Handler) CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Not a member of this group")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	// Get form values
	content := r.FormValue("content")

	// Validate content
	if content == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Create post
	post := &models.GroupPost{
		GroupID: groupID,
		UserID:  userID,
		Content: content,
	}

	// Check if image was uploaded
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Save image
		imagePath, err := utils.SaveImage(file, header, "group_posts")
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		post.Image = imagePath
	}

	// Save post
	if err := h.GroupPostService.Create(post); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	// Get user for response
	user, err := h.UserService.GetByID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}
	user.Password = ""

	// Add user to post for response
	post.User = user

	utils.RespondWithSuccess(w, http.StatusCreated, "Post created successfully", map[string]interface{}{
		"post": post,
	})
}

// GetGroupEvents handles retrieving events for a group
func (h *Handler) GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get current user ID from context
	currentUserID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set default values
	limit := 20
	offset := 0

	// Parse limit and offset
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get events
	events, err := h.EventService.GetByGroup(groupID, currentUserID, limit, offset)
	if err != nil {
		if err.Error() == "not authorized to view events in this group" {
			utils.RespondWithError(w, http.StatusForbidden, "Not authorized to view events in this group")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get group events")
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Group events retrieved successfully", map[string]interface{}{
		"events": events,
	})
}

// CreateGroupEvent handles creating an event in a group
func (h *Handler) CreateGroupEvent(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Not a member of this group")
		return
	}

	// Parse request body
	var req CreateGroupEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Title == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if req.StartTime == "" || req.EndTime == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Start time and end time are required")
		return
	}

	// Parse times
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid start time format")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid end time format")
		return
	}

	// Create event
	event := &models.Event{
		GroupID:     groupID,
		CreatorID:   userID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	// Save event
	if err := h.EventService.Create(event); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create event")
		return
	}

	// Get user for response
	user, err := h.UserService.GetByID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}
	user.Password = ""

	// Add creator to event for response
	event.Creator = user

	utils.RespondWithSuccess(w, http.StatusCreated, "Event created successfully", map[string]interface{}{
		"event": event,
	})
}

// RespondToEvent handles responding to an event
func (h *Handler) RespondToEvent(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get event ID from URL
	vars := mux.Vars(r)
	eventID := vars["id"]

	// Parse request body
	var req EventResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate response
	if req.Response != models.EventResponseGoing &&
		req.Response != models.EventResponseMaybe &&
		req.Response != models.EventResponseNotGoing {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid response")
		return
	}

	// Check if event exists
	_, err = h.EventService.GetByID(eventID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Create or update response
	response := &models.EventResponse{
		EventID:  eventID,
		UserID:   userID,
		Response: req.Response,
	}

	if err := h.EventResponseService.Create(response); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to respond to event")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Response saved successfully", map[string]interface{}{
		"response": req.Response,
	})
}

// GetGroupMembers handles retrieving members of a group
func (h *Handler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get current user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Not a member of this group")
		return
	}

	// Get members
	members, err := h.GroupMemberService.GetMembers(groupID, 100, 0) // Get up to 100 members
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get group members")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Group members retrieved successfully", map[string]interface{}{
		"members": members,
	})
}

// GetGroupPendingRequests handles retrieving pending join requests for a group
func (h *Handler) GetGroupPendingRequests(w http.ResponseWriter, r *http.Request) {
	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get current user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user is an admin of the group
	isAdmin, err := h.GroupMemberService.IsGroupAdmin(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check admin status")
		return
	}

	if !isAdmin {
		utils.RespondWithError(w, http.StatusForbidden, "Only group admins can view pending requests")
		return
	}

	// Get pending requests
	requests, err := h.GroupMemberService.GetPendingRequests(groupID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get pending requests")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Pending requests retrieved successfully", map[string]interface{}{
		"requests": requests,
	})
}

// ApproveJoinRequest handles approving a join request
func (h *Handler) ApproveJoinRequest(w http.ResponseWriter, r *http.Request) {
	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get current user ID from context
	currentUserID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if current user is an admin of the group
	isAdmin, err := h.GroupMemberService.IsGroupAdmin(groupID, currentUserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check admin status")
		return
	}

	if !isAdmin {
		utils.RespondWithError(w, http.StatusForbidden, "Only group admins can approve join requests")
		return
	}

	// Get the pending member request
	member, err := h.GroupMemberService.GetByGroupAndUser(groupID, req.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Join request not found")
		return
	}

	if member.Status != models.GroupMemberStatusPending {
		utils.RespondWithError(w, http.StatusBadRequest, "Request is not pending")
		return
	}

	// Update status to accepted
	if err := h.GroupMemberService.UpdateStatus(member.ID, models.GroupMemberStatusAccepted); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to approve request")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Join request approved successfully", nil)
}

// RejectJoinRequest handles rejecting a join request
func (h *Handler) RejectJoinRequest(w http.ResponseWriter, r *http.Request) {
	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get current user ID from context
	currentUserID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if current user is an admin of the group
	isAdmin, err := h.GroupMemberService.IsGroupAdmin(groupID, currentUserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check admin status")
		return
	}

	if !isAdmin {
		utils.RespondWithError(w, http.StatusForbidden, "Only group admins can reject join requests")
		return
	}

	// Get the pending member request
	member, err := h.GroupMemberService.GetByGroupAndUser(groupID, req.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Join request not found")
		return
	}

	if member.Status != models.GroupMemberStatusPending {
		utils.RespondWithError(w, http.StatusBadRequest, "Request is not pending")
		return
	}

	// Delete the request (reject it)
	if err := h.GroupMemberService.Delete(groupID, req.UserID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to reject request")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Join request rejected successfully", nil)
}

// InviteToGroup handles inviting a user to a group
func (h *Handler) InviteToGroup(w http.ResponseWriter, r *http.Request) {
	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get current user ID from context
	currentUserID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if current user is an admin of the group
	isAdmin, err := h.GroupMemberService.IsGroupAdmin(groupID, currentUserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check admin status")
		return
	}

	if !isAdmin {
		utils.RespondWithError(w, http.StatusForbidden, "Only group admins can invite users")
		return
	}

	// Check if user is already a member
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, req.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if isMember {
		utils.RespondWithError(w, http.StatusConflict, "User is already a member of this group")
		return
	}

	// Check if there's already a pending request or invitation
	existingMember, err := h.GroupMemberService.GetByGroupAndUser(groupID, req.UserID)
	if err == nil {
		if existingMember.Status == models.GroupMemberStatusPending {
			utils.RespondWithError(w, http.StatusConflict, "User already has a pending request for this group")
			return
		}
		if existingMember.Status == models.GroupMemberStatusInvited {
			utils.RespondWithError(w, http.StatusConflict, "User already has a pending invitation for this group")
			return
		}
	}

	// Create group member with invited status (user needs to accept the invitation)
	member := &models.GroupMember{
		GroupID: groupID,
		UserID:  req.UserID,
		Role:    models.GroupMemberRoleMember,
		Status:  models.GroupMemberStatusInvited,
	}

	if err := h.GroupMemberService.Create(member); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to invite user")
		return
	}

	// Get group info for notification
	group, err := h.GroupService.GetByID(groupID, currentUserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get group info")
		return
	}

	// Create notification for invited user
	notification := &models.Notification{
		UserID:   req.UserID,
		SenderID: currentUserID,
		Type:     models.NotificationTypeGroupInvite,
		Content:  "invited you to join the group",
		Data:     `{"groupId":"` + groupID + `","groupName":"` + group.Name + `"}`,
	}

	if err := h.NotificationService.Create(notification); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	utils.RespondWithSuccess(w, http.StatusOK, "User invited successfully", nil)
}

// RespondToGroupInvitation handles accepting or declining a group invitation
func (h *Handler) RespondToGroupInvitation(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get notification ID from URL
	vars := mux.Vars(r)
	notificationID := vars["id"]

	// Parse request body
	var req struct {
		Accept bool `json:"accept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get the notification to extract group information
	notification, err := h.NotificationService.GetByID(notificationID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Notification not found")
		return
	}

	// Verify this is a group invite notification for the current user
	if notification.UserID != userID || notification.Type != models.NotificationTypeGroupInvite {
		utils.RespondWithError(w, http.StatusForbidden, "Invalid notification")
		return
	}

	// Parse group data from notification
	var notificationData struct {
		GroupID   string `json:"groupId"`
		GroupName string `json:"groupName"`
	}
	if err := json.Unmarshal([]byte(notification.Data), &notificationData); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to parse notification data")
		return
	}

	// Get the invitation record
	invitation, err := h.GroupMemberService.GetByGroupAndUser(notificationData.GroupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Invitation not found")
		return
	}

	// Verify the invitation is still pending
	if invitation.Status != models.GroupMemberStatusInvited {
		utils.RespondWithError(w, http.StatusBadRequest, "Invitation is no longer pending")
		return
	}

	// Update the invitation status based on user's response
	var newStatus models.GroupMemberStatus
	if req.Accept {
		newStatus = models.GroupMemberStatusAccepted
	} else {
		newStatus = models.GroupMemberStatusRejected
	}

	if err := h.GroupMemberService.UpdateStatus(invitation.ID, newStatus); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update invitation status")
		return
	}

	// Mark the notification as read
	if err := h.NotificationService.MarkAsRead(notificationID, userID); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	responseMessage := "Invitation declined"
	if req.Accept {
		responseMessage = "Successfully joined the group"
	}

	utils.RespondWithSuccess(w, http.StatusOK, responseMessage, nil)
}

// GetGroupMessages handles retrieving messages for a group
func (h *Handler) GetGroupMessages(w http.ResponseWriter, r *http.Request) {
	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get current user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Not a member of this group")
		return
	}

	// Get messages
	messages, err := h.MessageService.GetGroupMessages(groupID, 50, 0)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get group messages")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Group messages retrieved successfully", map[string]interface{}{
		"messages": messages,
	})
}

// SendGroupMessage handles sending a message to a group
func (h *Handler) SendGroupMessage(w http.ResponseWriter, r *http.Request) {
	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get current user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Content == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Not a member of this group")
		return
	}

	// Create message
	message := &models.Message{
		SenderID: userID,
		GroupID:  groupID,
		Content:  req.Content,
	}

	// Save to database
	if err := h.MessageService.Create(message); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	// Get user for response
	user, err := h.UserService.GetByID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}
	user.Password = ""

	// Add sender to message for response
	message.Sender = user

	// Broadcast the message via WebSocket for real-time delivery
	roomID := "group-" + groupID
	responseMsg := map[string]interface{}{
		"roomId": roomID,
		"message": map[string]interface{}{
			"id":        message.ID,
			"content":   req.Content,
			"sender":    userID,
			"groupId":   groupID,
			"timestamp": message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"senderInfo": map[string]interface{}{
				"id":             user.ID,
				"username":       user.Username,
				"fullName":       user.FullName,
				"profilePicture": user.ProfilePicture,
			},
		},
	}

	// Serialize the response message
	data, err := json.Marshal(map[string]interface{}{
		"type":    "new_message",
		"payload": responseMsg,
	})
	if err == nil {
		// Broadcast to the room via WebSocket
		h.Hub.Broadcast <- &websocket.Broadcast{
			RoomID:  roomID,
			Message: data,
			Sender:  nil, // No specific sender client since this is from HTTP API
		}
	}

	utils.RespondWithSuccess(w, http.StatusCreated, "Message sent successfully", map[string]interface{}{
		"message": message,
	})
}
