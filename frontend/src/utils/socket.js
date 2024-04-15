import { io } from 'socket.io-client';
import { getToken } from './auth';

let socket = null;

/**
 * Initialize WebSocket connection
 * @returns {object} - Socket.io instance
 */
export const initializeSocket = () => {
  if (!socket) {
    const token = getToken();
    
    socket = io(process.env.NEXT_PUBLIC_SOCKET_URL || 'http://localhost:8080', {
      auth: {
        token
      },
      autoConnect: false,
    });
    
    // Socket event listeners
    socket.on('connect', () => {
      console.log('Socket connected');
    });
    
    socket.on('disconnect', () => {
      console.log('Socket disconnected');
    });
    
    socket.on('connect_error', (error) => {
      console.error('Socket connection error:', error);
    });
    
    // Connect to the socket server
    socket.connect();
  }
  
  return socket;
};

/**
 * Get the socket instance
 * @returns {object|null} - Socket.io instance or null if not initialized
 */
export const getSocket = () => socket;

/**
 * Disconnect the socket
 */
export const disconnectSocket = () => {
  if (socket) {
    socket.disconnect();
    socket = null;
  }
};

/**
 * Subscribe to a chat room
 * @param {string} roomId - Room ID to join
 */
export const joinChatRoom = (roomId) => {
  if (socket) {
    socket.emit('join_room', { roomId });
  }
};

/**
 * Leave a chat room
 * @param {string} roomId - Room ID to leave
 */
export const leaveChatRoom = (roomId) => {
  if (socket) {
    socket.emit('leave_room', { roomId });
  }
};

/**
 * Send a message to a chat room
 * @param {string} roomId - Room ID to send message to
 * @param {string} message - Message content
 */
export const sendMessage = (roomId, message) => {
  if (socket) {
    socket.emit('send_message', { roomId, message });
  }
};

/**
 * Subscribe to new messages in a chat room
 * @param {function} callback - Function to call when a new message is received
 * @returns {function} - Function to unsubscribe
 */
export const subscribeToMessages = (callback) => {
  if (socket) {
    socket.on('new_message', callback);
    return () => socket.off('new_message', callback);
  }
  return () => {};
};

/**
 * Subscribe to notifications
 * @param {function} callback - Function to call when a new notification is received
 * @returns {function} - Function to unsubscribe
 */
export const subscribeToNotifications = (callback) => {
  if (socket) {
    socket.on('notification', callback);
    return () => socket.off('notification', callback);
  }
  return () => {};
};
