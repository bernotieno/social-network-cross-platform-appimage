# Social Network Messenger - Desktop App

A cross-platform desktop messenger built with Electron that connects to the existing Social Network backend infrastructure. Provides a native desktop experience with secure messaging, real-time communication, and offline capabilities.

## Features

- 🔐 **Secure Authentication**: Reuses existing API endpoints with encrypted session storage
- 💬 **Real-time Messaging**: WebSocket-based chat with typing indicators and presence detection
- 🖥️ **Native Desktop App**: Built with Electron for Windows, macOS, and Linux
- 🔄 **Offline Support**: Message caching and offline mode with queue for pending messages
- 🔍 **Message Search**: Real-time search across cached messages
- 🎨 **Modern UI**: Clean, responsive interface optimized for desktop
- 😊 **Emoji Support**: Built-in emoji picker for enhanced messaging
- 🔔 **Desktop Notifications**: Native system notifications for new messages
- 🔒 **Secure Storage**: Encrypted local storage using Electron's safeStorage API
- ⌨️ **Keyboard Shortcuts**: Full keyboard navigation and shortcuts

## Architecture

The app is built as an Electron application with a secure, modular structure:

```
desktop-messenger/
├── main/               # Electron main process
│   └── main.js         # Main process entry point, window management, IPC handlers
├── preload/            # Preload scripts for secure IPC
│   └── preload.js      # Secure bridge between main and renderer processes
├── renderer/           # Renderer process (UI)
│   ├── index.html      # Main UI structure
│   ├── styles/         # CSS styling
│   ├── scripts/        # Frontend JavaScript modules
│   │   ├── auth.js     # Authentication management
│   │   ├── chat.js     # Chat interface and messaging
│   │   ├── storage.js  # Local data storage (IndexedDB)
│   │   ├── websocket.js # WebSocket connection handling
│   │   ├── utils.js    # Utility functions
│   │   ├── theme.js    # Theme management
│   │   ├── web-compat.js # Browser compatibility layer
│   │   └── main.js     # Application initialization
│   └── manifest.json   # PWA manifest
├── assets/             # Static assets (icons, etc.)
├── build.js            # Build script for all platforms
└── package.json        # Dependencies and build configuration
```

## Security Features

- **Process Isolation**: Main and renderer processes are isolated for security
- **Context Isolation**: Renderer process runs in isolated context with no Node.js access
- **Secure Storage**: Session data encrypted using Electron's safeStorage API
- **IPC Security**: All communication between processes uses secure IPC channels
- **No Remote Module**: Remote module disabled to prevent security vulnerabilities
- **Content Security Policy**: CSP headers prevent XSS attacks
- **HTTPS Ready**: Designed to work with HTTPS backends in production
- **Secure Preload**: Preload scripts provide controlled API access to renderer

## Getting Started

### Prerequisites

- Node.js 18+ (for development builds)
- Access to the existing Social Network backend (running on localhost:8080)
- Access to the existing frontend (for registration at localhost:3000)

### Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/bernotieno/social-network-cross-platform-appimage.git
   cd social-network-cross-platform-appimage
   ```
2. Run the Backend server:
   ```bash
   make
   ```

3. Run the Frontend server:
   ```bash
   make frontend
   ```

4. Start the application:

   **Production mode:**
   ```bash
   make messenger
   ```

5. The application will open as a native desktop window

### Building for Production

Build for all platforms:
```bash
npm run build
```

Build for specific platforms:
```bash
npm run build:win    # Windows
npm run build:mac    # macOS  
npm run build:linux  # Linux
```

## Configuration

The app connects to the backend at `http://localhost:8080` by default. To change this:

1. Update the `apiBaseUrl` in `renderer/scripts/auth.js`
2. Update the `wsUrl` in `renderer/scripts/websocket.js`

For production builds, you can also set environment variables or modify the configuration in the main process.

## Data Storage

The app uses multiple storage mechanisms for different types of data:

