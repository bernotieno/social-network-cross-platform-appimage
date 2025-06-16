'use client';

import { useState, useEffect } from 'react';
import Image from 'next/image';
import { groupAPI } from '@/utils/api';
import { useAlert } from '@/contexts/AlertContext';
import { getImageUrl, isGif, validateImageFile } from '@/utils/images';
import Button from '@/components/Button';
import styles from '@/styles/GroupEditModal.module.css';

export default function GroupEditModal({ group, isOpen, onClose, onUpdate }) {
  const { showAlert, showSuccess, showError } = useAlert();
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    privacy: 'public'
  });
  const [coverPhoto, setCoverPhoto] = useState(null);
  const [coverPhotoPreview, setCoverPhotoPreview] = useState(null);

  useEffect(() => {
    if (group && isOpen) {
      setFormData({
        name: group.name || '',
        description: group.description || '',
        privacy: group.privacy || 'public'
      });
      setCoverPhotoPreview(group.coverPhoto ? getImageUrl(group.coverPhoto) : null);
      setCoverPhoto(null);
    }
  }, [group, isOpen]);

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

  const handleRemoveCoverPhoto = () => {
    setCoverPhoto(null);
    setCoverPhotoPreview(null);
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

      const response = await groupAPI.updateGroup(group.id, submitData);

      if (response.data.success) {
        showSuccess('Group updated successfully!');
        onUpdate(); // Refresh group data
        onClose(); // Close modal
      } else {
        throw new Error(response.data.message || 'Failed to update group');
      }
    } catch (error) {
      console.error('Error updating group:', error);
      showError(error.response?.data?.message || 'Failed to update group. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    if (!isLoading) {
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div className={styles.modalOverlay} onClick={handleClose}>
      <div className={styles.modalContent} onClick={(e) => e.stopPropagation()}>
        <div className={styles.modalHeader}>
          <h2>Edit Group</h2>
          <button
            className={styles.closeButton}
            onClick={handleClose}
            disabled={isLoading}
          >
            ×
          </button>
        </div>

        <form onSubmit={handleSubmit} className={styles.form}>
          {/* Cover Photo Section */}
          <div className={styles.coverPhotoSection}>
            <label className={styles.label}>Cover Photo</label>
            <div className={styles.coverPhotoContainer}>
              {coverPhotoPreview ? (
                <div className={styles.coverPhotoPreview}>
                  {coverPhoto && isGif(coverPhoto.name) ? (
                    // Use regular img tag for GIFs to preserve animation
                    <img
                      src={coverPhotoPreview}
                      alt="Cover photo preview"
                      className={styles.coverPhotoImage}
                      style={{
                        width: '400px',
                        height: '200px',
                        objectFit: 'cover',
                        borderRadius: '8px'
                      }}
                    />
                  ) : (
                    // Use Next.js Image for static images
                    <Image
                      src={coverPhotoPreview}
                      alt="Cover photo preview"
                      width={400}
                      height={200}
                      className={styles.coverPhotoImage}
                    />
                  )}
                  <button
                    type="button"
                    className={styles.removeCoverButton}
                    onClick={handleRemoveCoverPhoto}
                    disabled={isLoading}
                  >
                    Remove
                  </button>
                </div>
              ) : (
                <div className={styles.coverPhotoPlaceholder}>
                  <span>No cover photo</span>
                </div>
              )}
              <input
                type="file"
                accept="image/jpeg,image/jpg,image/png,image/gif"
                onChange={handleCoverPhotoChange}
                className={styles.fileInput}
                disabled={isLoading}
              />
            </div>
          </div>

          {/* Group Name */}
          <div className={styles.formGroup}>
            <label htmlFor="name" className={styles.label}>
              Group Name *
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              className={styles.input}
              placeholder="Enter group name"
              required
              disabled={isLoading}
            />
          </div>

          {/* Description */}
          <div className={styles.formGroup}>
            <label htmlFor="description" className={styles.label}>
              Description
            </label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleInputChange}
              className={styles.textarea}
              placeholder="Describe your group..."
              rows={4}
              disabled={isLoading}
            />
          </div>

          {/* Privacy Settings */}
          <div className={styles.formGroup}>
            <label className={styles.label}>Privacy</label>
            <div className={styles.privacyOptions}>
              <label className={styles.radioOption}>
                <input
                  type="radio"
                  name="privacy"
                  value="public"
                  checked={formData.privacy === 'public'}
                  onChange={handleInputChange}
                  disabled={isLoading}
                />
                <div className={styles.radioContent}>
                  <div className={styles.radioTitle}>🌎 Public</div>
                  <div className={styles.radioDescription}>
                    Anyone can see posts and join the group.
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
                  disabled={isLoading}
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
              onClick={handleClose}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              disabled={isLoading || !formData.name.trim()}
            >
              {isLoading ? 'Updating...' : 'Update Group'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
