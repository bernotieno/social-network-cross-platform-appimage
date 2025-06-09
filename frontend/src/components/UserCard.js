'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { useAuth } from '@/hooks/useAuth';
import { userAPI } from '@/utils/api';
import { getUserProfilePictureUrl, getFallbackAvatar } from '@/utils/images';
import { getFollowButtonState } from '@/utils/privacy';
import Button from './Button';
import styles from '@/styles/UserCard.module.css';

export default function UserCard({ user, showFollowButton = true, onFollowChange }) {
  const { user: currentUser } = useAuth();
  const [isFollowing, setIsFollowing] = useState(user.isFollowing || false);
  const [hasPendingRequest, setHasPendingRequest] = useState(user.hasPendingFollowRequest || false);
  const [isLoading, setIsLoading] = useState(false);

  const isOwnProfile = currentUser?.id === user.id;

  // Update follow state when user prop changes
  useEffect(() => {
    setIsFollowing(user.isFollowing || false);
    setHasPendingRequest(user.hasPendingFollowRequest || false);
  }, [user.isFollowing, user.hasPendingFollowRequest]);

  const handleFollow = async () => {
    if (isLoading || isOwnProfile) return;

    try {
      setIsLoading(true);

      // Create a profile-like object for the button state function
      const profileData = {
        user: user,
        isFollowedByCurrentUser: isFollowing,
        hasPendingFollowRequest: hasPendingRequest,
        followStatus: hasPendingRequest ? 'pending' : (isFollowing ? 'accepted' : '')
      };

      const buttonState = getFollowButtonState(profileData, currentUser);

      switch (buttonState.action) {
        case 'unfollow':
          await userAPI.unfollow(user.id);
          setIsFollowing(false);
          setHasPendingRequest(false);
          if (onFollowChange) {
            onFollowChange(user.id, false);
          }
          break;

        case 'follow':
        case 'request_follow':
          await userAPI.follow(user.id);
          if (user.isPrivate) {
            // For private users, set pending request
            setHasPendingRequest(true);
            setIsFollowing(false);
          } else {
            // For public users, set following
            setIsFollowing(true);
            setHasPendingRequest(false);
          }
          if (onFollowChange) {
            onFollowChange(user.id, !user.isPrivate);
          }
          break;

        case 'cancel_request':
          await userAPI.unfollow(user.id);
          setHasPendingRequest(false);
          setIsFollowing(false);
          if (onFollowChange) {
            onFollowChange(user.id, false);
          }
          break;

        default:
          break;
      }
    } catch (error) {
      console.error('Error following/unfollowing user:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className={styles.userCard}>
      <Link href={`/profile/${user.id}`} className={styles.userLink}>
        <div className={styles.userAvatar}>
          {user.profilePicture ? (
            <img
              src={getUserProfilePictureUrl(user)}
              alt={user.username}
              className={styles.avatar}
              onError={(e) => {
                e.target.src = getFallbackAvatar(user);
              }}
            />
          ) : (
            <img
              src={getFallbackAvatar(user)}
              alt={user.username}
              className={styles.avatar}
            />
          )}
        </div>

        <div className={styles.userInfo}>
          <div className={styles.userNameRow}>
            <h3 className={styles.userName}>
              {user.fullName || `${user.firstName || ''} ${user.lastName || ''}`.trim() || user.username}
            </h3>
            <span className={`${styles.privacyTag} ${user.isPrivate ? styles.privateTag : styles.publicTag}`}>
              {user.isPrivate ? '🔒' : '🌐'}
            </span>
          </div>
          <p className={styles.userUsername}>@{user.username}</p>
          {user.bio && (
            <p className={styles.userBio}>{user.bio}</p>
          )}
        </div>
      </Link>

      {showFollowButton && !isOwnProfile && (
        <div className={styles.userActions}>
          {(() => {
            // Create a profile-like object for the button state function
            const profileData = {
              user: user,
              isFollowedByCurrentUser: isFollowing,
              hasPendingFollowRequest: hasPendingRequest,
              followStatus: hasPendingRequest ? 'pending' : (isFollowing ? 'accepted' : '')
            };

            const buttonState = getFollowButtonState(profileData, currentUser);

            return (
              <Button
                variant={buttonState.variant}
                size="small"
                onClick={handleFollow}
                disabled={isLoading || buttonState.disabled}
              >
                {isLoading ? 'Loading...' : buttonState.text}
              </Button>
            );
          })()}
        </div>
      )}
    </div>
  );
}
