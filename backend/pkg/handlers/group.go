package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bernaotieno/social-network/backend/pkg/middleware"
	"github.com/bernaotieno/social-network/backend/pkg/models"
	"github.com/bernaotieno/social-network/backend/pkg/utils"
	"github.com/bernaotieno/social-network/backend/pkg/websocket"
	"github.com/gorilla/mux"
)

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
		utils.RespondWithError(w, http.StatusBadRequest, "Group name is required")
		return
	}

	// Validate privacy
	if privacy != string(models.GroupPrivacyPublic) && privacy != string(models.GroupPrivacyPrivate) {
		privacy = string(models.GroupPrivacyPublic) // Default to public
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

		// Save cover photo
		coverPhotoPath, err := utils.SaveImage(file, header, "groups")
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		group.CoverPhoto = coverPhotoPath
	}

	// Create group
	if err := h.GroupService.Create(group); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	// Get creator for response
	creator, err := h.UserService.GetByID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get creator")
		return
	}
	creator.Password = ""

	// Add creator to group for response
	group.Creator = creator
	group.MembersCount = 1 // Creator is the first member
	group.IsJoined = true  // Creator is automatically joined
	group.IsAdmin = true   // Creator is automatically admin

	utils.RespondWithSuccess(w, http.StatusCreated, "Group created successfully", map[string]interface{}{
		"group": group,
	})
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
	var groups []*models.Group
	var getGroupsErr error

	if query != "" {
		groups, getGroupsErr = h.GroupService.SearchGroups(query, userID, limit, offset)
	} else {
		groups, getGroupsErr = h.GroupService.GetGroups(userID, limit, offset)
	}

	if getGroupsErr != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get groups")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Groups retrieved successfully", map[string]interface{}{
		"groups": groups,
	})
}

// GetGroup handles retrieving a group by ID
func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
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

	// Get group to check ownership
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Group not found")
		return
	}

	// Check if user is the creator
	if group.CreatorID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Only the group creator can update the group")
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
		utils.RespondWithError(w, http.StatusBadRequest, "Group name is required")
		return
	}

	// Validate privacy
	if privacy != string(models.GroupPrivacyPublic) && privacy != string(models.GroupPrivacyPrivate) {
		privacy = string(group.Privacy) // Keep current privacy if invalid
	}

	// Update group fields
	group.Name = name
	group.Description = description
	group.Privacy = models.GroupPrivacy(privacy)

	// Check if cover photo was uploaded
	file, header, err := r.FormFile("coverPhoto")
	if err == nil {
		defer file.Close()

		// Save cover photo
		coverPhotoPath, err := utils.SaveImage(file, header, "groups")
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Delete old cover photo if exists
		if group.CoverPhoto != "" {
			utils.DeleteImage(group.CoverPhoto)
		}

		group.CoverPhoto = coverPhotoPath
	}

	// Update group
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
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
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

	// Get group to check privacy
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Group not found")
		return
	}

	// Check if user is already a member
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if isMember {
		utils.RespondWithError(w, http.StatusConflict, "You are already a member of this group")
		return
	}

	// Check if user already has a pending request
	member, err := h.GroupMemberService.GetByGroupAndUser(groupID, userID)
	if err == nil && member.Status == models.GroupMemberStatusPending {
		utils.RespondWithError(w, http.StatusConflict, "You already have a pending request to join this group")
		return
	}

	// Determine initial status based on group privacy
	status := models.GroupMemberStatusAccepted
	if group.Privacy == models.GroupPrivacyPrivate {
		status = models.GroupMemberStatusPending
	}

	// Create group member
	member = &models.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    models.GroupMemberRoleMember,
		Status:  status,
	}

	if err := h.GroupMemberService.Create(member); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to join group")
		return
	}

	// If private group, create notification for group creator
	if group.Privacy == models.GroupPrivacyPrivate {
		// Create notification data
		notificationData := map[string]interface{}{
			"groupId":   groupID,
			"groupName": group.Name,
		}
		dataJSON, _ := json.Marshal(notificationData)

		// Create notification
		notification := &models.Notification{
			UserID:   group.CreatorID,
			SenderID: userID,
			Type:     models.NotificationTypeGroupJoinRequest,
			Content:  "requested to join your group",
			Data:     string(dataJSON),
		}

		if err := h.NotificationService.Create(notification); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to create notification: %v", err)
		}

		utils.RespondWithSuccess(w, http.StatusOK, "Join request sent successfully", map[string]interface{}{
			"status": "pending",
		})
	} else {
		utils.RespondWithSuccess(w, http.StatusOK, "Joined group successfully", map[string]interface{}{
			"status": "accepted",
		})
	}
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

	// Get group to check if user is the creator
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Group not found")
		return
	}

	// Check if user is the creator
	if group.CreatorID == userID {
		utils.RespondWithError(w, http.StatusForbidden, "Group creator cannot leave the group. Delete the group instead.")
		return
	}

	// Leave group
	if err := h.GroupMemberService.Delete(groupID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to leave group")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Left group successfully", nil)
}

