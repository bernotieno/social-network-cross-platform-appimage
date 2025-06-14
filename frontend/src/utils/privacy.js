/**
 * Privacy utility functions for handling user permissions and access control
 */

/**
 * Check if the current user can view a profile's full information
 * @param {Object} profile - The profile object from API
 * @param {Object} currentUser - The current logged-in user
 * @returns {boolean} - Whether the user can view full profile
 */
export const canViewFullProfile = (profile, currentUser) => {
  if (!profile || !currentUser) return false;

  // User can always view their own profile
  if (profile.user?.id === currentUser.id || profile.id === currentUser.id) {
    return true;
  }

  // If profile is public, everyone can view
  const user = profile.user || profile;
  if (!user.isPrivate) {
    return true;
  }

  // For private profiles, check if current user is following
  return profile.isFollowedByCurrentUser || user.isFollowedByCurrentUser || false;
};

/**
 * Check if the current user can view a profile's posts
 * @param {Object} profile - The profile object from API
 * @param {Object} currentUser - The current logged-in user
 * @returns {boolean} - Whether the user can view posts
 */
export const canViewPosts = (profile, currentUser) => {
  return canViewFullProfile(profile, currentUser);
};

/**
 * Check if the current user can view a profile's followers list
 * @param {Object} profile - The profile object from API
 * @param {Object} currentUser - The current logged-in user
 * @returns {boolean} - Whether the user can view followers
 */
export const canViewFollowers = (profile, currentUser) => {
  return canViewFullProfile(profile, currentUser);
};

/**
 * Check if the current user can view a profile's following list
 * @param {Object} profile - The profile object from API
 * @param {Object} currentUser - The current logged-in user
 * @returns {boolean} - Whether the user can view following
 */
export const canViewFollowing = (profile, currentUser) => {
  return canViewFullProfile(profile, currentUser);
};

/**
 * Check if the current user can send a message to another user
 * @param {Object} targetUser - The user to send message to
 * @param {Object} currentUser - The current logged-in user
 * @returns {boolean} - Whether the user can send message
 */
export const canSendMessage = (targetUser, currentUser) => {
  if (!targetUser || !currentUser) return false;

  // Can't message yourself
  if (targetUser.id === currentUser.id) return false;

  // If target user is public, allow messaging
  if (!targetUser.isPrivate) return true;

  // For private users, check if there's a follow relationship
  return targetUser.isFollowedByCurrentUser || false;
};

/**
 * Get the appropriate privacy message for restricted content
 * @param {string} contentType - Type of content (profile, posts, followers, following)
 * @param {Object} profile - The profile object
 * @returns {string} - Privacy message to display
 */
export const getPrivacyMessage = (contentType, profile) => {
  const user = profile?.user || profile;
  const username = user?.username || 'User';

  switch (contentType) {
    case 'profile':
      return `This account is private. Follow @${username} to see their profile.`;
    case 'posts':
      return `This account is private. Follow @${username} to see their posts.`;
    case 'followers':
      return `This account is private. Follow @${username} to see who follows them.`;
    case 'following':
      return `This account is private. Follow @${username} to see who they're following.`;
    default:
      return `This account is private. Follow @${username} to see their content.`;
  }
};

/**
 * Check if a user profile is the current user's own profile
 * @param {Object} profile - The profile object
 * @param {Object} currentUser - The current logged-in user
 * @returns {boolean} - Whether this is the user's own profile
 */
export const isOwnProfile = (profile, currentUser) => {
  if (!profile || !currentUser) return false;
  const user = profile.user || profile;
  return user.id === currentUser.id;
};

/**
 * Get privacy status information for a user
 * @param {Object} user - The user object
 * @returns {Object} - Privacy status with icon and label
 */
export const getPrivacyStatus = (user) => {
  if (!user) return { icon: '🌐', label: 'Public', isPrivate: false };

  return user.isPrivate
    ? { icon: '🔒', label: 'Private', isPrivate: true }
    : { icon: '🌐', label: 'Public', isPrivate: false };
};

/**
 * Get follow button state and text based on user relationship
 * @param {Object} profile - The profile object from API
 * @param {Object} currentUser - The current logged-in user
 * @returns {Object} - Button state with text, variant, and action
 */
export const getFollowButtonState = (profile, currentUser) => {
  if (!profile || !currentUser) {
    return { text: 'Follow', variant: 'primary', action: 'follow', disabled: true };
  }

  const user = profile.user || profile;

  // Can't follow yourself
  if (user.id === currentUser.id) {
    return { text: 'Edit Profile', variant: 'outline', action: 'edit', disabled: false };
  }

  // Check if already following
  if (profile.isFollowedByCurrentUser) {
    return { text: 'Unfollow', variant: 'secondary', action: 'unfollow', disabled: false };
  }

  // Check if there's a pending request
  if (profile.hasPendingFollowRequest || profile.followStatus === 'pending') {
    return { text: 'Pending Request', variant: 'secondary', action: 'cancel_request', disabled: false };
  }

  // For private profiles, show "Request to Follow"
  if (user.isPrivate) {
    return { text: 'Request to Follow', variant: 'primary', action: 'request_follow', disabled: false };
  }

  // Default follow button for public profiles
  return { text: 'Follow', variant: 'primary', action: 'follow', disabled: false };
};
