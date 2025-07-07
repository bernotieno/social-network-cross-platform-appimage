// Web compatibility layer for Electron APIs
// This provides mock implementations when running in a browser

if (typeof window !== 'undefined' && !window.electronAPI) {
    // Mock Electron APIs for web compatibility
    window.electronAPI = {
        // Authentication APIs
        auth: {
            login: async (credentials) => {
                localStorage.setItem('socialNetworkSession', JSON.stringify(credentials));
                return { success: true };
            },
            logout: async () => {
                localStorage.removeItem('socialNetworkSession');
                return { success: true };
            },
            getStoredSession: async () => {
                const session = localStorage.getItem('socialNetworkSession');
                return session ? JSON.parse(session) : null;
            },
            clearStoredSession: async () => {
                localStorage.removeItem('socialNetworkSession');
                return { success: true };
            },
            openRegistration: async () => {
                window.open('http://localhost:3000/auth/register', '_blank');
                return { success: true };
            }
        },

        // Storage APIs
        storage: {
            setItem: async (key, value) => {
                localStorage.setItem(key, JSON.stringify(value));
                return { success: true };
            },
            getItem: async (key) => {
                const item = localStorage.getItem(key);
                return item ? JSON.parse(item) : null;
            },
            removeItem: async (key) => {
                localStorage.removeItem(key);
                return { success: true };
            },
            clear: async () => {
                localStorage.clear();
                return { success: true };
            }
        },

        // Secure storage APIs (same as regular storage in web)
        secureStorage: {
            setItem: async (key, value) => {
                localStorage.setItem(`secure_${key}`, JSON.stringify(value));
                return { success: true };
            },
            getItem: async (key) => {
                const item = localStorage.getItem(`secure_${key}`);
                return item ? JSON.parse(item) : null;
            },
            removeItem: async (key) => {
                localStorage.removeItem(`secure_${key}`);
                return { success: true };
            }
        },

        // Network APIs
        network: {
            isOnline: async () => ({ isOnline: navigator.onLine }),
            onStatusChange: (callback) => {
                const handleOnline = () => callback(true);
                const handleOffline = () => callback(false);
                window.addEventListener('online', handleOnline);
                window.addEventListener('offline', handleOffline);
                return () => {
                    window.removeEventListener('online', handleOnline);
                    window.removeEventListener('offline', handleOffline);
                };
            }
        },

        // Notification APIs
        notifications: {
            show: async (options) => {
                if ('Notification' in window && Notification.permission === 'granted') {
                    new Notification(options.title, {
                        body: options.body,
                        icon: options.icon
                    });
                    return { success: true };
                }
                return { success: false, error: 'Notifications not available' };
            },
            requestPermission: async () => {
                if ('Notification' in window) {
                    return await Notification.requestPermission();
                }
                return 'denied';
            },
            onNotificationClick: (callback) => {
                // Mock implementation - in real Electron this would handle notification clicks
                return () => {};
            }
        },

        // Window APIs (no-op in web)
        window: {
            minimize: async () => ({ success: true }),
            maximize: async () => ({ success: true }),
            close: async () => {
                if (confirm('Close the application?')) {
                    window.close();
                }
                return { success: true };
            },
            focus: async () => {
                window.focus();
                return { success: true };
            },
            isMaximized: async () => false,
            onWindowStateChange: (callback) => () => {}
        },

        // App APIs
        app: {
            getVersion: async () => '1.0.0',
            quit: async () => {
                if (confirm('Quit the application?')) {
                    window.close();
                }
                return { success: true };
            },
            openExternal: async (url) => {
                window.open(url, '_blank');
                return { success: true };
            }
        },

        // Message cache APIs (using localStorage)
        messageCache: {
            saveMessages: async (userId, messages) => {
                localStorage.setItem(`messages_${userId}`, JSON.stringify(messages));
                return { success: true };
            },
            getMessages: async (userId) => {
                const messages = localStorage.getItem(`messages_${userId}`);
                return messages ? JSON.parse(messages) : [];
            },
            searchMessages: async (query, userId) => {
                const messages = await window.electronAPI.messageCache.getMessages(userId);
                return messages.filter(msg => 
                    msg.content.toLowerCase().includes(query.toLowerCase())
                );
            },
            clearMessages: async (userId) => {
                localStorage.removeItem(`messages_${userId}`);
                return { success: true };
            },
            clearAllMessages: async () => {
                const keys = Object.keys(localStorage);
                keys.forEach(key => {
                    if (key.startsWith('messages_')) {
                        localStorage.removeItem(key);
                    }
                });
                return { success: true };
            }
        },

        // User data APIs
        userData: {
            saveUser: async (user) => {
                localStorage.setItem('userData_user', JSON.stringify(user));
                return { success: true };
            },
            getUser: async () => {
                const user = localStorage.getItem('userData_user');
                return user ? JSON.parse(user) : null;
            },
            saveContacts: async (contacts) => {
                localStorage.setItem('userData_contacts', JSON.stringify(contacts));
                return { success: true };
            },
            getContacts: async () => {
                const contacts = localStorage.getItem('userData_contacts');
                return contacts ? JSON.parse(contacts) : [];
            },
            clearUserData: async () => {
                const keys = Object.keys(localStorage);
                keys.forEach(key => {
                    if (key.startsWith('userData_')) {
                        localStorage.removeItem(key);
                    }
                });
                return { success: true };
            }
        },

        // Development APIs
        dev: {
            openDevTools: async () => {
                console.log('DevTools would open in Electron');
                return { success: true };
            },
            reload: async () => {
                window.location.reload();
                return { success: true };
            }
        }
    };

    // Mock Node.js APIs
    window.nodeAPI = {
        platform: navigator.platform,
        arch: 'x64', // Mock value
        versions: {
            node: '18.0.0', // Mock value
            chrome: '120.0.0', // Mock value
            electron: '28.0.0' // Mock value
        }
    };

    // Mock WebSocket API (simplified)
    window.websocketAPI = {
        createConnection: async (url, protocols) => {
            // Return a mock connection ID
            return Math.random().toString(36).substr(2, 9);
        },
        closeConnection: async (connectionId) => ({ success: true }),
        sendMessage: async (connectionId, message) => ({ success: true }),
        onMessage: (connectionId, callback) => () => {},
        onClose: (connectionId, callback) => () => {},
        onError: (connectionId, callback) => () => {}
    };

    console.log('Web compatibility layer loaded - running in browser mode');
}