// GetGroupMembers handles retrieving members of a group
func (h *Handler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		// Check if the group is public
		group, err := h.GroupService.GetByID(groupID, userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Group not found")
			return
		}

		if group.Privacy == models.GroupPrivacyPrivate {
			utils.RespondWithError(w, http.StatusForbidden, "You must be a member to view this group's members")
			return
		}
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set default values
	limit := 50
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

	// Get members
	members, err := h.GroupMemberService.GetMembers(groupID, limit, offset)
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
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

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
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Parse request body
	var req struct {
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if user is an admin of the group
	isAdmin, err := h.GroupMemberService.IsGroupAdmin(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check admin status")
		return
	}

	if !isAdmin {
		utils.RespondWithError(w, http.StatusForbidden, "Only group admins can approve join requests")
		return
	}

	// Get the member to check status
	member, err := h.GroupMemberService.GetByGroupAndUser(groupID, req.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Join request not found")
		return
	}

	if member.Status != models.GroupMemberStatusPending {
		utils.RespondWithError(w, http.StatusBadRequest, "Join request is not pending")
		return
	}

	// Update member status
	member.Status = models.GroupMemberStatusAccepted
	if err := h.GroupMemberService.Update(member); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to approve join request")
		return
	}

	// Get group for notification
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		// Log error but continue
		log.Printf("Failed to get group for notification: %v", err)
	} else {
		// Create notification data
		notificationData := map[string]interface{}{
			"groupId":   groupID,
			"groupName": group.Name,
		}
		dataJSON, _ := json.Marshal(notificationData)

		// Create notification for the user
		notification := &models.Notification{
			UserID:   req.UserID,
			SenderID: userID,
			Type:     models.NotificationTypeGroupJoinApproved,
			Content:  "approved your request to join the group",
			Data:     string(dataJSON),
		}

		if err := h.NotificationService.Create(notification); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to create notification: %v", err)
		}

		// Update the original join request notification status
		if err := h.NotificationService.UpdateStatusByTypeAndSender(userID, req.UserID, models.NotificationTypeGroupJoinRequest, models.NotificationStatusApproved); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to update notification status: %v", err)
		}
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Join request approved successfully", nil)
}

// RejectJoinRequest handles rejecting a join request
func (h *Handler) RejectJoinRequest(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Parse request body
	var req struct {
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if user is an admin of the group
	isAdmin, err := h.GroupMemberService.IsGroupAdmin(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check admin status")
		return
	}

	if !isAdmin {
		utils.RespondWithError(w, http.StatusForbidden, "Only group admins can reject join requests")
		return
	}

	// Get the member to check status
	member, err := h.GroupMemberService.GetByGroupAndUser(groupID, req.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Join request not found")
		return
	}

	if member.Status != models.GroupMemberStatusPending {
		utils.RespondWithError(w, http.StatusBadRequest, "Join request is not pending")
		return
	}

	// Update member status
	member.Status = models.GroupMemberStatusRejected
	if err := h.GroupMemberService.Update(member); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to reject join request")
		return
	}

	// Get group for notification
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		// Log error but continue
		log.Printf("Failed to get group for notification: %v", err)
	} else {
		// Create notification data
		notificationData := map[string]interface{}{
			"groupId":   groupID,
			"groupName": group.Name,
		}
		dataJSON, _ := json.Marshal(notificationData)

		// Create notification for the user
		notification := &models.Notification{
			UserID:   req.UserID,
			SenderID: userID,
			Type:     models.NotificationTypeGroupJoinRejected,
			Content:  "declined your request to join the group",
			Data:     string(dataJSON),
		}

		if err := h.NotificationService.Create(notification); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to create notification: %v", err)
		}

		// Update the original join request notification status
		if err := h.NotificationService.UpdateStatusByTypeAndSender(userID, req.UserID, models.NotificationTypeGroupJoinRequest, models.NotificationStatusRejected); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to update notification status: %v", err)
		}
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Join request rejected successfully", nil)
}

