'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { useAuth } from '@/hooks/useAuth';
import { groupAPI } from '@/utils/api';
import Button from '@/components/Button';
import ProtectedRoute from '@/components/ProtectedRoute';
import styles from '@/styles/Groups.module.css';

export default function Groups() {
  const { user } = useAuth();
  const [groups, setGroups] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('my-groups');
  
  useEffect(() => {
    fetchGroups();
  }, [activeTab]);
  
  const fetchGroups = async () => {
    try {
      setIsLoading(true);
      
      // This would be an API call to get groups based on the active tab
      // For now, we'll just set an empty array
      const response = await groupAPI.getGroups();
      
      // Filter groups based on active tab
      let filteredGroups = [];
      if (activeTab === 'my-groups') {
        filteredGroups = response.data.groups.filter(group => 
          group.members.some(member => member.id === user.id)
        );
      } else if (activeTab === 'discover') {
        filteredGroups = response.data.groups.filter(group => 
          !group.members.some(member => member.id === user.id)
        );
      }
      
      setGroups(filteredGroups || []);
    } catch (error) {
      console.error('Error fetching groups:', error);
    } finally {
      setIsLoading(false);
    }
  };
  
  const handleJoinGroup = async (groupId) => {
    try {
      await groupAPI.joinGroup(groupId);
      
      // Update local state
      setGroups(prev => 
        prev.map(group => 
          group.id === groupId 
            ? { ...group, isJoined: true } 
            : group
        )
      );
    } catch (error) {
      console.error('Error joining group:', error);
    }
  };
  
  const handleLeaveGroup = async (groupId) => {
    if (window.confirm('Are you sure you want to leave this group?')) {
      try {
        await groupAPI.leaveGroup(groupId);
        
        // Update local state
        if (activeTab === 'my-groups') {
          setGroups(prev => prev.filter(group => group.id !== groupId));
        } else {
          setGroups(prev => 
            prev.map(group => 
              group.id === groupId 
                ? { ...group, isJoined: false } 
                : group
            )
          );
        }
      } catch (error) {
        console.error('Error leaving group:', error);
      }
    }
  };
  
  return (
    <ProtectedRoute>
      <div className={styles.groupsContainer}>
        <div className={styles.groupsHeader}>
          <h1 className={styles.groupsTitle}>Groups</h1>
          
          <Link href="/groups/create">
            <Button variant="primary">Create Group</Button>
          </Link>
        </div>
        
        <div className={styles.groupsTabs}>
          <button
            className={`${styles.tabButton} ${activeTab === 'my-groups' ? styles.activeTab : ''}`}
            onClick={() => setActiveTab('my-groups')}
          >
            My Groups
          </button>
          <button
            className={`${styles.tabButton} ${activeTab === 'discover' ? styles.activeTab : ''}`}
            onClick={() => setActiveTab('discover')}
          >
            Discover
          </button>
        </div>
        
        {isLoading ? (
          <div className={styles.loading}>Loading groups...</div>
        ) : groups.length === 0 ? (
          <div className={styles.emptyGroups}>
            {activeTab === 'my-groups' ? (
              <>
                <p>You haven't joined any groups yet</p>
                <Link href="/groups?tab=discover">
                  <Button variant="primary">Discover Groups</Button>
                </Link>
              </>
            ) : (
              <>
                <p>No groups to discover at the moment</p>
                <Link href="/groups/create">
                  <Button variant="primary">Create a Group</Button>
                </Link>
              </>
            )}
          </div>
        ) : (
          <div className={styles.groupsGrid}>
            {groups.map(group => (
              <div key={group.id} className={styles.groupCard}>
                <div className={styles.groupCover}>
                  {group.coverPhoto ? (
                    <Image
                      src={group.coverPhoto}
                      alt={group.name}
                      fill
                      style={{ objectFit: 'cover' }}
                    />
                  ) : (
                    <div className={styles.defaultCover} />
                  )}
                </div>
                
                <div className={styles.groupInfo}>
                  <h2 className={styles.groupName}>{group.name}</h2>
                  
                  <p className={styles.groupPrivacy}>
                    {group.privacy === 'public' ? '🌎 Public' : '🔒 Private'}
                  </p>
                  
                  <p className={styles.groupMembers}>
                    {group.membersCount} {group.membersCount === 1 ? 'member' : 'members'}
                  </p>
                  
                  {group.description && (
                    <p className={styles.groupDescription}>{group.description}</p>
                  )}
                </div>
                
                <div className={styles.groupActions}>
                  <Link href={`/groups/${group.id}`} className={styles.viewGroupLink}>
                    <Button variant="secondary" fullWidth>View Group</Button>
                  </Link>
                  
                  {group.isJoined ? (
                    <Button 
                      variant="outline" 
                      fullWidth
                      onClick={() => handleLeaveGroup(group.id)}
                    >
                      Leave Group
                    </Button>
                  ) : (
                    <Button 
                      variant="primary" 
                      fullWidth
                      onClick={() => handleJoinGroup(group.id)}
                    >
                      Join Group
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </ProtectedRoute>
  );
}
