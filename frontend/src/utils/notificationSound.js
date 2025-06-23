'use client';

// Notification sound utility
class NotificationSound {
  constructor() {
    this.audioContext = null;
    this.isEnabled = true;
    this.volume = 0.5;
    
    // Initialize audio context on first user interaction
    this.initializeAudioContext();
    
    // Load user preferences
    this.loadPreferences();
  }

  initializeAudioContext() {
    if (typeof window !== 'undefined') {
      try {
        // Create audio context on first user interaction
        const initAudio = () => {
          if (!this.audioContext) {
            this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
            console.log('Audio context initialized for notifications');
          }
          // Resume audio context if it's suspended
          if (this.audioContext.state === 'suspended') {
            this.audioContext.resume();
          }
          document.removeEventListener('click', initAudio);
          document.removeEventListener('keydown', initAudio);
          document.removeEventListener('touchstart', initAudio);
        };

        document.addEventListener('click', initAudio, { once: true });
        document.addEventListener('keydown', initAudio, { once: true });
        document.addEventListener('touchstart', initAudio, { once: true });
      } catch (error) {
        console.warn('Audio context not supported:', error);
      }
    }
  }

  loadPreferences() {
    if (typeof window !== 'undefined') {
      const savedEnabled = localStorage.getItem('notificationSoundEnabled');
      const savedVolume = localStorage.getItem('notificationSoundVolume');
      
      this.isEnabled = savedEnabled !== null ? JSON.parse(savedEnabled) : true;
      this.volume = savedVolume !== null ? parseFloat(savedVolume) : 0.5;
    }
  }

  savePreferences() {
    if (typeof window !== 'undefined') {
      localStorage.setItem('notificationSoundEnabled', JSON.stringify(this.isEnabled));
      localStorage.setItem('notificationSoundVolume', this.volume.toString());
    }
  }

  // Create a simple notification sound using Web Audio API
  createNotificationTone(frequency = 800, duration = 200, type = 'sine') {
    if (!this.audioContext || !this.isEnabled) return;

    try {
      // Resume audio context if suspended
      if (this.audioContext.state === 'suspended') {
        this.audioContext.resume();
      }

      const oscillator = this.audioContext.createOscillator();
      const gainNode = this.audioContext.createGain();

      oscillator.connect(gainNode);
      gainNode.connect(this.audioContext.destination);

      oscillator.frequency.setValueAtTime(frequency, this.audioContext.currentTime);
      oscillator.type = type;

      // Create envelope for smooth sound
      gainNode.gain.setValueAtTime(0, this.audioContext.currentTime);
      gainNode.gain.linearRampToValueAtTime(this.volume * 0.3, this.audioContext.currentTime + 0.01);
      gainNode.gain.exponentialRampToValueAtTime(0.001, this.audioContext.currentTime + duration / 1000);

      oscillator.start(this.audioContext.currentTime);
      oscillator.stop(this.audioContext.currentTime + duration / 1000);
    } catch (error) {
      console.warn('Error playing notification sound:', error);
    }
  }

  // Play different sounds for different notification types
  playNotificationSound(type = 'default') {
    if (!this.isEnabled) return;

    switch (type) {
      case 'message':
      case 'post_comment':
        // Higher pitch for messages/comments
        this.createNotificationTone(900, 150, 'sine');
        break;
      
      case 'like':
      case 'post_like':
        // Pleasant tone for likes
        this.createNotificationTone(700, 200, 'sine');
        break;
      
      case 'follow':
      case 'follow_request':
      case 'new_follower':
        // Friendly tone for follows
        this.createNotificationTone(600, 250, 'sine');
        break;
      
      case 'group':
      case 'group_invite':
      case 'group_join_request':
        // Group-related notifications
        this.createNotificationTone(750, 200, 'triangle');
        break;
      
      case 'event':
      case 'event_invite':
      case 'group_event_created':
        // Event notifications - two-tone
        this.createNotificationTone(650, 150, 'sine');
        setTimeout(() => {
          this.createNotificationTone(800, 150, 'sine');
        }, 100);
        break;
      
      case 'success':
        // Success sound - ascending tone
        this.createNotificationTone(600, 100, 'sine');
        setTimeout(() => {
          this.createNotificationTone(800, 100, 'sine');
        }, 80);
        break;
      
      case 'error':
        // Error sound - descending tone
        this.createNotificationTone(400, 200, 'sawtooth');
        break;
      
      case 'warning':
        // Warning sound - pulsing
        this.createNotificationTone(500, 100, 'square');
        setTimeout(() => {
          this.createNotificationTone(500, 100, 'square');
        }, 150);
        break;
      
      default:
        // Default notification sound
        this.createNotificationTone(800, 200, 'sine');
    }
  }

  // Enable/disable sounds
  setEnabled(enabled) {
    this.isEnabled = enabled;
    this.savePreferences();
  }

  // Set volume (0.0 to 1.0)
  setVolume(volume) {
    this.volume = Math.max(0, Math.min(1, volume));
    this.savePreferences();
  }

  // Get current settings
  getSettings() {
    return {
      isEnabled: this.isEnabled,
      volume: this.volume,
    };
  }

  // Test sound
  testSound() {
    this.playNotificationSound('default');
  }
}

// Create singleton instance
const notificationSound = new NotificationSound();

export default notificationSound;

// Convenience functions
export const playNotificationSound = (type) => {
  notificationSound.playNotificationSound(type);
};

export const setNotificationSoundEnabled = (enabled) => {
  notificationSound.setEnabled(enabled);
};

export const setNotificationSoundVolume = (volume) => {
  notificationSound.setVolume(volume);
};

export const getNotificationSoundSettings = () => {
  return notificationSound.getSettings();
};

export const testNotificationSound = () => {
  notificationSound.testSound();
};
