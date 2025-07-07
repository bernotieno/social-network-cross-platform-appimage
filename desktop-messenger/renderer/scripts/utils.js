// Utility functions for the Social Network Messenger

class Utils {
    // Format timestamp to human readable format
    static formatTimestamp(timestamp) {
        const date = new Date(timestamp);
        const now = new Date();
        const diffInMs = now - date;
        const diffInMinutes = Math.floor(diffInMs / (1000 * 60));
        const diffInHours = Math.floor(diffInMs / (1000 * 60 * 60));
        const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));

        if (diffInMinutes < 1) {
            return 'Just now';
        } else if (diffInMinutes < 60) {
            return `${diffInMinutes}m ago`;
        } else if (diffInHours < 24) {
            return `${diffInHours}h ago`;
        } else if (diffInDays < 7) {
            return `${diffInDays}d ago`;
        } else {
            return date.toLocaleDateString();
        }
    }

    // Format time for message display
    static formatMessageTime(timestamp) {
        const date = new Date(timestamp);
        return date.toLocaleTimeString([], { 
            hour: '2-digit', 
            minute: '2-digit' 
        });
    }

    // Escape HTML to prevent XSS
    static escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Debounce function for search and typing indicators
    static debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    // Throttle function for scroll events
    static throttle(func, limit) {
        let inThrottle;
        return function() {
            const args = arguments;
            const context = this;
            if (!inThrottle) {
                func.apply(context, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }

    // Generate avatar URL or fallback
    static getAvatarUrl(user, baseUrl = 'http://localhost:8080') {
        if (user && user.profilePicture) {
            // Ensure the URL is properly formatted
            const cleanBaseUrl = baseUrl.replace(/\/+$/, ''); // Remove trailing slashes
            const cleanProfilePicture = user.profilePicture.startsWith('/') ? user.profilePicture : `/${user.profilePicture}`;
            return `${cleanBaseUrl}${cleanProfilePicture}`;
        }
        return Utils.generateFallbackAvatar(user);
    }

    // Generate fallback avatar
    static generateFallbackAvatar(user) {
        if (!user) {
            return Utils.generateDefaultAvatar();
        }

        let initials = 'U';
        if (user.fullName) {
            initials = user.fullName.split(' ')
                .map(n => n.trim())
                .filter(n => n.length > 0)
                .map(n => n[0])
                .join('')
                .toUpperCase()
                .substring(0, 2); // Max 2 initials
        } else if (user.username) {
            initials = user.username.substring(0, 2).toUpperCase();
        }

        const canvas = document.createElement('canvas');
        canvas.width = 80; // Higher resolution for better quality
        canvas.height = 80;
        const ctx = canvas.getContext('2d');

        // Background color based on user ID or username
        const colors = ['#3498db', '#e74c3c', '#2ecc71', '#f39c12', '#9b59b6', '#1abc9c', '#e67e22', '#34495e'];
        let colorIndex = 0;
        if (user.id) {
            colorIndex = user.id.charCodeAt(0) % colors.length;
        } else if (user.username) {
            colorIndex = user.username.charCodeAt(0) % colors.length;
        }

        ctx.fillStyle = colors[colorIndex];
        ctx.fillRect(0, 0, 80, 80);

        // Text
        ctx.fillStyle = 'white';
        ctx.font = 'bold 28px Arial, sans-serif';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(initials, 40, 40);

        return canvas.toDataURL();
    }

    // Generate a default avatar when no user data is available
    static generateDefaultAvatar() {
        const canvas = document.createElement('canvas');
        canvas.width = 80;
        canvas.height = 80;
        const ctx = canvas.getContext('2d');

        // Gray background
        ctx.fillStyle = '#95a5a6';
        ctx.fillRect(0, 0, 80, 80);

        // User icon
        ctx.fillStyle = 'white';
        ctx.font = 'bold 32px Arial, sans-serif';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText('👤', 40, 40);

        return canvas.toDataURL();
    }

    // Validate email format
    static isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    // Sanitize message content
    static sanitizeMessage(content) {
        // Remove any HTML tags and trim whitespace
        return content.replace(/<[^>]*>/g, '').trim();
    }

    // Check if string contains only emojis
    static isOnlyEmojis(text) {
        const emojiRegex = /^[\u{1F600}-\u{1F64F}]|[\u{1F300}-\u{1F5FF}]|[\u{1F680}-\u{1F6FF}]|[\u{1F1E0}-\u{1F1FF}]|[\u{2600}-\u{26FF}]|[\u{2700}-\u{27BF}]+$/u;
        return emojiRegex.test(text.trim());
    }

    // Format file size
    static formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    // Copy text to clipboard
    static async copyToClipboard(text) {
        try {
            await navigator.clipboard.writeText(text);
            return true;
        } catch (err) {
            console.error('Failed to copy text: ', err);
            return false;
        }
    }

    // Show toast notification
    static showToast(message, type = 'info', duration = 3000) {
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.textContent = message;
        
        // Style the toast
        Object.assign(toast.style, {
            position: 'fixed',
            top: '20px',
            right: '20px',
            padding: '12px 20px',
            borderRadius: '6px',
            color: 'white',
            fontWeight: '500',
            zIndex: '10000',
            opacity: '0',
            transform: 'translateY(-20px)',
            transition: 'all 0.3s ease'
        });

        // Set background color based on type
        const colors = {
            info: '#3498db',
            success: '#2ecc71',
            warning: '#f39c12',
            error: '#e74c3c'
        };
        toast.style.backgroundColor = colors[type] || colors.info;

        document.body.appendChild(toast);

        // Animate in
        setTimeout(() => {
            toast.style.opacity = '1';
            toast.style.transform = 'translateY(0)';
        }, 10);

        // Remove after duration
        setTimeout(() => {
            toast.style.opacity = '0';
            toast.style.transform = 'translateY(-20px)';
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.parentNode.removeChild(toast);
                }
            }, 300);
        }, duration);
    }

    // Generate room ID for two users
    static generateRoomId(userId1, userId2) {
        return [userId1, userId2].sort().join('-');
    }

    // Check if user is online
    static isUserOnline(userId, onlineUsers) {
        return onlineUsers.has(userId);
    }

    // Format user status
    static formatUserStatus(user, onlineUsers, typingUsers, roomId) {
        const isOnline = Utils.isUserOnline(user.id, onlineUsers);
        const isTyping = typingUsers.has(roomId) && typingUsers.get(roomId).has(user.id);
        
        if (isTyping) {
            return 'typing...';
        } else if (isOnline) {
            return 'Online';
        } else {
            return 'Offline';
        }
    }

    // Scroll element to bottom smoothly
    static scrollToBottom(element, smooth = true) {
        if (smooth) {
            element.scrollTo({
                top: element.scrollHeight,
                behavior: 'smooth'
            });
        } else {
            element.scrollTop = element.scrollHeight;
        }
    }

    // Check if element is scrolled to bottom
    static isScrolledToBottom(element, threshold = 50) {
        return element.scrollHeight - element.clientHeight <= element.scrollTop + threshold;
    }

    // Highlight search terms in text
    static highlightSearchTerms(text, searchTerm) {
        if (!searchTerm) return text;
        
        const regex = new RegExp(`(${Utils.escapeRegex(searchTerm)})`, 'gi');
        return text.replace(regex, '<mark>$1</mark>');
    }

    // Escape regex special characters
    static escapeRegex(string) {
        return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    }

    // Get platform-specific keyboard shortcut text
    static getShortcutText(shortcut) {
        const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
        return shortcut.replace('Ctrl', isMac ? 'Cmd' : 'Ctrl');
    }

    // Check if current platform is macOS
    static isMacOS() {
        return navigator.platform.toUpperCase().indexOf('MAC') >= 0;
    }

    // Check if current platform is Windows
    static isWindows() {
        return navigator.platform.toUpperCase().indexOf('WIN') >= 0;
    }

    // Check if current platform is Linux
    static isLinux() {
        return navigator.platform.toUpperCase().indexOf('LINUX') >= 0;
    }

    // Format notification text
    static formatNotificationText(message, maxLength = 100) {
        if (message.length <= maxLength) {
            return message;
        }
        return message.substring(0, maxLength - 3) + '...';
    }

    // Create loading spinner element
    static createLoadingSpinner(size = 'medium') {
        const spinner = document.createElement('div');
        spinner.className = `loading-spinner loading-spinner-${size}`;
        
        const sizeMap = {
            small: '16px',
            medium: '24px',
            large: '32px'
        };
        
        Object.assign(spinner.style, {
            width: sizeMap[size] || sizeMap.medium,
            height: sizeMap[size] || sizeMap.medium,
            border: '2px solid #f3f3f3',
            borderTop: '2px solid #3498db',
            borderRadius: '50%',
            animation: 'spin 1s linear infinite'
        });
        
        return spinner;
    }

    // Add CSS animation for spinner if not already added
    static addSpinnerAnimation() {
        if (!document.getElementById('spinner-animation')) {
            const style = document.createElement('style');
            style.id = 'spinner-animation';
            style.textContent = `
                @keyframes spin {
                    0% { transform: rotate(0deg); }
                    100% { transform: rotate(360deg); }
                }
            `;
            document.head.appendChild(style);
        }
    }
}

// Add spinner animation on load
Utils.addSpinnerAnimation();

// Export Utils for global use
window.Utils = Utils;
