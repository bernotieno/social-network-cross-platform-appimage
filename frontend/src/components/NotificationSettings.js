'use client';

import React, { useState, useEffect } from 'react';
import Button from '@/components/Button';
import { 
  getNotificationSoundSettings, 
  setNotificationSoundEnabled, 
  setNotificationSoundVolume,
  testNotificationSound 
} from '@/utils/notificationSound';
import styles from '@/styles/NotificationSettings.module.css';

const NotificationSettings = ({ isOpen, onClose }) => {
  const [settings, setSettings] = useState({
    isEnabled: true,
    volume: 0.5,
  });

  useEffect(() => {
    // Load current settings
    const currentSettings = getNotificationSoundSettings();
    setSettings(currentSettings);
  }, []);

  const handleToggleSound = () => {
    const newEnabled = !settings.isEnabled;
    setNotificationSoundEnabled(newEnabled);
    setSettings(prev => ({ ...prev, isEnabled: newEnabled }));
  };

  const handleVolumeChange = (e) => {
    const newVolume = parseFloat(e.target.value);
    setNotificationSoundVolume(newVolume);
    setSettings(prev => ({ ...prev, volume: newVolume }));
  };

  const handleTestSound = () => {
    testNotificationSound();
  };

  if (!isOpen) return null;

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>Notification Settings</h2>
          <button 
            className={styles.closeButton}
            onClick={onClose}
            aria-label="Close settings"
          >
            ×
          </button>
        </div>
        
        <div className={styles.content}>
          <div className={styles.setting}>
            <div className={styles.settingInfo}>
              <h3 className={styles.settingTitle}>Sound Notifications</h3>
              <p className={styles.settingDescription}>
                Play sounds when you receive new notifications
              </p>
            </div>
            <label className={styles.toggle}>
              <input
                type="checkbox"
                checked={settings.isEnabled}
                onChange={handleToggleSound}
                className={styles.toggleInput}
              />
              <span className={styles.toggleSlider}></span>
            </label>
          </div>

          {settings.isEnabled && (
            <>
              <div className={styles.setting}>
                <div className={styles.settingInfo}>
                  <h3 className={styles.settingTitle}>Volume</h3>
                  <p className={styles.settingDescription}>
                    Adjust the volume of notification sounds
                  </p>
                </div>
                <div className={styles.volumeControl}>
                  <input
                    type="range"
                    min="0"
                    max="1"
                    step="0.1"
                    value={settings.volume}
                    onChange={handleVolumeChange}
                    className={styles.volumeSlider}
                  />
                  <span className={styles.volumeValue}>
                    {Math.round(settings.volume * 100)}%
                  </span>
                </div>
              </div>

              <div className={styles.setting}>
                <div className={styles.settingInfo}>
                  <h3 className={styles.settingTitle}>Test Sound</h3>
                  <p className={styles.settingDescription}>
                    Play a sample notification sound
                  </p>
                </div>
                <Button
                  variant="secondary"
                  size="small"
                  onClick={handleTestSound}
                >
                  Test Sound
                </Button>
              </div>
            </>
          )}

          <div className={styles.info}>
            <h4 className={styles.infoTitle}>About Notification Sounds</h4>
            <ul className={styles.infoList}>
              <li>Different notification types have unique sounds</li>
              <li>Sounds are generated using Web Audio API</li>
              <li>Your browser must support audio playback</li>
              <li>Settings are saved locally in your browser</li>
            </ul>
          </div>
        </div>

        <div className={styles.footer}>
          <Button
            variant="primary"
            onClick={onClose}
          >
            Done
          </Button>
        </div>
      </div>
    </div>
  );
};

export default NotificationSettings;
