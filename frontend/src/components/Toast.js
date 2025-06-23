'use client';

import React from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useToast } from '@/contexts/ToastContext';
import { getUserProfilePictureUrl, getFallbackAvatar } from '@/utils/images';
import styles from '@/styles/Toast.module.css';

const Toast = ({ toast }) => {
  const { removeToast } = useToast();
  const router = useRouter();

  const handleClose = () => {
    removeToast(toast.id);
  };

  const handleToastClick = () => {
    if (toast.type === 'notification') {
      // Navigate to notifications page when notification toast is clicked
      router.push('/notifications');
      removeToast(toast.id);
    }
  };

  const getToastIcon = () => {
    switch (toast.type) {
      case 'success':
        return '✓';
      case 'error':
        return '✕';
      case 'warning':
        return '⚠';
      case 'info':
        return 'ℹ';
      case 'notification':
        return '🔔';
      default:
        return 'ℹ';
    }
  };

  const getToastClass = () => {
    return `${styles.toast} ${styles[toast.type]} ${styles.slideIn}`;
  };

  // For notification toasts, show user avatar and more detailed layout
  if (toast.type === 'notification' && toast.notification?.sender) {
    const sender = toast.notification.sender;

    return (
      <div className={getToastClass()} onClick={handleToastClick} style={{ cursor: 'pointer' }}>
        <div className={styles.notificationToast}>
          <Link href={`/profile/${sender.id}`} className={styles.senderAvatar} onClick={(e) => e.stopPropagation()}>
            {sender.profilePicture ? (
              <Image
                src={getUserProfilePictureUrl(sender)}
                alt={sender.username}
                width={40}
                height={40}
                className={styles.avatar}
                onError={(e) => {
                  e.target.src = getFallbackAvatar(sender);
                }}
              />
            ) : (
              <Image
                src={getFallbackAvatar(sender)}
                alt={sender.username}
                width={40}
                height={40}
                className={styles.avatar}
              />
            )}
          </Link>

          <div className={styles.notificationContent}>
            <div className={styles.notificationHeader}>
              <Link href={`/profile/${sender.id}`} className={styles.senderName} onClick={(e) => e.stopPropagation()}>
                {toast.title}
              </Link>
              <button
                className={styles.closeButton}
                onClick={(e) => {
                  e.stopPropagation();
                  handleClose();
                }}
                aria-label="Close notification"
              >
                ×
              </button>
            </div>
            <div className={styles.notificationMessage}>
              {toast.message}
            </div>
            <div className={styles.clickHint}>
              Click to view all notifications
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Standard toast layout for other types
  return (
    <div className={getToastClass()}>
      <div className={styles.toastContent}>
        <div className={styles.toastIcon}>
          {getToastIcon()}
        </div>
        <div className={styles.toastText}>
          {toast.title && <div className={styles.toastTitle}>{toast.title}</div>}
          <div className={styles.toastMessage}>{toast.message}</div>
        </div>
        <button
          className={styles.closeButton}
          onClick={handleClose}
          aria-label="Close toast"
        >
          ×
        </button>
      </div>
    </div>
  );
};

const ToastContainer = () => {
  const { toasts } = useToast();

  if (toasts.length === 0) return null;

  return (
    <div className={styles.toastContainer}>
      {toasts.map(toast => (
        <Toast key={toast.id} toast={toast} />
      ))}
    </div>
  );
};

export default ToastContainer;
