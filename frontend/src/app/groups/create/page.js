'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import { useAuth } from '@/hooks/useAuth';
import { groupAPI } from '@/utils/api';
import { isGif, validateImageFile } from '@/utils/images';
import { useAlert } from '@/contexts/AlertContext';
import Button from '@/components/Button';
import ProtectedRoute from '@/components/ProtectedRoute';
import styles from '@/styles/CreateGroup.module.css';

export default function CreateGroup() {
  const { user } = useAuth();
  const { showSuccess, showError } = useAlert();
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    privacy: 'public'
  });
  const [coverPhoto, setCoverPhoto] = useState(null);
  const [coverPhotoPreview, setCoverPhotoPreview] = useState(null);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleCoverPhotoChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      // Validate file using utility function
      const validation = validateImageFile(file);
      if (!validation.isValid) {
        showError(validation.error, 'Invalid File');
        return;
      }

      setCoverPhoto(file);
      const reader = new FileReader();
      reader.onload = (e) => {
        setCoverPhotoPreview(e.target.result);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!formData.name.trim()) {
      showError('Group name is required', 'Validation Error');
      return;
    }

    setIsLoading(true);

    try {
      // Create FormData for multipart/form-data
      const submitData = new FormData();
      submitData.append('name', formData.name);
      submitData.append('description', formData.description);
      submitData.append('privacy', formData.privacy);

      if (coverPhoto) {
        submitData.append('coverPhoto', coverPhoto);
      }

      const response = await groupAPI.createGroup(submitData);

      if (response.data.success) {
        // Show success message
        await showSuccess('Your group has been created successfully!', 'Group Created');

        // Redirect to the newly created group
        router.push(`/groups/${response.data.data.group.id}`);
      } else {
        throw new Error(response.data.message || 'Failed to create group');
      }
    } catch (error) {
      console.error('Error creating group:', error);
      const errorMessage = error.response?.data?.message || 'Failed to create group. Please try again.';
      showError(errorMessage, 'Failed to Create Group');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <ProtectedRoute>
      <div className={styles.createGroupContainer}>
        <div className={styles.createGroupCard}>
          <div className={styles.header}>
            <h1 className={styles.title}>Create New Group</h1>
            <p className={styles.subtitle}>
              Create a group to connect with people who share your interests
            </p>
          </div>

          <form onSubmit={handleSubmit} className={styles.form}>
            {/* Cover Photo Section */}
            <div className={styles.coverPhotoSection}>
              <label className={styles.label}>Cover Photo (Optional)</label>
              <div className={styles.coverPhotoContainer}>
                {coverPhotoPreview ? (
                  <div className={styles.coverPhotoPreview}>
                    {coverPhoto && isGif(coverPhoto.name) ? (
                      // Use regular img tag for GIFs to preserve animation
                      <img
                        src={coverPhotoPreview}
                        alt="Cover photo preview"
                        style={{
                          width: '100%',
                          height: '100%',
                          objectFit: 'cover',
                          borderRadius: '8px'
                        }}
                      />
                    ) : (
                      // Use Next.js Image for static images
                      <Image
                        src={coverPhotoPreview}
                        alt="Cover photo preview"
                        fill
                        sizes="(max-width: 768px) 100vw, (max-width: 1200px) 80vw, 60vw"
                        style={{ objectFit: 'cover' }}
                      />
                    )}
                    <button
                      type="button"
                      className={styles.removeCoverPhoto}
                      onClick={() => {
                        setCoverPhoto(null);
                        setCoverPhotoPreview(null);
                      }}
                    >
                      ✕
                    </button>
                  </div>
                ) : (
                  <div className={styles.coverPhotoPlaceholder}>
                    <div className={styles.uploadIcon}>📷</div>
                    <p>Add a cover photo</p>
                  </div>
                )}
                <input
                  type="file"
                  accept="image/jpeg,image/jpg,image/png,image/gif"
                  onChange={handleCoverPhotoChange}
                  className={styles.fileInput}
                  id="coverPhoto"
                />
                <label htmlFor="coverPhoto" className={styles.fileInputLabel}>
                  {coverPhotoPreview ? 'Change Photo' : 'Add Photo'}
                </label>
              </div>
            </div>

            {/* Group Name */}
            <div className={styles.inputGroup}>
              <label htmlFor="name" className={styles.label}>
                Group Name *
              </label>
              <input
                type="text"
                id="name"
                name="name"
                value={formData.name}
                onChange={handleInputChange}
                placeholder="Enter group name"
                className={styles.input}
                maxLength={100}
                required
              />
            </div>

            {/* Description */}
            <div className={styles.inputGroup}>
              <label htmlFor="description" className={styles.label}>
                Description
              </label>
              <textarea
                id="description"
                name="description"
                value={formData.description}
                onChange={handleInputChange}
                placeholder="Describe what your group is about..."
                className={styles.textarea}
                rows={4}
                maxLength={500}
              />
              <div className={styles.charCount}>
                {formData.description.length}/500
              </div>
            </div>

            {/* Privacy Settings */}
            <div className={styles.inputGroup}>
              <label className={styles.label}>Privacy</label>
              <div className={styles.privacyOptions}>
                <label className={styles.radioOption}>
                  <input
                    type="radio"
                    name="privacy"
                    value="public"
                    checked={formData.privacy === 'public'}
                    onChange={handleInputChange}
                  />
                  <div className={styles.radioContent}>
                    <div className={styles.radioTitle}>🌎 Public</div>
                    <div className={styles.radioDescription}>
                      Anyone can see the group and join immediately
                    </div>
                  </div>
                </label>

                <label className={styles.radioOption}>
                  <input
                    type="radio"
                    name="privacy"
                    value="private"
                    checked={formData.privacy === 'private'}
                    onChange={handleInputChange}
                  />
                  <div className={styles.radioContent}>
                    <div className={styles.radioTitle}>🔒 Private</div>
                    <div className={styles.radioDescription}>
                      Only members can see posts. People must request to join.
                    </div>
                  </div>
                </label>
              </div>
            </div>

            {/* Action Buttons */}
            <div className={styles.actions}>
              <Button
                type="button"
                variant="secondary"
                onClick={() => router.back()}
                disabled={isLoading}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="primary"
                disabled={isLoading || !formData.name.trim()}
              >
                {isLoading ? 'Creating...' : 'Create Group'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </ProtectedRoute>
  );
}