// PromoteGroupMember handles promoting a group member to admin
func (h *Handler) PromoteGroupMember(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID and member ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]
	memberID := vars["memberId"]

	// Check if user is the group creator
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Group not found")
		return
	}

	if group.CreatorID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Only the group creator can promote members")
		return
	}

	// Promote member
	if err := h.GroupMemberService.PromoteToAdmin(groupID, memberID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Member promoted to admin successfully", nil)
}

// DemoteGroupMember handles demoting a group admin to member
func (h *Handler) DemoteGroupMember(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID and member ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]
	memberID := vars["memberId"]

	// Check if user is the group creator
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Group not found")
		return
	}

	if group.CreatorID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Only the group creator can demote admins")
		return
	}

	// Demote admin
	if err := h.GroupMemberService.DemoteFromAdmin(groupID, memberID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Admin demoted to member successfully", nil)
}

// RemoveGroupMember handles removing a member from a group
func (h *Handler) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID and member ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]
	memberID := vars["memberId"]

	// Remove member
	if err := h.GroupMemberService.RemoveMember(groupID, memberID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Member removed successfully", nil)
}

// InviteToGroup handles inviting a user to a group
func (h *Handler) InviteToGroup(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Parse request body
	var req struct {
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "You must be a member to invite others to this group")
		return
	}

	// Check if invited user is already a member
	isInvitedUserMember, err := h.GroupMemberService.IsGroupMember(groupID, req.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check if invited user is already a member")
		return
	}

	if isInvitedUserMember {
		utils.RespondWithError(w, http.StatusConflict, "User is already a member of this group")
		return
	}

	// Check if invited user already has a pending request or invitation
	member, err := h.GroupMemberService.GetByGroupAndUser(groupID, req.UserID)
	if err == nil {
		if member.Status == models.GroupMemberStatusPending {
			utils.RespondWithError(w, http.StatusConflict, "User already has a pending request to join this group")
			return
		}
		if member.Status == models.GroupMemberStatusInvited {
			utils.RespondWithError(w, http.StatusConflict, "User has already been invited to this group")
			return
		}
	}

	// Create group member with invited status
	member = &models.GroupMember{
		GroupID: groupID,
		UserID:  req.UserID,
		Role:    models.GroupMemberRoleMember,
		Status:  models.GroupMemberStatusInvited,
	}

	if err := h.GroupMemberService.Create(member); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to invite user")
		return
	}

	// Get group for notification
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		// Log error but continue
		log.Printf("Failed to get group for notification: %v", err)
	} else {
		// Create notification data
		notificationData := map[string]interface{}{
			"groupId":   groupID,
			"groupName": group.Name,
		}
		dataJSON, _ := json.Marshal(notificationData)

		// Create notification for the invited user
		notification := &models.Notification{
			UserID:   req.UserID,
			SenderID: userID,
			Type:     models.NotificationTypeGroupInvite,
			Content:  "invited you to join the group",
			Data:     string(dataJSON),
		}

		if err := h.NotificationService.Create(notification); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to create notification: %v", err)
		}
	}

	utils.RespondWithSuccess(w, http.StatusOK, "User invited successfully", nil)
}

