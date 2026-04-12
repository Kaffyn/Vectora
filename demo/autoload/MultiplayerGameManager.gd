extends Node
## Central multiplayer manager for Ability System demo
## Handles: ENet setup, player spawning, synchronization, snapshots

const TICK_RATE = 60
const SNAPSHOT_INTERVAL = 10

var is_server: bool = false
var is_client: bool = false
var current_tick: int = 0
var connected_players: Dictionary = {}

signal player_spawned(player_id: int, player_node: Node)
signal player_despawned(player_id: int)
signal tick_processed(tick: int)

@export var server_port: int = 7777
@export var server_host: String = "127.0.0.1"
@export var max_players: int = 11  # 1 server + 10 clients

var player_container: Node
var player_scene: PackedScene = preload("res://player/player.tscn")
var game_mode_configured: bool = false
var current_mode: String = ""  # "singleplayer", "multiplayer_server", "multiplayer_client"
var lan_server_port: int = -1  # Port when opened to LAN (-1 = not open)

func _ready() -> void:
	print("[MP-MANAGER] Initializing multiplayer manager...")

	# Create player container
	player_container = Node.new()
	player_container.name = "Players"
	add_child(player_container)

	# Check if running in CI/CD mode (GitHub Actions)
	var player_id = _get_player_id()
	if player_id >= 0:
		print("[MP-MANAGER] CI/CD mode detected, using automated setup")
		_configure_from_environment(player_id)
		game_mode_configured = true
	else:
		print("[MP-MANAGER] Waiting for menu to configure game mode...")

func start_game(mode: String, host: String = "127.0.0.1", port: int = 7777) -> void:
	"""Start game with specified configuration (called from menu)"""
	print("[MP-MANAGER] Starting game with mode: %s" % mode)

	game_mode_configured = true
	current_mode = mode
	server_host = host
	server_port = port

	match mode:
		"singleplayer":
			print("[MP-MANAGER] Singleplayer mode")
			is_server = false
			is_client = false
			# Load level and spawn local player
			var error = await get_tree().change_scene_to_file("res://scenes/level.tscn")
			if error == OK:
				# Spawn local player after level loads
				await get_tree().process_frame
				var spawn_pos = _get_spawn_position(1)
				_spawn_player(1, spawn_pos)
		"multiplayer_server":
			print("[MP-MANAGER] Multiplayer SERVER mode")
			await _setup_server()
			await get_tree().change_scene_to_file("res://scenes/level.tscn")
		"multiplayer_client":
			print("[MP-MANAGER] Multiplayer CLIENT mode")
			await _setup_client(1)
			await get_tree().change_scene_to_file("res://scenes/level.tscn")
		_:
			print("[MP-MANAGER] ERROR: Unknown mode: %s" % mode)

func _configure_from_environment(player_id: int) -> void:
	"""Configure from CI/CD environment variables"""
	if player_id == 0:
		current_mode = "multiplayer_server"
		_setup_server()
	elif player_id > 0:
		current_mode = "multiplayer_client"
		_setup_client(player_id)

func _setup_server() -> void:
	print("[MP-MANAGER] Setting up SERVER mode...")
	is_server = true

	var peer = ENetMultiplayerPeer.new()
	var error = peer.create_server(server_port, max_players)

	if error != OK:
		print("[MP-MANAGER] ERROR: Failed to create server: %d" % error)
		get_tree().quit(1)
		return

	multiplayer.multiplayer_peer = peer
	multiplayer.peer_connected.connect(_on_peer_connected)
	multiplayer.peer_disconnected.connect(_on_peer_disconnected)

	print("[MP-MANAGER] Server listening on %s:%d" % [server_host, server_port])

