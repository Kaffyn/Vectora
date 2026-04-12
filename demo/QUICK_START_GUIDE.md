# Ability System Demo - Quick Start Guide

Get started with the Ability System multiplayer demo featuring RBAC and admin controls.

## 🎮 Starting the Game

### From Editor
1. Open the project in Godot 4.6+
2. Select the main scene (user_registration.tscn is default)
3. Press F5 to run

### From Built Executable
- Windows: `demo_windows_client.exe`
- Linux: `./demo_linux_client`
- macOS: `./demo_macos_client`
- Android: `demo_android_client.apk`

## 👤 User Registration

1. **Enter Player Name**
   - Type your desired player name (max 20 characters)
   - Press "Continue" or Enter

2. **Main Menu Appears**
   - Two buttons: "Singleplayer" and "Multiplayer"

## 🎮 Singleplayer Mode

1. Click **"Singleplayer"** button
2. Level loads with your player
3. Play the game!

### Open World to LAN
While playing:
1. Press **ESC** to open pause menu
2. Click **"Open to LAN"** button
   - Port is automatically assigned (7778-7888)
   - Chat shows: `[PORT] Listening on port: XXXX`
3. Share port number with other players
4. Other players can now connect!

### Close Server
1. Press **ESC** to open pause menu
2. Click **"Close Server"** button
   - Existing players will be disconnected
   - Cannot accept new connections

## 🌐 Multiplayer Mode

### Host (Server)
1. Click **"Singleplayer"** on first instance
2. Press **ESC** → **"Open to LAN"**
3. Note the port shown in chat
4. Share this port with other players

### Client (Join)
1. Click **"Multiplayer"** button
2. Configuration panel appears
3. Enter host IP: `127.0.0.1` (same machine) or actual IP
4. Enter port: The port shown by host
5. Click **"Connect"**
   - "Waiting for connection..." message appears
   - Level loads when connected

## 🛡️ Admin Controls (Host Only)

### Admin Panel
Appears when host pauses the game:
- **Player List**: Shows all connected players
- **Kick Button**: Remove a player immediately
- **Ban Button**: Prevent player from rejoining
- **Broadcast**: Send message to all players
- **Server Status**: View connected player count

### Chat Commands
Admin can send special messages:
- Type message in chat (press T)
- System will forward to all players
- Use chat for announcements

### In-Game Pause Menu (Host)
- **Resume**: Continue playing
- **Open to LAN**: Allow other players to join (singleplayer only)
- **Close Server**: Stop accepting new connections
- **Quit to Menu**: Return to main menu

## 🎯 Features

### Player Management
- ✅ Real-time player list
- ✅ Kick disruptive players
- ✅ Ban players (24 hour default, configurable)
- ✅ Mute/Unmute players
- ✅ Auto-expiring bans

### Communication
- ✅ Chat system (press T)
- ✅ Admin broadcasts
- ✅ System notifications
- ✅ Join/Leave messages

### Roles & Permissions
- **Admin**: Full server control
- **Moderator**: Kick, mute, broadcast (if set)
- **User**: Chat, play, move

### Game Modes
- ✅ Singleplayer (local)
- ✅ Singleplayer→LAN (open existing world)
- ✅ Multiplayer Server (dedicated host)
- ✅ Multiplayer Client (join server)

## 🔧 Configuration

### Change Default Port
Edit `demo/autoload/MultiplayerGameManager.gd`:
```gdscript
@export var server_port: int = 7777  # Change to different port
```

### Change Player Roles
Edit `demo/autoload/RBAC.gd`:
```gdscript
const ROLES := {
    "admin": [list of permissions],
    "moderator": [list of permissions],
    "user": [list of permissions]
}
```

### Enable Server Logging to File
Edit `demo/autoload/ServerLogger.gd`:
```gdscript
const ENABLE_FILE_LOGGING = true  # Enable file logging
const LOG_FILE_PATH = "user://server.log"
```

## 📱 Platform Notes

### Windows
- Executables are `.exe` files
- Run from Command Prompt or double-click
- No special requirements

