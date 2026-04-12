extends Node
## Global game data - player identity and game state
## Persists across scene changes

var player_name: String = "Player"
var player_id: int = -1
var game_mode: String = ""  # "singleplayer", "multiplayer_server", "multiplayer_client"
var server_host: String = "127.0.0.1"
var server_port: int = 7777
var lan_port: int = -1  # Port when opened to LAN (-1 = not open)

func get_player_identifier() -> String:
	return "%s (ID: %d)" % [player_name, player_id]

func set_game_config(mode: String, name: String = "", host: String = "127.0.0.1", port: int = 7777) -> void:
	"""Set game configuration from menu"""
	game_mode = mode
	if name:
		player_name = name
	server_host = host
	server_port = port
	print("[GameData] Configured: mode=%s, name=%s, host=%s, port=%d" % [mode, name, host, port])

func reset() -> void:
	"""Reset to defaults"""
	player_name = "Player"
	player_id = -1
	game_mode = ""
	server_host = "127.0.0.1"
	server_port = 7777
	lan_port = -1
