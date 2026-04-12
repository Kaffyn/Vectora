extends Node
## Role-Based Access Control System
## Gerencia roles, permissões e controle de acesso para jogadores

# Definições de roles e suas permissões
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

# Mapeamento de peer_id para role
var peer_roles: Dictionary = {}  # {peer_id: "role_name"}
var peer_data: Dictionary = {}   # {peer_id: {name, role, joined_at}}
var banned_peers: Dictionary = {}  # {peer_id: {banned_at, expires_at, reason}}

func _ready() -> void:
	if multiplayer.is_server():
		peer_roles[1] = "admin"
		peer_data[1] = {
			"name": "Server",
			"role": "admin",
			"peer_id": 1,
			"joined_at": Time.get_ticks_msec()
		}
		print("[RBAC] Server initialized as admin")

## Retorna a role de um peer
func get_player_role(peer_id: int) -> String:
	return peer_roles.get(peer_id, "user")

## Define a role de um peer
func set_player_role(peer_id: int, role: String) -> bool:
	if role not in ROLES:
		print("[RBAC] Invalid role: %s" % role)
		return false

	if not multiplayer.is_server():
		print("[RBAC] Only server can set roles")
		return false

	peer_roles[peer_id] = role
	if peer_id in peer_data:
		peer_data[peer_id]["role"] = role

	_broadcast_role_change.rpc(peer_id, role)
	print("[RBAC] Set %d role to %s" % [peer_id, role])
	return true

## Verifica se um peer pode executar uma ação
func can_execute(peer_id: int, action: String) -> bool:
	# Se está banido, nega tudo exceto reconexão
	if is_banned(peer_id):
		return false

	var role = get_player_role(peer_id)
	var permissions = ROLES.get(role, [])
	return action in permissions

## Registra um novo peer conectado
func register_peer(peer_id: int, player_name: String) -> void:
	if not multiplayer.is_server():
		return

	var initial_role = "admin" if peer_id == 1 else "user"
	peer_roles[peer_id] = initial_role
	peer_data[peer_id] = {
		"name": player_name,
		"role": initial_role,
		"peer_id": peer_id,
		"joined_at": Time.get_ticks_msec()
	}

	print("[RBAC] Registered peer %d: %s (role: %s)" % [peer_id, player_name, initial_role])
	if ServerLogger:
		ServerLogger.log_peer_connected(peer_id, player_name)

## Remove um peer desconectado
func unregister_peer(peer_id: int) -> void:
	if peer_id in peer_roles:
		peer_roles.erase(peer_id)
	if peer_id in peer_data:
		peer_data.erase(peer_id)
	print("[RBAC] Unregistered peer %d" % peer_id)

## Retorna lista de peers conectados
func get_connected_peers() -> Array:
	return peer_data.keys()

## Retorna número de peers conectados
func get_connected_count() -> int:
	return peer_data.size()

## Retorna dados de um peer específico
func get_peer_data(peer_id: int) -> Dictionary:
	return peer_data.get(peer_id, {})

## Retorna lista completa de peers com dados
func get_all_peers() -> Array:
	var result := []
	for peer_id in peer_data.keys():
		result.append({
			"peer_id": peer_id,
			"name": peer_data[peer_id].get("name", "Unknown"),
			"role": peer_data[peer_id].get("role", "user"),
			"joined_at": peer_data[peer_id].get("joined_at", 0)
		})
	return result

## Bane um peer (apenas server)
func ban_peer(peer_id: int, hours: float = 24.0, reason: String = "") -> bool:
	if not multiplayer.is_server():
		return false

	var now = Time.get_ticks_msec()
	var expires_at = now + int(hours * 3600 * 1000)

	banned_peers[peer_id] = {
		"banned_at": now,
		"expires_at": expires_at,
		"reason": reason
	}

	print("[RBAC] Banned peer %d for %.1f hours: %s" % [peer_id, hours, reason])
	return true

## Desban um peer
func unban_peer(peer_id: int) -> bool:
	if peer_id in banned_peers:
		banned_peers.erase(peer_id)
		print("[RBAC] Unbanned peer %d" % peer_id)
		return true
	return false

## Verifica se um peer está banido
func is_banned(peer_id: int) -> bool:
	if peer_id not in banned_peers:
		return false

	var ban_data = banned_peers[peer_id]
	var now = Time.get_ticks_msec()

	# Check se o ban expirou
	if now > ban_data["expires_at"]:
		banned_peers.erase(peer_id)
		return false

	return true

## Retorna informações de ban de um peer
func get_ban_info(peer_id: int) -> Dictionary:
	if peer_id not in banned_peers:
		return {}

	var ban_data = banned_peers[peer_id]
	var now = Time.get_ticks_msec()
	var remaining_ms = ban_data["expires_at"] - now

	return {
		"banned": true,
		"reason": ban_data.get("reason", ""),
		"remaining_hours": max(0.0, remaining_ms / (3600.0 * 1000.0))
	}

## RPC: Broadcast mudança de role para todos clientes
@rpc("authority", "call_local")
func _broadcast_role_change(peer_id: int, role: String) -> void:
	if peer_id in peer_data:
		peer_data[peer_id]["role"] = role
	print("[RBAC] Broadcasted role change: %d → %s" % [peer_id, role])

## Retorna string descritiva de um role
func get_role_description(role: String) -> String:
	match role:
		"admin":
			return "Administrador do servidor"
		"moderator":
			return "Moderador"
		"user":
			return "Usuário normal"
		_:
			return "Desconhecido"

## Debug: Imprime estado atual do RBAC
func debug_print_state() -> void:
	print("\n=== RBAC State ===")
	print("Connected peers: %d" % get_connected_count())
	for peer_info in get_all_peers():
		print("  [%d] %s - %s" % [
			peer_info["peer_id"],
			peer_info["name"],
			peer_info["role"]
		])
	print("Banned peers: %d" % banned_peers.size())
	for peer_id in banned_peers.keys():
		var ban_info = get_ban_info(peer_id)
		print("  [%d] Banned for %.1f more hours" % [peer_id, ban_info["remaining_hours"]])
	print("==================\n")
