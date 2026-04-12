# RBAC & Admin System Guide

Complete documentation for the Role-Based Access Control (RBAC) system and admin features in the Ability System Demo.

## System Overview

The demo implements a complete multiplayer admin system with:
- **Role-Based Access Control (RBAC)** - Define roles (admin, moderator, user) with permissions
- **Admin Commands** - Execute server commands with permission validation
- **Admin Terminal** - CLI interface for headless server management
- **Admin Panel UI** - In-game UI for host to manage players
- **Player Management** - Kick, ban, mute players with configurable durations
- **Server Logging** - Centralized logging of all events
- **Open to LAN** - Allow singleplayer worlds to accept multiplayer connections

## Architecture

### 1. RBAC System (`demo/autoload/RBAC.gd`)

**Purpose**: Manages roles, permissions, and player access control.

**Key Concepts**:
- **Roles**: Predefined groups with specific permissions
  - `admin` - Full server control
  - `moderator` - Limited moderation (kick, mute, broadcast)
  - `user` - Basic gameplay (chat, move, attack)

- **Peer Registration**: Track connected players and their roles
- **Ban System**: Temporary bans with automatic expiration
- **Permission Checking**: Validate if a peer can execute an action

**Key Methods**:
```gdscript
# Registration
register_peer(peer_id: int, player_name: String) -> void
unregister_peer(peer_id: int) -> void

# Role Management
get_player_role(peer_id: int) -> String
set_player_role(peer_id: int, role: String) -> bool

# Permission Checking
can_execute(peer_id: int, action: String) -> bool

# Player Management
get_all_peers() -> Array
get_peer_data(peer_id: int) -> Dictionary
get_connected_count() -> int

# Ban Management
ban_peer(peer_id: int, hours: float = 24.0, reason: String = "") -> bool
unban_peer(peer_id: int) -> bool
is_banned(peer_id: int) -> bool
get_ban_info(peer_id: int) -> Dictionary
```

**Usage Example**:
```gdscript
# Register a new player
RBAC.register_peer(2, "Alice")

# Check if player can perform action
if RBAC.can_execute(2, "kick_player"):
    # Allow kick command
    pass

# Ban a player for 24 hours
RBAC.ban_peer(2, 24.0, "Disruptive behavior")

# Get player info
var peers = RBAC.get_all_peers()
for peer_info in peers:
    print("%s (%s)" % [peer_info["name"], peer_info["role"]])
```

### 2. AdminCommand System (`demo/autoload/AdminCommand.gd`)

**Purpose**: Process admin commands with RBAC permission validation.

**Supported Commands**:
- `kick_player` - Expel a player from the server
- `ban_player` - Ban a player for a duration
- `mute_player` - Prevent player from chatting
- `unmute_player` - Restore player chat access
- `shutdown_server` - Shutdown the server
- `broadcast_message` - Send admin message to all players
- `list_players` - List all connected players
- `get_server_status` - Show server status

**Key Methods**:
```gdscript
# Execute a command with permission checking
execute_command(executor_id: int, command: String, args: Array = []) -> bool

# RPC for remote command execution
execute_command_rpc(command: String, args: Array = []) -> void
```

**Usage Example**:
```gdscript
# Kick player 5 with reason
AdminCommand.execute_command(1, "kick_player", [5, "Toxic behavior"])

# Ban player 5 for 48 hours
AdminCommand.execute_command(1, "ban_player", [5, 48.0, "Multiple violations"])

# Mute player 3
AdminCommand.execute_command(1, "mute_player", [3])

# Get server status
AdminCommand.execute_command(1, "get_server_status", [])

# Broadcast message to all players
AdminCommand.execute_command(1, "broadcast_message", ["Server will restart in 5 minutes"])
```

**Signals**:
```gdscript
signal command_executed(command: String, executor_id: int, success: bool)
signal player_kicked(peer_id: int, reason: String)
signal player_banned(peer_id: int, hours: float, reason: String)
signal broadcast_message_sent(message: String)
signal server_shutdown_initiated(reason: String)
```

### 3. AdminTerminal (`demo/autoload/AdminTerminal.gd`)

**Purpose**: CLI interface for headless server management.

**Architecture**: Non-blocking command processing using `_process()` frame callback instead of blocking stdin read.

**Supported CLI Commands**:
```
help              - Show this help text
list, players     - List all connected players
kick <id> [reason] - Kick a player
ban <id> [hours] [reason] - Ban a player
unban <id>        - Unban a player
say <message>     - Broadcast a message
mute <id>         - Mute a player
unmute <id>       - Unmute a player
status            - Show server status
shutdown [reason] - Shutdown the server
```

