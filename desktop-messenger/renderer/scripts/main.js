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
            this.handleOnlineEvent();
        });

        window.addEventListener('offline', () => {
            this.handleOfflineEvent();
        });

        // Handle unload to cleanup
        window.addEventListener('beforeunload', () => {
            this.cleanup();
        });

        // Setup theme toggle functionality
        this.setupThemeToggle();
    }

    setupThemeToggle() {
        const themeToggleBtn = document.getElementById('theme-toggle-btn');
        const themeIcon = document.getElementById('theme-icon');

        if (themeToggleBtn && themeIcon) {
            // Update icon based on current theme
            this.updateThemeIcon();

            // Listen for theme changes
            window.addEventListener('themeChanged', (e) => {
                this.updateThemeIcon();
            });

            // Handle theme toggle click
            themeToggleBtn.addEventListener('click', async () => {
                try {
                    const newTheme = await window.themeManager.toggleTheme();
                    console.log('Theme switched to:', newTheme);
                } catch (error) {
                    console.error('Failed to toggle theme:', error);
                }
            });
        }
    }

    updateThemeIcon() {
        const themeIcon = document.getElementById('theme-icon');
        if (themeIcon && window.themeManager) {
            const isDark = window.themeManager.isDarkMode();
            themeIcon.textContent = isDark ? '☀️' : '🌙';

            const themeToggleBtn = document.getElementById('theme-toggle-btn');
            if (themeToggleBtn) {
                themeToggleBtn.title = isDark ? 'Switch to light mode' : 'Switch to dark mode';
            }
        }
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
        // Setup contact search (sidebar)
        this.setupContactSearch();

        // Setup message search (chat area)
        this.setupMessageSearch();
    }

    setupContactSearch() {
        const contactSearchInput = document.getElementById('contact-search-input');

        contactSearchInput.addEventListener('input', Utils.debounce((e) => {
            this.handleContactSearch(e.target.value);
        }, 300));

        contactSearchInput.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.clearContactSearch();
            }
        });
    }

    setupMessageSearch() {
        const messageSearchInput = document.getElementById('message-search-input');

        // Initialize message search state
        this.messageSearchState = {
            query: '',
            results: [],
            currentIndex: -1,
            isActive: false
        };

        messageSearchInput.addEventListener('input', Utils.debounce((e) => {
            this.handleMessageSearch(e.target.value);
        }, 300));

        // Handle message search keyboard shortcuts
        messageSearchInput.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.clearMessageSearch();
            } else if (e.key === 'Enter') {
                e.preventDefault();
                if (e.shiftKey) {
                    this.navigateMessageSearchResults('prev');
                } else {
                    this.navigateMessageSearchResults('next');
                }
            }
        });

        // Setup message search navigation buttons
        this.setupMessageSearchNavigation();
    }

    async startChat() {
        try {
            console.log('Starting chat...');

            // Initialize chat manager
            await chatManager.start();

            console.log('Chat started successfully');
        } catch (error) {
            console.error('Error starting chat:', error);
            Utils.showToast('Failed to start chat', 'error');
        }
    }

    async handleContactSearch(query) {
        // Only filter contacts in the sidebar
        if (query.length >= 1) {
            chatManager.filterContacts(query);
        } else {
            this.clearContactSearch();
        }
    }

    clearContactSearch() {
        const contactSearchInput = document.getElementById('contact-search-input');
        contactSearchInput.value = '';
        chatManager.filterContacts(''); // Show all contacts
    }

    async handleMessageSearch(query) {
        if (query.length >= 2) {
            await this.performMessageSearch(query);
        } else {
            this.clearMessageSearchResults();
        }
    }

    setupMessageSearchNavigation() {
        const searchPrevBtn = document.getElementById('message-search-prev-btn');
        const searchNextBtn = document.getElementById('message-search-next-btn');
        const searchCloseBtn = document.getElementById('message-search-close-btn');

        searchPrevBtn.addEventListener('click', () => this.navigateMessageSearchResults('prev'));
        searchNextBtn.addEventListener('click', () => this.navigateMessageSearchResults('next'));
        searchCloseBtn.addEventListener('click', () => this.clearMessageSearch());
    }

    async performMessageSearch(query) {
        try {
            this.messageSearchState.query = query;

            // Search in current conversation if one is selected
            if (chatManager.selectedContact) {
                const messages = await chatManager.searchMessages(query);
                this.messageSearchState.results = messages;
                this.messageSearchState.currentIndex = -1;
                this.displayMessageSearchResults(messages);
            } else {
                // If no conversation selected, clear results
                this.messageSearchState.results = [];
                this.hideMessageSearchResults();
            }

        } catch (error) {
            console.error('Error performing message search:', error);
        }
    }

    displayMessageSearchResults(results) {
        const searchResultsPanel = document.getElementById('message-search-results-panel');
        const searchResultsCount = document.getElementById('message-search-results-count');
        const searchResultsList = document.getElementById('message-search-results-list');

        if (results.length > 0) {
            this.messageSearchState.isActive = true;
            searchResultsPanel.style.display = 'block';
            searchResultsCount.textContent = `${results.length} result${results.length > 1 ? 's' : ''}`;

            // Populate search results list
            this.populateMessageSearchResultsList(results);

            // Highlight first result
            if (results.length > 0) {
                this.messageSearchState.currentIndex = 0;
                this.highlightMessageSearchResult(0);
            }
        } else {
            this.hideMessageSearchResults();
            Utils.showToast('No messages found', 'info');
        }
    }

    populateMessageSearchResultsList(results) {
        const searchResultsList = document.getElementById('message-search-results-list');
        searchResultsList.innerHTML = '';

        results.forEach((message, index) => {
            const resultItem = document.createElement('div');
            resultItem.className = 'search-result-item';
            resultItem.setAttribute('data-index', index);

            const time = new Date(message.timestamp).toLocaleString();
            const preview = this.getMessagePreview(message.content, this.messageSearchState.query);

            resultItem.innerHTML = `
                <div class="search-result-content">
                    <div class="search-result-preview">${preview}</div>
                    <div class="search-result-time">${time}</div>
                </div>
            `;

            resultItem.addEventListener('click', () => {
                this.messageSearchState.currentIndex = index;
                this.highlightMessageSearchResult(index);
                this.scrollToMessage(message);
            });

            searchResultsList.appendChild(resultItem);
        });
    }

    getMessagePreview(content, query) {
        const maxLength = 100;
        const queryIndex = content.toLowerCase().indexOf(query.toLowerCase());

        if (queryIndex === -1) return Utils.escapeHtml(content.substring(0, maxLength));

        // Show context around the found term
        const start = Math.max(0, queryIndex - 30);
        const end = Math.min(content.length, queryIndex + query.length + 30);
        let preview = content.substring(start, end);

        if (start > 0) preview = '...' + preview;
        if (end < content.length) preview = preview + '...';

        // Highlight the search term
        const regex = new RegExp(`(${Utils.escapeHtml(query)})`, 'gi');
        preview = Utils.escapeHtml(preview).replace(regex, '<mark>$1</mark>');

        return preview;
    }

    navigateSearchResults(direction) {
        if (!this.searchState.isActive || this.searchState.results.length === 0) return;

        const totalResults = this.searchState.results.length;

        if (direction === 'next') {
            this.searchState.currentIndex = (this.searchState.currentIndex + 1) % totalResults;
        } else if (direction === 'prev') {
            this.searchState.currentIndex = this.searchState.currentIndex <= 0
                ? totalResults - 1
                : this.searchState.currentIndex - 1;
        }

        this.highlightSearchResult(this.searchState.currentIndex);
        this.scrollToMessage(this.searchState.results[this.searchState.currentIndex]);
    }

    highlightSearchResult(index) {
        // Update search results list highlighting
        const resultItems = document.querySelectorAll('.search-result-item');
        resultItems.forEach((item, i) => {
            if (i === index) {
                item.classList.add('active');
            } else {
                item.classList.remove('active');
            }
        });

        // Update counter
        const searchResultsCount = document.getElementById('search-results-count');
        const total = this.searchState.results.length;
        searchResultsCount.textContent = `${index + 1} of ${total} result${total > 1 ? 's' : ''}`;
    }

    scrollToMessage(message) {
        // Use ChatManager's findMessageElement method for better accuracy
        const targetElement = chatManager.findMessageElement(message);

        if (targetElement) {
            // Remove previous highlights
            document.querySelectorAll('.message.search-highlighted').forEach(el => {
                el.classList.remove('search-highlighted');
            });

            // Highlight current message
            targetElement.classList.add('search-highlighted');

            // Scroll to message
            targetElement.scrollIntoView({
                behavior: 'smooth',
                block: 'center'
            });

            // Highlight search term within the message
            this.highlightSearchTermInMessage(targetElement, this.searchState.query);

            // Show success feedback
            Utils.showToast('Message found', 'success', 1500);
        } else {
            Utils.showToast('Message not visible in current conversation', 'warning');
        }
    }

    highlightSearchTermInMessage(messageElement, query) {
        const contentElement = messageElement.querySelector('.message-content');
        if (!contentElement) return;

        const originalText = contentElement.textContent;
        const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
        const highlightedText = Utils.escapeHtml(originalText).replace(regex, '<mark class="search-highlight">$1</mark>');

        contentElement.innerHTML = highlightedText;

        // Remove highlight after 3 seconds
        setTimeout(() => {
            if (contentElement.innerHTML.includes('search-highlight')) {
                contentElement.textContent = originalText;
            }
        }, 3000);
    }

    hideSearchResults() {
        // This method is kept for compatibility but the old search panel no longer exists
        // The functionality is now handled by hideMessageSearchResults()
        this.searchState.isActive = false;
        this.searchState.currentIndex = -1;

        // Remove message highlights
        document.querySelectorAll('.message.search-highlighted').forEach(el => {
            el.classList.remove('search-highlighted');
        });
    }

    clearSearchResults() {
        this.searchState = {
            query: '',
            results: [],
            currentIndex: -1,
            isActive: false
        };

        this.hideSearchResults();

        // Reset contact list
        if (chatManager) {
            chatManager.renderContacts();
        }
    }

    clearSearch() {
        // This method is kept for compatibility but now delegates to specific search clearing
        this.clearContactSearch();
        this.clearMessageSearch();
    }

    // Message search methods
    navigateMessageSearchResults(direction) {
        if (!this.messageSearchState.isActive || this.messageSearchState.results.length === 0) return;

        const totalResults = this.messageSearchState.results.length;

        if (direction === 'next') {
            this.messageSearchState.currentIndex = (this.messageSearchState.currentIndex + 1) % totalResults;
        } else {
            this.messageSearchState.currentIndex = this.messageSearchState.currentIndex <= 0
                ? totalResults - 1
                : this.messageSearchState.currentIndex - 1;
        }

        this.highlightMessageSearchResult(this.messageSearchState.currentIndex);
        this.scrollToMessage(this.messageSearchState.results[this.messageSearchState.currentIndex]);
    }

    highlightMessageSearchResult(index) {
        // Remove previous highlights
        document.querySelectorAll('.search-result-item.active').forEach(el => {
            el.classList.remove('active');
        });

        // Highlight current result
        const resultItems = document.querySelectorAll('#message-search-results-list .search-result-item');
        if (resultItems[index]) {
            resultItems[index].classList.add('active');
        }
    }

    clearMessageSearchResults() {
        this.messageSearchState = {
            query: '',
            results: [],
            currentIndex: -1,
            isActive: false
        };

        this.hideMessageSearchResults();
    }

    clearMessageSearch() {
        const messageSearchInput = document.getElementById('message-search-input');
        messageSearchInput.value = '';
        this.clearMessageSearchResults();
    }

    hideMessageSearchResults() {
        const searchResultsPanel = document.getElementById('message-search-results-panel');
        searchResultsPanel.style.display = 'none';

        // Remove message highlights
        document.querySelectorAll('.message.search-highlighted').forEach(el => {
            el.classList.remove('search-highlighted');
        });
    }

    handleKeyboardShortcuts(e) {
        // Ctrl/Cmd + K: Focus contact search
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            const contactSearchInput = document.getElementById('contact-search-input');
            if (contactSearchInput) {
                contactSearchInput.focus();
            }
        }

        // Ctrl/Cmd + F: Focus message search (if chat is open)
        if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
            e.preventDefault();
            const messageSearchInput = document.getElementById('message-search-input');
            if (messageSearchInput && chatManager.selectedContact) {
                messageSearchInput.focus();
            } else {
                // Fallback to contact search if no chat is open
                const contactSearchInput = document.getElementById('contact-search-input');
                if (contactSearchInput) {
                    contactSearchInput.focus();
                }
            }
        }

        // F3 or Ctrl/Cmd + G: Next search result
        if (e.key === 'F3' || ((e.ctrlKey || e.metaKey) && e.key === 'g' && !e.shiftKey)) {
            e.preventDefault();
            if (this.messageSearchState.isActive) {
                this.navigateMessageSearchResults('next');
            }
        }

        // Shift + F3 or Ctrl/Cmd + Shift + G: Previous search result
        if ((e.key === 'F3' && e.shiftKey) || ((e.ctrlKey || e.metaKey) && e.key === 'g' && e.shiftKey)) {
            e.preventDefault();
            if (this.messageSearchState.isActive) {
                this.navigateMessageSearchResults('prev');
            }
        }

        // Escape: Clear search or close modals
        if (e.key === 'Escape') {
            const contactSearchInput = document.getElementById('contact-search-input');
            const messageSearchInput = document.getElementById('message-search-input');

            // Clear contact search if it has value and is not focused
            if (contactSearchInput && contactSearchInput.value && document.activeElement !== contactSearchInput) {
                this.clearContactSearch();
            }

            // Clear message search if it has value and is not focused
            if (messageSearchInput && messageSearchInput.value && document.activeElement !== messageSearchInput) {
                this.clearMessageSearch();
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

    handleOnlineEvent() {
        Utils.showToast('Connection restored', 'success');

        // Update online status in chat manager
        if (chatManager) {
            chatManager.isOnline = true;
            // Flush message queue and retry failed messages
            chatManager.flushMessageQueue();
        }

        // Hide offline indicator
        const offlineIndicator = document.getElementById('offline-indicator');
        if (offlineIndicator) {
            offlineIndicator.style.display = 'none';
        }

        // Reconnect WebSocket if needed
        if (wsManager && !wsManager.isConnected && authManager.isLoggedIn()) {
            authManager.getCurrentSession().then(session => {
                if (session && session.token) {
                    wsManager.connect(session.token);
                }
            });
        }
    }

    handleOfflineEvent() {
        Utils.showToast('Connection lost - messages will be queued', 'warning');

        // Update online status in chat manager
        if (chatManager) {
            chatManager.isOnline = false;
        }

        // Show offline indicator
        const offlineIndicator = document.getElementById('offline-indicator');
        if (offlineIndicator) {
            offlineIndicator.style.display = 'block';
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
