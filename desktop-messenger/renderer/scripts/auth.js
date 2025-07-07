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
                // Validate the session with the backend
                const isValid = await this.validateSession(storedSession.token);
                if (isValid) {
                    this.currentUser = storedSession.user;
                    this.isAuthenticated = true;
                    this.showChatScreen();
                    return;
                }
            }
        } catch (error) {
            console.error('Error initializing auth:', error);
        }

        // Show login screen if no valid session
        this.showLoginScreen();
    }

    async validateSession(token) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/users/profile`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                credentials: 'include'
            });

            if (response.ok) {
                const data = await response.json();
                if (data.success && data.user) {
                    this.currentUser = data.user;
                    return true;
                }
            }
            return false;
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
                this.currentUser = data.user;
                this.isAuthenticated = true;

                // Store session securely
                if (this.isElectron) {
                    await window.electronAPI.auth.login({
                        token: data.token,
                        user: data.user
                    });
                    // Store user data
                    await window.electronAPI.userData.saveUser(data.user);
                } else {
                    // Store in localStorage for web version
                    localStorage.setItem('socialNetworkSession', JSON.stringify({
                        token: data.token,
                        user: data.user
                    }));
                    localStorage.setItem('socialNetworkUser', JSON.stringify(data.user));
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
        document.getElementById('loading-screen').style.display = 'none';
        document.getElementById('login-screen').style.display = 'none';
        document.getElementById('chat-screen').style.display = 'flex';
        
        // Update user info in the sidebar
        if (this.currentUser) {
            document.getElementById('user-name').textContent = this.currentUser.fullName || this.currentUser.username;
            const userAvatar = document.getElementById('user-avatar');
            if (this.currentUser.profilePicture) {
                userAvatar.src = `http://localhost:8080${this.currentUser.profilePicture}`;
            } else {
                userAvatar.src = this.getFallbackAvatar(this.currentUser);
            }
        }
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
