'use client';

import React from 'react';
import { getPrivacyMessage } from '@/utils/privacy';
import Button from './Button';
import styles from '@/styles/PrivacyRestricted.module.css';

/**
 * Component for displaying privacy-restricted content with appropriate messages and actions
 */
const PrivacyRestricted = ({ 
  contentType = 'content',
  profile,
  showFollowButton = false,
  onFollowClick,
  isFollowing = false,
  isLoadingFollow = false,
  className = '',
  children
}) => {
  const message = getPrivacyMessage(contentType, profile);
  const user = profile?.user || profile;

  return (
    <div className={`${styles.restrictedContainer} ${className}`}>
      <div className={styles.restrictedContent}>
        <div className={styles.lockIcon}>🔒</div>
        <h3 className={styles.restrictedTitle}>This Account is Private</h3>
        <p className={styles.restrictedMessage}>{message}</p>
        
        {showFollowButton && user && (
          <div className={styles.actionContainer}>
            <Button
              variant={isFollowing ? 'secondary' : 'primary'}
              onClick={onFollowClick}
              disabled={isLoadingFollow}
              className={styles.followButton}
            >
              {isLoadingFollow ? 'Loading...' : (isFollowing ? 'Unfollow' : 'Follow')}
            </Button>
          </div>
        )}
        
        {children && (
          <div className={styles.additionalContent}>
            {children}
          </div>
        )}
      </div>
    </div>
  );
};

/**
 * Component for displaying blurred/restricted content with overlay
 */
export const BlurredContent = ({ 
  children, 
  isRestricted = true, 
  message = "This content is private",
  className = ''
}) => {
  if (!isRestricted) {
    return children;
  }

  return (
    <div className={`${styles.blurredContainer} ${className}`}>
      <div className={styles.blurredContent}>
        {children}
      </div>
      <div className={styles.blurredOverlay}>
        <div className={styles.overlayContent}>
          <div className={styles.lockIcon}>🔒</div>
          <p className={styles.overlayMessage}>{message}</p>
        </div>
      </div>
    </div>
  );
};

/**
 * Component for displaying empty state with privacy message
 */
export const PrivacyEmptyState = ({ 
  contentType = 'content',
  profile,
  icon = '🔒',
  className = ''
}) => {
  const message = getPrivacyMessage(contentType, profile);

  return (
    <div className={`${styles.emptyState} ${className}`}>
      <div className={styles.emptyIcon}>{icon}</div>
      <p className={styles.emptyMessage}>{message}</p>
    </div>
  );
};

/**
 * Component for displaying limited profile information for private accounts
 */
export const LimitedProfileInfo = ({ 
  user,
  showFollowButton = false,
  onFollowClick,
  isFollowing = false,
  isLoadingFollow = false,
  className = ''
}) => {
  return (
    <div className={`${styles.limitedProfile} ${className}`}>
      <div className={styles.limitedInfo}>
        <h2 className={styles.limitedName}>
          {user.fullName || `${user.firstName || ''} ${user.lastName || ''}`.trim() || user.username}
        </h2>
        <p className={styles.limitedUsername}>@{user.username}</p>
        <div className={styles.privacyIndicator}>
          <span className={styles.privateTag}>🔒 Private Account</span>
        </div>
      </div>
      
      {showFollowButton && (
        <div className={styles.limitedActions}>
          <Button
            variant={isFollowing ? 'secondary' : 'primary'}
            onClick={onFollowClick}
            disabled={isLoadingFollow}
          >
            {isLoadingFollow ? 'Loading...' : (isFollowing ? 'Unfollow' : 'Follow')}
          </Button>
        </div>
      )}
    </div>
  );
};

export default PrivacyRestricted;