// RespondToGroupInvitation handles responding to a group invitation
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

	// Get notification to extract group ID
	notification, err := h.NotificationService.GetByID(notificationID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Notification not found")
		return
	}

	// Check if notification is for the current user
	if notification.UserID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "Not authorized to respond to this invitation")
		return
	}

	// Check if notification is a group invitation
	if notification.Type != models.NotificationTypeGroupInvite {
		utils.RespondWithError(w, http.StatusBadRequest, "Not a group invitation")
		return
	}

	// Parse notification data to get group ID
	var notificationData map[string]interface{}
	if err := json.Unmarshal([]byte(notification.Data), &notificationData); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to parse notification data")
		return
	}

	groupID, ok := notificationData["groupId"].(string)
	if !ok {
		utils.RespondWithError(w, http.StatusInternalServerError, "Invalid notification data")
		return
	}

	// Get the member to check status
	member, err := h.GroupMemberService.GetByGroupAndUser(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Invitation not found")
		return
	}

	if member.Status != models.GroupMemberStatusInvited {
		utils.RespondWithError(w, http.StatusBadRequest, "Not an invitation")
		return
	}

	if req.Accept {
		// Accept invitation
		member.Status = models.GroupMemberStatusAccepted
		if err := h.GroupMemberService.Update(member); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to accept invitation")
			return
		}

		// Update notification status
		if err := h.NotificationService.UpdateStatus(notificationID, userID, models.NotificationStatusAccepted); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to update notification status: %v", err)
		}

		utils.RespondWithSuccess(w, http.StatusOK, "Invitation accepted successfully", nil)
	} else {
		// Decline invitation
		if err := h.GroupMemberService.Delete(groupID, userID); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to decline invitation")
			return
		}

		// Update notification status
		if err := h.NotificationService.UpdateStatus(notificationID, userID, models.NotificationStatusDeclined); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to update notification status: %v", err)
		}

		utils.RespondWithSuccess(w, http.StatusOK, "Invitation declined successfully", nil)
	}
}

// GetGroupPosts handles retrieving posts for a group
func (h *Handler) GetGroupPosts(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

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
	posts, err := h.GroupPostService.GetByGroup(groupID, userID, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get group posts")
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "You must be a member to post in this group")
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

	// Create post
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

// DeleteGroupPost handles deleting a post from a group
func (h *Handler) DeleteGroupPost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID and post ID from URL
	vars := mux.Vars(r)
	groupID := vars["groupId"]
	postID := vars["postId"]

	// Delete post
	if err := h.GroupPostService.Delete(postID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Post deleted successfully", nil)
}

// LikeGroupPost handles liking a post in a group
func (h *Handler) LikeGroupPost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID and post ID from URL
	vars := mux.Vars(r)
	groupID := vars["groupId"]
	postID := vars["postId"]

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "You must be a member to like posts in this group")
		return
	}

	// Like post
	if err := h.LikeService.Create(postID, userID); err != nil {
		if err.Error() == "post already liked by user" {
			utils.RespondWithError(w, http.StatusConflict, "Post already liked")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to like post")
		}
		return
	}

	// Get post to get author ID
	post, err := h.GroupPostService.GetByID(postID, userID)
	if err != nil {
		// Log error but continue
		log.Printf("Failed to get post for notification: %v", err)
	} else if post.UserID != userID {
		// Create notification for post owner (if not the same user)
		// Prepare notification data with post content
		postContent := post.Content
		if len(postContent) > 50 {
			postContent = postContent[:50] + "..."
		}

		notificationData := map[string]interface{}{
			"postId":      postID,
			"postContent": postContent,
			"groupId":     groupID,
			"groupName":   post.Group.Name,
		}
		dataJSON, _ := json.Marshal(notificationData)

		notification := &models.Notification{
			UserID:   post.UserID,
			SenderID: userID,
			Type:     models.NotificationTypePostLike,
			Content:  "liked your post",
			Data:     string(dataJSON),
		}

		if err := h.NotificationService.Create(notification); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to create notification: %v", err)
		}
	}

	// Broadcast like event via WebSocket
	likeEvent := map[string]interface{}{
		"postId":  postID,
		"groupId": groupID,
		"userId":  userID,
		"action":  "like",
	}

	message := map[string]interface{}{
		"type":    "group_post_like",
		"payload": likeEvent,
	}

	messageData, _ := json.Marshal(message)

	// Broadcast to all connected clients
	h.Hub.Broadcast <- &websocket.Broadcast{
		RoomID:  "", // Broadcast to default room (all users)
		Message: messageData,
		Sender:  nil, // No specific sender for server events
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Post liked successfully", nil)
}

