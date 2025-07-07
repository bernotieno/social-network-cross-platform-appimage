// Main application script for Social Network Messenger

class MessengerApp {
    constructor() {
        this.isInitialized = false;
        this.searchTimeout = null;
        this.currentSearchQuery = '';
        this.searchResults = [];
        
        this.init();
    }

    async init() {
        try {
            console.log('Initializing Social Network Messenger...');
            
            // Wait for DOM to be ready
            if (document.readyState === 'loading') {
                document.addEventListener('DOMContentLoaded', () => this.start());
            } else {
                await this.start();
            }
        } catch (error) {
            console.error('Error initializing app:', error);
            this.showError('Failed to initialize application');
        }
    }

    async start() {
        try {
            // Initialize storage first
            await storageManager.init();
            
            // Set up global event listeners
            this.setupGlobalEventListeners();
            
            // Set up login form
            this.setupLoginForm();
            
            // Set up search functionality
            this.setupSearchFunctionality();
            
            // Initialize auth manager (this will show appropriate screen)
            await authManager.init();
            
            // If user is logged in, start chat
            if (authManager.isLoggedIn()) {
                await this.startChat();
            }
            
            this.isInitialized = true;
            console.log('Social Network Messenger initialized successfully');
            
        } catch (error) {
            console.error('Error starting app:', error);
            this.showError('Failed to start application');
        }
    }

