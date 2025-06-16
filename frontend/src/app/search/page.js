'use client';

import { useState, useEffect } from 'react';
import { userAPI } from '@/utils/api';
import { useAuth } from '@/hooks/useAuth';
import UserCard from '@/components/UserCard';
import ProtectedRoute from '@/components/ProtectedRoute';
import styles from '@/styles/Search.module.css';

export default function SearchPage() {
  const { user } = useAuth();
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState([]);
  const [isSearching, setIsSearching] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);

  // Debounced search function
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (searchQuery.trim()) {
        searchUsers(searchQuery.trim());
      } else {
        setSearchResults([]);
        setHasSearched(false);
      }
    }, 300); // 300ms debounce

    return () => clearTimeout(timeoutId);
  }, [searchQuery]);

  const searchUsers = async (query) => {
    if (!query.trim()) {
      setSearchResults([]);
      setHasSearched(false);
      return;
    }

    setIsSearching(true);
    setHasSearched(true);

    try {
      const response = await userAPI.searchUsers(query);

      // Filter out current user from results
      const filteredResults = response.data.data.users.filter(
        (searchUser) => searchUser.id !== user?.id
      );

      setSearchResults(filteredResults);
    } catch (error) {
      console.error('Error searching users:', error);
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  };

  return (
    <ProtectedRoute>
      <div className={styles.searchContainer}>
        <div className={styles.searchHeader}>
          <h1 className={styles.searchTitle}>Discover People</h1>
          <p className={styles.searchSubtitle}>
            Find and connect with other users on the platform
          </p>
        </div>

        <div className={styles.searchInputContainer}>
          <div className={styles.searchInputWrapper}>
            <input
              type="text"
              placeholder="Search for users by name or username..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className={styles.searchInput}
            />
            <div className={styles.searchIcon}>🔍</div>
          </div>
        </div>

        <div className={styles.searchContent}>
          {isSearching && (
            <div className={styles.loading}>
              <div className={styles.loadingSpinner}></div>
              <p>Searching users...</p>
            </div>
          )}

          {!isSearching && hasSearched && searchResults.length === 0 && (
            <div className={styles.noResults}>
              <div className={styles.noResultsIcon}>👥</div>
              <h3>No users found</h3>
              <p>
                Try searching with different keywords or check your spelling.
              </p>
            </div>
          )}

          {!isSearching && !hasSearched && (
            <div className={styles.searchPrompt}>
              <div className={styles.searchPromptIcon}>🔍</div>
              <h3>Start searching</h3>
              <p>Enter a name or username to find other users to follow.</p>
            </div>
          )}

          {!isSearching && searchResults.length > 0 && (
            <div className={styles.searchResults}>
              <div className={styles.resultsHeader}>
                <h3>Search Results ({searchResults.length})</h3>
              </div>
              <div className={styles.usersList}>
                {searchResults.map((searchUser) => (
                  <UserCard
                    key={searchUser.id}
                    user={searchUser}
                    showFollowButton={false}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </ProtectedRoute>
  );
}
