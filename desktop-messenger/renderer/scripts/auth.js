class AuthManager {
    constructor() {
        this.currentUser = null;
        this.isAuthenticated = false;
        this.apiBaseUrl = 'http://localhost:8080/api'; // Default backend URL
        this.isElectron = typeof window !== 'undefined' && window.electronAPI;

        // Initialize auth state
        this.init();
    }

    async init() {
        try {
            // Try to get stored session
            let storedSession;
            if (this.isElectron) {
                storedSession = await window.electronAPI.auth.getStoredSession();
            } else {
                const sessionData = localStorage.getItem('socialNetworkSession');
                storedSession = sessionData ? JSON.parse(sessionData) : null;
            }

            if (storedSession && storedSession.token && storedSession.user) {
                console.log('=== STORED SESSION FOUND ===');
                console.log('Session data:', JSON.stringify(storedSession, null, 2));
                console.log('============================');

                // Validate the session with the backend
                const isValid = await this.validateSession(storedSession.token, storedSession.user);
                if (isValid) {
                    this.currentUser = storedSession.user; // Ensure currentUser is set
                    this.isAuthenticated = true;
                    this.showChatScreen();
                    return;
                }
            } else {
                console.log('No valid stored session found');
            }
        } catch (error) {
            console.error('Error initializing auth:', error);
        }

        // Show login screen if no valid session
        this.showLoginScreen();
    }

    async validateSession(token, storedUser = null) {
        try {
            console.log('Validating session with token:', token ? 'Token present' : 'No token');
            console.log('Stored user:', storedUser);

            // Try to validate with any authenticated endpoint
            const response = await fetch(`${this.apiBaseUrl}/messages/online-users`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                credentials: 'include'
            });

            console.log('Session validation response status:', response.status);

            if (response.ok) {
                // Token is valid, set the user from stored data
                if (storedUser) {
                    this.currentUser = storedUser;
                    console.log('Session validation successful, user set:', this.currentUser);
                    return true;
                } else {
                    console.log('Token valid but no stored user data');
                    return false;
                }
            } else {
                console.log('Session validation failed, response not ok');
                return false;
            }
        } catch (error) {
            console.error('Session validation error:', error);
            return false;
        }
    }

    async login(email, password) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ email, password }),
                credentials: 'include'
            });

            const data = await response.json();

            if (response.ok && data.success) {
                console.log('=== LOGIN RESPONSE DATA ===');
                console.log('Full response:', JSON.stringify(data, null, 2));
                console.log('User data:', JSON.stringify(data.data?.user, null, 2));
                console.log('Token:', data.data?.token ? 'Present' : 'Missing');
                console.log('===========================');

                // Extract user and token from the nested data structure
                const user = data.data?.user;
                const token = data.data?.token;

                if (!user || !token) {
                    console.error('Missing user or token in response');
                    return { success: false, error: 'Invalid response format' };
                }

                this.currentUser = user;
                this.isAuthenticated = true;

                // Store session securely
                if (this.isElectron) {
                    await window.electronAPI.auth.login({
                        token: token,
                        user: user
                    });
                    // Store user data
                    await window.electronAPI.userData.saveUser(user);
                } else {
                    // Store in localStorage for web version
                    localStorage.setItem('socialNetworkSession', JSON.stringify({
                        token: token,
                        user: user
                    }));
                    localStorage.setItem('socialNetworkUser', JSON.stringify(user));
                }

                this.showChatScreen();
                return { success: true };
            } else {
                return { 
                    success: false, 
                    error: data.message || 'Login failed' 
                };
            }
        } catch (error) {
            console.error('Login error:', error);
            return { 
                success: false, 
                error: 'Network error. Please check your connection.' 
            };
        }
    }

    async logout() {
        try {
            // Call backend logout
            await fetch(`${this.apiBaseUrl}/auth/logout`, {
                method: 'POST',
                credentials: 'include'
            });
        } catch (error) {
            console.error('Logout API error:', error);
        }

        // Clear local data
        if (this.isElectron) {
            await window.electronAPI.auth.logout();
            await window.electronAPI.userData.clearUserData();
            await window.electronAPI.messageCache.clearAllMessages();
        } else {
            // Clear localStorage for web version
            localStorage.removeItem('socialNetworkSession');
            localStorage.removeItem('socialNetworkUser');
            localStorage.clear(); // Clear all app data
        }

        this.currentUser = null;
        this.isAuthenticated = false;
        
        this.showLoginScreen();
    }

    async openRegistration() {
        // Open the web registration page in the default browser
        if (this.isElectron) {
            await window.electronAPI.app.openExternal('http://localhost:3000/auth/register');
        } else {
            // Open in new tab for web version
            window.open('http://localhost:3000/auth/register', '_blank');
        }
    }

    showLoginScreen() {
        document.getElementById('loading-screen').style.display = 'none';
        document.getElementById('chat-screen').style.display = 'none';
        document.getElementById('login-screen').style.display = 'block';
    }

    showChatScreen() {
        console.log('Showing chat screen...');
        document.getElementById('loading-screen').style.display = 'none';
        document.getElementById('login-screen').style.display = 'none';
        document.getElementById('chat-screen').style.display = 'flex';

        // Update user info in the sidebar
        if (this.currentUser) {
            console.log('=== CURRENT USER DATA ===');
            console.log(JSON.stringify(this.currentUser, null, 2));
            console.log('========================');

            document.getElementById('user-name').textContent = this.currentUser.fullName || this.currentUser.username;
            const userAvatar = document.getElementById('user-avatar');

            let avatarUrl;
            if (this.currentUser.profilePicture) {
                avatarUrl = `${this.apiBaseUrl.replace('/api', '')}${this.currentUser.profilePicture}`;
                console.log('User avatar URL:', avatarUrl);
                console.log('User profilePicture:', this.currentUser.profilePicture);
                console.log('API base URL:', this.apiBaseUrl);
            } else {
                avatarUrl = this.getFallbackAvatar(this.currentUser);
                console.log('Using fallback avatar for user (no profilePicture)');
            }

            // Set up error handler for fallback
            userAvatar.onerror = (error) => {
                console.error('❌ User avatar failed to load:', error);
                console.log('Failed URL was:', avatarUrl);
                userAvatar.src = this.getFallbackAvatar(this.currentUser);
                userAvatar.onerror = null; // Prevent infinite loop
            };

            userAvatar.src = avatarUrl;
        } else {
            console.warn('No current user found when showing chat screen');
        }

        // Set up logout button event listener now that the chat screen is visible
        this.setupLogoutButton();

        // Add debug function to window for easy access
        this.addDebugFunctions();
    }

    setupLogoutButton() {
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) {
            // Remove any existing event listeners
            logoutBtn.replaceWith(logoutBtn.cloneNode(true));
            const newLogoutBtn = document.getElementById('logout-btn');

            newLogoutBtn.addEventListener('click', () => {
                this.logout();
            });
        }
    }

    addDebugFunctions() {
        // Add global debug functions
        window.printUserData = () => {
            console.log('=== CURRENT USER DATA ===');
            console.log(JSON.stringify(this.currentUser, null, 2));
            console.log('========================');
            return this.currentUser;
        };

        window.printStoredSession = async () => {
            try {
                const session = await window.electronAPI.auth.getStoredSession();
                console.log('=== STORED SESSION DATA ===');
                console.log(JSON.stringify(session, null, 2));
                console.log('===========================');
                return session;
            } catch (error) {
                console.error('Error getting stored session:', error);
                return null;
            }
        };

        window.printAllUserData = async () => {
            console.log('=== ALL USER DATA ===');
            console.log('Current User:', this.currentUser);
            console.log('Is Authenticated:', this.isAuthenticated);

            try {
                const session = await window.electronAPI.auth.getStoredSession();
                console.log('Stored Session:', session);
            } catch (error) {
                console.error('Error getting stored session:', error);
            }

            console.log('====================');
        };

        // Add avatar debug functions
        window.debugAvatars = {
            refreshAvatars: () => {
                if (window.chatManager) {
                    window.chatManager.refreshAllAvatars();
                    console.log('Avatars refreshed');
                } else {
                    console.log('Chat manager not available');
                }
            },
            showAvatarCache: () => {
                if (window.chatManager && window.chatManager.avatarCache) {
                    console.log('Avatar cache:', window.chatManager.avatarCache);
                    return window.chatManager.avatarCache;
                } else {
                    console.log('Avatar cache not available');
                }
            },
            testAvatarUrl: (contact) => {
                if (window.chatManager) {
                    const url = window.chatManager.getContactAvatar(contact);
                    console.log('Avatar URL for', contact.username, ':', url);
                    return url;
                } else {
                    console.log('Chat manager not available');
                }
            },
            testImageLoad: (url) => {
                console.log('Testing image load for URL:', url);
                const img = new Image();
                img.onload = () => console.log('✅ Image loaded successfully:', url);
                img.onerror = (error) => console.error('❌ Image failed to load:', url, error);
                img.src = url;
                return img;
            },
            checkBackendConnection: async () => {
                try {
                    const response = await fetch('http://localhost:8080/uploads/avatars/4c4cc51c-658f-4ce7-80fb-87b376a8e74e.jpg');
                    console.log('Backend avatar test response:', response.status, response.statusText);
                    if (response.ok) {
                        console.log('✅ Backend avatar endpoint is accessible');
                    } else {
                        console.log('❌ Backend avatar endpoint returned error:', response.status);
                    }
                } catch (error) {
                    console.error('❌ Failed to connect to backend avatar endpoint:', error);
                }
            },
            testContactsData: () => {
                if (window.chatManager && window.chatManager.contacts) {
                    console.log('=== CONTACTS DATA ===');
                    window.chatManager.contacts.forEach(contact => {
                        console.log(`Contact: ${contact.username} (${contact.fullName})`);
                        console.log(`  ID: ${contact.id}`);
                        console.log(`  Profile Picture: ${contact.profilePicture}`);
                        if (contact.profilePicture) {
                            const avatarUrl = window.chatManager.getContactAvatar(contact);
                            console.log(`  Generated Avatar URL: ${avatarUrl}`);
                        }
                        console.log('---');
                    });
                    console.log('===================');
                } else {
                    console.log('No contacts data available');
                }
            }
        };

        console.log('Debug functions added to window:');
        console.log('- printUserData() - prints current user data');
        console.log('- printStoredSession() - prints stored session data');
        console.log('- printAllUserData() - prints all user-related data');
        console.log('- debugAvatars.refreshAvatars() - refresh all avatars');
        console.log('- debugAvatars.showAvatarCache() - show avatar cache');
        console.log('- debugAvatars.testAvatarUrl(contact) - test avatar URL for contact');
    }

    getFallbackAvatar(user) {
        // Generate a simple avatar based on user initials
        const initials = user.fullName 
            ? user.fullName.split(' ').map(n => n[0]).join('').toUpperCase()
            : user.username[0].toUpperCase();
        
        // Create a data URL for a simple colored avatar
        const canvas = document.createElement('canvas');
        canvas.width = 40;
        canvas.height = 40;
        const ctx = canvas.getContext('2d');
        
        // Background color based on user ID
        const colors = ['#3498db', '#e74c3c', '#2ecc71', '#f39c12', '#9b59b6', '#1abc9c'];
        const colorIndex = user.id ? user.id.charCodeAt(0) % colors.length : 0;
        ctx.fillStyle = colors[colorIndex];
        ctx.fillRect(0, 0, 40, 40);
        
        // Text
        ctx.fillStyle = 'white';
        ctx.font = '16px Arial';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(initials, 20, 20);
        
        return canvas.toDataURL();
    }

    getCurrentUser() {
        return this.currentUser;
    }

    isLoggedIn() {
        return this.isAuthenticated;
    }
}

// Initialize auth manager
const authManager = new AuthManager();

// Export for use in other modules
window.authManager = authManager;
