# Installation Guide - Social Network Messenger

This guide will help you install and set up the Social Network Messenger desktop application.

## Prerequisites

Before installing the desktop messenger, ensure you have:

1. **Backend Server Running**: The Social Network backend must be running on `localhost:8080`
2. **Frontend Available**: The web frontend should be accessible at `localhost:3000` for user registration
3. **Node.js 18+**: Required for development builds

## Quick Start

### Option 1: Pre-built Binaries (Recommended)

1. Download the appropriate installer for your platform from the releases:
   - **Windows**: `Social-Network-Messenger-Setup-1.0.0.exe`
   - **macOS**: `Social-Network-Messenger-1.0.0.dmg`
   - **Linux**: `Social-Network-Messenger-1.0.0.AppImage` or `social-network-messenger_1.0.0_amd64.deb`

2. Install the application:
   - **Windows**: Run the `.exe` installer and follow the setup wizard
   - **macOS**: Open the `.dmg` file and drag the app to Applications folder
   - **Linux (AppImage)**: Make executable and run: `chmod +x *.AppImage && ./Social-Network-Messenger-1.0.0.AppImage`
   - **Linux (DEB)**: Install with: `sudo dpkg -i social-network-messenger_1.0.0_amd64.deb`

3. Launch the application and sign in with your existing Social Network credentials

### Option 2: Build from Source

1. Clone the repository and navigate to the desktop messenger:
   ```bash
   cd social-network-cross-platform-appimage/desktop-messenger
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Run in development mode:
   ```bash
   npm run dev
   ```

4. Or build for production:
   ```bash
   # Build for current platform
   npm run build
   
   # Build for specific platform
   npm run build:win    # Windows
   npm run build:mac    # macOS
   npm run build:linux  # Linux
   ```

## Configuration

### Backend Connection

By default, the app connects to:
- **API**: `http://localhost:8080/api`
- **WebSocket**: `ws://localhost:8080/ws`

To change these endpoints, edit:
- `renderer/scripts/auth.js` - Update `apiBaseUrl`
- `renderer/scripts/websocket.js` - Update `wsUrl`

### First Time Setup

1. **Start Backend**: Ensure your Social Network backend is running
2. **Launch App**: Open the desktop messenger
3. **Sign In**: Use your existing email and password
4. **Registration**: If you don't have an account, click "Create one on our website" to open the web registration

## Platform-Specific Instructions

### Windows

**System Requirements:**
- Windows 10 or later
- 100 MB free disk space

**Installation:**
1. Download `Social-Network-Messenger-Setup-1.0.0.exe`
2. Right-click and select "Run as administrator" if needed
3. Follow the installation wizard
4. The app will be available in Start Menu and Desktop

**Uninstallation:**
- Use "Add or Remove Programs" in Windows Settings
- Or run the uninstaller from the installation directory

### macOS

**System Requirements:**
- macOS 10.14 (Mojave) or later
- 100 MB free disk space

**Installation:**
1. Download `Social-Network-Messenger-1.0.0.dmg`
2. Open the DMG file
3. Drag "Social Network Messenger" to Applications folder
4. Launch from Applications or Spotlight

**Security Note:**
If you see "App can't be opened because it is from an unidentified developer":
1. Right-click the app and select "Open"
2. Click "Open" in the security dialog
3. Or go to System Preferences > Security & Privacy and click "Open Anyway"

### Linux

**System Requirements:**
- Ubuntu 18.04+ / Debian 10+ / CentOS 8+ or equivalent
- 100 MB free disk space

**AppImage (Universal):**
1. Download `Social-Network-Messenger-1.0.0.AppImage`
2. Make it executable: `chmod +x Social-Network-Messenger-1.0.0.AppImage`
3. Run: `./Social-Network-Messenger-1.0.0.AppImage`

**DEB Package (Debian/Ubuntu):**
1. Download `social-network-messenger_1.0.0_amd64.deb`
2. Install: `sudo dpkg -i social-network-messenger_1.0.0_amd64.deb`
3. Fix dependencies if needed: `sudo apt-get install -f`
4. Launch from applications menu or run: `social-network-messenger`

## Troubleshooting

### Common Issues

**App won't start:**
- Check if Node.js 18+ is installed (for development builds)
- Verify backend server is running on localhost:8080
- Check system requirements for your platform

**Can't connect to server:**
- Ensure backend is running and accessible
- Check firewall settings
- Verify API endpoints in configuration

**Login fails:**
- Verify credentials are correct
- Check backend logs for authentication errors
- Ensure CORS is properly configured on backend

**Messages not syncing:**
- Check WebSocket connection in developer tools
- Verify backend WebSocket server is running
- Check network connectivity

**Notifications not working:**
- Grant notification permissions when prompted
- Check system notification settings
- Restart the app if permissions were recently changed

### Getting Help

1. **Check Logs:**
   - Development: Check terminal output and browser developer tools
   - Production: Check app data directory for log files

2. **Reset App Data:**
   - Close the app completely
   - Delete app data directory:
     - Windows: `%APPDATA%/social-network-messenger`
     - macOS: `~/Library/Application Support/social-network-messenger`
     - Linux: `~/.config/social-network-messenger`
   - Restart the app

3. **Reinstall:**
   - Uninstall the current version
   - Download and install the latest version
   - Your messages will be re-downloaded from the server

### Development Mode

For developers and advanced users:

```bash
# Run with debug features
npm run dev

# Debug shortcuts (development only)
Ctrl+Shift+D  # Show debug info
Ctrl+Shift+C  # Clear all local data
Ctrl+Shift+I  # Open developer tools
```

## Data Storage

The app stores data locally in:
- **Windows**: `%APPDATA%/social-network-messenger`
- **macOS**: `~/Library/Application Support/social-network-messenger`
- **Linux**: `~/.config/social-network-messenger`

This includes:
- Encrypted session tokens
- Cached messages and contacts
- User preferences
- Application settings

## Security

The app implements several security measures:
- Session tokens are encrypted using Electron's safeStorage
- No Node.js access from renderer process
- Content Security Policy prevents XSS attacks
- Secure communication with backend APIs

## Updates

The app will check for updates automatically (when implemented). For now:
1. Download the latest version from releases
2. Install over the existing version
3. Your data and settings will be preserved

## Uninstallation

**Windows:**
- Use "Add or Remove Programs" in Settings
- Or run uninstaller from installation directory

**macOS:**
- Drag app from Applications to Trash
- Delete app data: `~/Library/Application Support/social-network-messenger`

**Linux:**
- DEB: `sudo apt remove social-network-messenger`
- AppImage: Delete the AppImage file
- Delete app data: `~/.config/social-network-messenger`
