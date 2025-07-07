class WebSocketManager {
    constructor() {
        this.socket = null;
        this.isConnected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // Start with 1 second
        this.messageQueue = [];
        this.eventListeners = new Map();
        this.currentRoomId = null;
        this.wsUrl = 'ws://localhost:8080'; // Default WebSocket URL
    }

    async connect(token) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            return;
        }

        try {
            const url = `${this.wsUrl}/ws?token=${encodeURIComponent(token)}`;
            console.log('Connecting to WebSocket:', url);
            
            this.socket = new WebSocket(url);
            
            this.socket.onopen = () => {
                console.log('WebSocket connected');
                this.isConnected = true;
                this.reconnectAttempts = 0;
                this.reconnectDelay = 1000;
                
                // Send queued messages
                this.flushMessageQueue();
                
                // Trigger connect event
                this.emit('connect');
            };

            this.socket.onclose = (event) => {
                console.log('WebSocket disconnected:', event.code, event.reason);
                this.isConnected = false;
                this.emit('disconnect');
                
                // Attempt to reconnect if not a clean close
                if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
                    this.scheduleReconnect(token);
                }
            };

            this.socket.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.emit('error', error);
            };

            this.socket.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    console.log('WebSocket message received:', data);
                    
                    // Emit the specific event type
                    if (data.type) {
                        this.emit(data.type, data.payload || data);
                    }
                    
                    // Also emit a general message event
                    this.emit('message', data);
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            };

        } catch (error) {
            console.error('Error creating WebSocket connection:', error);
            this.emit('error', error);
        }
    }

    scheduleReconnect(token) {
        this.reconnectAttempts++;
        console.log(`Scheduling reconnect attempt ${this.reconnectAttempts} in ${this.reconnectDelay}ms`);
        
        setTimeout(() => {
            this.connect(token);
        }, this.reconnectDelay);
        
        // Exponential backoff
        this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000);
    }

    disconnect() {
        if (this.socket) {
            this.socket.close(1000, 'User disconnected');
            this.socket = null;
        }
        this.isConnected = false;
        this.currentRoomId = null;
    }

    send(message) {
        if (this.isConnected && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(message));
        } else {
            // Queue message for later
            this.messageQueue.push(message);
            console.log('WebSocket not connected, queuing message:', message);
        }
    }

    flushMessageQueue() {
        while (this.messageQueue.length > 0) {
            const message = this.messageQueue.shift();
            this.send(message);
        }
    }

    joinRoom(roomId) {
        this.currentRoomId = roomId;
        this.send({
            type: 'join_room',
            roomId: roomId
        });
        console.log('Joined room:', roomId);
    }

    leaveRoom(roomId) {
        if (this.currentRoomId === roomId) {
            this.currentRoomId = null;
        }
        this.send({
            type: 'leave_room',
            roomId: roomId
        });
        console.log('Left room:', roomId);
    }

    sendMessage(roomId, content) {
        const message = {
            type: 'message',
            roomId: roomId,
            content: {
                content: content,
                timestamp: new Date().toISOString()
            }
        };
        
        this.send(message);
        console.log('Sent message:', message);
    }

    sendTypingStatus(roomId, isTyping) {
        this.send({
            type: 'typing_status',
            roomId: roomId,
            content: {
                isTyping: isTyping
            }
        });
    }

    // Event system
    on(event, callback) {
        if (!this.eventListeners.has(event)) {
            this.eventListeners.set(event, []);
        }
        this.eventListeners.get(event).push(callback);
    }

    off(event, callback) {
        if (this.eventListeners.has(event)) {
            const listeners = this.eventListeners.get(event);
            const index = listeners.indexOf(callback);
            if (index > -1) {
                listeners.splice(index, 1);
            }
        }
    }

    emit(event, data) {
        if (this.eventListeners.has(event)) {
            this.eventListeners.get(event).forEach(callback => {
                try {
                    callback(data);
                } catch (error) {
                    console.error('Error in event listener:', error);
                }
            });
        }
    }

    // Utility methods
    getRoomId(userId1, userId2) {
        return [userId1, userId2].sort().join('-');
    }

    isConnectedToRoom(roomId) {
        return this.currentRoomId === roomId;
    }

    getConnectionStatus() {
        return {
            isConnected: this.isConnected,
            currentRoom: this.currentRoomId,
            queuedMessages: this.messageQueue.length
        };
    }
}

// Create global instance
const wsManager = new WebSocketManager();

// Export for use in other modules
window.wsManager = wsManager;
