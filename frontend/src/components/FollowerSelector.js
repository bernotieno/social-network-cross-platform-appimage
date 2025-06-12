import React, { useState, useEffect } from 'react';
import { userAPI } from '@/utils/api';
import { useAuth } from '@/hooks/useAuth';
import { useAlert } from '@/contexts/AlertContext';
import { getUserProfilePictureUrl, getFallbackAvatar } from '@/utils/images';
import styles from '@/styles/FollowerSelector.module.css';

const FollowerSelector = ({ selectedFollowers, onSelectionChange, isVisible, onClose }) => {
  const [followers, setFollowers] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [loading, setLoading] = useState(false);
  const { user } = useAuth();
  const { showError } = useAlert();

  useEffect(() => {
    if (isVisible) {
      fetchFollowers();
    }
  }, [isVisible]);

  const fetchFollowers = async () => {
    if (!user?.id) {
      console.error('No user ID available');
      return;
    }

    setLoading(true);
    try {
      // Get current user's followers
      const response = await userAPI.getFollowers(user.id);
      if (response.data.success) {
        setFollowers(response.data.followers || []);
      }
    } catch (error) {
      console.error('Error fetching followers:', error);
      showError('Failed to load followers', 'Error');
    } finally {
      setLoading(false);
    }
  };

  const filteredFollowers = followers.filter(follower =>
    follower.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
    follower.fullName.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleFollowerToggle = (followerId) => {
    const newSelection = selectedFollowers.includes(followerId)
      ? selectedFollowers.filter(id => id !== followerId)
      : [...selectedFollowers, followerId];
    
    onSelectionChange(newSelection);
  };

  const handleSelectAll = () => {
    if (selectedFollowers.length === filteredFollowers.length) {
      onSelectionChange([]);
    } else {
      onSelectionChange(filteredFollowers.map(f => f.id));
    }
  };

  if (!isVisible) return null;

  return (
    <div className={styles.overlay}>
      <div className={styles.modal}>
        <div className={styles.header}>
          <h3>Select Followers</h3>
          <button className={styles.closeButton} onClick={onClose}>
            ✕
          </button>
        </div>

        <div className={styles.searchContainer}>
          <input
            type="text"
            placeholder="Search followers..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className={styles.searchInput}
          />
        </div>

        <div className={styles.actions}>
          <button
            className={styles.selectAllButton}
            onClick={handleSelectAll}
            disabled={filteredFollowers.length === 0}
          >
            {selectedFollowers.length === filteredFollowers.length ? 'Deselect All' : 'Select All'}
          </button>
          <span className={styles.selectedCount}>
            {selectedFollowers.length} selected
          </span>
        </div>

        <div className={styles.followersList}>
          {loading ? (
            <div className={styles.loading}>Loading followers...</div>
          ) : filteredFollowers.length === 0 ? (
            <div className={styles.noFollowers}>
              {searchTerm ? 'No followers found matching your search.' : 'You have no followers yet.'}
            </div>
          ) : (
            filteredFollowers.map(follower => (
              <div
                key={follower.id}
                className={`${styles.followerItem} ${
                  selectedFollowers.includes(follower.id) ? styles.selected : ''
                }`}
                onClick={() => handleFollowerToggle(follower.id)}
              >
                <div className={styles.followerInfo}>
                  <img
                    src={getUserProfilePictureUrl(follower.profilePicture) || getFallbackAvatar()}
                    alt={follower.username}
                    className={styles.avatar}
                    onError={(e) => {
                      e.target.src = getFallbackAvatar();
                    }}
                  />
                  <div className={styles.followerDetails}>
                    <div className={styles.username}>{follower.username}</div>
                    <div className={styles.fullName}>{follower.fullName}</div>
                  </div>
                </div>
                <div className={styles.checkbox}>
                  <input
                    type="checkbox"
                    checked={selectedFollowers.includes(follower.id)}
                    onChange={() => handleFollowerToggle(follower.id)}
                    onClick={(e) => e.stopPropagation()}
                  />
                </div>
              </div>
            ))
          )}
        </div>

        <div className={styles.footer}>
          <button className={styles.cancelButton} onClick={onClose}>
            Cancel
          </button>
          <button className={styles.confirmButton} onClick={onClose}>
            Done ({selectedFollowers.length})
          </button>
        </div>
      </div>
    </div>
  );
};

export default FollowerSelector;
