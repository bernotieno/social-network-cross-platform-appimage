package handlers

import (
	"net/http"
	"strconv"

	"github.com/bernaotieno/social-network/backend/pkg/middleware"
	"github.com/bernaotieno/social-network/backend/pkg/utils"
	"github.com/gorilla/mux"
)

// GetNotifications handles retrieving notifications for a user
func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
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

	// Get notifications
	notifications, err := h.NotificationService.GetByUser(userID, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get notifications")
		return
	}

	// Get unread count
	unreadCount, err := h.NotificationService.GetUnreadCount(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get unread count")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Notifications retrieved successfully", map[string]interface{}{
		"notifications": notifications,
		"unreadCount":   unreadCount,
	})
}

// MarkNotificationAsRead handles marking a notification as read
func (h *Handler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get notification ID from URL
	vars := mux.Vars(r)
	notificationID := vars["id"]

	// Mark notification as read
	if err := h.NotificationService.MarkAsRead(notificationID, userID); err != nil {
		if err.Error() == "notification not found or not authorized" {
			utils.RespondWithError(w, http.StatusNotFound, "Notification not found")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to mark notification as read")
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Notification marked as read", nil)
}

// MarkAllNotificationsAsRead handles marking all notifications as read
func (h *Handler) MarkAllNotificationsAsRead(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Mark all notifications as read
	if err := h.NotificationService.MarkAllAsRead(userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to mark all notifications as read")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "All notifications marked as read", nil)
}

// DeleteNotification handles deleting a specific notification
func (h *Handler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get notification ID from URL
	vars := mux.Vars(r)
	notificationID := vars["id"]

	// Delete notification
	if err := h.NotificationService.Delete(notificationID, userID); err != nil {
		if err.Error() == "notification not found or not authorized" {
			utils.RespondWithError(w, http.StatusNotFound, "Notification not found")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete notification")
		}
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "Notification deleted successfully", nil)
}

// DeleteAllNotifications handles deleting all notifications for a user
func (h *Handler) DeleteAllNotifications(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Delete all notifications
	if err := h.NotificationService.DeleteAll(userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete all notifications")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, "All notifications deleted successfully", nil)
}
