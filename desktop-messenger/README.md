# Social Network Messenger - Desktop App

A cross-platform desktop messenger that connects to the existing Social Network backend infrastructure. Currently running as a web application with plans for Electron packaging.

## Features

- 🔐 **Secure Authentication**: Reuses existing API endpoints with secure session storage
- 💬 **Real-time Messaging**: WebSocket-based chat with typing indicators and presence detection
- 🌐 **Web-based**: Runs in any modern browser with responsive design
- 🔄 **Offline Support**: Message caching and offline mode with queue for pending messages
- 🔍 **Message Search**: Real-time search across cached messages
- 🎨 **Modern UI**: Clean, responsive interface that works across devices
- 😊 **Emoji Support**: Built-in emoji picker for enhanced messaging
- 📱 **Progressive Web App**: Can be installed as a desktop app from the browser

## Architecture

The app is built as a modern web application with a modular structure:

```
desktop-messenger/
├── renderer/           # Web application files
│   ├── index.html      # Main UI structure
│   ├── styles/         # CSS styling
│   └── scripts/        # Frontend JavaScript modules
│       ├── web-compat.js   # Browser compatibility layer
│       ├── auth.js         # Authentication management
│       ├── chat.js         # Chat interface and messaging
│       ├── storage.js      # Local data storage (IndexedDB)
│       ├── websocket.js    # WebSocket connection handling
│       ├── utils.js        # Utility functions
│       └── main.js         # Application initialization
├── main/               # Electron files (for future packaging)
├── preload/            # Preload scripts (for future packaging)
├── assets/             # Static assets (icons, etc.)
├── start.sh            # Linux/macOS startup script
└── start.bat           # Windows startup script
```

## Security Features

- **Browser Security**: Runs in browser sandbox with standard web security
- **Local Storage**: Session data stored securely in browser localStorage
- **CSP Headers**: Content Security Policy prevents XSS attacks
- **HTTPS Ready**: Designed to work with HTTPS backends in production
- **No Server-side Storage**: All data stored locally or on your existing backend

## Getting Started

### Prerequisites

- Python 3 (for local web server)
- Access to the existing Social Network backend (running on localhost:8080)
- Access to the existing frontend (for registration at localhost:3000)
- Modern web browser (Chrome, Firefox, Safari, Edge)

### Quick Start

1. Navigate to the desktop messenger directory:
   ```bash
   cd desktop-messenger
   ```

2. Start the application:

   **Linux/macOS:**
   ```bash
   ./start.sh
   ```

   **Windows:**
   ```cmd
   start.bat
   ```

   **Manual start:**
   ```bash
   python3 -m http.server 8081 --directory renderer
   ```
   Then open http://localhost:8081 in your browser

3. The application will open in your default browser

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

## Data Storage

The app uses multiple storage mechanisms:

- **IndexedDB**: Message caching and search (renderer process)
- **Electron Store**: User preferences and settings (main process)
- **Safe Storage**: Encrypted session tokens (main process)

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

- **main.js**: Main process entry point, handles app lifecycle
- **preload.js**: Secure bridge between main and renderer processes
- **auth.js**: Authentication management and session handling
- **websocket.js**: WebSocket connection and message handling
- **chat.js**: Chat interface and message management
- **storage.js**: Local data storage using IndexedDB
- **utils.js**: Utility functions and helpers
- **main.js**: Application initialization and coordination

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

1. **WebSocket connection fails**: Ensure backend is running on localhost:8080
2. **Login fails**: Check backend API is accessible and CORS is configured
3. **Messages not syncing**: Verify WebSocket connection in developer tools
4. **App won't start**: Check Node.js version (requires 18+)

### Logs

- Main process logs: Check terminal where app was started
- Renderer logs: Open Developer Tools (Ctrl+Shift+I in dev mode)
- Storage issues: Check app data directory for corrupted files

## Contributing

1. Follow the existing code structure and patterns
2. Maintain security best practices (no Node.js in renderer)
3. Test on all target platforms before submitting
4. Update documentation for new features

## License

This project is part of the Social Network application suite.
