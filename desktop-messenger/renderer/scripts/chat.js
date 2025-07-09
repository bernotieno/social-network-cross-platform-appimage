class ChatManager {
    constructor() {
        this.contacts = [];
        this.selectedContact = null;
        this.messages = [];
        this.onlineUsers = new Set();
        this.typingUsers = new Map();
        this.currentRoomId = null;
        this.isTyping = false;
        this.typingTimeout = null;
        this.apiBaseUrl = 'http://localhost:8080/api';
        this.messageQueue = []; // For offline messages
        this.isOnline = navigator.onLine;
        this.isInitialized = false;
        this.avatarCache = new Map(); // Initialize avatar cache early

        // Don't call init() here - it will be called from main.js
    }

    async init() {
        if (this.isInitialized) return;

        // Initialize storage
        await storageManager.init();

        // Set up event listeners
        this.setupEventListeners();

        // Monitor network status
        this.setupNetworkMonitoring();

        // Load cached data
        await this.loadCachedData();

        // Load failed messages and retry if online
        await this.loadFailedMessagesOnStartup();

        this.isInitialized = true;
    }

    setupEventListeners() {
        // WebSocket events
        wsManager.on('connect', () => {
            console.log('WebSocket connected');
            this.updateConnectionStatus(true);
        });

        wsManager.on('disconnect', () => {
            console.log('WebSocket disconnected');
            this.updateConnectionStatus(false);
        });

        wsManager.on('message', (data) => {
            this.handleWebSocketMessage(data);
        });

        wsManager.on('new_message', (data) => {
            this.handleWebSocketMessage(data);
        });

        wsManager.on('user_presence', (data) => {
            this.handleUserPresence(data);
        });

        wsManager.on('typing_status', (data) => {
            this.handleTypingStatus(data);
        });

        // UI event listeners
        document.getElementById('search-input').addEventListener('input', (e) => {
            this.filterContacts(e.target.value);
        });

        document.getElementById('message-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.sendMessage();
        });

        document.getElementById('message-input').addEventListener('input', (e) => {
            this.handleTyping(e.target.value);
        });

        // Add keyboard shortcut for emoji picker (Ctrl/Cmd + E)
        document.getElementById('message-input').addEventListener('keydown', (e) => {
            if ((e.ctrlKey || e.metaKey) && e.key === 'e') {
                e.preventDefault();
                console.log('Emoji shortcut triggered (Ctrl/Cmd + E)');
                this.toggleEmojiPicker();
            }
        });

        // Use event delegation for emoji button since it might be in a hidden container initially
        document.addEventListener('click', (e) => {
            if (e.target.id === 'emoji-btn' || e.target.closest('#emoji-btn')) {
                console.log('Emoji button clicked via delegation!');
                e.preventDefault();
                e.stopPropagation();
                this.toggleEmojiPicker();
            }
        });
        console.log('Emoji button event delegation set up');
    }

    setupNetworkMonitoring() {
        window.addEventListener('online', () => {
            this.isOnline = true;
            this.updateOfflineIndicator();
            this.reconnectWebSocket();
            this.flushMessageQueue();
        });

        window.addEventListener('offline', () => {
            this.isOnline = false;
            this.updateOfflineIndicator();
        });

        this.updateOfflineIndicator();
    }

    updateOfflineIndicator() {
        const offlineIndicator = document.getElementById('offline-indicator');
        const offlineNotification = document.getElementById('offline-notification');
        
        if (this.isOnline) {
            offlineIndicator.style.display = 'none';
            offlineNotification.style.display = 'none';
        } else {
            offlineIndicator.style.display = 'block';
            offlineNotification.style.display = 'block';
        }
    }

    async loadCachedData() {
        try {
            // Load cached contacts
            const cachedContacts = await storageManager.getContacts();
            if (cachedContacts.length > 0) {
                this.contacts = cachedContacts;
                this.renderContacts();
            }
        } catch (error) {
            console.error('Error loading cached data:', error);
        }
    }

    async loadContacts() {
        try {
            console.log('Loading contacts...');
            const user = authManager.getCurrentUser();
            if (!user) {
                console.warn('No current user found, cannot load contacts');
                return;
            }
            console.log('Current user:', user);

            // Get the stored session token
            const session = await window.electronAPI.auth.getStoredSession();
            if (!session || !session.token) {
                console.error('No valid session token found');
                return;
            }
            console.log('Session token found, making API calls...');

            const headers = {
                'Authorization': `Bearer ${session.token}`,
                'Content-Type': 'application/json'
            };

            // Get both followers and following
            const [followingResponse, followersResponse, onlineUsersResponse] = await Promise.all([
                fetch(`${this.apiBaseUrl}/users/${user.id}/following`, {
                    headers,
                    credentials: 'include'
                }),
                fetch(`${this.apiBaseUrl}/users/${user.id}/followers`, {
                    headers,
                    credentials: 'include'
                }),
                fetch(`${this.apiBaseUrl}/messages/online-users`, {
                    headers,
                    credentials: 'include'
                })
            ]);

            const followingData = await followingResponse.json();
            const followersData = await followersResponse.json();
            const onlineUsersData = await onlineUsersResponse.json();

            console.log('API responses:', {
                following: followingData,
                followers: followersData,
                onlineUsers: onlineUsersData
            });

            // Combine and deduplicate contacts
            const contactsMap = new Map();

            if (followingData.success && followingData.data?.following) {
                followingData.data.following.forEach(user => {
                    contactsMap.set(user.id, user);
                });
            }

            if (followersData.success && followersData.data?.followers) {
                followersData.data.followers.forEach(user => {
                    contactsMap.set(user.id, user);
                });
            }

            this.contacts = Array.from(contactsMap.values());
            console.log('Loaded contacts:', this.contacts);

            // Update online users
            if (onlineUsersData.success && onlineUsersData.onlineUsers) {
                this.onlineUsers = new Set(onlineUsersData.onlineUsers.map(user => user.id));
            }

            // Preload avatars for better performance
            this.preloadAvatars(this.contacts);

            // Cache contacts
            await storageManager.saveContacts(this.contacts);

            this.renderContacts();
        } catch (error) {
            console.error('Error loading contacts:', error);
            // Use cached contacts if available
            const cachedContacts = await storageManager.getContacts();
            if (cachedContacts.length > 0) {
                this.contacts = cachedContacts;
                this.renderContacts();
            }
        }
    }

    renderContacts() {
        const contactsList = document.getElementById('contacts-list');
        contactsList.innerHTML = '';

        if (this.contacts.length === 0) {
            contactsList.innerHTML = `
                <div class="no-contacts">
                    <p>No contacts found</p>
                    <p>Follow people to start chatting</p>
                </div>
            `;
            return;
        }

        this.contacts.forEach(contact => {
            const contactElement = this.createContactElement(contact);
            contactsList.appendChild(contactElement);
        });
    }

    createContactElement(contact) {
        const div = document.createElement('div');
        div.className = 'contact-item';
        div.dataset.contactId = contact.id;
        
        const isOnline = this.onlineUsers.has(contact.id);
        const roomId = this.getRoomId(authManager.getCurrentUser().id, contact.id);
        const isTyping = this.typingUsers.has(roomId) && this.typingUsers.get(roomId).has(contact.id);
        
        const avatarUrl = this.getContactAvatar(contact);
        const fallbackUrl = Utils.generateFallbackAvatar(contact);

        div.innerHTML = `
            <div class="contact-avatar-container">
                <img class="contact-avatar" src="${avatarUrl}" alt="${contact.username}"
                     data-fallback="${fallbackUrl}" data-contact-id="${contact.id}">
                ${isOnline ? '<div class="online-indicator"></div>' : ''}
            </div>
            <div class="contact-info">
                <h3 class="contact-name">${contact.fullName || contact.username}</h3>
                <p class="contact-preview">
                    @${contact.username}
                    ${isTyping ? ' • typing...' : ''}
                </p>
            </div>
        `;

        // Set up proper error handling for the avatar
        const avatarImg = div.querySelector('.contact-avatar');
        avatarImg.onerror = () => {
            console.log('Avatar failed to load for', contact.username, 'using fallback');
            avatarImg.src = fallbackUrl;
            avatarImg.onerror = null; // Prevent infinite loop
        };

        div.addEventListener('click', () => {
            this.selectContact(contact);
        });

        return div;
    }

    getContactAvatar(contact) {
        // Check cache first
        if (this.avatarCache && this.avatarCache.has(contact.id)) {
            return this.avatarCache.get(contact.id);
        }

        if (contact.profilePicture) {
            // Use the API base URL instead of hardcoded localhost
            const avatarUrl = `${this.apiBaseUrl.replace('/api', '')}${contact.profilePicture}`;
            console.log('Generated avatar URL for', contact.username, ':', avatarUrl);
            console.log('Contact profilePicture:', contact.profilePicture);
            console.log('API base URL:', this.apiBaseUrl);
            return avatarUrl;
        }

        // Generate fallback avatar immediately
        const fallbackUrl = Utils.generateFallbackAvatar(contact);
        console.log('Generated fallback avatar for', contact.username);
        return fallbackUrl;
    }

    preloadAvatars(contacts) {
        console.log('Preloading avatars for', contacts.length, 'contacts');
        if (!this.avatarCache) {
            this.avatarCache = new Map();
        }

        contacts.forEach(contact => {
            if (contact.profilePicture) {
                const avatarUrl = `${this.apiBaseUrl.replace('/api', '')}${contact.profilePicture}`;
                console.log('Preloading avatar for', contact.username, 'from URL:', avatarUrl);
                const img = new Image();

                img.onload = () => {
                    this.avatarCache.set(contact.id, avatarUrl);
                    console.log('✅ Avatar preloaded successfully for:', contact.username);
                    // Update any existing avatar elements
                    this.updateExistingAvatars(contact.id, avatarUrl);
                };

                img.onerror = (error) => {
                    console.error('❌ Avatar failed to preload for:', contact.username, 'Error:', error);
                    console.log('Failed URL was:', avatarUrl);
                    const fallbackUrl = Utils.generateFallbackAvatar(contact);
                    this.avatarCache.set(contact.id, fallbackUrl);
                    // Update any existing avatar elements with fallback
                    this.updateExistingAvatars(contact.id, fallbackUrl);
                };

                img.src = avatarUrl;
            } else {
                // Generate and cache fallback avatar immediately
                const fallbackUrl = Utils.generateFallbackAvatar(contact);
                this.avatarCache.set(contact.id, fallbackUrl);
                console.log('Generated fallback avatar for:', contact.username, '(no profilePicture)');
            }
        });
    }

    updateExistingAvatars(contactId, avatarUrl) {
        // Update contact list avatars
        const contactAvatars = document.querySelectorAll(`[data-contact-id="${contactId}"]`);
        contactAvatars.forEach(avatar => {
            avatar.src = avatarUrl;
        });

        // Update chat header avatar if this contact is currently selected
        if (this.selectedContact && this.selectedContact.id === contactId) {
            const chatHeaderAvatar = document.getElementById('contact-avatar');
            if (chatHeaderAvatar) {
                chatHeaderAvatar.src = avatarUrl;
            }
        }
    }

    // Force refresh all avatars (useful for debugging)
    refreshAllAvatars() {
        console.log('Refreshing all avatars...');
        this.avatarCache.clear();
        this.preloadAvatars(this.contacts);
        this.renderContacts();

        // Refresh current chat header avatar
        if (this.selectedContact) {
            this.updateChatHeader(this.selectedContact);
        }
    }

    async selectContact(contact) {
        console.log(`Selecting contact: ${contact.username} (ID: ${contact.id})`);

        // Update UI
        document.querySelectorAll('.contact-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-contact-id="${contact.id}"]`).classList.add('active');

        this.selectedContact = contact;
        this.currentRoomId = this.getRoomId(authManager.getCurrentUser().id, contact.id);
        console.log(`Room ID set to: ${this.currentRoomId}`);

        // Show chat container
        document.getElementById('no-chat-selected').style.display = 'none';
        document.getElementById('chat-container').style.display = 'flex';

        // Update chat header
        this.updateChatHeader(contact);

        // Join WebSocket room
        if (wsManager.isConnected) {
            wsManager.joinRoom(this.currentRoomId);
            console.log(`Joined WebSocket room: ${this.currentRoomId}`);
        } else {
            console.log('WebSocket not connected, skipping room join');
        }

        // Load messages
        console.log('Starting to load messages...');
        await this.loadMessages(contact.id);
        console.log('Finished loading messages');
    }

    updateChatHeader(contact) {
        document.getElementById('contact-name').textContent = contact.fullName || contact.username;

        // Set up avatar with error handling
        const avatarElement = document.getElementById('contact-avatar');
        const avatarUrl = this.getContactAvatar(contact);
        const fallbackUrl = Utils.generateFallbackAvatar(contact);

        console.log('Updating chat header avatar for', contact.username, 'with URL:', avatarUrl);

        // Remove any existing error handlers
        avatarElement.onerror = null;

        // Set up error handler for fallback
        avatarElement.onerror = () => {
            console.log('Chat header avatar failed to load, using fallback for:', contact.username);
            avatarElement.src = fallbackUrl;
            avatarElement.onerror = null; // Prevent infinite loop
        };

        avatarElement.src = avatarUrl;

        const isOnline = this.onlineUsers.has(contact.id);
        const roomId = this.getRoomId(authManager.getCurrentUser().id, contact.id);
        const isTyping = this.typingUsers.has(roomId) && this.typingUsers.get(roomId).has(contact.id);

        let statusText = '@' + contact.username;
        if (isTyping) {
            statusText += ' • typing...';
        } else if (isOnline) {
            statusText += ' • Online';
        }

        document.getElementById('contact-status').textContent = statusText;
    }

    async loadMessages(contactId) {
        console.log(`Loading messages for contact ${contactId}, room ${this.currentRoomId}`);

        try {
            // First, load cached messages
            const cachedMessages = await storageManager.getMessages(this.currentRoomId);
            console.log(`Found ${cachedMessages.length} cached messages for room ${this.currentRoomId}`);

            if (cachedMessages.length > 0) {
                this.messages = cachedMessages;
                this.renderMessages();
                console.log('Rendered cached messages');
            }

            // Always try to fetch fresh messages from API, regardless of online status
            // This ensures we get the latest messages if the connection is available
            try {
                // Get the stored session token
                const session = await window.electronAPI.auth.getStoredSession();
                if (!session || !session.token) {
                    console.warn('No valid session token found for loading messages, using cached messages only');
                    // If we have cached messages, that's fine, otherwise show empty state
                    if (cachedMessages.length === 0) {
                        this.messages = [];
                        this.renderMessages();
                    }
                    return;
                }

                const headers = {
                    'Authorization': `Bearer ${session.token}`,
                    'Content-Type': 'application/json'
                };

                console.log(`Fetching fresh messages from API for contact ${contactId}`);
                const response = await fetch(`${this.apiBaseUrl}/messages/${contactId}`, {
                    headers,
                    credentials: 'include'
                });

                if (response.ok) {
                    const data = await response.json();
                    console.log('API response received:', data);
                    console.log('API response data.success:', data.success);
                    console.log('API response data.messages length:', data.messages ? data.messages.length : 'undefined');

                    if (data.success && data.messages) {
                        console.log('Processing API messages...');
                        console.log('Raw API messages (first 3):', data.messages.slice(0, 3));

                        // Convert to our format
                        const formattedMessages = data.messages.map((msg, index) => {
                            const formatted = {
                                senderId: msg.senderId || msg.sender_id,
                                content: msg.content,
                                timestamp: msg.createdAt || msg.created_at,
                                roomId: this.currentRoomId
                            };
                            if (index < 3) {
                                console.log(`Formatted message ${index}:`, formatted);
                            }
                            return formatted;
                        });

                        console.log(`Formatted ${formattedMessages.length} messages, reversing order...`);
                        this.messages = formattedMessages.reverse(); // API returns newest first
                        console.log(`After reverse, messages length: ${this.messages.length}`);
                        console.log('Final messages (first 3):', this.messages.slice(0, 3));

                        // Cache messages
                        console.log('Saving messages to cache...');
                        await storageManager.saveMessages(this.currentRoomId, this.messages);
                        console.log('Messages saved to cache');

                        this.renderMessages();
                        console.log('Rendered fresh messages from API');
                    } else {
                        console.warn('API response indicates failure or no messages:', data);
                        console.warn('data.success:', data.success);
                        console.warn('data.messages:', data.messages);
                        // Keep using cached messages if API doesn't return valid data
                        if (cachedMessages.length === 0) {
                            this.messages = [];
                            this.renderMessages();
                        }
                    }
                } else {
                    console.error(`API request failed with status ${response.status}: ${response.statusText}`);
                    // Keep using cached messages if API fails
                    if (cachedMessages.length === 0) {
                        this.messages = [];
                        this.renderMessages();
                    }
                }
            } catch (apiError) {
                console.error('Error fetching messages from API:', apiError);
                // Keep using cached messages if API call fails
                if (cachedMessages.length === 0) {
                    this.messages = [];
                    this.renderMessages();
                }
            }
        } catch (error) {
            console.error('Error loading messages:', error);
            // Fallback to empty messages if everything fails
            this.messages = [];
            this.renderMessages();
        }
    }

    renderMessages() {
        const messagesList = document.getElementById('messages-list');
        messagesList.innerHTML = '';

        console.log(`Rendering ${this.messages.length} messages for room ${this.currentRoomId}`);

        if (this.messages.length === 0) {
            messagesList.innerHTML = `
                <div class="no-messages">
                    <p>No messages yet</p>
                    <p>Send a message to start the conversation</p>
                </div>
            `;
            console.log('Displayed no messages placeholder');
            return;
        }

        this.messages.forEach(message => {
            const messageElement = this.createMessageElement(message);
            messagesList.appendChild(messageElement);
        });

        console.log(`Successfully rendered ${this.messages.length} messages`);

        // Scroll to bottom
        this.scrollToBottom();
    }

    createMessageElement(message) {
        const div = document.createElement('div');
        const isOwn = message.senderId === authManager.getCurrentUser().id;

        div.className = `message ${isOwn ? 'own' : 'other'}`;
        div.setAttribute('data-message-id', message.id || '');

        // Add status classes
        const status = message.status || (message.pending ? 'sending' : 'sent');
        div.classList.add(`status-${status}`);

        const time = new Date(message.timestamp).toLocaleTimeString([], {
            hour: '2-digit',
            minute: '2-digit'
        });

        let statusIndicator = '';
        let retryButton = '';

        // Add status indicators for own messages
        if (isOwn) {
            switch (status) {
                case 'sending':
                    statusIndicator = '<span class="message-status sending">⏳ Sending...</span>';
                    break;
                case 'queued':
                    statusIndicator = '<span class="message-status queued">📤 Queued</span>';
                    break;
                case 'sent':
                    statusIndicator = '<span class="message-status sent">✓ Sent</span>';
                    break;
                case 'failed':
                    statusIndicator = `<span class="message-status failed">❌ Failed${message.errorMessage ? `: ${message.errorMessage}` : ''}</span>`;
                    retryButton = `<button class="retry-btn" onclick="chatManager.retryMessage('${message.id}')" title="Retry sending message">🔄 Retry</button>`;
                    break;
            }
        }

        div.innerHTML = `
            <div class="message-content">${this.escapeHtml(message.content)}</div>
            <div class="message-footer">
                <div class="message-time">${time}</div>
                ${statusIndicator}
                ${retryButton}
            </div>
        `;

        return div;
    }

    async sendMessage() {
        const input = document.getElementById('message-input');
        const content = input.value.trim();

        if (!content || !this.selectedContact) return;

        // Clear input immediately
        input.value = '';

        // Stop typing indicator
        this.stopTyping();

        const messageId = this.generateMessageId();
        const message = {
            id: messageId,
            senderId: authManager.getCurrentUser().id,
            content: content,
            timestamp: new Date().toISOString(),
            roomId: this.currentRoomId,
            status: this.isOnline ? 'sending' : 'queued', // 'sending', 'sent', 'failed', 'queued'
            retryCount: 0
        };

        // Add to messages immediately for optimistic UI
        this.messages.push(message);
        this.renderMessages();

        if (this.isOnline) {
            await this.attemptSendMessage(message);
        } else {
            // Add to queue for when back online
            this.addToMessageQueue(message);
            Utils.showToast('Message queued - will send when online', 'warning');
        }
    }

    generateMessageId() {
        return `msg_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;
    }

    async attemptSendMessage(message, isRetry = false) {
        try {
            // Get the stored session token
            const session = await window.electronAPI.auth.getStoredSession();
            if (!session || !session.token) {
                console.error('No valid session token found for sending message');
                this.markMessageAsFailed(message.id, 'Authentication failed');
                return;
            }

            // Update status to sending if it's a retry
            if (isRetry) {
                message.status = 'sending';
                this.renderMessages();
            }

            // Send via API
            const response = await fetch(`${this.apiBaseUrl}/messages`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${session.token}`,
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    receiverId: this.selectedContact.id,
                    content: message.content
                })
            });

            if (response.ok) {
                // Send via WebSocket for real-time delivery
                if (wsManager.isConnected) {
                    wsManager.sendMessage(this.currentRoomId, message.content);
                }

                // Mark as sent
                message.status = 'sent';
                this.renderMessages();

                // Cache message
                await storageManager.saveMessage(this.currentRoomId, message);

                // Remove from failed messages storage if it was there
                await this.removeFromFailedMessages(message.id);

                if (isRetry) {
                    Utils.showToast('Message sent successfully', 'success');
                }
            } else {
                throw new Error(`Server error: ${response.status}`);
            }
        } catch (error) {
            console.error('Error sending message:', error);
            this.markMessageAsFailed(message.id, error.message);

            // Add to persistent failed messages storage
            await this.saveFailedMessage(message);
        }
    }

    markMessageAsFailed(messageId, errorMessage = 'Failed to send') {
        const message = this.messages.find(msg => msg.id === messageId);
        if (message) {
            message.status = 'failed';
            message.errorMessage = errorMessage;
            this.renderMessages();
            Utils.showToast(`Message failed to send: ${errorMessage}`, 'error');
        }
    }

    addToMessageQueue(message) {
        this.messageQueue.push({
            id: message.id,
            receiverId: this.selectedContact.id,
            content: message.content,
            timestamp: message.timestamp,
            roomId: message.roomId
        });
    }

    async saveFailedMessage(message) {
        try {
            const failedMessages = await this.getFailedMessages();
            failedMessages.push({
                id: message.id,
                senderId: message.senderId,
                receiverId: this.selectedContact.id,
                content: message.content,
                timestamp: message.timestamp,
                roomId: message.roomId,
                retryCount: message.retryCount || 0,
                errorMessage: message.errorMessage
            });
            await window.electronAPI.secureStorage.setItem('failedMessages', JSON.stringify(failedMessages));
        } catch (error) {
            console.error('Error saving failed message:', error);
        }
    }

    async getFailedMessages() {
        try {
            const result = await window.electronAPI.secureStorage.getItem('failedMessages');
            return result.value ? JSON.parse(result.value) : [];
        } catch (error) {
            console.error('Error getting failed messages:', error);
            return [];
        }
    }

    async removeFromFailedMessages(messageId) {
        try {
            const failedMessages = await this.getFailedMessages();
            const updatedMessages = failedMessages.filter(msg => msg.id !== messageId);
            await window.electronAPI.secureStorage.setItem('failedMessages', JSON.stringify(updatedMessages));
        } catch (error) {
            console.error('Error removing failed message:', error);
        }
    }

    async retryMessage(messageId) {
        const message = this.messages.find(msg => msg.id === messageId);
        if (!message) {
            console.error('Message not found for retry:', messageId);
            return;
        }

        if (!this.isOnline) {
            Utils.showToast('Cannot retry - you are offline', 'warning');
            return;
        }

        message.retryCount = (message.retryCount || 0) + 1;
        if (message.retryCount > 3) {
            Utils.showToast('Maximum retry attempts reached', 'error');
            return;
        }

        Utils.showToast('Retrying message...', 'info');
        await this.attemptSendMessage(message, true);
    }

    async loadFailedMessagesOnStartup() {
        try {
            const failedMessages = await this.getFailedMessages();

            if (failedMessages.length > 0) {
                console.log(`Found ${failedMessages.length} failed messages from previous session`);

                // Show notification about failed messages
                Utils.showToast(`${failedMessages.length} messages failed to send previously`, 'warning', 5000);

                // If online, attempt to retry them
                if (this.isOnline) {
                    setTimeout(() => {
                        this.retryFailedMessages();
                    }, 2000); // Wait 2 seconds before retrying
                }
            }
        } catch (error) {
            console.error('Error loading failed messages on startup:', error);
        }
    }

    async flushMessageQueue() {
        if (!this.isOnline) return;

        // Process regular message queue
        if (this.messageQueue.length > 0) {
            const queue = [...this.messageQueue];
            this.messageQueue = [];

            for (const queuedMessage of queue) {
                const message = this.messages.find(msg => msg.id === queuedMessage.id);
                if (message) {
                    await this.attemptSendMessage(message, true);
                }
            }
        }

        // Process failed messages
        await this.retryFailedMessages();
    }

    async retryFailedMessages() {
        try {
            const failedMessages = await this.getFailedMessages();

            for (const failedMsg of failedMessages) {
                // Find the message in current messages
                const message = this.messages.find(msg => msg.id === failedMsg.id);
                if (message && message.status === 'failed') {
                    // Only retry if not exceeded max attempts
                    if ((failedMsg.retryCount || 0) < 3) {
                        await this.attemptSendMessage(message, true);
                    }
                }
            }
        } catch (error) {
            console.error('Error retrying failed messages:', error);
        }
    }

    handleWebSocketMessage(data) {
        // Handle both 'message' and 'new_message' types
        if ((data.type === 'message' || data.type === 'new_message') && data.payload) {
            const messageData = data.payload.message;
            const roomId = data.payload.roomId;

            if (roomId === this.currentRoomId) {
                const message = {
                    senderId: messageData.sender,
                    content: messageData.content,
                    timestamp: messageData.timestamp,
                    roomId: roomId
                };

                // Check if message already exists (avoid duplicates)
                const exists = this.messages.some(msg =>
                    msg.content === message.content &&
                    msg.senderId === message.senderId &&
                    Math.abs(new Date(msg.timestamp) - new Date(message.timestamp)) < 5000
                );

                if (!exists) {
                    this.messages.push(message);
                    this.renderMessages();

                    // Cache message
                    storageManager.saveMessage(roomId, message);

                    // Show notification if not focused
                    if (message.senderId !== authManager.getCurrentUser().id) {
                        this.showNotification(message);
                    }
                }
            }
        }
    }

    handleUserPresence(data) {
        if (data.userId && data.status) {
            if (data.status === 'online') {
                this.onlineUsers.add(data.userId);
            } else {
                this.onlineUsers.delete(data.userId);
            }
            
            // Update UI
            this.renderContacts();
            if (this.selectedContact && this.selectedContact.id === data.userId) {
                this.updateChatHeader(this.selectedContact);
            }
        }
    }

    handleTypingStatus(data) {
        if (data.userId && data.roomId) {
            if (!this.typingUsers.has(data.roomId)) {
                this.typingUsers.set(data.roomId, new Set());
            }

            const roomTypingUsers = this.typingUsers.get(data.roomId);
            
            if (data.isTyping) {
                roomTypingUsers.add(data.userId);
            } else {
                roomTypingUsers.delete(data.userId);
            }

            if (roomTypingUsers.size === 0) {
                this.typingUsers.delete(data.roomId);
            }

            // Update UI
            this.renderContacts();
            if (this.selectedContact && data.roomId === this.currentRoomId) {
                this.updateChatHeader(this.selectedContact);
            }
        }
    }

    handleTyping(value) {
        if (!this.selectedContact || !this.currentRoomId) return;

        if (value.trim() && !this.isTyping) {
            this.isTyping = true;
            if (wsManager.isConnected) {
                wsManager.sendTypingStatus(this.currentRoomId, true);
            }
        } else if (!value.trim() && this.isTyping) {
            this.stopTyping();
        }

        // Clear existing timeout
        if (this.typingTimeout) {
            clearTimeout(this.typingTimeout);
        }

        // Set timeout to stop typing after 3 seconds
        if (value.trim()) {
            this.typingTimeout = setTimeout(() => {
                this.stopTyping();
            }, 3000);
        }
    }

    stopTyping() {
        if (this.isTyping) {
            this.isTyping = false;
            if (wsManager.isConnected && this.currentRoomId) {
                wsManager.sendTypingStatus(this.currentRoomId, false);
            }
        }
        
        if (this.typingTimeout) {
            clearTimeout(this.typingTimeout);
            this.typingTimeout = null;
        }
    }

    toggleEmojiPicker() {
        console.log('toggleEmojiPicker called');
        const emojiPicker = document.getElementById('emoji-picker');

        if (!emojiPicker) {
            console.error('Emoji picker element not found');
            return;
        }

        const isVisible = emojiPicker.style.display !== 'none';
        console.log('Emoji picker current visibility:', isVisible);

        if (isVisible) {
            emojiPicker.style.display = 'none';
            console.log('Hiding emoji picker');
        } else {
            this.renderEmojiPicker();
            emojiPicker.style.display = 'grid';
            console.log('Showing emoji picker');
        }
    }

    renderEmojiPicker() {
        console.log('Rendering emoji picker');
        const emojiPicker = document.getElementById('emoji-picker');

        if (!emojiPicker) {
            console.error('Emoji picker container not found');
            return;
        }

        const emojis = ['😀', '😂', '😍', '🤔', '👍', '👎', '❤️', '🎉', '🔥', '💯', '😊', '😎', '🤗', '😘', '🥰', '😋', '🤪', '😜', '🙃', '😇', '🤩', '🥳', '😴', '🤤', '😪', '😵', '🤯', '🤠', '🥸', '😈', '👿', '💀'];

        emojiPicker.innerHTML = '';
        emojis.forEach(emoji => {
            const button = document.createElement('button');
            button.className = 'emoji-btn';
            button.textContent = emoji;
            button.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                console.log('Emoji clicked:', emoji);
                this.insertEmoji(emoji);
                this.toggleEmojiPicker();
            });
            emojiPicker.appendChild(button);
        });
        console.log('Emoji picker rendered with', emojis.length, 'emojis');
    }

    insertEmoji(emoji) {
        const input = document.getElementById('message-input');
        input.value += emoji;
        input.focus();
    }

    filterContacts(query) {
        const contactItems = document.querySelectorAll('.contact-item');
        
        contactItems.forEach(item => {
            const contactName = item.querySelector('.contact-name').textContent.toLowerCase();
            const contactUsername = item.querySelector('.contact-preview').textContent.toLowerCase();
            
            if (contactName.includes(query.toLowerCase()) || contactUsername.includes(query.toLowerCase())) {
                item.style.display = 'flex';
            } else {
                item.style.display = 'none';
            }
        });
    }

    async showNotification(message) {
        if (!('Notification' in window)) return;

        if (Notification.permission === 'granted') {
            const contact = this.contacts.find(c => c.id === message.senderId);
            const title = contact ? contact.fullName || contact.username : 'New Message';
            
            new Notification(title, {
                body: message.content,
                icon: contact ? this.getContactAvatar(contact) : null,
                tag: 'message-' + message.senderId
            });
        } else if (Notification.permission !== 'denied') {
            const permission = await Notification.requestPermission();
            if (permission === 'granted') {
                this.showNotification(message);
            }
        }
    }

    updateConnectionStatus(isConnected) {
        // Update UI to show connection status
        const userStatus = document.getElementById('user-status');
        if (userStatus) {
            userStatus.textContent = isConnected ? 'Online' : 'Connecting...';
        }
    }

    async reconnectWebSocket() {
        if (authManager.isLoggedIn()) {
            const session = await window.electronAPI.auth.getStoredSession();
            if (session && session.token) {
                wsManager.connect(session.token);
            }
        }
    }

    getRoomId(userId1, userId2) {
        return [userId1, userId2].sort().join('-');
    }

    scrollToBottom() {
        const messagesContainer = document.getElementById('messages-container');
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Public methods for external use
    async start() {
        // Initialize if not already done
        if (!this.isInitialized) {
            await this.init();
        }

        if (authManager.isLoggedIn()) {
            await this.loadContacts();

            // Connect WebSocket
            const session = await window.electronAPI.auth.getStoredSession();
            if (session && session.token) {
                wsManager.connect(session.token);
            }
        }
    }

    async searchMessages(query) {
        const results = await storageManager.searchMessages(query, this.currentRoomId);

        // Ensure results have IDs for navigation
        return results.map(message => {
            if (!message.id) {
                // Generate a temporary ID based on content and timestamp for older messages
                message.id = `temp_${btoa(message.content + message.timestamp).replace(/[^a-zA-Z0-9]/g, '').substring(0, 16)}`;
            }
            return message;
        });
    }

    // Method to find message element by content and timestamp (for messages without IDs)
    findMessageElement(message) {
        const messageElements = document.querySelectorAll('.message');

        for (const element of messageElements) {
            const messageId = element.getAttribute('data-message-id');
            if (messageId === message.id) {
                return element;
            }

            // Fallback: match by content and approximate timestamp
            const contentElement = element.querySelector('.message-content');
            const timeElement = element.querySelector('.message-time');

            if (contentElement && timeElement) {
                const elementContent = contentElement.textContent.trim();
                const elementTime = timeElement.textContent;

                if (elementContent === message.content) {
                    // Check if timestamps are close (within 1 minute)
                    const messageTime = new Date(message.timestamp);
                    const displayTime = messageTime.toLocaleTimeString([], {
                        hour: '2-digit',
                        minute: '2-digit'
                    });

                    if (elementTime === displayTime) {
                        return element;
                    }
                }
            }
        }

        return null;
    }
}

// Create global instance
const chatManager = new ChatManager();

// Export for use in other modules
window.chatManager = chatManager;
