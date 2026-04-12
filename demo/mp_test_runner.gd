extends SceneTree

# mp_test_runner.gd
# Entry point for multiplayer tests in GHA / CLI
# Supports both single-player (headless tests) and multiplayer (10 clients coordinated via ENet)

var connected_peers = {}
var test_results = {}

# Latency simulation
var simulated_latency_ms: int = 0
var rpc_delay_queue: Array = []

func _init():
	print("[MP-RUNNER] Starting AbilitySystem Test Runner...")

	# Get player configuration from environment
	var player_id = _get_player_id()
	var is_multiplayer = player_id >= 0

	print("[MP-RUNNER] Player ID: %d" % player_id if is_multiplayer else "Player ID: -1 (Singleplayer)")
	print("[MP-RUNNER] Mode: %s" % ["MULTIPLAYER" if is_multiplayer else "UNIT_TESTS"])

	# Load simulated latency if configured
	var latency_env = OS.get_environment("MP_SIMULATED_LATENCY_MS")
	if latency_env:
		simulated_latency_ms = int(latency_env)
		print("[MP-RUNNER] Simulated latency: %dms" % simulated_latency_ms)

	if is_multiplayer:
		_setup_network(player_id)
	else:
		_run_unit_tests()

# Extract player ID from environment variable MP_GHA_PLAYER_ID
func _get_player_id() -> int:
	var env_id = OS.get_environment("MP_GHA_PLAYER_ID")
	return int(env_id) if env_id else -1

# Setup network for multiplayer tests
func _setup_network(player_id: int) -> void:
	print("[MP-RUNNER] Setting up network for Player %d..." % player_id)

	if player_id == 0:
		_server_mode()
	else:
		_client_mode(player_id)

# Server mode: Listen for clients and coordinate tests
func _server_mode() -> void:
	print("[MP-RUNNER] Starting in SERVER mode...")

	var peer = ENetMultiplayerPeer.new()
	var port_env = OS.get_environment("MP_SERVER_PORT")
	var port = int(port_env) if port_env else 7777
	var max_peers = 11  # 1 server + 10 clients

	var error = peer.create_server(port, max_peers)
	if error != OK:
		print("[ERROR] Failed to create server: %d" % error)
		quit(1)

	get_root().multiplayer.multiplayer_peer = peer
	get_root().multiplayer.peer_connected.connect(_on_peer_connected)
	get_root().multiplayer.peer_disconnected.connect(_on_peer_disconnected)

	print("[MP-RUNNER] Server listening on 127.0.0.1:%d" % port)
	print("[MP-RUNNER] Waiting for all 10 clients to connect...")

	# Await all clients to connect
	if not await _wait_for_all_clients():
		print("[ERROR] Failed to connect all clients")
		quit(1)

	# All clients connected, run tests
	_run_multiplayer_tests()

# Client mode: Connect to server and run tests
func _client_mode(player_id: int) -> void:
	print("[MP-RUNNER] Starting in CLIENT mode (Player %d)..." % player_id)

	var host = OS.get_environment("MP_SERVER_HOST") if OS.get_environment("MP_SERVER_HOST") else "127.0.0.1"
	var port_env = OS.get_environment("MP_SERVER_PORT")
	var port = int(port_env) if port_env else 7777
	var max_retries = 5

	for attempt in range(max_retries):
		print("[MP-RUNNER] Connecting to %s:%d (attempt %d/%d)..." % [host, port, attempt + 1, max_retries])

		var peer = ENetMultiplayerPeer.new()
		var error = peer.create_client(host, port)

		if error != OK:
			print("[WARN] Connection attempt %d failed: %d" % [attempt + 1, error])
			if attempt < max_retries - 1:
				await create_timer(0.5).timeout
			continue

		get_root().multiplayer.multiplayer_peer = peer
		get_root().multiplayer.connected_to_server.connect(_on_connected_to_server)
		get_root().multiplayer.server_disconnected.connect(_on_server_disconnected)

		# Wait for connection
		if await _wait_for_connection():
			print("[MP-RUNNER] Successfully connected to server!")
			_run_multiplayer_tests()
			return

	# All retry attempts failed
	print("[ERROR] Failed to connect after %d attempts" % max_retries)
	quit(1)

# Peer connection handlers
func _on_peer_connected(peer_id: int) -> void:
	print("[MP-RUNNER] Peer %d connected" % peer_id)
	connected_peers[peer_id] = true

func _on_peer_disconnected(peer_id: int) -> void:
	print("[WARN] Peer %d disconnected" % peer_id)
	connected_peers.erase(peer_id)

func _on_connected_to_server() -> void:
	print("[MP-RUNNER] Connected to server!")

func _on_server_disconnected() -> void:
	print("[ERROR] Disconnected from server!")
	quit(1)

# Wait for all 10 clients to connect (server-side)
func _wait_for_all_clients() -> bool:
	var expected_peers = 10
	var timeout = 30.0
	var elapsed = 0.0

	while len(connected_peers) < expected_peers:
		await root.process_frame
		elapsed += 0.016  # Approximate delta time (60 FPS)

		if int(elapsed) % 5 == 0 and elapsed > 0:
			print("[MP-RUNNER] Waiting... Connected %d/%d peers" % [len(connected_peers), expected_peers])

		if elapsed > timeout:
			print("[ERROR] Timeout waiting for clients (got %d/%d after %fs)" % [len(connected_peers), expected_peers, elapsed])
			return false

	print("[MP-RUNNER] All %d clients connected!" % expected_peers)
	return true

# Wait for server connection (client-side)
func _wait_for_connection() -> bool:
	var timeout = 30.0
	var elapsed = 0.0

	while not get_root().multiplayer.is_connected_to_server():
		await root.process_frame
		elapsed += 0.016  # Approximate delta time (60 FPS)

		if elapsed > timeout:
			print("[ERROR] Connection timeout after %f seconds" % timeout)
			return false

	return true

# Run unit tests only (no multiplayer)
func _run_unit_tests() -> void:
	print("[MP-RUNNER] Executing unit tests...")
	var result = AbilitySystem.run_tests()
	print("[MP-RUNNER] Unit tests completed with code %d" % result)
	quit(result)

# Run multiplayer tests
func _run_multiplayer_tests() -> void:
	print("[MP-RUNNER] Executing multiplayer tests...")

	# All clients run the same test suite
	var result = AbilitySystem.run_tests()

	print("[MP-RUNNER] Tests completed with code %d" % result)

	# Brief delay for log flush
	await root.process_frame

	quit(result)

# ============= LATENCY SIMULATION =============

func apply_simulated_latency() -> void:
	"""Apply simulated network latency (for debugging network behavior)"""
	if simulated_latency_ms > 0:
		var seconds = simulated_latency_ms / 1000.0
		await create_timer(seconds).timeout

func _process_tick_with_latency(delta: float) -> void:
	"""Process tick with latency simulation"""
	if simulated_latency_ms > 0:
		# Queue RPC calls and execute them after latency
		await apply_simulated_latency()

func get_latency_ms() -> int:
	"""Get current simulated latency"""
	return simulated_latency_ms
