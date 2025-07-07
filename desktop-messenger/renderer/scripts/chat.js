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
        
        this.init();
    }

    async init() {
        // Initialize storage
        await storageManager.init();
        
        // Set up event listeners
        this.setupEventListeners();
        
        // Monitor network status
        this.setupNetworkMonitoring();
        
        // Load cached data
        await this.loadCachedData();
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

        document.getElementById('emoji-btn').addEventListener('click', () => {
            this.toggleEmojiPicker();
        });

        document.getElementById('logout-btn').addEventListener('click', () => {
            authManager.logout();
        });
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
            const user = authManager.getCurrentUser();
            if (!user) return;

            // Get both followers and following
            const [followingResponse, followersResponse, onlineUsersResponse] = await Promise.all([
                fetch(`${this.apiBaseUrl}/users/${user.id}/following`, {
                    credentials: 'include'
                }),
                fetch(`${this.apiBaseUrl}/users/${user.id}/followers`, {
                    credentials: 'include'
                }),
                fetch(`${this.apiBaseUrl}/messages/online-users`, {
                    credentials: 'include'
                })
            ]);

            const followingData = await followingResponse.json();
            const followersData = await followersResponse.json();
            const onlineUsersData = await onlineUsersResponse.json();

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
            
            // Update online users
            if (onlineUsersData.success && onlineUsersData.onlineUsers) {
                this.onlineUsers = new Set(onlineUsersData.onlineUsers.map(user => user.id));
            }

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
        
        div.innerHTML = `
            <div class="contact-avatar-container">
                <img class="contact-avatar" src="${this.getContactAvatar(contact)}" alt="${contact.username}">
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

        div.addEventListener('click', () => {
            this.selectContact(contact);
        });

        return div;
    }

    getContactAvatar(contact) {
        if (contact.profilePicture) {
            return `http://localhost:8080${contact.profilePicture}`;
        }
        return authManager.getFallbackAvatar(contact);
    }

    async selectContact(contact) {
        // Update UI
        document.querySelectorAll('.contact-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-contact-id="${contact.id}"]`).classList.add('active');

        this.selectedContact = contact;
        this.currentRoomId = this.getRoomId(authManager.getCurrentUser().id, contact.id);

        // Show chat container
        document.getElementById('no-chat-selected').style.display = 'none';
        document.getElementById('chat-container').style.display = 'flex';

        // Update chat header
        this.updateChatHeader(contact);

        // Join WebSocket room
        if (wsManager.isConnected) {
            wsManager.joinRoom(this.currentRoomId);
        }

        // Load messages
        await this.loadMessages(contact.id);
    }

    updateChatHeader(contact) {
        document.getElementById('contact-name').textContent = contact.fullName || contact.username;
        document.getElementById('contact-avatar').src = this.getContactAvatar(contact);
        
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
        try {
            // First, load cached messages
            const cachedMessages = await storageManager.getMessages(this.currentRoomId);
            if (cachedMessages.length > 0) {
                this.messages = cachedMessages;
                this.renderMessages();
            }

            // Then, fetch fresh messages from API
            if (this.isOnline) {
                const response = await fetch(`${this.apiBaseUrl}/messages/${contactId}`, {
                    credentials: 'include'
                });

                if (response.ok) {
                    const data = await response.json();
                    if (data.success && data.messages) {
                        // Convert to our format
                        const formattedMessages = data.messages.map(msg => ({
                            senderId: msg.senderId || msg.sender_id,
                            content: msg.content,
                            timestamp: msg.createdAt || msg.created_at,
                            roomId: this.currentRoomId
                        }));

                        this.messages = formattedMessages.reverse(); // API returns newest first
                        
                        // Cache messages
                        await storageManager.saveMessages(this.currentRoomId, this.messages);
                        
                        this.renderMessages();
                    }
                }
            }
        } catch (error) {
            console.error('Error loading messages:', error);
        }
    }

    renderMessages() {
        const messagesList = document.getElementById('messages-list');
        messagesList.innerHTML = '';

        if (this.messages.length === 0) {
            messagesList.innerHTML = `
                <div class="no-messages">
                    <p>No messages yet</p>
                    <p>Send a message to start the conversation</p>
                </div>
            `;
            return;
        }

        this.messages.forEach(message => {
            const messageElement = this.createMessageElement(message);
            messagesList.appendChild(messageElement);
        });

        // Scroll to bottom
        this.scrollToBottom();
    }

    createMessageElement(message) {
        const div = document.createElement('div');
        const isOwn = message.senderId === authManager.getCurrentUser().id;
        
        div.className = `message ${isOwn ? 'own' : 'other'}`;
        if (message.pending) {
            div.classList.add('sending');
        }

        const time = new Date(message.timestamp).toLocaleTimeString([], { 
            hour: '2-digit', 
            minute: '2-digit' 
        });

        div.innerHTML = `
            <div class="message-content">${this.escapeHtml(message.content)}</div>
            <div class="message-time">${time}</div>
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

        const message = {
            senderId: authManager.getCurrentUser().id,
            content: content,
            timestamp: new Date().toISOString(),
            roomId: this.currentRoomId,
            pending: !this.isOnline
        };

        // Add to messages immediately for optimistic UI
        this.messages.push(message);
        this.renderMessages();

        if (this.isOnline) {
            try {
                // Send via API
                const response = await fetch(`${this.apiBaseUrl}/messages`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify({
                        receiverId: this.selectedContact.id,
                        content: content
                    })
                });

                if (response.ok) {
                    // Send via WebSocket for real-time delivery
                    if (wsManager.isConnected) {
                        wsManager.sendMessage(this.currentRoomId, content);
                    }

                    // Remove pending flag
                    message.pending = false;
                    this.renderMessages();

                    // Cache message
                    await storageManager.saveMessage(this.currentRoomId, message);
                } else {
                    throw new Error('Failed to send message');
                }
            } catch (error) {
                console.error('Error sending message:', error);
                // Add to queue for retry
                this.messageQueue.push({
                    receiverId: this.selectedContact.id,
                    content: content,
                    timestamp: message.timestamp
                });
            }
        } else {
            // Add to queue for when back online
            this.messageQueue.push({
                receiverId: this.selectedContact.id,
                content: content,
                timestamp: message.timestamp
            });
        }
    }

    async flushMessageQueue() {
        if (!this.isOnline || this.messageQueue.length === 0) return;

        const queue = [...this.messageQueue];
        this.messageQueue = [];

        for (const queuedMessage of queue) {
            try {
                const response = await fetch(`${this.apiBaseUrl}/messages`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify(queuedMessage)
                });

                if (!response.ok) {
                    // Re-add to queue if failed
                    this.messageQueue.push(queuedMessage);
                }
            } catch (error) {
                console.error('Error sending queued message:', error);
                this.messageQueue.push(queuedMessage);
            }
        }
    }

    handleWebSocketMessage(data) {
        if (data.type === 'message' && data.payload) {
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
        const emojiPicker = document.getElementById('emoji-picker');
        const isVisible = emojiPicker.style.display !== 'none';
        
        if (isVisible) {
            emojiPicker.style.display = 'none';
        } else {
            this.renderEmojiPicker();
            emojiPicker.style.display = 'grid';
        }
    }

    renderEmojiPicker() {
        const emojiPicker = document.getElementById('emoji-picker');
        const emojis = ['😀', '😂', '😍', '🤔', '👍', '👎', '❤️', '🎉', '🔥', '💯', '😊', '😎', '🤗', '😘', '🥰', '😋', '🤪', '😜', '🙃', '😇', '🤩', '🥳', '😴', '🤤', '😪', '😵', '🤯', '🤠', '🥸', '😈', '👿', '💀'];
        
        emojiPicker.innerHTML = '';
        emojis.forEach(emoji => {
            const button = document.createElement('button');
            button.className = 'emoji-btn';
            button.textContent = emoji;
            button.addEventListener('click', () => {
                this.insertEmoji(emoji);
                this.toggleEmojiPicker();
            });
            emojiPicker.appendChild(button);
        });
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
        return await storageManager.searchMessages(query, this.currentRoomId);
    }
}

// Create global instance
const chatManager = new ChatManager();

// Export for use in other modules
window.chatManager = chatManager;
