import React, { useState, useEffect } from 'react';
import { userAPI } from '@/utils/api';
import { useAuth } from '@/hooks/useAuth';
import { getUserProfilePictureUrl, getFallbackAvatar } from '@/utils/images';
import styles from '@/styles/SelectedFollowersTags.module.css';

const SelectedFollowersTags = ({ selectedFollowerIds, onRemoveFollower }) => {
  const [followers, setFollowers] = useState([]);
  const [loading, setLoading] = useState(false);
  const { user } = useAuth();

  useEffect(() => {
    if (selectedFollowerIds.length > 0) {
      fetchFollowerDetails();
    } else {
      setFollowers([]);
    }
  }, [selectedFollowerIds]);

  const fetchFollowerDetails = async () => {
    if (!user?.id) {
      console.error('No user ID available');
      return;
    }

    setLoading(true);
    try {
      const response = await userAPI.getFollowers(user.id);
      if (response.data.success) {
        const allFollowers = response.data.followers || [];
        const selectedFollowers = allFollowers.filter(follower =>
          selectedFollowerIds.includes(follower.id)
        );
        setFollowers(selectedFollowers);
      }
    } catch (error) {
      console.error('Error fetching follower details:', error);
    } finally {
      setLoading(false);
    }
  };

  if (selectedFollowerIds.length === 0) {
    return null;
  }

  if (loading) {
    return (
      <div className={styles.container}>
        <div className={styles.loading}>Loading selected followers...</div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.label}>Selected followers:</div>
      <div className={styles.tagsContainer}>
        {followers.map(follower => (
          <div key={follower.id} className={styles.tag}>
            <img
              src={getUserProfilePictureUrl(follower.profilePicture) || getFallbackAvatar()}
              alt={follower.username}
              className={styles.avatar}
              onError={(e) => {
                e.target.src = getFallbackAvatar();
              }}
            />
            <span className={styles.username}>{follower.username}</span>
            <button
              className={styles.removeButton}
              onClick={() => onRemoveFollower(follower.id)}
              title={`Remove ${follower.username}`}
            >
              ✕
            </button>
          </div>
        ))}
      </div>
    </div>
  );
};

export default SelectedFollowersTags;
