const { app, BrowserWindow, Menu, ipcMain, shell, safeStorage, Notification } = require('electron');
const path = require('path');
const isDev = process.argv.includes('--dev');

// Keep a global reference of the window object
let mainWindow;

// Security: Prevent new window creation
app.on('web-contents-created', (event, contents) => {
  contents.on('new-window', (event, navigationUrl) => {
    event.preventDefault();
    shell.openExternal(navigationUrl);
  });
});

function createWindow() {
  // Create the browser window
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 800,
    minHeight: 600,
    icon: path.join(__dirname, '../assets/icon.png'),
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      enableRemoteModule: false,
      preload: path.join(__dirname, '../preload/preload.js'),
      webSecurity: true,
      allowRunningInsecureContent: false,
      experimentalFeatures: false
    },
    show: false, // Don't show until ready
    titleBarStyle: process.platform === 'darwin' ? 'hiddenInset' : 'default'
  });

  // Load the app
  mainWindow.loadFile(path.join(__dirname, '../renderer/index.html'));

  // Show window when ready to prevent visual flash
  mainWindow.once('ready-to-show', () => {
    mainWindow.show();
    
    // Focus on window creation
    if (isDev) {
      mainWindow.webContents.openDevTools();
    }
  });

  // Handle window closed
  mainWindow.on('closed', () => {
    mainWindow = null;
  });

  // Prevent navigation to external URLs
  mainWindow.webContents.on('will-navigate', (event, url) => {
    if (url !== mainWindow.webContents.getURL()) {
      event.preventDefault();
      shell.openExternal(url);
    }
  });

  // Handle external links
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: 'deny' };
  });
}

