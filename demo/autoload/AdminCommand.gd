extends Node
## Sistema de Processamento de Comandos de Admin
## Executa comandos validados pelo RBAC

signal command_executed(command: String, executor_id: int, success: bool)
signal player_kicked(peer_id: int, reason: String)
signal player_banned(peer_id: int, hours: float, reason: String)
signal broadcast_message_sent(message: String)
signal server_shutdown_initiated(reason: String)

## Dicionário de handlers de comando
var _command_handlers: Dictionary = {}

func _ready() -> void:
	# Registra handlers de comando
	_command_handlers = {
		"kick_player": _cmd_kick_player,
		"ban_player": _cmd_ban_player,
		"shutdown_server": _cmd_shutdown_server,
		"broadcast_message": _cmd_broadcast_message,
		"list_players": _cmd_list_players,
		"mute_player": _cmd_mute_player,
		"unmute_player": _cmd_unmute_player,
		"get_server_status": _cmd_get_server_status
	}

	print("[AdminCommand] Initialized with %d command handlers" % _command_handlers.size())

## Executa um comando se o executor tiver permissão
func execute_command(executor_id: int, command: String, args: Array = []) -> bool:
	if not RBAC.can_execute(executor_id, command):
		print("[AdminCommand] Permission denied for %d to execute '%s'" % [executor_id, command])
		if ServerLogger:
			ServerLogger.log_message(ServerLogger.LogLevel.WARN, "AdminCommand", "Permission denied for peer %d: %s" % [executor_id, command])
		command_executed.emit(command, executor_id, false)
		return false

	if command not in _command_handlers:
		print("[AdminCommand] Unknown command: %s" % command)
		if ServerLogger:
			ServerLogger.log_message(ServerLogger.LogLevel.WARN, "AdminCommand", "Unknown command: %s" % command)
		command_executed.emit(command, executor_id, false)
		return false

	var handler = _command_handlers[command]
	var success = handler.call(executor_id, args)

	if ServerLogger:
		ServerLogger.log_command(executor_id, command, success)

	command_executed.emit(command, executor_id, success)
	return success

## RPC: Executa comando de qualquer cliente (validado no server)
@rpc("any_peer")
func execute_command_rpc(command: String, args: Array = []) -> void:
	var executor_id = multiplayer.get_remote_sender_id()
	if multiplayer.is_server():
		execute_command(executor_id, command, args)

# ============= COMMAND HANDLERS =============

func _cmd_kick_player(executor_id: int, args: Array) -> bool:
	if args.size() < 1:
		print("[AdminCommand] kick_player requires peer_id")
		return false

	var target_id = int(args[0])
	var reason = args[1] if args.size() > 1 else "Kicked by admin"

	if target_id == 1:
		print("[AdminCommand] Cannot kick server")
		return false

	var player = _get_player_by_id(target_id)
	if not player:
		print("[AdminCommand] Player %d not found" % target_id)
		return false

	print("[AdminCommand] Kicking player %d: %s" % [target_id, reason])

	# Disconnect o player
	if multiplayer.is_server():
		multiplayer.peer_disconnected.emit(target_id)

	player_kicked.emit(target_id, reason)
	return true

func _cmd_ban_player(executor_id: int, args: Array) -> bool:
	if args.size() < 1:
		print("[AdminCommand] ban_player requires peer_id")
		return false

	var target_id = int(args[0])
	var hours = float(args[1]) if args.size() > 1 else 24.0
	var reason = args[2] if args.size() > 2 else "Banned by admin"

	if target_id == 1:
		print("[AdminCommand] Cannot ban server")
		return false

	RBAC.ban_peer(target_id, hours, reason)

	# Kick o player também
	var player = _get_player_by_id(target_id)
	if player:
		multiplayer.peer_disconnected.emit(target_id)

	player_banned.emit(target_id, hours, reason)
	return true

func _cmd_shutdown_server(executor_id: int, args: Array) -> bool:
	if not multiplayer.is_server():
		print("[AdminCommand] Only server can shutdown")
		return false

	var reason = args[0] if args.size() > 0 else "Server shutdown by admin"
	print("[AdminCommand] Shutting down server: %s" % reason)

	server_shutdown_initiated.emit(reason)

	# Aguarda um frame e desliga
	get_tree().call_deferred("quit")
	return true

