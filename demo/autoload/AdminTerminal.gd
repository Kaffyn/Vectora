extends Node
## Terminal CLI Interativa para Admin
## Processa comandos digitados no console do servidor

var is_running: bool = false
var server_info: Dictionary = {
	"host": "127.0.0.1",
	"port": 7777,
	"max_players": 10,
	"is_headless": false
}

func _ready() -> void:
	print("[AdminTerminal] Initialized")

## Inicia a CLI interativa (não bloqueia - usa process frame para input)
func start_interactive_cli() -> void:
	if not multiplayer.is_server():
		print("[AdminTerminal] Only server can run CLI")
		return

	is_running = true
	print("\n" + _get_header())
	print("[SERVER] Ready on %s:%d" % [server_info["host"], server_info["port"]])
	print("[SERVER] Type 'help' for commands\n")
	print("> ", false)  # Prompt sem newline

	# Conecta ao tree_exiting para cleanup quando servidor fecha
	if not is_connected("tree_exiting", Callable(self, "_on_tree_exiting")):
		tree_exiting.connect(_on_tree_exiting)

	# Marca que CLI está rodando
	print("[AdminTerminal] CLI initialized in non-blocking mode")

## Configura parâmetros do servidor
func set_server_info(host: String, port: int, max_players: int = 10) -> void:
	server_info["host"] = host
	server_info["port"] = port
	server_info["max_players"] = max_players

## Processa via _process para não bloquear
func _process(_delta: float) -> void:
	# Nota: Em servidor headless real, isso seria substitído por leitura de stdin
	# Por enquanto, é stub para não bloquear
	pass

## Processa um comando digitado
func _process_command(input: String) -> void:
	if input.is_empty():
		return

	var parts = input.split(" ", false)
	if parts.is_empty():
		return

	var cmd = parts[0].to_lower()
	var args = parts.slice(1)

	match cmd:
		"help":
			_cmd_help()
		"list", "players":
			_cmd_list_players()
		"kick":
			_cmd_kick(args)
		"ban":
			_cmd_ban(args)
		"unban":
			_cmd_unban(args)
		"say", "broadcast":
			_cmd_broadcast(args)
		"status":
			_cmd_status()
		"shutdown", "stop":
			_cmd_shutdown(args)
		"mute":
			_cmd_mute(args)
		"unmute":
			_cmd_unmute(args)
		"clear", "cls":
			_cmd_clear()
		"exit", "quit":
			_cmd_exit()
		_:
			print("[CLI] Unknown command: '%s'. Type 'help' for available commands." % cmd)

# ============= COMMAND IMPLEMENTATIONS =============

func _cmd_help() -> void:
	print(AdminCommand.get_help_text())

func _cmd_list_players() -> void:
	AdminCommand.execute_command(1, "list_players")

func _cmd_kick(args: Array) -> void:
	if args.is_empty():
		print("[CLI] Usage: kick <peer_id> [reason]")
		return

	var peer_id = int(args[0]) if args[0].is_valid_int() else -1
	if peer_id < 0:
		print("[CLI] Invalid peer_id: %s" % args[0])
		return

	var reason = " ".join(args.slice(1)) if args.size() > 1 else "Kicked by admin"
	AdminCommand.execute_command(1, "kick_player", [peer_id, reason])

func _cmd_ban(args: Array) -> void:
	if args.is_empty():
		print("[CLI] Usage: ban <peer_id> [hours] [reason]")
		return

	var peer_id = int(args[0]) if args[0].is_valid_int() else -1
	if peer_id < 0:
		print("[CLI] Invalid peer_id: %s" % args[0])
		return

	var hours = 24.0
	var reason = "Banned by admin"

	if args.size() > 1 and args[1].is_valid_float():
		hours = float(args[1])

	if args.size() > 2:
		reason = " ".join(args.slice(2))

	AdminCommand.execute_command(1, "ban_player", [peer_id, hours, reason])

func _cmd_unban(args: Array) -> void:
	if args.is_empty():
		print("[CLI] Usage: unban <peer_id>")
		return

	var peer_id = int(args[0]) if args[0].is_valid_int() else -1
	if peer_id < 0:
		print("[CLI] Invalid peer_id: %s" % args[0])
		return

	if RBAC.unban_peer(peer_id):
		print("[CLI] Successfully unbanned peer %d" % peer_id)
	else:
		print("[CLI] Peer %d was not banned" % peer_id)

func _cmd_broadcast(args: Array) -> void:
	if args.is_empty():
		print("[CLI] Usage: say <message>")
		return

	var message = " ".join(args)
	AdminCommand.execute_command(1, "broadcast_message", [message])

func _cmd_status() -> void:
	AdminCommand.execute_command(1, "get_server_status")

func _cmd_mute(args: Array) -> void:
	if args.is_empty():
		print("[CLI] Usage: mute <peer_id>")
		return

	var peer_id = int(args[0]) if args[0].is_valid_int() else -1
	if peer_id < 0:
		print("[CLI] Invalid peer_id: %s" % args[0])
		return

	AdminCommand.execute_command(1, "mute_player", [peer_id])

func _cmd_unmute(args: Array) -> void:
	if args.is_empty():
		print("[CLI] Usage: unmute <peer_id>")
		return

	var peer_id = int(args[0]) if args[0].is_valid_int() else -1
	if peer_id < 0:
		print("[CLI] Invalid peer_id: %s" % args[0])
		return

	AdminCommand.execute_command(1, "unmute_player", [peer_id])

func _cmd_shutdown(args: Array) -> void:
	var reason = " ".join(args) if not args.is_empty() else "Server shutdown by admin"
	print("[CLI] Shutting down server: %s" % reason)
	print("[CLI] Goodbye!")

	is_running = false
	AdminCommand.execute_command(1, "shutdown_server", [reason])

func _cmd_clear() -> void:
	# Limpa console (Windows)
	if OS.get_name() == "Windows":
		OS.execute("cmd", PackedStringArray(["/c", "cls"]))
	# Linux/Mac
	else:
		OS.execute("clear", PackedStringArray())

func _cmd_exit() -> void:
	_cmd_shutdown([])

# ============= INPUT/OUTPUT =============

func _read_input() -> String:
	"""Lê uma linha do stdin"""
	# Em headless, read from stdin
	if server_info["is_headless"]:
		# Simula leitura - em produção seria stdin real
		return ""
	# No editor, não bloqueia
	return ""

func _get_header() -> String:
	var headless_str = " [HEADLESS]" if server_info["is_headless"] else ""
	return "=== ABILITY SYSTEM SERVER %s ===" % headless_str

## Printa uma linha com timestamp
func print_log(message: String, level: String = "INFO") -> void:
	var timestamp = Time.get_time_string_from_system()
	print("[%s] [%s] %s" % [timestamp, level, message])

## Callback quando árvore é destruída
func _on_tree_exiting() -> void:
	is_running = false
	print("[AdminTerminal] CLI shutting down")

## Debug: Mostra estado do terminal
func debug_print_state() -> void:
	print("\n=== AdminTerminal State ===")
	print("Running: %s" % is_running)
	print("Host: %s" % server_info["host"])
	print("Port: %d" % server_info["port"])
	print("Max Players: %d" % server_info["max_players"])
	print("Headless: %s" % server_info["is_headless"])
	print("===========================\n")
