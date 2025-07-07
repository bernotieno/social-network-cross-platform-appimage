class StorageManager {
    constructor() {
        this.dbName = 'SocialNetworkMessenger';
        this.dbVersion = 1;
        this.db = null;
        this.isInitialized = false;
        this.isElectron = typeof window !== 'undefined' && window.electronAPI;
    }

    async init() {
        if (this.isInitialized) return;

        return new Promise((resolve, reject) => {
            const request = indexedDB.open(this.dbName, this.dbVersion);

            request.onerror = () => {
                console.error('Error opening IndexedDB:', request.error);
                reject(request.error);
            };

            request.onsuccess = () => {
                this.db = request.result;
                this.isInitialized = true;
                console.log('IndexedDB initialized successfully');
                resolve();
            };

            request.onupgradeneeded = (event) => {
                const db = event.target.result;

                // Messages store
                if (!db.objectStoreNames.contains('messages')) {
                    const messagesStore = db.createObjectStore('messages', { 
                        keyPath: 'id', 
                        autoIncrement: true 
                    });
                    messagesStore.createIndex('roomId', 'roomId', { unique: false });
                    messagesStore.createIndex('timestamp', 'timestamp', { unique: false });
                    messagesStore.createIndex('senderId', 'senderId', { unique: false });
                    messagesStore.createIndex('content', 'content', { unique: false });
                }

                // Contacts store
                if (!db.objectStoreNames.contains('contacts')) {
                    const contactsStore = db.createObjectStore('contacts', { 
                        keyPath: 'id' 
                    });
                    contactsStore.createIndex('username', 'username', { unique: false });
                    contactsStore.createIndex('fullName', 'fullName', { unique: false });
                }

                // User data store
                if (!db.objectStoreNames.contains('userData')) {
                    db.createObjectStore('userData', { keyPath: 'key' });
                }

                // Settings store
                if (!db.objectStoreNames.contains('settings')) {
                    db.createObjectStore('settings', { keyPath: 'key' });
                }
            };
        });
    }

    async ensureInitialized() {
        if (!this.isInitialized) {
            await this.init();
        }
    }

    // Message operations
    async saveMessage(roomId, message) {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['messages'], 'readwrite');
        const store = transaction.objectStore('messages');
        
        const messageData = {
            roomId,
            senderId: message.senderId || message.sender,
            content: message.content,
            timestamp: message.timestamp || new Date().toISOString(),
            createdAt: message.createdAt || new Date().toISOString()
        };

        return new Promise((resolve, reject) => {
            const request = store.add(messageData);
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async saveMessages(roomId, messages) {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['messages'], 'readwrite');
        const store = transaction.objectStore('messages');
        
        // Clear existing messages for this room first
        await this.clearMessagesForRoom(roomId);
        
        const promises = messages.map(message => {
            const messageData = {
                roomId,
                senderId: message.senderId || message.sender,
                content: message.content,
                timestamp: message.timestamp || new Date().toISOString(),
                createdAt: message.createdAt || new Date().toISOString()
            };
            
            return new Promise((resolve, reject) => {
                const request = store.add(messageData);
                request.onsuccess = () => resolve(request.result);
                request.onerror = () => reject(request.error);
            });
        });

        return Promise.all(promises);
    }

    async getMessages(roomId, limit = 100) {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['messages'], 'readonly');
        const store = transaction.objectStore('messages');
        const index = store.index('roomId');
        
        return new Promise((resolve, reject) => {
            const request = index.getAll(roomId);
            request.onsuccess = () => {
                const messages = request.result
                    .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
                    .slice(-limit); // Get the last N messages
                resolve(messages);
            };
            request.onerror = () => reject(request.error);
        });
    }

    async searchMessages(query, roomId = null) {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['messages'], 'readonly');
        const store = transaction.objectStore('messages');
        
        return new Promise((resolve, reject) => {
            const request = store.getAll();
            request.onsuccess = () => {
                const allMessages = request.result;
                const filteredMessages = allMessages.filter(message => {
                    const matchesQuery = message.content.toLowerCase().includes(query.toLowerCase());
                    const matchesRoom = roomId ? message.roomId === roomId : true;
                    return matchesQuery && matchesRoom;
                });
                
                // Sort by timestamp
                filteredMessages.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
                resolve(filteredMessages);
            };
            request.onerror = () => reject(request.error);
        });
    }

    async clearMessagesForRoom(roomId) {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['messages'], 'readwrite');
        const store = transaction.objectStore('messages');
        const index = store.index('roomId');
        
        return new Promise((resolve, reject) => {
            const request = index.openCursor(roomId);
            request.onsuccess = (event) => {
                const cursor = event.target.result;
                if (cursor) {
                    cursor.delete();
                    cursor.continue();
                } else {
                    resolve();
                }
            };
            request.onerror = () => reject(request.error);
        });
    }

    async clearAllMessages() {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['messages'], 'readwrite');
        const store = transaction.objectStore('messages');
        
        return new Promise((resolve, reject) => {
            const request = store.clear();
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }

    // Contact operations
    async saveContacts(contacts) {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['contacts'], 'readwrite');
        const store = transaction.objectStore('contacts');
        
        // Clear existing contacts first
        await new Promise((resolve, reject) => {
            const clearRequest = store.clear();
            clearRequest.onsuccess = () => resolve();
            clearRequest.onerror = () => reject(clearRequest.error);
        });
        
        const promises = contacts.map(contact => {
            return new Promise((resolve, reject) => {
                const request = store.add(contact);
                request.onsuccess = () => resolve(request.result);
                request.onerror = () => reject(request.error);
            });
        });

        return Promise.all(promises);
    }

    async getContacts() {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['contacts'], 'readonly');
        const store = transaction.objectStore('contacts');
        
        return new Promise((resolve, reject) => {
            const request = store.getAll();
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    // User data operations
    async saveUserData(key, data) {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['userData'], 'readwrite');
        const store = transaction.objectStore('userData');
        
        return new Promise((resolve, reject) => {
            const request = store.put({ key, data });
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async getUserData(key) {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['userData'], 'readonly');
        const store = transaction.objectStore('userData');
        
        return new Promise((resolve, reject) => {
            const request = store.get(key);
            request.onsuccess = () => {
                const result = request.result;
                resolve(result ? result.data : null);
            };
            request.onerror = () => reject(request.error);
        });
    }

    async clearUserData() {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['userData'], 'readwrite');
        const store = transaction.objectStore('userData');
        
        return new Promise((resolve, reject) => {
            const request = store.clear();
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }

    // Settings operations
    async saveSetting(key, value) {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['settings'], 'readwrite');
        const store = transaction.objectStore('settings');
        
        return new Promise((resolve, reject) => {
            const request = store.put({ key, value });
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async getSetting(key, defaultValue = null) {
        await this.ensureInitialized();
        
        const transaction = this.db.transaction(['settings'], 'readonly');
        const store = transaction.objectStore('settings');
        
        return new Promise((resolve, reject) => {
            const request = store.get(key);
            request.onsuccess = () => {
                const result = request.result;
                resolve(result ? result.value : defaultValue);
            };
            request.onerror = () => reject(request.error);
        });
    }
}

// Create global instance
const storageManager = new StorageManager();

// Export for use in other modules
window.storageManager = storageManager;