// UnlikeGroupPost handles unliking a post in a group
func (h *Handler) UnlikeGroupPost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID and post ID from URL
	vars := mux.Vars(r)
	groupID := vars["groupId"]
	postID := vars["postId"]

	// Unlike post
	if err := h.LikeService.Delete(postID, userID); err != nil {
		if err.Error() == "like not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Post not liked")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to unlike post")
		}
		return
	}

	// Broadcast unlike event via WebSocket
	unlikeEvent := map[string]interface{}{
		"postId":  postID,
		"groupId": groupID,
		"userId":  userID,
		"action":  "unlike",
	}

	message := map[string]interface{}{
		"type":    "group_post_like",
		"payload": unlikeEvent,
	}

	messageData, _ := json.Marshal(message)

	// Broadcast to all connected clients
	h.Hub.Broadcast <- &websocket.Broadcast{
		RoomID:  "", // Broadcast to default room (all users)
		Message: messageData,
		Sender:  nil, // No specific sender for server events
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Post unliked successfully", nil)
}

// GetGroupPostComments handles retrieving comments for a post in a group
func (h *Handler) GetGroupPostComments(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID and post ID from URL
	vars := mux.Vars(r)
	groupID := vars["groupId"]
	postID := vars["postId"]

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		// Check if the group is public
		group, err := h.GroupService.GetByID(groupID, userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Group not found")
			return
		}

		if group.Privacy == models.GroupPrivacyPrivate {
			utils.RespondWithError(w, http.StatusForbidden, "You must be a member to view comments in this group")
			return
		}
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set default values
	limit := 50
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

	// Get comments
	comments, err := h.CommentService.GetCommentsByPost(postID, userID, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get comments")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Comments retrieved successfully", map[string]interface{}{
		"comments": comments,
	})
}

// AddGroupPostComment handles adding a comment to a post in a group
func (h *Handler) AddGroupPostComment(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID and post ID from URL
	vars := mux.Vars(r)
	groupID := vars["groupId"]
	postID := vars["postId"]

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "You must be a member to comment on posts in this group")
		return
	}

	var content string
	var imagePath string

	// Check content type to determine how to parse the request
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Parse multipart form (for comments with images)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Failed to parse form")
			return
		}

		// Get content from form
		content = r.FormValue("content")

		// Check if image was uploaded
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			// Save image
			imagePath, err = utils.SaveImage(file, header, "comments")
			if err != nil {
				utils.RespondWithError(w, http.StatusBadRequest, err.Error())
				return
			}
		}
	} else {
		// Parse JSON request body (for text-only comments)
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		content = req.Content
	}

	// Validate content (allow empty content if image is provided)
	if content == "" && imagePath == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Content or image is required")
		return
	}

	// Check if post exists and user can view it
	post, err := h.GroupPostService.GetByID(postID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Post not found")
		return
	}

	// Create comment
	comment := &models.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: content,
		Image:   imagePath,
	}

	if err := h.CommentService.Create(comment); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	// Get user for response
	user, err := h.UserService.GetByID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}
	user.Password = ""

	// Add user to comment for response
	comment.Author = user

	// Create notification for post owner (if not the same user)
	if post.UserID != userID {
		// Get group for notification
		group, err := h.GroupService.GetByID(groupID, userID)
		if err != nil {
			// Log error but continue
			log.Printf("Failed to get group for notification: %v", err)
		} else {
			// Create notification data
			notificationData := map[string]interface{}{
				"postId":      postID,
				"comment":     content,
				"groupId":     groupID,
				"groupName":   group.Name,
				"postContent": post.Content,
			}
			dataJSON, _ := json.Marshal(notificationData)

			notification := &models.Notification{
				UserID:   post.UserID,
				SenderID: userID,
				Type:     models.NotificationTypePostComment,
				Content:  "commented on your post",
				Data:     string(dataJSON),
			}

			if err := h.NotificationService.Create(notification); err != nil {
				// Log error but don't fail the request
				log.Printf("Failed to create notification: %v", err)
			}
		}
	}

	// Broadcast new comment event via WebSocket
	newCommentEvent := map[string]interface{}{
		"postId":  postID,
		"groupId": groupID,
		"comment": comment,
	}

	message := map[string]interface{}{
		"type":    "group_post_comment",
		"payload": newCommentEvent,
	}

	messageData, _ := json.Marshal(message)

	// Broadcast to all connected clients
	h.Hub.Broadcast <- &websocket.Broadcast{
		RoomID:  "", // Broadcast to default room (all users)
		Message: messageData,
		Sender:  nil, // No specific sender for server events
	}

	utils.RespondWithSuccess(w, http.StatusCreated, "Comment added successfully", map[string]interface{}{
		"comment": comment,
	})
}