func _setup_client(player_id: int) -> void:
	print("[MP-MANAGER] Setting up CLIENT mode (player %d)..." % player_id)
	is_client = true

	# Retry logic for connection
	var max_retries = 5
	for attempt in range(max_retries):
		print("[MP-MANAGER] Connecting to %s:%d (attempt %d/%d)..." % [server_host, server_port, attempt + 1, max_retries])

		var peer = ENetMultiplayerPeer.new()
		var error = peer.create_client(server_host, server_port)

		if error != OK:
			print("[MP-MANAGER] Connection attempt %d failed: %d" % [attempt + 1, error])
			if attempt < max_retries - 1:
				await get_tree().create_timer(0.5).timeout
			continue

		multiplayer.multiplayer_peer = peer
		multiplayer.connected_to_server.connect(_on_connected_to_server)
		multiplayer.server_disconnected.connect(_on_server_disconnected)

		# Wait for connection
		var timeout = 30.0
		var elapsed = 0.0
		while not multiplayer.is_connected_to_server() and elapsed < timeout:
			await get_tree().process_frame
			elapsed += get_process_delta_time()

		if multiplayer.is_connected_to_server():
			print("[MP-MANAGER] Successfully connected to server!")
			return

	print("[MP-MANAGER] ERROR: Failed to connect after %d attempts" % max_retries)
	get_tree().quit(1)

func _on_peer_connected(peer_id: int) -> void:
	print("[MP-MANAGER] Peer %d connected" % peer_id)
	connected_players[peer_id] = true

	# Register peer in RBAC
	if is_server and RBAC:
		var player_name = "Player_%d" % peer_id
		# Only use GameData.player_name for host (peer_id=1)
		# Other peers get default name or will be set via RPC later
		if peer_id == 1:
			var game_data = get_node_or_null("/root/GameData")
			if game_data and game_data.player_name != "Player":
				player_name = game_data.player_name
		RBAC.register_peer(peer_id, player_name)

	# Server spawns a player for each connected client
	if is_server and peer_id > 0:
		var spawn_pos = _get_spawn_position(peer_id)
		_spawn_player_for_client.rpc_id(peer_id, peer_id, spawn_pos)

func _on_peer_disconnected(peer_id: int) -> void:
	print("[MP-MANAGER] Peer %d disconnected" % peer_id)
	connected_players.erase(peer_id)

	# Unregister peer from RBAC
	if is_server and RBAC:
		RBAC.unregister_peer(peer_id)

	player_despawned.emit(peer_id)

func _on_connected_to_server() -> void:
	print("[MP-MANAGER] Connected to server!")

func _on_server_disconnected() -> void:
	print("[MP-MANAGER] ERROR: Server disconnected!")
	get_tree().quit(1)

@rpc("call_remote")
func _spawn_player_for_client(player_id: int, position: Vector2) -> void:
	"""Server RPC: Tell client to spawn their player"""
	if not multiplayer.is_server():
		_spawn_player(player_id, position)

func _spawn_player(player_id: int, position: Vector2) -> void:
	"""Spawn a player node locally"""
	var player = player_scene.instantiate()
	player.name = "Player_%d" % player_id
	player.position = position
	player.network_id = player_id
	player_container.add_child(player)

	# Setup network sync if exists
	if player.has_method("setup_network_sync"):
		player.setup_network_sync(player_id)

	print("[MP-MANAGER] Spawned player %d at %s" % [player_id, position])
	player_spawned.emit(player_id, player)

func _get_player_id() -> int:
	var env = OS.get_environment("MP_GHA_PLAYER_ID")
	return int(env) if env else -1

func _get_spawn_position(player_id: int) -> Vector2:
	"""Get spawn position for player (spread across level)"""
	var positions = [
		Vector2(100, 300),   # Player 1
		Vector2(500, 300),   # Player 2
		Vector2(900, 300),   # Player 3
		Vector2(200, 500),   # Player 4
		Vector2(600, 500),   # Player 5
		Vector2(200, 100),   # Player 6
		Vector2(600, 100),   # Player 7
		Vector2(400, 400),   # Player 8
		Vector2(800, 200),   # Player 9
		Vector2(300, 600),   # Player 10
	]

	var idx = player_id - 1  # 1-indexed to 0-indexed
	if idx >= 0 and idx < positions.size():
		return positions[idx]
	return Vector2(400, 300)  # Fallback