- **IndexedDB**: Message caching and search (renderer process)
- **Electron Store**: User preferences and settings (main process)
- **Safe Storage**: Encrypted session tokens and sensitive data (main process)
- **File System**: Message cache files stored in user data directory
- **Memory**: Temporary data and application state

## API Integration

The desktop app integrates with existing backend endpoints:

- `POST /api/auth/login` - User authentication
- `POST /api/auth/logout` - User logout
- `GET /api/users/{id}/following` - Get user's following list
- `GET /api/users/{id}/followers` - Get user's followers
- `GET /api/messages/{userId}` - Get message history
- `POST /api/messages` - Send new message
- `GET /api/messages/online-users` - Get online users
- `WebSocket /ws` - Real-time messaging

## WebSocket Events

The app handles these WebSocket message types:

- `message` - New chat message
- `user_presence` - User online/offline status
- `typing_status` - Typing indicators
- `join_room` - Join chat room
- `leave_room` - Leave chat room

## Keyboard Shortcuts

- `Ctrl/Cmd + K` - Focus search
- `Ctrl/Cmd + Enter` - Send message
- `Ctrl/Cmd + L` - Logout
- `Ctrl/Cmd + R` - Refresh/Reload
- `Ctrl/Cmd + Q` - Quit application
- `F11` - Toggle fullscreen
- `Ctrl/Cmd + Shift + I` - Toggle Developer Tools (dev mode)
- `Escape` - Clear search or close modals

## Development

### Debug Mode

Run with debug features enabled:
```bash
npm run dev
```

Debug shortcuts (development only):
- `Ctrl + Shift + D` - Show debug info in console
- `Ctrl + Shift + C` - Clear all local data

### Project Structure

- **main/main.js**: Main process entry point, handles app lifecycle, window management, and IPC
- **preload/preload.js**: Secure bridge between main and renderer processes
- **renderer/scripts/auth.js**: Authentication management and session handling
- **renderer/scripts/websocket.js**: WebSocket connection and message handling
- **renderer/scripts/chat.js**: Chat interface and message management
- **renderer/scripts/storage.js**: Local data storage using IndexedDB
- **renderer/scripts/utils.js**: Utility functions and helpers
- **renderer/scripts/theme.js**: Theme management and UI customization
- **renderer/scripts/main.js**: Application initialization and coordination

## Offline Mode

When offline, the app:
- Shows offline indicators in the UI
- Disables message sending (with visual feedback)
- Allows viewing of cached messages
- Queues messages for sending when back online
- Maintains read-only access to conversation history

## Notifications

The app supports native desktop notifications:
- New message notifications (when app is not focused)
- Click notifications to focus the app
- Respects system notification preferences

## Building and Distribution

The app uses electron-builder for packaging:

- **Windows**: NSIS installer and portable executable
- **macOS**: DMG disk image with universal binary (x64 + ARM64)
- **Linux**: AppImage and DEB packages

## Troubleshooting

### Common Issues

1. **App won't start**: Check Node.js version (requires 18+) and run `npm install`
2. **WebSocket connection fails**: Ensure backend is running on localhost:8080
3. **Login fails**: Check backend API is accessible and CORS is configured
4. **Messages not syncing**: Verify WebSocket connection in developer tools
5. **Build fails**: Ensure all dependencies are installed and platform requirements are met
6. **Notifications not working**: Check system notification permissions

### Logs

- **Main process logs**: Check terminal where app was started
- **Renderer logs**: Open Developer Tools (Ctrl+Shift+I in dev mode)
- **Storage issues**: Check app data directory for corrupted files
- **Build logs**: Check console output during `npm run build`
- **IPC errors**: Check both main and renderer process logs for communication issues

## Contributing

1. Follow the existing code structure and patterns
2. Maintain security best practices (no Node.js in renderer, use IPC for main process communication)
3. Test on all target platforms before submitting
4. Update documentation for new features
5. Use the preload script for secure API exposure to renderer
6. Follow Electron security guidelines and best practices

## License

This project is part of the Social Network application suite.
