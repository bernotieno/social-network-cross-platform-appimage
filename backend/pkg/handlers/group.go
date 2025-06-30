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

// GetGroups handles retrieving a list of groups
func (h *Handler) GetGroups(w http.ResponseWriter, r *http.Request) {
	// Get current user ID from context (authenticated user required)
	currentUserID, err := middleware.GetUserID(r)
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
	var getErr error

	if query != "" {
		groups, getErr = h.GroupService.SearchGroups(query, currentUserID, limit, offset)
	} else {
		groups, getErr = h.GroupService.GetGroups(currentUserID, limit, offset)
	}

	if getErr != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get groups")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Groups retrieved successfully", map[string]interface{}{
		"groups": groups,
	})
}

// GetGroup handles retrieving a group by ID
func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	// Get group ID from URL
	vars := mux.Vars(r)
	groupID := vars["id"]

	// Get current user ID from context (authenticated user required)
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

		// Save cover photo
		coverPhotoPath, err := utils.SaveImage(file, header, "groups")
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		group.CoverPhoto = coverPhotoPath
	}

	// Save group
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
	group.IsJoined = true
	group.IsAdmin = true
	group.MembersCount = 1

	utils.RespondWithSuccess(w, http.StatusCreated, "Group created successfully", map[string]interface{}{
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

	// Get group
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
		privacy = string(models.GroupPrivacyPublic)
	}

	// Update group fields
	group.Name = name
	group.Description = description
	group.Privacy = models.GroupPrivacy(privacy)

	// Check if cover photo was uploaded
	file, header, err := r.FormFile("coverPhoto")
	if err == nil {
		defer file.Close()

		// Delete old cover photo if exists
		if group.CoverPhoto != "" {
			utils.DeleteImage(group.CoverPhoto)
		}

		// Save new cover photo
		coverPhotoPath, err := utils.SaveImage(file, header, "groups")
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		group.CoverPhoto = coverPhotoPath
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

	// Create group member
	member := &models.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    models.GroupMemberRoleMember,
	}

	// Set status based on group privacy
	if group.Privacy == models.GroupPrivacyPublic {
		member.Status = models.GroupMemberStatusAccepted
	} else {
		member.Status = models.GroupMemberStatusPending
	}

	// Save member
	if err := h.GroupMemberService.Create(member); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to join group")
		return
	}

	// If private group, create notification for group creator
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

	utils.RespondWithSuccess(w, http.StatusOK, "Join request sent successfully", map[string]interface{}{
		"status": member.Status,
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

	// Check if user is the creator
	if group.CreatorID == userID {
		utils.RespondWithError(w, http.StatusForbidden, "Group creator cannot leave the group. Delete the group instead.")
		return
	}

	// Check if user is a member
	if !group.IsJoined {
		utils.RespondWithError(w, http.StatusBadRequest, "Not a member of this group")
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

	// Get group
	group, err := h.GroupService.GetByID(groupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Group not found")
		return
	}

	// Check if user is a member (for private groups)
	if group.Privacy == models.GroupPrivacyPrivate && !group.IsJoined {
		utils.RespondWithError(w, http.StatusForbidden, "Not authorized to view members of this private group")
		return
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

	// Promote member
	if err := h.GroupMemberService.PromoteToAdmin(groupID, memberID, userID); err != nil {
		if err.Error() == "only the group creator can promote members" {
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to promote member")
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Member promoted to admin successfully", nil)
}

// DemoteGroupMember handles demoting a group admin to regular member
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

	// Demote member
	if err := h.GroupMemberService.DemoteFromAdmin(groupID, memberID, userID); err != nil {
		if err.Error() == "only the group creator can demote admins" {
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to demote admin")
		}
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
		if err.Error() == "not authorized to remove members from this group" {
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to remove member")
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Member removed successfully", nil)
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

	// Get the member record
	member, err := h.GroupMemberService.GetByGroupAndUser(groupID, req.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Join request not found")
		return
	}

	// Check if the request is pending
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
	if err == nil {
		// Create notification for the user
		notification := &models.Notification{
			UserID:   req.UserID,
			SenderID: userID,
			Type:     models.NotificationTypeGroupJoinApproved,
			Content:  "approved your request to join the group",
			Data:     `{"groupId":"` + groupID + `","groupName":"` + group.Name + `"}`,
		}

		if err := h.NotificationService.Create(notification); err != nil {
			// Log error but don't fail the request
			// TODO: Add proper logging
		}

		// Update the original join request notification status
		if err := h.NotificationService.UpdateStatusByTypeAndSender(userID, req.UserID, models.NotificationTypeGroupJoinRequest, models.NotificationStatusApproved); err != nil {
			// Log error but don't fail the request
			// TODO: Add proper logging
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

	// Get the member record
	member, err := h.GroupMemberService.GetByGroupAndUser(groupID, req.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Join request not found")
		return
	}

	// Check if the request is pending
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
	if err == nil {
		// Create notification for the user
		notification := &models.Notification{
			UserID:   req.UserID,
			SenderID: userID,
			Type:     models.NotificationTypeGroupJoinRejected,
			Content:  "declined your request to join the group",
			Data:     `{"groupId":"` + groupID + `","groupName":"` + group.Name + `"}`,
		}

		if err := h.NotificationService.Create(notification); err != nil {
			// Log error but don't fail the request
			// TODO: Add proper logging
		}

		// Update the original join request notification status
		if err := h.NotificationService.UpdateStatusByTypeAndSender(userID, req.UserID, models.NotificationTypeGroupJoinRequest, models.NotificationStatusRejected); err != nil {
			// Log error but don't fail the request
			// TODO: Add proper logging
		}
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Join request rejected successfully", nil)
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
		utils.RespondWithError(w, http.StatusForbidden, "Only group members can invite others")
		return
	}

	// Check if invited user is already a member
	isAlreadyMember, err := h.GroupMemberService.IsGroupMember(groupID, req.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check if user is already a member")
		return
	}

	if isAlreadyMember {
		utils.RespondWithError(w, http.StatusConflict, "User is already a member of this group")
		return
	}

	// Check if there's already a pending invitation
	member, err := h.GroupMemberService.GetByGroupAndUser(groupID, req.UserID)
	if err == nil {
		// Member record exists
		if member.Status == models.GroupMemberStatusInvited {
			utils.RespondWithError(w, http.StatusConflict, "User has already been invited to this group")
			return
		}
		if member.Status == models.GroupMemberStatusPending {
			utils.RespondWithError(w, http.StatusConflict, "User has already requested to join this group")
			return
		}
	}

	// Create invitation
	invitation := &models.GroupMember{
		GroupID: groupID,
		UserID:  req.UserID,
		Role:    models.GroupMemberRoleMember,
		Status:  models.GroupMemberStatusInvited,
	}

	if err := h.GroupMemberService.Create(invitation); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create invitation")
		return
	}

	// Get group for notification
	group, err := h.GroupService.GetByID(groupID, userID)
	if err == nil {
		// Create notification for the invited user
		notification := &models.Notification{
			UserID:   req.UserID,
			SenderID: userID,
			Type:     models.NotificationTypeGroupInvite,
			Content:  "invited you to join a group",
			Data:     `{"groupId":"` + groupID + `","groupName":"` + group.Name + `"}`,
		}

		if err := h.NotificationService.Create(notification); err != nil {
			// Log error but don't fail the request
			// TODO: Add proper logging
		}
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Invitation sent successfully", nil)
}

// RespondToGroupInvitation handles accepting or rejecting a group invitation
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

	// Get notification
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
	var notificationData struct {
		GroupID   string `json:"groupId"`
		GroupName string `json:"groupName"`
	}
	if err := json.Unmarshal([]byte(notification.Data), &notificationData); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to parse notification data")
		return
	}

	// Get the invitation
	invitation, err := h.GroupMemberService.GetByGroupAndUser(notificationData.GroupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Invitation not found")
		return
	}

	// Check if invitation is still valid
	if invitation.Status != models.GroupMemberStatusInvited {
		utils.RespondWithError(w, http.StatusBadRequest, "Invitation is no longer valid")
		return
	}

	// Update invitation status
	if req.Accept {
		invitation.Status = models.GroupMemberStatusAccepted
	} else {
		invitation.Status = models.GroupMemberStatusRejected
	}

	if err := h.GroupMemberService.Update(invitation); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update invitation")
		return
	}

	// Update notification status
	var notificationStatus models.NotificationStatus
	if req.Accept {
		notificationStatus = models.NotificationStatusAccepted
	} else {
		notificationStatus = models.NotificationStatusDeclined
	}

	if err := h.NotificationService.UpdateStatus(notificationID, userID, notificationStatus); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Invitation response recorded successfully", map[string]interface{}{
		"accepted": req.Accept,
	})
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Only group members can create posts")
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
		if err.Error() == "not authorized to delete this post" {
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete post")
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Post deleted successfully", nil)
}

// LikeGroupPost handles liking a group post
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Only group members can like posts")
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
	if err == nil && post.UserID != userID {
		// Create notification for post owner
		notification := &models.Notification{
			UserID:   post.UserID,
			SenderID: userID,
			Type:     models.NotificationTypePostLike,
			Content:  "liked your post in a group",
			Data:     `{"postId":"` + postID + `","groupId":"` + groupID + `"}`,
		}

		if err := h.NotificationService.Create(notification); err != nil {
			// Log error but don't fail the request
			// TODO: Add proper logging
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

// UnlikeGroupPost handles unliking a group post
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

// GetGroupPostComments handles retrieving comments for a group post
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Only group members can view comments")
		return
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

// AddGroupPostComment handles adding a comment to a group post
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Only group members can comment on posts")
		return
	}

	var content string
	var imagePath string

	// Check content type to determine how to parse the request
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		// Parse JSON request body (for text-only comments)
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		content = req.Content
	} else {
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
	}

	// Validate content (allow empty content if image is provided)
	if content == "" && imagePath == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Content or image is required")
		return
	}

	// Check if post exists
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
		notification := &models.Notification{
			UserID:   post.UserID,
			SenderID: userID,
			Type:     models.NotificationTypePostComment,
			Content:  "commented on your post in a group",
			Data:     `{"postId":"` + postID + `","groupId":"` + groupID + `","comment":"` + content + `"}`,
		}

		if err := h.NotificationService.Create(notification); err != nil {
			// Log error but don't fail the request
			// TODO: Add proper logging
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

// DeleteGroupPostComment handles deleting a comment from a group post
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Only group members can delete comments")
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Only group members can create events")
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

	// Get group for response
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
		members, err := h.GroupMemberService.GetMembers(groupID, 1000, 0)
		if err != nil {
			return
		}

		// Create notifications for all members except the creator
		for _, member := range members {
			memberID := member.UserID
			if memberID != userID {
				notification := &models.Notification{
					UserID:   memberID,
					SenderID: userID,
					Type:     models.NotificationTypeGroupEventCreated,
					Content:  "created a new event in a group",
					Data:     `{"groupId":"` + groupID + `","groupName":"` + group.Name + `","eventId":"` + event.ID + `","eventTitle":"` + event.Title + `","eventStartTime":"` + event.StartTime.Format(time.RFC3339) + `","eventLocation":"` + event.Location + `"}`,
				}

				if err := h.NotificationService.Create(notification); err != nil {
					// Log error but continue
					// TODO: Add proper logging
				}
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get events")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Events retrieved successfully", map[string]interface{}{
		"events": events,
	})
}

// UpdateGroupEvent handles updating an event
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

	// Get event
	event, err := h.EventService.GetByID(eventID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Check if user is the creator
	if event.CreatorID != userID {
		// Check if user is a group admin
		isAdmin, err := h.GroupMemberService.IsGroupAdmin(event.GroupID, userID)
		if err != nil || !isAdmin {
			utils.RespondWithError(w, http.StatusForbidden, "Only the event creator or group admins can update the event")
			return
		}
	}

	// Update event fields
	event.Title = req.Title
	event.Description = req.Description
	event.Location = req.Location
	event.StartTime = startTime
	event.EndTime = endTime

	// Save changes
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

// DeleteGroupEvent handles deleting an event
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
		if err.Error() == "not authorized to delete this event" {
			utils.RespondWithError(w, http.StatusForbidden, err.Error())
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete event")
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Event deleted successfully", nil)
}

// RespondToEvent handles responding to an event (going, maybe, not going)
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
		Response string `json:"response"` // going, maybe, not_going
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate response
	if req.Response != "going" && req.Response != "maybe" && req.Response != "not_going" {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid response. Must be 'going', 'maybe', or 'not_going'")
		return
	}

	// Get event
	event, err := h.EventService.GetByID(eventID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Check if user is a member of the group
	isMember, err := h.GroupMemberService.IsGroupMember(event.GroupID, userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Only group members can respond to events")
		return
	}

	// Create or update response
	response := &models.EventResponse{
		EventID:  eventID,
		UserID:   userID,
		Response: models.EventResponseType(req.Response),
	}

	if err := h.EventResponseService.Create(response); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save response")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Response saved successfully", map[string]interface{}{
		"response": req.Response,
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Only group members can view messages")
		return
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

	// Get messages
	messages, err := h.MessageService.GetGroupMessages(groupID, limit, offset)
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check group membership")
		return
	}

	if !isMember {
		utils.RespondWithError(w, http.StatusForbidden, "Only group members can send messages")
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
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create message")
		return
	}

	// Get sender information for the response
	sender, err := h.UserService.GetByID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get sender info")
		return
	}
	sender.Password = ""

	// Create sender info for WebSocket broadcast
	senderInfo := map[string]interface{}{
		"id":             userID,
		"fullName":       sender.FullName,
		"username":       sender.Username,
		"profilePicture": sender.ProfilePicture,
	}

	// Broadcast message via WebSocket
	roomID := "group-" + groupID
	responseMsg := map[string]interface{}{
		"roomId": roomID,
		"message": map[string]interface{}{
			"id":         message.ID,
			"content":    req.Content,
			"sender":     userID,
			"senderInfo": senderInfo,
			"timestamp":  message.CreatedAt.Format(time.RFC3339),
			"groupId":    groupID,
		},
	}

	// Serialize the response message
	data, err := json.Marshal(map[string]interface{}{
		"type":    "new_message",
		"payload": responseMsg,
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to serialize message")
		return
	}

	// Broadcast to the room via WebSocket
	h.Hub.Broadcast <- &websocket.Broadcast{
		RoomID:  roomID,
		Message: data,
		Sender:  nil, // No specific sender client since this is from HTTP API
	}

	utils.RespondWithSuccess(w, http.StatusCreated, "Message sent successfully", map[string]interface{}{
		"message": map[string]interface{}{
			"id":        message.ID,
			"content":   message.Content,
			"senderId":  message.SenderID,
			"groupId":   message.GroupID,
			"createdAt": message.CreatedAt,
			"sender":    sender,
		},
	})
}