### Linux
- Executables require execute permission: `chmod +x demo_linux_client`
- Run from terminal: `./demo_linux_client`
- May need X11 or Wayland display server

### macOS
- Executables are signed for Apple Silicon (M1+) and Intel
- May need to allow in Security settings on first run
- Run from terminal: `./demo_macos_client`

### Android
- APK can be sideloaded or installed via adb
- `adb install demo_android_client.apk`
- Touch controls work on all menus
- Requires network permission for multiplayer

## 🖥️ Dedicated Server

Run headless server on Linux/Windows:

```bash
# Windows (Powershell)
.\demo_windows_server.exe

# Linux
./demo_linux_server

# With custom port
./demo_linux_server --port 9999
```

Server will:
- Start listening on specified port
- Log all events to console
- Accept client connections
- Support CLI commands (future: stdin integration)

## 🐛 Troubleshooting

### Can't Connect to Server
1. Check firewall allows the port
2. Verify correct IP address
3. Verify correct port number
4. Check server is actually running
5. Look for error messages in chat

### Player Appears But Can't Move
1. Check network connection
2. Try pressing ESC and resuming
3. Look in chat for system messages
4. May be banned from server

### Admin Panel Not Visible
1. Must be the server host (open to LAN or multiplayer_server mode)
2. Must press ESC to pause game
3. Check pause_menu.gd is loaded

### Chat Not Working
1. Press **T** to activate chat input
2. Type message and press **Enter**
3. Chat should appear for all players
4. Check for any bans/mutes in admin panel

### Network Lag
1. Close unnecessary applications
2. Check internet connection quality
3. Reduce player count if many connected
4. Check ServerLogger for bottlenecks

## 📞 Support

### Check Logs
- **In-Game**: Chat system shows system messages
- **Console**: Check godot console output
- **File**: Enable ENABLE_FILE_LOGGING in ServerLogger.gd

### Common Issues
| Issue | Solution |
|-------|----------|
| Port in use | Change port in MultiplayerGameManager |
| Can't join LAN | Firewall blocking port |
| Lag | Network congestion or distance |
| Admin panel missing | Must be server host, press ESC |
| Chat doesn't send | Press T, type message, press Enter |

## 📚 Advanced Usage

### Running Multiple Instances
```bash
# Terminal 1: Server
./demo_linux_server --port 7777

# Terminal 2: Client 1
./demo_linux_client

# Terminal 3: Client 2
./demo_linux_client
```

### Network Testing
```bash
# Test connection to server
ping 192.168.1.100

# Check port is open
nc -zv 192.168.1.100 7777

# Monitor network traffic
netstat -an | grep 7777
```

### Debugging
```gdscript
# In any script:
print("[DEBUG] Message: %s" % value)

# Or via ServerLogger:
if ServerLogger:
    ServerLogger.log(ServerLogger.LogLevel.DEBUG, "MySystem", "Debug message")
```

## 🎓 Learning Resources

- **RBAC System Guide**: Read `RBAC_SYSTEM_GUIDE.md` for detailed docs
- **Implementation Checklist**: See `IMPLEMENTATION_CHECKLIST.md` for all features
- **Test Suite**: Run `res://tests/test_rbac_scene.tscn` for automated tests
- **Source Code**: Explore `demo/autoload/` for system implementations

## 🚀 Next Steps

1. **Play the Game**: Experience the demo as a player
2. **Test Admin Controls**: Try kicking/banning players (if hosting)
3. **Read Documentation**: Understand RBAC system in detail
4. **Customize**: Modify roles, permissions, or UI
5. **Deploy**: Build executables for platforms you need

## ✨ Tips & Tricks

- **Open World Tip**: Open to LAN allows friends to join your singleplayer game
- **Admin Tip**: Admin panel only shows during pause, visible to host
- **Network Tip**: Open to LAN assigns random port 7778-7888
- **Chat Tip**: Press T to chat, no input field shown, just type
- **Ban Tip**: Bans auto-expire after specified hours

Enjoy the Ability System Demo! 🎮