**Implementation Note**: The current implementation uses a non-blocking stub for stdin handling. To implement true interactive CLI on a headless server, you would need platform-specific stdin reading (e.g., using threads or async I/O).

### 4. ServerLogger (`demo/autoload/ServerLogger.gd`)

**Purpose**: Centralized logging for all server events.

**Log Levels**:
- `DEBUG` - Detailed debug information
- `INFO` - General information
- `WARN` - Warning messages
- `ERROR` - Error messages

**Key Methods**:
```gdscript
log(level: LogLevel, source: String, message: String) -> void
log_peer_connected(peer_id: int, player_name: String) -> void
log_peer_disconnected(peer_id: int, player_name: String) -> void
log_command(executor_id: int, command: String, success: bool) -> void
log_chat(peer_id: int, player_name: String, message: String) -> void
log_error(source: String, error_msg: String) -> void
get_recent_logs(count: int = 50) -> PackedStringArray
```

**Configuration**:
```gdscript
const ENABLE_FILE_LOGGING = false  # Set to true for file persistence
const LOG_FILE_PATH = "user://server.log"
```

### 5. Admin Panel UI (`demo/ui/admin_panel.tscn` / `admin_panel.gd`)

**Purpose**: In-game UI for host to manage server and players.

**Features**:
- View list of connected players with roles
- Kick/Ban individual players
- Send broadcast messages to all players
- Server status display
- Open/Close LAN functionality

**Visibility**: Only visible to server host (multiplayer.is_server() == true)

**Signals**:
```gdscript
signal close_to_lan_requested
signal shutdown_requested
```

### 6. Pause Menu Integration (`demo/ui/pause_menu.gd`)

**Purpose**: Integrate admin panel and LAN controls into pause menu.

**Features**:
- Show admin panel when paused (if host)
- "Open to LAN" button for singleplayer games
- "Close Server" button to stop accepting new connections
- Quit to menu option

### 7. Multiplayer Game Manager (`demo/autoload/MultiplayerGameManager.gd`)

**Purpose**: Core multiplayer setup and player lifecycle management.

**Game Modes**:
- `singleplayer` - Single player local game
- `multiplayer_server` - Server mode (host)
- `multiplayer_client` - Client mode (join existing server)

**LAN Features**:
```gdscript
# Open current singleplayer world to LAN for other players to join
open_to_lan(port: int = 0) -> bool

# Close LAN server
close_lan_server() -> void

# Get current LAN port
get_lan_port() -> int
```

**Peer Lifecycle**:
1. Peer connects → `_on_peer_connected()` → RBAC registers peer
2. Player spawns → `player_spawned` signal emitted
3. Peer disconnects → `_on_peer_disconnected()` → RBAC unregisters peer

## Integration Points

### On Server Startup
1. `MultiplayerGameManager._ready()` checks CI/CD environment or waits for menu
2. If multiplayer mode, `_setup_server()` creates ENetMultiplayerPeer
3. `multiplayer.peer_connected` signal connects to `_on_peer_connected()`
4. `RBAC` autoload is ready and waiting for peer connections

### On Player Connection
1. `MultiplayerGameManager._on_peer_connected()` fires
2. RBAC registers peer via `RBAC.register_peer()`
3. AdminCommand handlers are available for use
4. Admin panel (if host) gets updated with new player

### On Admin Command Execution
1. Admin calls `AdminCommand.execute_command()`
2. RBAC permission check via `RBAC.can_execute()`
3. If permission granted, command handler executes
4. ServerLogger records command execution
5. Signal emitted for UI updates

## Security Considerations

### Peer ID Validation
- Always validate `executor_id` matches the RPC caller
- Server-side permission checks prevent client-side cheating

### RPC Security
```gdscript
# Commands are RPC-callable but checked server-side
@rpc("any_peer")
func execute_command_rpc(command: String, args: Array = []) -> void:
    var executor_id = multiplayer.get_remote_sender_id()
    if multiplayer.is_server():
        execute_command(executor_id, command, args)
```

### Ban Expiration
- Bans are checked on connection attempt
- Expired bans automatically removed from memory
- Optional: Implement file persistence for permanent ban history

## Usage Examples

### Starting a Game

