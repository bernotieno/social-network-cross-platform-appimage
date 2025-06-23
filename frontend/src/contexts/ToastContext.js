'use client';

import React, { createContext, useContext, useState, useCallback } from 'react';

const ToastContext = createContext();

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};

export const ToastProvider = ({ children }) => {
  const [toasts, setToasts] = useState([]);

  const removeToast = useCallback((id) => {
    setToasts(prev => prev.filter(toast => toast.id !== id));
  }, []);

  const addToast = useCallback((toast) => {
    const id = Date.now() + Math.random();
    const newToast = {
      id,
      type: 'info',
      duration: 5000,
      ...toast,
    };

    setToasts(prev => [...prev, newToast]);

    // Auto remove toast after duration
    if (newToast.duration > 0) {
      setTimeout(() => {
        setToasts(prev => prev.filter(toast => toast.id !== id));
      }, newToast.duration);
    }

    return id;
  }, []);

  const showSuccess = useCallback((message, options = {}) => {
    return addToast({
      type: 'success',
      message,
      ...options,
    });
  }, [addToast]);

  const showError = useCallback((message, options = {}) => {
    return addToast({
      type: 'error',
      message,
      duration: 7000, // Longer duration for errors
      ...options,
    });
  }, [addToast]);

  const showWarning = useCallback((message, options = {}) => {
    return addToast({
      type: 'warning',
      message,
      ...options,
    });
  }, [addToast]);

  const showInfo = useCallback((message, options = {}) => {
    return addToast({
      type: 'info',
      message,
      ...options,
    });
  }, [addToast]);

  const showNotification = useCallback((notification, options = {}) => {
    // Format notification for toast display
    let message = '';
    let title = '';

    if (notification.sender) {
      title = notification.sender.fullName || notification.sender.username;
    }

    // Parse notification data
    let notificationData = {};
    if (notification.data) {
      try {
        notificationData = typeof notification.data === 'string'
          ? JSON.parse(notification.data)
          : notification.data;
      } catch (error) {
        console.error('Error parsing notification data:', error);
      }
    }

    // Generate message based on notification type
    switch (notification.type) {
      case 'follow_request':
        message = 'requested to follow you';
        break;
      case 'follow_accepted':
        message = 'accepted your follow request';
        break;
      case 'new_follower':
        message = 'started following you';
        break;
      case 'post_like':
        const postContent = notificationData.postContent;
        const groupName = notificationData.groupName;
        message = `liked your ${groupName ? `post in "${groupName}"` : 'post'}${postContent ? `: "${postContent}"` : ''}`;
        break;
      case 'post_comment':
        const comment = notificationData.comment;
        const postContentForComment = notificationData.postContent;
        const groupNameForComment = notificationData.groupName;
        message = `commented on your ${groupNameForComment ? `post in "${groupNameForComment}"` : 'post'}${postContentForComment ? ` "${postContentForComment}"` : ''}`;
        break;
      case 'group_invite':
        message = `invited you to join "${notificationData.groupName || 'a group'}"`;
        break;
      case 'group_join_request':
        message = `requested to join your group "${notificationData.groupName || 'Unknown Group'}"`;
        break;
      case 'group_join_approved':
        message = `approved your request to join "${notificationData.groupName || 'a group'}"`;
        break;
      case 'group_join_rejected':
        message = `declined your request to join "${notificationData.groupName || 'a group'}"`;
        break;
      case 'event_invite':
        const eventTitle = notificationData.eventTitle || notificationData.eventName || 'an event';
        message = `invited you to "${eventTitle}"`;
        break;
      case 'group_event_created':
        const eventTitleCreated = notificationData.eventTitle || 'Untitled Event';
        const eventGroupName = notificationData.groupName;
        message = `created a new event "${eventTitleCreated}"${eventGroupName ? ` in "${eventGroupName}"` : ''}`;
        break;
      default:
        message = notification.content || 'sent you a notification';
    }

    return addToast({
      type: 'notification',
      title,
      message,
      duration: 6000,
      notification,
      ...options,
    });
  }, [addToast]);

  const clearAll = useCallback(() => {
    setToasts([]);
  }, []);

  const value = {
    toasts,
    addToast,
    removeToast,
    showSuccess,
    showError,
    showWarning,
    showInfo,
    showNotification,
    clearAll,
  };

  return (
    <ToastContext.Provider value={value}>
      {children}
    </ToastContext.Provider>
  );
};

export default ToastProvider;