    setupGlobalEventListeners() {
        // Handle keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            this.handleKeyboardShortcuts(e);
        });

        // Handle window focus/blur for notifications
        window.addEventListener('focus', () => {
            this.handleWindowFocus();
        });

        window.addEventListener('blur', () => {
            this.handleWindowBlur();
        });

        // Handle online/offline events
        window.addEventListener('online', () => {
            Utils.showToast('Connection restored', 'success');
        });

        window.addEventListener('offline', () => {
            Utils.showToast('Connection lost', 'warning');
        });

        // Handle unload to cleanup
        window.addEventListener('beforeunload', () => {
            this.cleanup();
        });
    }

    setupLoginForm() {
        const loginForm = document.getElementById('login-form');
        const emailInput = document.getElementById('email');
        const passwordInput = document.getElementById('password');
        const loginBtn = document.getElementById('login-btn');
        const registerLink = document.getElementById('register-link');
        const errorDiv = document.getElementById('login-error');

        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const email = emailInput.value.trim();
            const password = passwordInput.value;

            // Validate inputs
            if (!email || !password) {
                this.showLoginError('Please fill in all fields');
                return;
            }

            if (!Utils.isValidEmail(email)) {
                this.showLoginError('Please enter a valid email address');
                return;
            }

            // Show loading state
            loginBtn.disabled = true;
            loginBtn.textContent = 'Signing In...';
            errorDiv.style.display = 'none';

            try {
                const result = await authManager.login(email, password);
                
                if (result.success) {
                    // Start chat after successful login
                    await this.startChat();
                } else {
                    this.showLoginError(result.error);
                }
            } catch (error) {
                console.error('Login error:', error);
                this.showLoginError('An unexpected error occurred');
            } finally {
                loginBtn.disabled = false;
                loginBtn.textContent = 'Sign In';
            }
        });

        registerLink.addEventListener('click', (e) => {
            e.preventDefault();
            authManager.openRegistration();
        });
    }

    setupSearchFunctionality() {
        const searchInput = document.getElementById('search-input');
        
        searchInput.addEventListener('input', Utils.debounce((e) => {
            this.handleSearch(e.target.value);
        }, 300));

        // Handle search keyboard shortcuts
        searchInput.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                searchInput.value = '';
                this.clearSearch();
            }
        });
    }

    async startChat() {
        try {
            console.log('Starting chat...');
            
            // Initialize chat manager
            await chatManager.start();
            
            // Set up message search
            this.setupMessageSearch();
            
            console.log('Chat started successfully');
        } catch (error) {
            console.error('Error starting chat:', error);
            Utils.showToast('Failed to start chat', 'error');
        }
    }

    setupMessageSearch() {
        // Add search functionality to the existing search input
        const searchInput = document.getElementById('search-input');
        
        // Override the existing search to include message search
        searchInput.addEventListener('input', Utils.debounce(async (e) => {
            const query = e.target.value.trim();
            
            if (query.length >= 2) {
                await this.performMessageSearch(query);
            } else {
                this.clearSearch();
            }
        }, 300));
    }

    async performMessageSearch(query) {
        try {
            this.currentSearchQuery = query;
            
            // Search in current conversation if one is selected
            if (chatManager.selectedContact) {
                const messages = await chatManager.searchMessages(query);
                this.displaySearchResults(messages, 'messages');
            }
            
            // Also filter contacts
            chatManager.filterContacts(query);
            
        } catch (error) {
            console.error('Error performing search:', error);
        }
    }

    displaySearchResults(results, type) {
        // For now, we'll just highlight matching messages
        // In a more advanced implementation, we could show a search results panel
        
        if (type === 'messages' && results.length > 0) {
            console.log(`Found ${results.length} messages matching "${this.currentSearchQuery}"`);
            
            // You could implement a search results overlay here
            Utils.showToast(`Found ${results.length} messages`, 'info');
        }
    }

    clearSearch() {
        this.currentSearchQuery = '';
        this.searchResults = [];
        
        // Reset contact list
        if (chatManager) {
            chatManager.renderContacts();
        }
    }

    handleKeyboardShortcuts(e) {
        // Ctrl/Cmd + K: Focus search
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            const searchInput = document.getElementById('search-input');
            if (searchInput) {
                searchInput.focus();
            }
        }

        // Escape: Clear search or close modals
        if (e.key === 'Escape') {
            const searchInput = document.getElementById('search-input');
            if (searchInput && searchInput.value) {
                searchInput.value = '';
                this.clearSearch();
            }
            
            // Close emoji picker if open
            const emojiPicker = document.getElementById('emoji-picker');
            if (emojiPicker && emojiPicker.style.display !== 'none') {
                emojiPicker.style.display = 'none';
            }
        }

        // Ctrl/Cmd + Enter: Send message
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            const messageInput = document.getElementById('message-input');
            if (messageInput && messageInput.value.trim()) {
                chatManager.sendMessage();
            }
        }

        // Ctrl/Cmd + L: Logout
        if ((e.ctrlKey || e.metaKey) && e.key === 'l') {
            e.preventDefault();
            if (authManager.isLoggedIn()) {
                authManager.logout();
            }
        }
    }

    handleWindowFocus() {
        // Clear any notification badges or update presence
        if (chatManager && wsManager.isConnected) {
            // Could implement presence updates here
        }
    }

    handleWindowBlur() {
        // Could implement away status here
    }

    showLoginError(message) {
        const errorDiv = document.getElementById('login-error');
        errorDiv.textContent = message;
        errorDiv.style.display = 'block';
    }

    showError(message) {
        Utils.showToast(message, 'error');
    }

    cleanup() {
        // Cleanup WebSocket connections
        if (wsManager) {
            wsManager.disconnect();
        }

        // Clear any timeouts
        if (this.searchTimeout) {
            clearTimeout(this.searchTimeout);
        }

        // Stop typing indicators
        if (chatManager) {
            chatManager.stopTyping();
        }
    }

    // Public methods for external use
    getCurrentUser() {
        return authManager.getCurrentUser();
    }

    isOnline() {
        return navigator.onLine;
    }

    getConnectionStatus() {
        return {
            online: this.isOnline(),
            websocket: wsManager ? wsManager.getConnectionStatus() : null,
            authenticated: authManager ? authManager.isLoggedIn() : false
        };
    }

    async refreshData() {
        if (authManager.isLoggedIn() && chatManager) {
            await chatManager.loadContacts();
        }
    }

    // Development helpers
    async clearAllData() {
        if (confirm('Are you sure you want to clear all local data? This cannot be undone.')) {
            await storageManager.clearAllMessages();
            await storageManager.clearUserData();
            Utils.showToast('All local data cleared', 'success');
            
            // Logout and restart
            authManager.logout();
        }
    }

    getDebugInfo() {
        return {
            initialized: this.isInitialized,
            user: this.getCurrentUser(),
            connectionStatus: this.getConnectionStatus(),
            searchQuery: this.currentSearchQuery,
            searchResults: this.searchResults.length
        };
    }
}

// Initialize the app when the script loads
const messengerApp = new MessengerApp();

// Export for global access and debugging
window.messengerApp = messengerApp;

// Add some global keyboard shortcuts for development
if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    document.addEventListener('keydown', (e) => {
        // Ctrl+Shift+D: Show debug info
        if (e.ctrlKey && e.shiftKey && e.key === 'D') {
            console.log('Debug Info:', messengerApp.getDebugInfo());
        }
        
        // Ctrl+Shift+C: Clear all data (development only)
        if (e.ctrlKey && e.shiftKey && e.key === 'C') {
            messengerApp.clearAllData();
        }
    });
}

console.log('Social Network Messenger loaded successfully');
