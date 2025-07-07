const { contextBridge, ipcRenderer } = require('electron');

// Expose protected methods that allow the renderer process to use
// the ipcRenderer without exposing the entire object
contextBridge.exposeInMainWorld('electronAPI', {
  // Authentication APIs
  auth: {
    login: (credentials) => ipcRenderer.invoke('auth:login', credentials),
    logout: () => ipcRenderer.invoke('auth:logout'),
    getStoredSession: () => ipcRenderer.invoke('auth:getStoredSession'),
    clearStoredSession: () => ipcRenderer.invoke('auth:clearStoredSession'),
    openRegistration: () => ipcRenderer.invoke('auth:openRegistration')
  },

  // Storage APIs
  storage: {
    setItem: (key, value) => ipcRenderer.invoke('storage:setItem', key, value),
    getItem: (key) => ipcRenderer.invoke('storage:getItem', key),
    removeItem: (key) => ipcRenderer.invoke('storage:removeItem', key),
    clear: () => ipcRenderer.invoke('storage:clear')
  },

  // Secure storage APIs (for sensitive data like tokens)
  secureStorage: {
    setItem: (key, value) => ipcRenderer.invoke('secureStorage:setItem', key, value),
    getItem: (key) => ipcRenderer.invoke('secureStorage:getItem', key),
    removeItem: (key) => ipcRenderer.invoke('secureStorage:removeItem', key)
  },

  // Network APIs
  network: {
    isOnline: () => ipcRenderer.invoke('network:isOnline'),
    onStatusChange: (callback) => {
      ipcRenderer.on('network:statusChanged', (event, isOnline) => callback(isOnline));
      return () => ipcRenderer.removeAllListeners('network:statusChanged');
    }
  },

  // Notification APIs
  notifications: {
    show: (options) => ipcRenderer.invoke('notifications:show', options),
    requestPermission: () => ipcRenderer.invoke('notifications:requestPermission'),
    onNotificationClick: (callback) => {
      ipcRenderer.on('notification:clicked', (event, data) => callback(data));
      return () => ipcRenderer.removeAllListeners('notification:clicked');
    }
  },

  // Window APIs
  window: {
    minimize: () => ipcRenderer.invoke('window:minimize'),
    maximize: () => ipcRenderer.invoke('window:maximize'),
    close: () => ipcRenderer.invoke('window:close'),
    focus: () => ipcRenderer.invoke('window:focus'),
    isMaximized: () => ipcRenderer.invoke('window:isMaximized'),
    onWindowStateChange: (callback) => {
      ipcRenderer.on('window:stateChanged', (event, state) => callback(state));
      return () => ipcRenderer.removeAllListeners('window:stateChanged');
    }
  },

  // App APIs
  app: {
    getVersion: () => ipcRenderer.invoke('app:getVersion'),
    quit: () => ipcRenderer.invoke('app:quit'),
    openExternal: (url) => ipcRenderer.invoke('app:openExternal', url)
  },

  // Message cache APIs
  messageCache: {
    saveMessages: (userId, messages) => ipcRenderer.invoke('messageCache:saveMessages', userId, messages),
    getMessages: (userId) => ipcRenderer.invoke('messageCache:getMessages', userId),
    searchMessages: (query, userId) => ipcRenderer.invoke('messageCache:searchMessages', query, userId),
    clearMessages: (userId) => ipcRenderer.invoke('messageCache:clearMessages', userId),
    clearAllMessages: () => ipcRenderer.invoke('messageCache:clearAllMessages')
  },

  // User data APIs
  userData: {
    saveUser: (user) => ipcRenderer.invoke('userData:saveUser', user),
    getUser: () => ipcRenderer.invoke('userData:getUser'),
    saveContacts: (contacts) => ipcRenderer.invoke('userData:saveContacts', contacts),
    getContacts: () => ipcRenderer.invoke('userData:getContacts'),
    clearUserData: () => ipcRenderer.invoke('userData:clearUserData')
  },

  // Development APIs (only available in dev mode)
  dev: {
    openDevTools: () => ipcRenderer.invoke('dev:openDevTools'),
    reload: () => ipcRenderer.invoke('dev:reload')
  }
});

// Expose a limited set of Node.js APIs
contextBridge.exposeInMainWorld('nodeAPI', {
  platform: process.platform,
  arch: process.arch,
  versions: process.versions
});

// WebSocket API - we'll handle WebSocket connections in the renderer
// but provide utilities for connection management
contextBridge.exposeInMainWorld('websocketAPI', {
  // These will be implemented to work with the existing WebSocket infrastructure
  createConnection: (url, protocols) => {
    // Return a promise that resolves to a WebSocket-like object
    return ipcRenderer.invoke('websocket:createConnection', url, protocols);
  },
  closeConnection: (connectionId) => ipcRenderer.invoke('websocket:closeConnection', connectionId),
  sendMessage: (connectionId, message) => ipcRenderer.invoke('websocket:sendMessage', connectionId, message),
  onMessage: (connectionId, callback) => {
    const channel = `websocket:message:${connectionId}`;
    ipcRenderer.on(channel, (event, data) => callback(data));
    return () => ipcRenderer.removeAllListeners(channel);
  },
  onClose: (connectionId, callback) => {
    const channel = `websocket:close:${connectionId}`;
    ipcRenderer.on(channel, (event, data) => callback(data));
    return () => ipcRenderer.removeAllListeners(channel);
  },
  onError: (connectionId, callback) => {
    const channel = `websocket:error:${connectionId}`;
    ipcRenderer.on(channel, (event, data) => callback(data));
    return () => ipcRenderer.removeAllListeners(channel);
  }
});
