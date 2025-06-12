'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import { useAuth } from '@/hooks/useAuth';
import { postAPI } from '@/utils/api';
import { getImageUrl, validateImageFile, getFileTypeDisplayName } from '@/utils/images';
import { useAlert } from '@/contexts/AlertContext';
import Button from '@/components/Button';
import ProtectedRoute from '@/components/ProtectedRoute';
import FollowerSelector from '@/components/FollowerSelector';
import SelectedFollowersTags from '@/components/SelectedFollowersTags';
import styles from '@/styles/CreatePost.module.css';

export default function CreatePost() {
  const { user } = useAuth();
  const { showSuccess, showError } = useAlert();
  const router = useRouter();

  const [content, setContent] = useState('');
  const [image, setImage] = useState(null);
  const [imagePreview, setImagePreview] = useState(null);
  const [visibility, setVisibility] = useState('public');
  const [selectedFollowers, setSelectedFollowers] = useState([]);
  const [showFollowerSelector, setShowFollowerSelector] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  const handleContentChange = (e) => {
    setContent(e.target.value);
  };

  const handleImageChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      // Validate file
      const validation = validateImageFile(file);
      if (!validation.isValid) {
        setError(validation.error);
        return;
      }

      setError(''); // Clear any previous errors
      setImage(file);

      // Create preview URL
      const reader = new FileReader();
      reader.onloadend = () => {
        setImagePreview(reader.result);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleRemoveImage = () => {
    setImage(null);
    setImagePreview(null);
  };

  const handleVisibilityChange = (e) => {
    const newVisibility = e.target.value;
    setVisibility(newVisibility);

    // Clear selected followers if not custom visibility
    if (newVisibility !== 'custom') {
      setSelectedFollowers([]);
    }

    // Show follower selector if custom visibility is selected
    if (newVisibility === 'custom') {
      setShowFollowerSelector(true);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!content.trim() && !image) {
      setError('Please add some content or an image to your post');
      return;
    }

    try {
      setIsSubmitting(true);
      setError('');

      // Create FormData for API request
      const formData = new FormData();
      formData.append('content', content);
      formData.append('visibility', visibility);

      // Add custom viewers if visibility is custom
      if (visibility === 'custom' && selectedFollowers.length > 0) {
        formData.append('customViewers', JSON.stringify(selectedFollowers));
      }

      if (image) {
        formData.append('image', image);
      }

      // Call API to create post
      const response = await postAPI.createPost(formData);
      console.log('Post created:', response.data);

      // Show success message
      await showSuccess('Your post has been created successfully!', 'Post Created');

      // Redirect to home page after successful post creation
      router.push('/');
    } catch (error) {
      console.error('Error creating post:', error);
      const errorMessage = error.response?.data?.message || 'Failed to create post. Please try again.';
      setError(errorMessage);
      showError(errorMessage, 'Failed to Create Post');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <ProtectedRoute>
      <div className={styles.createPostContainer}>
        <div className={styles.createPostCard}>
          <h1 className={styles.createPostTitle}>Create a Post</h1>

          {error && (
            <div className={styles.errorAlert}>
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className={styles.createPostForm}>
            <div className={styles.userInfo}>
              {user?.profilePicture ? (
                <Image
                  src={getImageUrl(user.profilePicture)}
                  alt={user.username}
                  width={40}
                  height={40}
                  className={styles.userAvatar}
                />
              ) : (
                <div className={styles.userAvatarPlaceholder}>
                  {user?.username?.charAt(0).toUpperCase() || 'U'}
                </div>
              )}
              <span className={styles.userName}>{user?.fullName}</span>
            </div>

            <textarea
              className={styles.contentInput}
              placeholder="What's on your mind?"
              value={content}
              onChange={handleContentChange}
              rows={5}
            />

            {imagePreview && (
              <div className={styles.imagePreviewContainer}>
                <img
                  src={imagePreview}
                  alt="Preview"
                  className={styles.imagePreview}
                  style={{ maxWidth: '100%', height: 'auto' }}
                />
                <button
                  type="button"
                  className={styles.removeImageButton}
                  onClick={handleRemoveImage}
                  title="Remove image"
                >
                  ✕
                </button>
                {image && (
                  <div className={styles.fileTypeIndicator}>
                    <span>📁 {getFileTypeDisplayName(image.type)}</span>
                  </div>
                )}
              </div>
            )}

            <div className={styles.postOptions}>
              <div className={styles.visibilityOption}>
                <label htmlFor="visibility" className={styles.visibilityLabel}>
                  Who can see your post?
                </label>
                <select
                  id="visibility"
                  className={styles.visibilitySelect}
                  value={visibility}
                  onChange={handleVisibilityChange}
                >
                  <option value="public">Everyone</option>
                  <option value="followers">Followers only</option>
                  <option value="custom">Custom (Select followers)</option>
                  <option value="private">Only me</option>
                </select>
              </div>

              {/* Custom Followers Selection */}
              {visibility === 'custom' && (
                <div className={styles.customVisibilitySection}>
                  <div className={styles.customVisibilityHeader}>
                    <span className={styles.customVisibilityLabel}>
                      Select specific followers who can see this post:
                    </span>
                    <button
                      type="button"
                      className={styles.selectFollowersButton}
                      onClick={() => setShowFollowerSelector(true)}
                    >
                      {selectedFollowers.length === 0 ? 'Select Followers' : `Edit Selection (${selectedFollowers.length})`}
                    </button>
                  </div>

                  <SelectedFollowersTags
                    selectedFollowerIds={selectedFollowers}
                    onRemoveFollower={(followerId) => {
                      setSelectedFollowers(prev => prev.filter(id => id !== followerId));
                    }}
                  />

                  {selectedFollowers.length === 0 && (
                    <div className={styles.noFollowersSelected}>
                      No followers selected. Click "Select Followers" to choose who can see this post.
                    </div>
                  )}
                </div>
              )}

              <div className={styles.addToPost}>
                <p className={styles.addToPostLabel}>Add to your post:</p>
                <div className={styles.addToPostOptions}>
                  <label className={styles.imageUploadLabel}>
                    <span className={styles.imageIcon}>🖼️</span>
                    <span>Photo/GIF</span>
                    <input
                      type="file"
                      accept="image/jpeg,image/jpg,image/png,image/gif"
                      onChange={handleImageChange}
                      className={styles.imageInput}
                    />
                  </label>
                </div>
              </div>
            </div>

            <Button
              type="submit"
              variant="primary"
              size="large"
              fullWidth
              disabled={isSubmitting || (!content.trim() && !image)}
            >
              {isSubmitting ? 'Posting...' : 'Post'}
            </Button>
          </form>

          {/* Follower Selector Modal */}
          <FollowerSelector
            selectedFollowers={selectedFollowers}
            onSelectionChange={setSelectedFollowers}
            isVisible={showFollowerSelector}
            onClose={() => setShowFollowerSelector(false)}
          />
        </div>
      </div>
    </ProtectedRoute>
  );
}
