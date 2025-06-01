'use client';

import React, { createContext, useContext, useState } from 'react';
import { AlertModal, ConfirmModal } from '@/components/Modal';

const AlertContext = createContext();

export const useAlert = () => {
  const context = useContext(AlertContext);
  if (!context) {
    throw new Error('useAlert must be used within an AlertProvider');
  }
  return context;
};

export const AlertProvider = ({ children }) => {
  // Alert state
  const [alertState, setAlertState] = useState({
    isOpen: false,
    title: 'Alert',
    message: '',
    type: 'info',
    buttonText: 'OK',
    onClose: null
  });

  // Confirm state
  const [confirmState, setConfirmState] = useState({
    isOpen: false,
    title: 'Confirm Action',
    message: '',
    confirmText: 'Confirm',
    cancelText: 'Cancel',
    confirmVariant: 'primary',
    onConfirm: null,
    onCancel: null,
    isLoading: false
  });

  // Alert functions
  const showAlert = (config) => {
    return new Promise((resolve) => {
      setAlertState({
        isOpen: true,
        title: config.title || 'Alert',
        message: config.message || '',
        type: config.type || 'info',
        buttonText: config.buttonText || 'OK',
        onClose: () => {
          setAlertState(prev => ({ ...prev, isOpen: false }));
          resolve();
          if (config.onClose) config.onClose();
        }
      });
    });
  };

  const showSuccess = (message, title = 'Success') => {
    return showAlert({
      title,
      message,
      type: 'success'
    });
  };

  const showError = (message, title = 'Error') => {
    return showAlert({
      title,
      message,
      type: 'error'
    });
  };

  const showWarning = (message, title = 'Warning') => {
    return showAlert({
      title,
      message,
      type: 'warning'
    });
  };

  const showInfo = (message, title = 'Information') => {
    return showAlert({
      title,
      message,
      type: 'info'
    });
  };

  // Confirm functions
  const showConfirm = (config) => {
    return new Promise((resolve) => {
      setConfirmState({
        isOpen: true,
        title: config.title || 'Confirm Action',
        message: config.message || '',
        confirmText: config.confirmText || 'Confirm',
        cancelText: config.cancelText || 'Cancel',
        confirmVariant: config.confirmVariant || 'primary',
        isLoading: false,
        onConfirm: async () => {
          if (config.onConfirm) {
            setConfirmState(prev => ({ ...prev, isLoading: true }));
            try {
              await config.onConfirm();
              setConfirmState(prev => ({ ...prev, isOpen: false, isLoading: false }));
              resolve(true);
            } catch (error) {
              setConfirmState(prev => ({ ...prev, isLoading: false }));
              throw error;
            }
          } else {
            setConfirmState(prev => ({ ...prev, isOpen: false }));
            resolve(true);
          }
        },
        onCancel: () => {
          setConfirmState(prev => ({ ...prev, isOpen: false }));
          resolve(false);
          if (config.onCancel) config.onCancel();
        }
      });
    });
  };

  const closeAlert = () => {
    setAlertState(prev => ({ ...prev, isOpen: false }));
  };

  const closeConfirm = () => {
    setConfirmState(prev => ({ ...prev, isOpen: false }));
  };

  const value = {
    // Alert methods
    showAlert,
    showSuccess,
    showError,
    showWarning,
    showInfo,
    closeAlert,
    
    // Confirm methods
    showConfirm,
    closeConfirm,
    
    // State
    alertState,
    confirmState
  };

  return (
    <AlertContext.Provider value={value}>
      {children}
      
      {/* Alert Modal */}
      <AlertModal
        isOpen={alertState.isOpen}
        onClose={alertState.onClose}
        title={alertState.title}
        message={alertState.message}
        type={alertState.type}
        buttonText={alertState.buttonText}
      />
      
      {/* Confirm Modal */}
      <ConfirmModal
        isOpen={confirmState.isOpen}
        onClose={confirmState.onCancel}
        onConfirm={confirmState.onConfirm}
        title={confirmState.title}
        message={confirmState.message}
        confirmText={confirmState.confirmText}
        cancelText={confirmState.cancelText}
        confirmVariant={confirmState.confirmVariant}
        isLoading={confirmState.isLoading}
      />
    </AlertContext.Provider>
  );
};

export default AlertProvider;
