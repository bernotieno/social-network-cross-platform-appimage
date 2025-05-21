'use client';

import { useState, useEffect } from 'react';
import Image from 'next/image';
import { useParams } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { userAPI, postAPI } from '@/utils/api';
import Button from '@/components/Button';
import ProtectedRoute from '@/components/ProtectedRoute';
import styles from '@/styles/Profile.module.css';

export default function ProfilePage() {
  const { id } = useParams();
  const { user: currentUser } = useAuth();
  
  const [profile, setProfile] = useState(null);
  const [posts, setPosts] = useState([]);
  const [activeTab, setActiveTab] = useState('posts');
  const [isLoading, setIsLoading] = useState(true);
  const [isFollowing, setIsFollowing] = useState(false);
  const [followersCount, setFollowersCount] = useState(0);
  const [followingCount, setFollowingCount] = useState(0);
  
  const isOwnProfile = currentUser?.id === id;

  useEffect(() => {
    fetchProfileData();
  }, [id]);

  const fetchProfileData = async () => {
    try {
      setIsLoading(true);
      
      // Fetch user profile
      const profileResponse = await userAPI.getProfile(id);
      setProfile(profileResponse.data);
      
      // Check if current user is following this profile
      if (currentUser && !isOwnProfile) {
        setIsFollowing(profileResponse.data.isFollowedByCurrentUser || false);
      }
      
      // Set followers and following counts
      setFollowersCount(profileResponse.data.followersCount || 0);
      setFollowingCount(profileResponse.data.followingCount || 0);
      
      // Fetch user posts
      const postsResponse = await postAPI.getPosts(id);
      setPosts(postsResponse.data.posts || []);
    } catch (error) {
      console.error('Error fetching profile data:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleFollow = async () => {
    try {
      if (isFollowing) {
        await userAPI.unfollow(id);
        setIsFollowing(false);
        setFollowersCount(prev => prev - 1);
      } else {
        await userAPI.follow(id);
        setIsFollowing(true);
        setFollowersCount(prev => prev + 1);
      }
    } catch (error) {
      console.error('Error following/unfollowing user:', error);
    }
  };

  const handleTabChange = (tab) => {
    setActiveTab(tab);
  };

  if (isLoading) {
    return (
      <div className={styles.loading}>
        Loading profile...
      </div>
    );
  }

  if (!profile) {
    return (
      <div className={styles.notFound}>
        <h1>Profile not found</h1>
        <p>The user you're looking for doesn't exist or has been removed.</p>
      </div>
    );
  }

  return (
    <ProtectedRoute>
      <div className={styles.profileContainer}>
        <div className={styles.profileHeader}>
          <div className={styles.profileCover}>
            {profile.coverPhoto ? (
              <Image
                src={profile.coverPhoto}
                alt={`${profile.username}'s cover`}
                fill
                style={{ objectFit: 'cover' }}
              />
            ) : (
              <div className={styles.defaultCover} />
            )}
          </div>
          
          <div className={styles.profileInfo}>
            <div className={styles.profilePicture}>
              {profile.profilePicture ? (
                <Image
                  src={profile.profilePicture}
                  alt={profile.username}
                  width={120}
                  height={120}
                  className={styles.avatar}
                />
              ) : (
                <div className={styles.avatarPlaceholder}>
                  {profile.username?.charAt(0).toUpperCase() || 'U'}
                </div>
              )}
            </div>
            
            <div className={styles.profileDetails}>
              <h1 className={styles.profileName}>{profile.fullName}</h1>
              <p className={styles.profileUsername}>@{profile.username}</p>
              
              {profile.bio && (
                <p className={styles.profileBio}>{profile.bio}</p>
              )}
              
              <div className={styles.profileStats}>
                <div className={styles.stat}>
                  <span className={styles.statCount}>{posts.length}</span>
                  <span className={styles.statLabel}>Posts</span>
                </div>
                <div className={styles.stat}>
                  <span className={styles.statCount}>{followersCount}</span>
                  <span className={styles.statLabel}>Followers</span>
                </div>
                <div className={styles.stat}>
                  <span className={styles.statCount}>{followingCount}</span>
                  <span className={styles.statLabel}>Following</span>
                </div>
              </div>
            </div>
            
            <div className={styles.profileActions}>
              {isOwnProfile ? (
                <Button variant="outline">Edit Profile</Button>
              ) : (
                <Button
                  variant={isFollowing ? 'secondary' : 'primary'}
                  onClick={handleFollow}
                >
                  {isFollowing ? 'Unfollow' : 'Follow'}
                </Button>
              )}
            </div>
          </div>
        </div>
        
        <div className={styles.profileContent}>
          <div className={styles.profileTabs}>
            <button
              className={`${styles.tabButton} ${activeTab === 'posts' ? styles.activeTab : ''}`}
              onClick={() => handleTabChange('posts')}
            >
              Posts
            </button>
            <button
              className={`${styles.tabButton} ${activeTab === 'followers' ? styles.activeTab : ''}`}
              onClick={() => handleTabChange('followers')}
            >
              Followers
            </button>
            <button
              className={`${styles.tabButton} ${activeTab === 'following' ? styles.activeTab : ''}`}
              onClick={() => handleTabChange('following')}
            >
              Following
            </button>
          </div>
          
          <div className={styles.tabContent}>
            {activeTab === 'posts' && (
              <div className={styles.postsGrid}>
                {posts.length === 0 ? (
                  <div className={styles.emptyState}>
                    <p>No posts yet</p>
                  </div>
                ) : (
                  <div className={styles.postsPlaceholder}>
                    <p>Posts will be displayed here</p>
                  </div>
                )}
              </div>
            )}
            
            {activeTab === 'followers' && (
              <div className={styles.followersGrid}>
                {followersCount === 0 ? (
                  <div className={styles.emptyState}>
                    <p>No followers yet</p>
                  </div>
                ) : (
                  <div className={styles.followersPlaceholder}>
                    <p>Followers will be displayed here</p>
                  </div>
                )}
              </div>
            )}
            
            {activeTab === 'following' && (
              <div className={styles.followingGrid}>
                {followingCount === 0 ? (
                  <div className={styles.emptyState}>
                    <p>Not following anyone yet</p>
                  </div>
                ) : (
                  <div className={styles.followingPlaceholder}>
                    <p>Following will be displayed here</p>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}
