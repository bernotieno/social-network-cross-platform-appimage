'use client';

import { useEffect } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { formatDistanceToNow } from 'date-fns';
import useNotifications from '@/hooks/useNotifications';
import { getUserProfilePictureUrl, getFallbackAvatar } from '@/utils/images';
import { groupAPI, userAPI } from '@/utils/api';
import { useAlert } from '@/contexts/AlertContext';
import Button from '@/components/Button';
import ProtectedRoute from '@/components/ProtectedRoute';
import styles from '@/styles/Notifications.module.css';

export default function Notifications() {
  const {
    notifications,
    isLoading,
    isRefreshing,
    fetchNotifications,
    markAsRead,
    markAllAsRead,
    deleteNotification,
    deleteAllNotifications
  } = useNotifications();

  const { showAlert } = useAlert();

  useEffect(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  const getNotificationContent = (notification) => {
    console.log("this is the notification", notification)

    // Parse notification data if it's a string
    let notificationData = {};
    if (notification.data) {
      try {
        notificationData = typeof notification.data === 'string'
          ? JSON.parse(notification.data)
          : notification.data;
      } catch (error) {
        console.error('Error parsing notification data:', error);
        notificationData = {};
      }
    }

    switch (notification.type) {
      case 'follow_request':
        return (
          <>
            <span className={styles.notificationText}>
              requested to follow you
            </span>
            <div className={styles.notificationActions}>
              <Button
                variant="primary"
                size="small"
                onClick={() => handleFollowResponse(notification, true)}
              >
                Accept
              </Button>
              <Button
                variant="secondary"
                size="small"
                onClick={() => handleFollowResponse(notification, false)}
              >
                Decline
              </Button>
            </div>
          </>
        );
      case 'follow_accepted':
        return (
          <span className={styles.notificationText}>
            accepted your follow request
          </span>
        );
      case 'new_follower':
        return (
          <span className={styles.notificationText}>
            started following you
          </span>
        );
      case 'post_like':
        const postContent = notificationData.postContent;
        return (
          <span className={styles.notificationText}>
            liked your post{postContent ? `: "${postContent}"` : ''}
          </span>
        );
      case 'post_comment':
        const comment = notificationData.comment;
        const postContentForComment = notificationData.postContent;
        return (
          <span className={styles.notificationText}>
            commented on your post{postContentForComment ? ` "${postContentForComment}"` : ''}:
            {comment ? ` "${comment}"` : ' (comment unavailable)'}
          </span>
        );
      case 'group_invite':
        return (
          <>
            <span className={styles.notificationText}>
              invited you to join the group "{notificationData.groupName || 'Unknown Group'}"
            </span>
            <div className={styles.notificationActions}>
              <Button
                variant="primary"
                size="small"
                onClick={() => handleGroupInviteResponse(notification.id, true)}
              >
                Join
              </Button>
              <Button
                variant="secondary"
                size="small"
                onClick={() => handleGroupInviteResponse(notification.id, false)}
              >
                Decline
              </Button>
            </div>
          </>
        );
      case 'group_join_request':
        return (
          <>
            <span className={styles.notificationText}>
              requested to join your group "{notificationData.groupName || 'Unknown Group'}"
            </span>
            <div className={styles.notificationActions}>
              <Button
                variant="primary"
                size="small"
                onClick={() => handleGroupJoinResponse(notification.id, true)}
              >
                Accept
              </Button>
              <Button
                variant="secondary"
                size="small"
                onClick={() => handleGroupJoinResponse(notification.id, false)}
              >
                Decline
              </Button>
            </div>
          </>
        );
      case 'group_join_approved':
        return (
          <span className={styles.notificationText}>
            approved your request to join the group "{notificationData.groupName || 'Unknown Group'}"
          </span>
        );
      case 'group_join_rejected':
        return (
          <span className={styles.notificationText}>
            declined your request to join the group "{notificationData.groupName || 'Unknown Group'}"
          </span>
        );
      case 'event_invite':
        return (
          <>
            <span className={styles.notificationText}>
              invited you to the event "{notificationData.eventName || 'Unknown Event'}"
            </span>
            <div className={styles.notificationActions}>
              <Button
                variant="primary"
                size="small"
                onClick={() => handleEventResponse(notification.id, 'going')}
              >
                Going
              </Button>
              <Button
                variant="secondary"
                size="small"
                onClick={() => handleEventResponse(notification.id, 'maybe')}
              >
                Maybe
              </Button>
              <Button
                variant="outline"
                size="small"
                onClick={() => handleEventResponse(notification.id, 'not_going')}
              >
                Decline
              </Button>
            </div>
          </>
        );
      case 'group_event_created':
        return (
          <span className={styles.notificationText}>
            created a new event "{notificationData.eventTitle || 'Untitled Event'}" in the group
          </span>
        );
      default:
        return (
          <span className={styles.notificationText}>
            {notification.content}
          </span>
        );
    }
  };

  // Handle follow request response
  const handleFollowResponse = async (notification, accept) => {
    try {
      // Extract follow request ID from notification data
      let followRequestId = notification.id; // fallback to notification ID

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
      markAsRead(notification.id);
      fetchNotifications();

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

  const handleGroupInviteResponse = async (notificationId, accept) => {
    try {
      await groupAPI.respondToGroupInvitation(notificationId, accept);

      // Mark notification as read and refresh notifications
      markAsRead(notificationId);
      fetchNotifications();

      // Show success message
      console.log(`Group invite ${accept ? 'accepted' : 'declined'} successfully`);
    } catch (error) {
      console.error('Error responding to group invitation:', error);
      // You might want to show an error message to the user here
    }
  };

  const handleGroupJoinResponse = async (notificationId, accept) => {
    try {
      // Get the notification to extract group and user information
      const notification = notifications.find(n => n.id === notificationId);
      if (!notification) {
        console.error('Notification not found');
        return;
      }

      // Parse the notification data to get group and user info
      let notificationData;
      try {
        notificationData = JSON.parse(notification.data);
      } catch (error) {
        console.error('Failed to parse notification data:', error);
        return;
      }

      const groupId = notificationData.groupId;
      const userId = notification.senderId; // The user who sent the request

      if (!groupId || !userId) {
        console.error('Missing group ID or user ID in notification data');
        return;
      }

      // Call the appropriate API endpoint
      if (accept) {
        await groupAPI.approveJoinRequest(groupId, userId);
        showAlert({
          type: 'success',
          title: 'Success',
          message: 'Join request approved successfully!'
        });
      } else {
        await groupAPI.rejectJoinRequest(groupId, userId);
        showAlert({
          type: 'success',
          title: 'Success',
          message: 'Join request rejected successfully!'
        });
      }

      // Mark notification as read and refresh notifications
      markAsRead(notificationId);
      fetchNotifications();

    } catch (error) {
      console.error('Error responding to group join request:', error);
      showAlert({
        type: 'error',
        title: 'Error',
        message: error.response?.data?.message || 'Failed to respond to join request. Please try again.'
      });
    }
  };

  const handleEventResponse = (notificationId, response) => {
    console.log(`Event response ${response}: ${notificationId}`);
    markAsRead(notificationId);
  };

  return (
    <ProtectedRoute>
      <div className={styles.notificationsContainer}>
        <div className={styles.notificationsHeader}>
          <div className={styles.titleContainer}>
            <h1 className={styles.notificationsTitle}>Notifications</h1>
            {isRefreshing && (
              <div className={styles.refreshIndicator}>
                <div className={styles.spinner}></div>
                <span>Updating...</span>
              </div>
            )}
          </div>

          {notifications && notifications.length > 0 && (
            <div className={styles.headerActions}>
              <Button
                variant="secondary"
                size="small"
                onClick={markAllAsRead}
              >
                Mark all as read
              </Button>
              <Button
                variant="outline"
                size="small"
                onClick={deleteAllNotifications}
                className={styles.deleteAllButton}
              >
                Clear all
              </Button>
            </div>
          )}
        </div>

        {isLoading ? (
          <div className={styles.loading}>Loading notifications...</div>
        ) : !notifications || notifications.length === 0 ? (
          <div className={styles.emptyNotifications}>
            <p>No notifications yet</p>
            <p>When you get notifications, they'll appear here</p>
          </div>
        ) : (
          <div className={styles.notificationsList}>
            {notifications.map(notification => (
              <div
                key={notification.id}
                className={`${styles.notificationItem} ${!notification.readAt ? styles.unread : ''}`}
              >
                <Link href={`/profile/${notification.sender.id}`} className={styles.notificationSender}>
                  {notification.sender.profilePicture ? (
                    <Image
                      src={getUserProfilePictureUrl(notification.sender)}
                      alt={notification.sender.username}
                      width={50}
                      height={50}
                      className={styles.senderAvatar}
                      onError={(e) => {
                        e.target.src = getFallbackAvatar(notification.sender);
                      }}
                    />
                  ) : (
                    <Image
                      src={getFallbackAvatar(notification.sender)}
                      alt={notification.sender.username}
                      width={50}
                      height={50}
                      className={styles.senderAvatar}
                    />
                  )}
                </Link>

                <div
                  className={styles.notificationContent}
                  onClick={() => markAsRead(notification.id)}
                >
                  <div className={styles.notificationHeader}>
                    <Link href={`/profile/${notification.sender.id}`} className={styles.senderName}>
                      {notification.sender.fullName}
                    </Link>
                    {getNotificationContent(notification)}
                  </div>

                  <div className={styles.notificationTime}>
                    {formatDistanceToNow(new Date(notification.createdAt), { addSuffix: true })}
                  </div>
                </div>

                <div className={styles.notificationActions}>
                  {!notification.readAt && (
                    <div className={styles.unreadIndicator} />
                  )}
                  <button
                    className={styles.deleteButton}
                    onClick={(e) => {
                      e.stopPropagation();
                      deleteNotification(notification.id);
                    }}
                    title="Delete notification"
                  >
                    ×
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </ProtectedRoute>
  );
}
