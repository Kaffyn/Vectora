extends Control
## Main menu - allows player to choose game mode

@onready var single_button: Button = $VBoxContainer/ButtonsBox/SingleplayerButton
@onready var multi_button: Button = $VBoxContainer/ButtonsBox/MultiplayerButton
@onready var config_panel: PanelContainer = $VBoxContainer/ConfigPanel
@onready var ip_input: LineEdit = $VBoxContainer/ConfigPanel/MarginContainer/VBoxContainer/IpInput
@onready var port_input: LineEdit = $VBoxContainer/ConfigPanel/MarginContainer/VBoxContainer/PortInput
@onready var connect_button: Button = $VBoxContainer/ConfigPanel/MarginContainer/VBoxContainer/ButtonsBox/ConnectButton
@onready var back_button: Button = $VBoxContainer/ConfigPanel/MarginContainer/VBoxContainer/ButtonsBox/BackButton
@onready var status_label: Label = $VBoxContainer/ConfigPanel/MarginContainer/VBoxContainer/StatusLabel
@onready var waiting_label: Label = $VBoxContainer/ConfigPanel/MarginContainer/VBoxContainer/WaitingLabel

func _ready() -> void:
	single_button.pressed.connect(_on_singleplayer_pressed)
	multi_button.pressed.connect(_on_multiplayer_pressed)
	connect_button.pressed.connect(_on_connect_pressed)
	back_button.pressed.connect(_on_back_pressed)

	ip_input.text_changed.connect(_on_ip_changed)
	port_input.text_changed.connect(_on_port_changed)

	config_panel.visible = false
	var game_data = get_node_or_null("/root/GameData")
	if game_data:
		ip_input.text = game_data.server_host
		port_input.text = str(game_data.server_port)

func _on_singleplayer_pressed() -> void:
	print("[MainMenu] Starting singleplayer...")
	var game_data = get_node_or_null("/root/GameData")
	if game_data:
		game_data.game_mode = "singleplayer"
	var mp_manager = get_node_or_null("/root/MultiplayerGameManager")
	if mp_manager and mp_manager.has_method("start_game"):
		mp_manager.start_game("singleplayer")

func _on_multiplayer_pressed() -> void:
	print("[MainMenu] Showing multiplayer configuration...")
	config_panel.visible = true
	ip_input.grab_focus()

func _on_back_pressed() -> void:
	print("[MainMenu] Back to main menu")
	config_panel.visible = false
	single_button.grab_focus()

func _on_connect_pressed() -> void:
	var host = ip_input.text.strip_edges()
	var port_str = port_input.text.strip_edges()

	status_label.text = ""

	if not _validate_ip(host):
		status_label.text = "Invalid IP address"
		return

	if port_str.is_empty():
		status_label.text = "Port cannot be empty"
		return

	var port = int(port_str) if port_str.is_valid_int() else -1
	if port < 1 or port > 65535:
		status_label.text = "Invalid port (1-65535)"
		return

	print("[MainMenu] Connecting to %s:%d" % [host, port])
	var game_data = get_node_or_null("/root/GameData")
	if game_data:
		game_data.server_host = host
		game_data.server_port = port
		game_data.game_mode = "multiplayer_client"

	status_label.text = ""
	waiting_label.visible = true
	connect_button.disabled = true
	back_button.disabled = true

	var mp_manager = get_node_or_null("/root/MultiplayerGameManager")
	if mp_manager and mp_manager.has_method("start_game"):
		mp_manager.start_game("multiplayer_client", host, port)

func _validate_ip(ip: String) -> bool:
	"""Check if IP is in valid format"""
	var parts = ip.split(".")
	if parts.size() != 4:
		return false

	for part in parts:
		if not part.is_valid_int():
			return false
		var num = int(part)
		if num < 0 or num > 255:
			return false

	return true

func _on_ip_changed(_new_text: String) -> void:
	"""Update connect button state based on input validation"""
	var host = ip_input.text.strip_edges()
	var port = port_input.text.strip_edges()
	connect_button.disabled = not (_validate_ip(host) and port.is_valid_int())

func _on_port_changed(_new_text: String) -> void:
	"""Update connect button state based on input validation"""
	var host = ip_input.text.strip_edges()
	var port = port_input.text.strip_edges()
	connect_button.disabled = not (_validate_ip(host) and port.is_valid_int())