```gdscript
# From menu selection
var mp_manager = get_node("/root/MultiplayerGameManager")

# Singleplayer mode
mp_manager.start_game("singleplayer")

# Multiplayer server (local host)
mp_manager.start_game("multiplayer_server")

# Multiplayer client (join server)
mp_manager.start_game("multiplayer_client", "192.168.1.100", 7777)

# Open singleplayer world to LAN
if mp_manager.open_to_lan(7778):
    print("World opened to LAN on port 7778")
```

### Admin Commands

```gdscript
# Kick a disruptive player
AdminCommand.execute_command(1, "kick_player", [player_id, "Disruptive behavior"])

# Ban a player
AdminCommand.execute_command(1, "ban_player", [player_id, 24.0, "Toxicity"])

# Mute a player
AdminCommand.execute_command(1, "mute_player", [player_id])

# Broadcast server message
AdminCommand.execute_command(1, "broadcast_message", ["Server restarting in 5 minutes"])

# Check permissions
if RBAC.can_execute(peer_id, "kick_player"):
    print("Player has kick permission")
```

### Admin Panel Display

```gdscript
# Admin panel automatically updates on pause menu open
# Server host sees:
# - List of connected players
# - Kick/Ban buttons for each player
# - Broadcast message input
# - Server status

# When player is kicked:
if admin_panel:
    admin_panel.refresh()  # Update player list
    chat_system.add_system_message("[KICK] Player removed")
```

## Configuration

### RBAC Roles (Edit in `RBAC.gd`)

```gdscript
const ROLES := {
    "admin": [
        "kick_player",
        "ban_player",
        "shutdown_server",
        "broadcast_message",
        "whitelist_add",
        "whitelist_remove",
        "mute_player",
        "unmute_player"
    ],
    "moderator": [
        "kick_player",
        "mute_player",
        "unmute_player",
        "broadcast_message"
    ],
    "user": [
        "chat",
        "move",
        "attack"
    ]
}
```

### GameData Configuration

```gdscript
var player_name: String = "Player"
var game_mode: String = ""  # "singleplayer", "multiplayer_server", "multiplayer_client"
var server_host: String = "127.0.0.1"
var server_port: int = 7777
var lan_port: int = -1  # Port when opened to LAN
```

## Testing

A test suite is provided in `demo/tests/test_rbac_system.gd`:

```bash
# Load test scene in editor
# Open: res://tests/test_rbac_scene.tscn
# Run the scene to execute tests
```

Test coverage includes:
- RBAC initialization
- Role-based permissions
- Player registration/unregistration
- Ban system with expiration
- AdminCommand execution

## Future Enhancements

- [ ] Persistent ban list (file-based or database)
- [ ] Whitelist system (allow only certain players)
- [ ] Admin audit log (permanent record of actions)
- [ ] Voting system (players vote to kick)
- [ ] Role hierarchy (moderator promoted from user)
- [ ] Permission customization (dynamic role creation)
- [ ] Anti-cheat system integration
- [ ] Rate limiting on commands
- [ ] Webhook notifications for bans/kicks

## Files Reference

| File | Purpose |
|------|---------|
| `demo/autoload/RBAC.gd` | Role-based access control system |
| `demo/autoload/AdminCommand.gd` | Admin command processor |
| `demo/autoload/AdminTerminal.gd` | CLI interface for headless server |
| `demo/autoload/ServerLogger.gd` | Centralized logging system |
| `demo/ui/admin_panel.tscn/.gd` | In-game admin UI |
| `demo/ui/pause_menu.gd` | Pause menu with admin integration |
| `demo/autoload/MultiplayerGameManager.gd` | Multiplayer setup and LAN features |
| `demo/autoload/GameData.gd` | Global game state |
| `demo/tests/test_rbac_system.gd` | RBAC test suite |

## Support & Debugging

### Enable Verbose Logging

Add to relevant systems:
```gdscript
# In RBAC.gd, AdminCommand.gd, etc.
print("[SYSTEM] Action: %s" % details)
if ServerLogger:
    ServerLogger.log(ServerLogger.LogLevel.DEBUG, "System", details)
```

### Check Server Status

```gdscript
# From pause menu or console
var peers = RBAC.get_all_peers()
print("Connected players: %d" % peers.size())
for peer in peers:
    print("  [%d] %s (%s)" % [peer["peer_id"], peer["name"], peer["role"]])
```

### Monitor LAN Status

```gdscript
var lan_port = MultiplayerGameManager.get_lan_port()
if lan_port > 0:
    print("LAN server active on port: %d" % lan_port)
else:
    print("LAN server not open")
```
