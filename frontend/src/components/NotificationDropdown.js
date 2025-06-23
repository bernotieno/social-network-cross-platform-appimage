'use client';

import { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { formatDistanceToNow } from 'date-fns';
import { getUserProfilePictureUrl, getFallbackAvatar } from '@/utils/images';
import { userAPI, groupAPI } from '@/utils/api';
import { useAlert } from '@/contexts/AlertContext';
import Button from '@/components/Button';
import styles from '@/styles/NotificationDropdown.module.css';

const NotificationDropdown = ({ 
  notifications, 
  isOpen, 
  onClose, 
  onMarkAsRead, 
  onRefresh 
}) => {
  const dropdownRef = useRef(null);
  const { showAlert } = useAlert();

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen, onClose]);

  // Get recent notifications (limit to 8)
  const recentNotifications = notifications.slice(0, 8);

  // Handle follow request response
  const handleFollowResponse = async (notification, accept) => {
    try {
      // Extract follow request ID from notification data
      let followRequestId = notification.id;

      if (notification.data) {
        try {
          const data = JSON.parse(notification.data);
          if (data.followRequestId) {
            followRequestId = data.followRequestId;
          }
        } catch (parseError) {
          console.warn('Failed to parse notification data:', parseError);
        }
      }

      // Call the API to respond to the follow request
      await userAPI.respondToFollowRequest(followRequestId, accept);

      // Mark notification as read and refresh notifications
      onMarkAsRead(notification.id);
      onRefresh();

      // Show success message
      showAlert({
        type: 'success',
        title: 'Success',
        message: `Follow request ${accept ? 'accepted' : 'declined'} successfully!`
      });
    } catch (error) {
      console.error('Error responding to follow request:', error);
      showAlert({
        type: 'error',
        title: 'Error',
        message: error.response?.data?.message || 'Failed to respond to follow request. Please try again.'
      });
    }
  };

  // Handle group invite response
  const handleGroupInviteResponse = async (notification, accept) => {
    try {
      await groupAPI.respondToGroupInvitation(notification.id, accept);

      // Mark notification as read and refresh notifications
      onMarkAsRead(notification.id);
      onRefresh();

      // Show success message
      showAlert({
        type: 'success',
        title: 'Success',
        message: `Group invite ${accept ? 'accepted' : 'declined'} successfully!`
      });
    } catch (error) {
      console.error('Error responding to group invitation:', error);
      showAlert({
        type: 'error',
        title: 'Error',
        message: error.response?.data?.message || 'Failed to respond to group invitation. Please try again.'
      });
    }
  };

  // Handle group join request response
  const handleGroupJoinResponse = async (notification, accept) => {
    try {
      // Parse the notification data to get group and user info
      let notificationData;
      try {
        notificationData = JSON.parse(notification.data);
      } catch (error) {
        console.error('Failed to parse notification data:', error);
        return;
      }

      const groupId = notificationData.groupId;
      const userId = notification.senderId;

      if (!groupId || !userId) {
        console.error('Missing group ID or user ID in notification data');
        return;
      }

      // Call the appropriate API endpoint
      if (accept) {
        await groupAPI.approveJoinRequest(groupId, userId);
      } else {
        await groupAPI.rejectJoinRequest(groupId, userId);
      }

      // Mark notification as read and refresh notifications
      onMarkAsRead(notification.id);
      onRefresh();

      // Show success message
      showAlert({
        type: 'success',
        title: 'Success',
        message: `Join request ${accept ? 'approved' : 'rejected'} successfully!`
      });
    } catch (error) {
      console.error('Error responding to group join request:', error);
      showAlert({
        type: 'error',
        title: 'Error',
        message: error.response?.data?.message || 'Failed to respond to join request. Please try again.'
      });
    }
  };

  // Get notification content and actions
  const getNotificationContent = (notification) => {
    switch (notification.type) {
      case 'follow_request':
        return {
          text: 'requested to follow you',
          actions: (
            <div className={styles.quickActions}>
              <Button
                variant="primary"
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  handleFollowResponse(notification, true);
                }}
              >
                Accept
              </Button>
              <Button
                variant="secondary"
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  handleFollowResponse(notification, false);
                }}
              >
                Decline
              </Button>
            </div>
          )
        };
      case 'group_invite':
        return {
          text: 'invited you to join a group',
          actions: (
            <div className={styles.quickActions}>
              <Button
                variant="primary"
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  handleGroupInviteResponse(notification, true);
                }}
              >
                Accept
              </Button>
              <Button
                variant="secondary"
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  handleGroupInviteResponse(notification, false);
                }}
              >
                Decline
              </Button>
            </div>
          )
        };
      case 'group_join_request':
        return {
          text: 'requested to join your group',
          actions: (
            <div className={styles.quickActions}>
              <Button
                variant="primary"
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  handleGroupJoinResponse(notification, true);
                }}
              >
                Approve
              </Button>
              <Button
                variant="secondary"
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  handleGroupJoinResponse(notification, false);
                }}
              >
                Reject
              </Button>
            </div>
          )
        };
      case 'follow_accepted':
        return { text: 'accepted your follow request' };
      case 'new_follower':
        return { text: 'started following you' };
      case 'post_like':
        return { text: 'liked your post' };
      case 'post_comment':
        return { text: 'commented on your post' };
      case 'group_join_approved':
        return { text: 'approved your request to join the group' };
      default:
        return { text: notification.content || 'sent you a notification' };
    }
  };

  if (!isOpen) return null;

  return (
    <div className={styles.dropdownContainer} ref={dropdownRef}>
      <div className={styles.dropdownHeader}>
        <h3>Notifications</h3>
      </div>
      
      <div className={styles.notificationsList}>
        {recentNotifications.length === 0 ? (
          <div className={styles.emptyState}>
            <p>No notifications yet</p>
          </div>
        ) : (
          recentNotifications.map(notification => {
            const content = getNotificationContent(notification);
            return (
              <div
                key={notification.id}
                className={`${styles.notificationItem} ${!notification.readAt ? styles.unread : ''}`}
                onClick={() => onMarkAsRead(notification.id)}
              >
                <div className={styles.notificationSender}>
                  {notification.sender?.profilePicture ? (
                    <Image
                      src={getUserProfilePictureUrl(notification.sender)}
                      alt={notification.sender.username}
                      width={40}
                      height={40}
                      className={styles.senderAvatar}
                      onError={(e) => {
                        e.target.src = getFallbackAvatar(notification.sender);
                      }}
                    />
                  ) : (
                    <Image
                      src={getFallbackAvatar(notification.sender)}
                      alt={notification.sender?.username || 'User'}
                      width={40}
                      height={40}
                      className={styles.senderAvatar}
                    />
                  )}
                </div>

                <div className={styles.notificationContent}>
                  <div className={styles.notificationText}>
                    <span className={styles.senderName}>
                      {notification.sender?.fullName || notification.sender?.username}
                    </span>
                    {' '}
                    <span className={styles.actionText}>
                      {content.text}
                    </span>
                  </div>
                  
                  <div className={styles.notificationTime}>
                    {formatDistanceToNow(new Date(notification.createdAt), { addSuffix: true })}
                  </div>

                  {content.actions && (
                    <div className={styles.actionsContainer}>
                      {content.actions}
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>

      <div className={styles.dropdownFooter}>
        <Link href="/notifications" className={styles.viewAllLink} onClick={onClose}>
          View All Notifications
        </Link>
      </div>
    </div>
  );
};

export default NotificationDropdown;