// App event handlers
app.whenReady().then(() => {
  createWindow();
  createMenu();

  app.on('activate', () => {
    // On macOS, re-create window when dock icon is clicked
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on('window-all-closed', () => {
  // On macOS, keep app running even when all windows are closed
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

// Security: Prevent new window creation
app.on('web-contents-created', (event, contents) => {
  contents.on('new-window', (event, navigationUrl) => {
    event.preventDefault();
    shell.openExternal(navigationUrl);
  });
});

function createMenu() {
  const template = [
    {
      label: 'File',
      submenu: [
        {
          label: 'Quit',
          accelerator: process.platform === 'darwin' ? 'Cmd+Q' : 'Ctrl+Q',
          click: () => {
            app.quit();
          }
        }
      ]
    },
    {
      label: 'Edit',
      submenu: [
        { role: 'undo' },
        { role: 'redo' },
        { type: 'separator' },
        { role: 'cut' },
        { role: 'copy' },
        { role: 'paste' },
        { role: 'selectall' }
      ]
    },
    {
      label: 'View',
      submenu: [
        { role: 'reload' },
        { role: 'forceReload' },
        { role: 'toggleDevTools' },
        { type: 'separator' },
        { role: 'resetZoom' },
        { role: 'zoomIn' },
        { role: 'zoomOut' },
        { type: 'separator' },
        { role: 'togglefullscreen' }
      ]
    },
    {
      label: 'Window',
      submenu: [
        { role: 'minimize' },
        { role: 'close' }
      ]
    }
  ];

  if (process.platform === 'darwin') {
    template.unshift({
      label: app.getName(),
      submenu: [
        { role: 'about' },
        { type: 'separator' },
        { role: 'services' },
        { type: 'separator' },
        { role: 'hide' },
        { role: 'hideOthers' },
        { role: 'unhide' },
        { type: 'separator' },
        { role: 'quit' }
      ]
    });

    // Window menu
    template[4].submenu = [
      { role: 'close' },
      { role: 'minimize' },
      { role: 'zoom' },
      { type: 'separator' },
      { role: 'front' }
    ];
  }

  const menu = Menu.buildFromTemplate(template);
  Menu.setApplicationMenu(menu);
}

// IPC handlers
setupIpcHandlers();

function setupIpcHandlers() {
  // Authentication handlers
  ipcMain.handle('auth:login', async (event, credentials) => {
    try {
      // Store credentials securely
      if (safeStorage.isEncryptionAvailable()) {
        const encryptedData = safeStorage.encryptString(JSON.stringify(credentials));
        // Store in app data directory
        const fs = require('fs');
        const path = require('path');
        const userDataPath = app.getPath('userData');
        const sessionPath = path.join(userDataPath, 'session.dat');
        fs.writeFileSync(sessionPath, encryptedData);
      }
      return { success: true };
    } catch (error) {
      console.error('Error storing session:', error);
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('auth:logout', async (event) => {
    try {
      // Clear stored session
      const fs = require('fs');
      const path = require('path');
      const userDataPath = app.getPath('userData');
      const sessionPath = path.join(userDataPath, 'session.dat');
      if (fs.existsSync(sessionPath)) {
        fs.unlinkSync(sessionPath);
      }
      return { success: true };
    } catch (error) {
      console.error('Error clearing session:', error);
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('auth:getStoredSession', async (event) => {
    try {
      const fs = require('fs');
      const path = require('path');
      const userDataPath = app.getPath('userData');
      const sessionPath = path.join(userDataPath, 'session.dat');

      if (fs.existsSync(sessionPath) && safeStorage.isEncryptionAvailable()) {
        const encryptedData = fs.readFileSync(sessionPath);
        const decryptedData = safeStorage.decryptString(encryptedData);
        return JSON.parse(decryptedData);
      }
      return null;
    } catch (error) {
      console.error('Error retrieving session:', error);
      return null;
    }
  });

  ipcMain.handle('auth:clearStoredSession', async (event) => {
    return ipcMain.handle('auth:logout', event);
  });

  ipcMain.handle('auth:openRegistration', async (event) => {
    shell.openExternal('http://localhost:3000/auth/register');
    return { success: true };
  });

  // Storage handlers
  ipcMain.handle('storage:setItem', async (event, key, value) => {
    try {
      const Store = require('electron-store');
      const store = new Store();
      store.set(key, value);
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('storage:getItem', async (event, key) => {
    try {
      const Store = require('electron-store');
      const store = new Store();
      return store.get(key);
    } catch (error) {
      return null;
    }
  });

  ipcMain.handle('storage:removeItem', async (event, key) => {
    try {
      const Store = require('electron-store');
      const store = new Store();
      store.delete(key);
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('storage:clear', async (event) => {
    try {
      const Store = require('electron-store');
      const store = new Store();
      store.clear();
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  // Secure storage handlers
  ipcMain.handle('secureStorage:setItem', async (event, key, value) => {
    try {
      if (safeStorage.isEncryptionAvailable()) {
        const encryptedValue = safeStorage.encryptString(JSON.stringify(value));
        const Store = require('electron-store');
        const store = new Store({ name: 'secure' });
        store.set(key, encryptedValue.toString('base64'));
        return { success: true };
      }
      return { success: false, error: 'Encryption not available' };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('secureStorage:getItem', async (event, key) => {
    try {
      if (safeStorage.isEncryptionAvailable()) {
        const Store = require('electron-store');
        const store = new Store({ name: 'secure' });
        const encryptedValue = store.get(key);
        if (encryptedValue) {
          const buffer = Buffer.from(encryptedValue, 'base64');
          const decryptedValue = safeStorage.decryptString(buffer);
          return JSON.parse(decryptedValue);
        }
      }
      return null;
    } catch (error) {
      return null;
    }
  });

  ipcMain.handle('secureStorage:removeItem', async (event, key) => {
    try {
      const Store = require('electron-store');
      const store = new Store({ name: 'secure' });
      store.delete(key);
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  // Network handlers
  ipcMain.handle('network:isOnline', async (event) => {
    return { isOnline: true }; // Electron doesn't have a built-in way to check this
  });

  // Notification handlers
  ipcMain.handle('notifications:show', async (event, options) => {
    try {
      if (Notification.isSupported()) {
        const notification = new Notification({
          title: options.title,
          body: options.body,
          icon: options.icon
        });

        notification.on('click', () => {
          if (mainWindow) {
            if (mainWindow.isMinimized()) mainWindow.restore();
            mainWindow.focus();
          }
          event.sender.send('notification:clicked', options.data);
        });

        notification.show();
        return { success: true };
      }
      return { success: false, error: 'Notifications not supported' };
    } catch (error) {
      return { success: false, error: error.message };
    }
  });

  ipcMain.handle('notifications:requestPermission', async (event) => {
    // Electron handles this automatically
    return 'granted';
  });

  // Window handlers
  ipcMain.handle('window:minimize', async (event) => {
    if (mainWindow) mainWindow.minimize();
    return { success: true };
  });

  ipcMain.handle('window:maximize', async (event) => {
    if (mainWindow) {
      if (mainWindow.isMaximized()) {
        mainWindow.unmaximize();
      } else {
        mainWindow.maximize();
      }
    }
    return { success: true };
  });

  ipcMain.handle('window:close', async (event) => {
    if (mainWindow) mainWindow.close();
    return { success: true };
  });

  ipcMain.handle('window:focus', async (event) => {
    if (mainWindow) {
      if (mainWindow.isMinimized()) mainWindow.restore();
      mainWindow.focus();
    }
    return { success: true };
  });

  ipcMain.handle('window:isMaximized', async (event) => {
    return mainWindow ? mainWindow.isMaximized() : false;
  });

  // App handlers
  ipcMain.handle('app:getVersion', async (event) => {
    return app.getVersion();
  });

  ipcMain.handle('app:quit', async (event) => {
    app.quit();
    return { success: true };
  });

  ipcMain.handle('app:openExternal', async (event, url) => {
    shell.openExternal(url);
    return { success: true };
  });

  // Development handlers
  ipcMain.handle('dev:openDevTools', async (event) => {
    if (mainWindow && isDev) {
      mainWindow.webContents.openDevTools();
    }
    return { success: true };
  });

  ipcMain.handle('dev:reload', async (event) => {
    if (mainWindow && isDev) {
      mainWindow.webContents.reload();
    }
    return { success: true };
  });
}

module.exports = { mainWindow };
