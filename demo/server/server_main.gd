extends Node
## Entry Point para Servidor Dedicado Headless
## Inicializa servidor multiplayer sem UI, com CLI admin

class_name ServerMain

# Configuração do servidor a partir de argumentos CLI
var server_config: Dictionary = {
	"host": "0.0.0.0",
	"port": 7777,
	"max_players": 10,
	"is_headless": true
}

func _ready() -> void:
	_parse_command_line_args()

	print("\n" + "=".repeat(50))
	print("ABILITY SYSTEM - DEDICATED SERVER")
	print("=".repeat(50) + "\n")

	# Configura Godot para headless
	if server_config["is_headless"]:
		get_tree().root.propagate_notification(NOTIFICATION_APPLICATION_FOCUS_IN)

	# Configura RBAC e AdminTerminal
	if RBAC:
		print("[Server] RBAC initialized")
	if AdminTerminal:
		AdminTerminal.set_server_info(
			server_config["host"],
			server_config["port"],
			server_config["max_players"]
		)
		AdminTerminal.server_info["is_headless"] = server_config["is_headless"]

	# Inicia servidor multiplayer
	_start_server()

	# Inicia CLI admin (bloqueia)
	if AdminTerminal:
		AdminTerminal.start_interactive_cli()

## Parse de argumentos de linha de comando
func _parse_command_line_args() -> void:
	var args = OS.get_cmdline_args()

	for i in range(0, args.size()):
		match args[i]:
			"--host":
				if i + 1 < args.size():
					server_config["host"] = args[i + 1]
			"--port":
				if i + 1 < args.size() and args[i + 1].is_valid_int():
					server_config["port"] = int(args[i + 1])
			"--max-players":
				if i + 1 < args.size() and args[i + 1].is_valid_int():
					server_config["max_players"] = int(args[i + 1])
			"--help", "-h":
				_print_help()
				get_tree().quit()

	print("[Server] Config:")
	print("  Host: %s" % server_config["host"])
	print("  Port: %d" % server_config["port"])
	print("  Max Players: %d" % server_config["max_players"])
	print()

## Inicia o servidor multiplayer
func _start_server() -> void:
	var peer = ENetMultiplayerPeer.new()

	# Bind de server
	var error = peer.create_server(
		server_config["port"],
		server_config["max_players"]
	)

	if error != OK:
		print("[Server ERROR] Failed to create server: %s" % error)
		get_tree().quit()
		return

	multiplayer.multiplayer_peer = peer

	# Conecta sinais
	multiplayer.peer_connected.connect(_on_peer_connected)
	multiplayer.peer_disconnected.connect(_on_peer_disconnected)
	multiplayer.connected_to_server.connect(_on_connected_to_server)
	multiplayer.connection_failed.connect(_on_connection_failed)

	# Registra server como peer 1
	RBAC.register_peer(1, "Server")

	print("[Server] Ready on %s:%d" % [server_config["host"], server_config["port"]])
	print("[Server] Type 'help' for commands\n")

## Callback: Peer conectado
func _on_peer_connected(peer_id: int) -> void:
	print("[Server] Player connected: peer_id=%d" % peer_id)
	RBAC.register_peer(peer_id, "Player_%d" % peer_id)

## Callback: Peer desconectado
func _on_peer_disconnected(peer_id: int) -> void:
	print("[Server] Player disconnected: peer_id=%d" % peer_id)
	RBAC.unregister_peer(peer_id)

## Callback: Conectado ao server (não aplica a servidor)
func _on_connected_to_server() -> void:
	pass

## Callback: Conexão falhou (não aplica a servidor)
func _on_connection_failed() -> void:
	pass

func _process(delta: float) -> void:
	# Log periódico de status
	if Engine.get_physics_frames() % 300 == 0:  # ~5 segundos
		_log_server_status()

## Loga status do servidor periodicamente
func _log_server_status() -> void:
	var uptime_sec = Time.get_ticks_msec() / 1000.0
	var peer_count = RBAC.get_connected_count()

	print("[Server] Uptime: %.1fs | Connected: %d/%d" % [
		uptime_sec,
		peer_count - 1,  # Subtrai o próprio server
		server_config["max_players"]
	])

## Imprime mensagem de ajuda
func _print_help() -> void:
	print("""
Usage: ability-system-server [options]

Options:
  --host <ip>           Listen on IP (default: 0.0.0.0)
  --port <port>         Listen on port (default: 7777)
  --max-players <n>     Maximum players (default: 10)
  --help                Show this message

Example:
  ability-system-server --port 8888 --max-players 20
""")

## Debug: Imprime estado do servidor
func debug_print_state() -> void:
	print("\n=== Server State ===")
	print("Is Server: %s" % multiplayer.is_server())
	print("Peer Count: %d" % RBAC.get_connected_count())
	print("Config: %s" % server_config)
	RBAC.debug_print_state()
	print("====================\n")