// DeleteGroupPostComment handles deleting a comment from a post in a group
func (h *Handler) DeleteGroupPostComment(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID, post ID, and comment ID from URL
	vars := mux.Vars(r)
	groupID := vars["groupId"]
	postID := vars["postId"]
	commentID := vars["commentId"]

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "You must be a member to delete comments in this group")
		return
	}

	// Delete comment
	if err := h.CommentService.Delete(commentID, userID); err != nil {
		if err.Error() == "comment not found or not authorized to delete" {
			utils.RespondWithError(w, http.StatusForbidden, "Not authorized to delete this comment")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete comment")
		}
		return
	}

	// Broadcast comment deletion event via WebSocket
	deleteCommentEvent := map[string]interface{}{
		"postId":    postID,
		"groupId":   groupID,
		"commentId": commentID,
	}

	message := map[string]interface{}{
		"type":    "group_post_comment_delete",
		"payload": deleteCommentEvent,
	}

	messageData, _ := json.Marshal(message)

	// Broadcast to all connected clients
	h.Hub.Broadcast <- &websocket.Broadcast{
		RoomID:  "", // Broadcast to default room (all users)
		Message: messageData,
		Sender:  nil, // No specific sender for server events
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Comment deleted successfully", nil)
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "You must be a member to create events in this group")
		return
	}

	// Parse request body
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Location    string `json:"location"`
		StartTime   string `json:"startTime"`
		EndTime     string `json:"endTime"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Title == "" || req.StartTime == "" || req.EndTime == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Title, start time, and end time are required")
		return
	}

	// Parse start and end times
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

	// Validate that end time is after start time
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		utils.RespondWithError(w, http.StatusBadRequest, "End time must be after start time")
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

	if err := h.EventService.Create(event); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create event")
		return
	}

	// Get creator for response
	creator, err := h.UserService.GetByID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get creator")
		return
	}
	creator.Password = ""

	// Add creator to event for response
	event.Creator = creator

	// Get group for response and notifications
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get group")
		return
	}

	// Add group to event for response
	event.Group = group

	// Create notifications for all group members
	go func() {
		// Get all group members
		members, err := h.GroupMemberService.GetMembers(groupID, 1000, 0) // Limit to 1000 members
		if err != nil {
			log.Printf("Failed to get group members for event notifications: %v", err)
			return
		}

		// Create notification data
		notificationData := map[string]interface{}{
			"eventId":      event.ID,
			"eventTitle":   event.Title,
			"groupId":      groupID,
			"groupName":    group.Name,
			"eventStartTime": event.StartTime.Format(time.RFC3339),
			"eventLocation":  event.Location,
		}
		dataJSON, _ := json.Marshal(notificationData)

		// Create notifications for all members except the creator
		var notifications []*models.Notification
		for _, member := range members {
			if member.UserID != userID { // Skip creator
				notification := &models.Notification{
					UserID:   member.UserID,
					SenderID: userID,
					Type:     models.NotificationTypeGroupEventCreated,
					Content:  "created a new event",
					Data:     string(dataJSON),
				}
				notifications = append(notifications, notification)
			}
		}

		// Create notifications in batch
		if len(notifications) > 0 {
			if err := h.NotificationService.CreateBatch(notifications); err != nil {
				log.Printf("Failed to create event notifications: %v", err)
			}
		}
	}()

	utils.RespondWithSuccess(w, http.StatusCreated, "Event created successfully", map[string]interface{}{
		"event": event,
	})
}

// GetGroupEvents handles retrieving events for a group
func (h *Handler) GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

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
	events, err := h.EventService.GetByGroup(groupID, userID, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get group events")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Group events retrieved successfully", map[string]interface{}{
		"events": events,
	})
}

// UpdateGroupEvent handles updating an event in a group
func (h *Handler) UpdateGroupEvent(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Location    string `json:"location"`
		StartTime   string `json:"startTime"`
		EndTime     string `json:"endTime"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Title == "" || req.StartTime == "" || req.EndTime == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Title, start time, and end time are required")
		return
	}

	// Parse start and end times
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

	// Validate that end time is after start time
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		utils.RespondWithError(w, http.StatusBadRequest, "End time must be after start time")
		return
	}

	// Get event to check ownership
	event, err := h.EventService.GetByID(eventID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Check if user is the event creator
	if event.CreatorID != userID {
		// Check if user is a group admin
		isAdmin, err := h.GroupMemberService.IsGroupAdmin(event.GroupID, userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check admin status")
			return
		}

		if !isAdmin {
			utils.RespondWithError(w, http.StatusForbidden, "Only the event creator or group admins can update this event")
			return
		}
	}

	// Update event fields
	event.Title = req.Title
	event.Description = req.Description
	event.Location = req.Location
	event.StartTime = startTime
	event.EndTime = endTime

	// Update event
	if err := h.EventService.Update(event); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update event")
		return
	}

	// Get creator for response
	creator, err := h.UserService.GetByID(event.CreatorID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get creator")
		return
	}
	creator.Password = ""

	// Add creator to event for response
	event.Creator = creator

	utils.RespondWithSuccess(w, http.StatusOK, "Event updated successfully", map[string]interface{}{
		"event": event,
	})
}