func open_to_lan(port: int = 0) -> bool:
	"""Open singleplayer world to LAN for other players to join"""
	if current_mode != "singleplayer":
		print("[MP-MANAGER] ERROR: Can only open to LAN in singleplayer mode")
		return false

	if lan_server_port >= 0:
		print("[MP-MANAGER] WARNING: World already open to LAN on port %d" % lan_server_port)
		return false

	# Use provided port or random port between 7778-7888
	if port <= 0:
		port = randi_range(7778, 7888)

	print("[MP-MANAGER] Opening world to LAN on port %d..." % port)

	is_server = true

	var peer = ENetMultiplayerPeer.new()
	var error = peer.create_server(port, max_players)

	if error != OK:
		print("[MP-MANAGER] ERROR: Failed to open to LAN: %d" % error)
		is_server = false
		return false

	multiplayer.multiplayer_peer = peer
	multiplayer.peer_connected.connect(_on_peer_connected)
	multiplayer.peer_disconnected.connect(_on_peer_disconnected)

	lan_server_port = port

	# Notify via chat system
	_broadcast_lan_info(port)

	print("[MP-MANAGER] World opened to LAN on port %d" % port)
	return true

func close_lan_server() -> void:
	"""Close LAN server and return to singleplayer"""
	if lan_server_port < 0:
		print("[MP-MANAGER] WARNING: LAN server not open")
		return

	print("[MP-MANAGER] Closing LAN server on port %d..." % lan_server_port)

	# Disconnect all clients gracefully
	if multiplayer.multiplayer_peer:
		for peer_id in multiplayer.get_peers():
			if peer_id != 1:  # Don't disconnect self
				multiplayer.peer_disconnected.emit(peer_id)

		multiplayer.multiplayer_peer = null

	lan_server_port = -1
	is_server = false

	# Notify via chat system
	_broadcast_lan_closed()

	print("[MP-MANAGER] LAN server closed")

func get_lan_port() -> int:
	"""Get current LAN port, or -1 if not open"""
	return lan_server_port

func _broadcast_lan_info(port: int) -> void:
	"""Send LAN info via chat system"""
	var chat_system = get_tree().root.get_node_or_null("Level/ChatSystem")
	if chat_system and chat_system.has_method("add_system_message"):
		var ip_addr = "127.0.0.1"  # For LAN, typically localhost
		# Try to get actual local IP if possible
		var env_ip = OS.get_environment("MULTIPLAYER_IP")
		if env_ip:
			ip_addr = env_ip

		chat_system.add_system_message("[SERVER] World opened to LAN!")
		chat_system.add_system_message("[PORT] Listening on port: %d" % port)
		chat_system.add_system_message("[IP] Local IP: %s" % ip_addr)
		if ServerLogger:
			ServerLogger.log(ServerLogger.LogLevel.INFO, "LAN", "World opened on %s:%d" % [ip_addr, port])

func _broadcast_lan_closed() -> void:
	"""Send LAN closed message via chat system"""
	var chat_system = get_tree().root.get_node_or_null("Level/ChatSystem")
	if chat_system and chat_system.has_method("add_system_message"):
		chat_system.add_system_message("[SERVER] LAN server closed")
		if ServerLogger:
			ServerLogger.log(ServerLogger.LogLevel.INFO, "LAN", "World closed")

func _process(_delta: float) -> void:
	"""Main game loop - tick processing"""
	if is_server or is_client:
		_process_tick()

func _process_tick() -> void:
	"""Process one game tick"""
	current_tick += 1
	tick_processed.emit(current_tick)

	# Every SNAPSHOT_INTERVAL ticks, capture snapshot (server only)
	if is_server and current_tick % SNAPSHOT_INTERVAL == 0:
		_capture_and_broadcast_snapshot()

func _capture_and_broadcast_snapshot() -> void:
	"""Capture game state and broadcast to all clients"""
	if not is_server:
		return

	# TODO: Implement snapshot capture and distribution
	# This would serialize all player states and send to clients for correction
	pass