func _cmd_broadcast_message(executor_id: int, args: Array) -> bool:
	if args.size() < 1:
		print("[AdminCommand] broadcast_message requires message text")
		return false

	var message = args[0]
	var sender_name = RBAC.get_peer_data(executor_id).get("name", "Server")
	var formatted_message = "[ADMIN] %s: %s" % [sender_name, message]

	print("[AdminCommand] Broadcasting: %s" % formatted_message)

	# Envia via RPC para chat - broadcast_message é a RPC method
	var chat_layer = _get_chat_system()
	if chat_layer:
		chat_layer.broadcast_message.rpc(sender_name, formatted_message)
	else:
		print("[AdminCommand] WARNING: Chat system not found")

	broadcast_message_sent.emit(formatted_message)
	return true

func _cmd_list_players(executor_id: int, args: Array) -> bool:
	var peers = RBAC.get_all_peers()
	print("\n=== Connected Players ===")
	for peer_info in peers:
		var role = peer_info.get("role", "user")
		print("[%d] %s - %s" % [peer_info["peer_id"], peer_info["name"], role])
	print("========================\n")
	return true

func _cmd_mute_player(executor_id: int, args: Array) -> bool:
	if args.size() < 1:
		print("[AdminCommand] mute_player requires peer_id")
		return false

	var target_id = int(args[0])
	var player = _get_player_by_id(target_id)

	if not player:
		print("[AdminCommand] Player %d not found" % target_id)
		return false

	# Tenta acessar ASComponent direto
	var asc = null
	if player.has_node("ASComponent"):
		asc = player.get_node("ASComponent")
	elif "asc" in player:
		asc = player.asc
	elif player.has_meta("asc"):
		asc = player.get_meta("asc")

	if asc and asc.has_method("add_tag"):
		asc.add_tag(&"status.muted")
		print("[AdminCommand] Muted player %d" % target_id)
		return true

	print("[AdminCommand] Could not mute player %d (no ASComponent)" % target_id)
	return false

func _cmd_unmute_player(executor_id: int, args: Array) -> bool:
	if args.size() < 1:
		print("[AdminCommand] unmute_player requires peer_id")
		return false

	var target_id = int(args[0])
	var player = _get_player_by_id(target_id)

	if not player:
		print("[AdminCommand] Player %d not found" % target_id)
		return false

	# Tenta acessar ASComponent direto
	var asc = null
	if player.has_node("ASComponent"):
		asc = player.get_node("ASComponent")
	elif "asc" in player:
		asc = player.asc
	elif player.has_meta("asc"):
		asc = player.get_meta("asc")

	if asc and asc.has_method("remove_tag"):
		asc.remove_tag(&"status.muted")
		print("[AdminCommand] Unmuted player %d" % target_id)
		return true

	print("[AdminCommand] Could not unmute player %d (no ASComponent)" % target_id)
	return false

func _cmd_get_server_status(executor_id: int, args: Array) -> bool:
	var status = {
		"is_server": multiplayer.is_server(),
		"connected_peers": RBAC.get_connected_count(),
		"uptime_ms": Time.get_ticks_msec(),
		"players": RBAC.get_all_peers()
	}

	print("\n=== Server Status ===")
	print("Is Server: %s" % status["is_server"])
	print("Connected Players: %d" % status["connected_peers"])
	print("Uptime: %.1f seconds" % (status["uptime_ms"] / 1000.0))
	print("=====================\n")
	return true

# ============= HELPERS =============

func _get_player_by_id(peer_id: int) -> Node:
	"""Encontra um player node pela peer_id"""
	var players = get_tree().get_nodes_in_group("Player")
	for player in players:
		if player.has_method("get_network_id"):
			if player.get_network_id() == peer_id:
				return player
		elif "network_id" in player:
			if player.network_id == peer_id:
				return player
	return null

func _get_chat_system() -> Node:
	"""Encontra o ChatSystem na árvore"""
	var chat_layers = get_tree().get_nodes_in_group("ChatSystem")
	if chat_layers.size() > 0:
		return chat_layers[0]

	# Fallback: procura por CanvasLayer com script ChatSystem
	for node in get_tree().get_nodes_in_group(""):
		if node.get_script() and "ChatLayer" in node.get_script().get_path():
			return node

	return null

## Retorna string formatada com ajuda de comandos
func get_help_text() -> String:
	var help = """
=== ADMIN COMMANDS ===
kick_player <peer_id> [reason]     - Expulsar jogador
ban_player <peer_id> [hours] [reason] - Banir jogador
shutdown_server [reason]           - Desligar servidor
broadcast_message <message>        - Enviar mensagem geral
list_players                       - Listar todos os jogadores
mute_player <peer_id>             - Mutar jogador
unmute_player <peer_id>           - Desmutar jogador
get_server_status                 - Status do servidor
======================
"""
	return help