// DeleteGroupEvent handles deleting an event from a group
func (h *Handler) DeleteGroupEvent(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get event ID from URL
	vars := mux.Vars(r)
	eventID := vars["id"]

	// Delete event
	if err := h.EventService.Delete(eventID, userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Event deleted successfully", nil)
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
	var req struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate response
	if req.Response != string(models.EventResponseGoing) &&
		req.Response != string(models.EventResponseMaybe) &&
		req.Response != string(models.EventResponseNotGoing) {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid response. Must be 'going', 'maybe', or 'not_going'")
		return
	}

	// Get event to check if it exists and user can respond
	event, err := h.EventService.GetByID(eventID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(event.GroupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "You must be a member to respond to events in this group")
		return
	}

	// Create or update response
	response := &models.EventResponse{
		EventID:  eventID,
		UserID:   userID,
		Response: models.EventResponseType(req.Response),
	}

	if err := h.EventResponseService.Create(response); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to respond to event")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Response saved successfully", map[string]interface{}{
		"response": response,
	})
}

// GetGroupMessages handles retrieving messages for a group
func (h *Handler) GetGroupMessages(w http.ResponseWriter, r *http.Request) {
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "You must be a member to view messages in this group")
		return
	}

	// Get messages
	messages, err := h.MessageService.GetGroupMessages(groupID, 50, 0)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Messages retrieved successfully", map[string]interface{}{
		"messages": messages,
	})
}

// SendGroupMessage handles sending a message to a group
func (h *Handler) SendGroupMessage(w http.ResponseWriter, r *http.Request) {
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "You must be a member to send messages in this group")
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

	// Validate content
	if req.Content == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Create message
	message := &models.Message{
		SenderID: userID,
		GroupID:  groupID,
		Content:  req.Content,
	}

	if err := h.MessageService.Create(message); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	// Get sender for response
	sender, err := h.UserService.GetByID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get sender")
		return
	}
	sender.Password = ""

	// Add sender to message for response
	message.Sender = sender

	// Create sender info for WebSocket broadcast
	senderInfo := map[string]interface{}{
		"id":             sender.ID,
		"fullName":       sender.FullName,
		"username":       sender.Username,
		"profilePicture": sender.ProfilePicture,
	}

	// Broadcast message via WebSocket
	roomID := fmt.Sprintf("group-%s", groupID)
	responseMsg := map[string]interface{}{
		"roomId": roomID,
		"message": map[string]interface{}{
			"id":         message.ID,
			"content":    message.Content,
			"sender":     userID,
			"senderInfo": senderInfo,
			"timestamp":  message.CreatedAt.Format(time.RFC3339),
		},
	}

	// Serialize the response message
	data, err := json.Marshal(map[string]interface{}{
		"type":    "new_message",
		"payload": responseMsg,
	})
	if err != nil {
		log.Printf("Error marshaling WebSocket message: %v", err)
		// Continue even if WebSocket broadcast fails
	} else {
		// Broadcast to the room via WebSocket
		h.Hub.Broadcast <- &websocket.Broadcast{
			RoomID:  roomID,
			Message: data,
			Sender:  nil, // No specific sender client since this is from HTTP API
		}
		log.Printf("Message broadcasted via WebSocket to room %s", roomID)
	}

	utils.RespondWithSuccess(w, http.StatusCreated, "Message sent successfully", map[string]interface{}{
		"message": message,
	})
}