extends Control
## Pause menu - shown when player presses ESC

@onready var resume_button: Button = $VBoxContainer/ResumeButton
@onready var open_lan_button: Button = $VBoxContainer/OpenLanButton
@onready var close_server_button: Button = $VBoxContainer/CloseServerButton
@onready var quit_button: Button = $VBoxContainer/QuitButton
@onready var status_label: Label = $VBoxContainer/StatusLabel

var is_open: bool = false
var lan_is_open: bool = false
var admin_panel: Node = null

func _ready() -> void:
	resume_button.pressed.connect(_on_resume_pressed)
	open_lan_button.pressed.connect(_on_open_lan_pressed)
	close_server_button.pressed.connect(_on_close_server_pressed)
	quit_button.pressed.connect(_on_quit_pressed)

	# Carrega admin panel se existir
	# Tenta múltiplos caminhos possíveis
	admin_panel = get_node_or_null("../AdminPanel")  # Relativo se pausemenu é irmão
	if not admin_panel:
		admin_panel = get_node_or_null("/root/level/AdminPanel")
	if not admin_panel:
		# Procura por CanvasLayer/PanelContainer com script admin_panel
		for node in get_tree().root.get_children():
			if node.name == "Level" and node.has_node("AdminPanel"):
				admin_panel = node.get_node("AdminPanel")
				break

	if admin_panel:
		if admin_panel.has_signal("close_to_lan_requested"):
			admin_panel.close_to_lan_requested.connect(_on_admin_close_to_lan)
		if admin_panel.has_signal("shutdown_requested"):
			admin_panel.shutdown_requested.connect(_on_admin_shutdown)
		print("[PauseMenu] Admin panel connected successfully")
	else:
		print("[PauseMenu] WARNING: Admin panel not found")

	visible = false
	_update_buttons()

func _process(_delta: float) -> void:
	if Input.is_action_just_pressed("ui_cancel"):
		if visible:
			_on_resume_pressed()
		else:
			_show_pause_menu()

func _show_pause_menu() -> void:
	visible = true
	is_open = true
	get_tree().paused = true
	resume_button.grab_focus()
	_update_buttons()

	# Mostra admin panel se for host/servidor
	if admin_panel and multiplayer.is_server():
		admin_panel.show_admin_panel()
		admin_panel.refresh()

func _on_resume_pressed() -> void:
	visible = false
	is_open = false
	get_tree().paused = false

	# Esconde admin panel
	if admin_panel:
		admin_panel.hide_admin_panel()

func _on_open_lan_pressed() -> void:
	var game_data = get_node_or_null("/root/GameData")
	if not game_data or game_data.game_mode != "singleplayer":
		status_label.text = "Only available in singleplayer mode"
		return

	print("[PauseMenu] Opening world to LAN...")

	var mp_manager = get_node_or_null("/root/MultiplayerGameManager")
	if mp_manager:
		if mp_manager.open_to_lan():
			status_label.text = ""
			lan_is_open = true
			_update_buttons()
		else:
			status_label.text = "Failed to open world to LAN"
	else:
		status_label.text = "Multiplayer manager not found"

func _on_close_server_pressed() -> void:
	print("[PauseMenu] Closing LAN server...")

	var mp_manager = get_node_or_null("/root/MultiplayerGameManager")
	if mp_manager:
		mp_manager.close_lan_server()
		lan_is_open = false
		_update_buttons()
		status_label.text = ""
	else:
		status_label.text = "Multiplayer manager not found"

func _on_quit_pressed() -> void:
	print("[PauseMenu] Quitting to menu...")
	get_tree().paused = false
	var game_data = get_node_or_null("/root/GameData")
	if game_data and game_data.has_method("reset"):
		game_data.reset()
	get_tree().change_scene_to_file("res://ui/main_menu.tscn")

func _update_buttons() -> void:
	"""Update button visibility based on game state"""
	var game_data = get_node_or_null("/root/GameData")
	var mode = game_data.game_mode if game_data else ""
	match mode:
		"singleplayer":
			open_lan_button.visible = not lan_is_open
			close_server_button.visible = lan_is_open
		"multiplayer_server", "multiplayer_client":
			open_lan_button.visible = false
			close_server_button.visible = false

func _on_admin_close_to_lan() -> void:
	"""Handler quando admin fecha LAN via admin panel"""
	print("[PauseMenu] Admin closed LAN")
	lan_is_open = false
	_update_buttons()

func _on_admin_shutdown() -> void:
	"""Handler quando admin desliga servidor via admin panel"""
	print("[PauseMenu] Admin shutdown server")
	_on_quit_pressed()
