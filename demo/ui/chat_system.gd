extends CanvasLayer
## In-game chat system with RPC message broadcasting
## Displays system messages and player messages

@onready var chat_container: VBoxContainer = $ChatPanel/MarginContainer/VBoxContainer
@onready var messages_box: VBoxContainer = $ChatPanel/MarginContainer/VBoxContainer/MessagesBox
@onready var input_box: LineEdit = $ChatPanel/MarginContainer/VBoxContainer/InputBox
@onready var chat_panel: PanelContainer = $ChatPanel

const MAX_MESSAGES = 10
const MESSAGE_SCENE: PackedScene = preload("res://ui/chat_message.tscn")

var messages: Array = []
var is_active: bool = false

func _ready() -> void:
	input_box.text_submitted.connect(_on_input_submitted)
	input_box.focus_exited.connect(_on_input_focus_lost)
	input_box.visible = false
	messages_box.custom_minimum_size = Vector2(300, 200)

	# Subscribe to multiplayer events
	var mp_manager = get_node_or_null("/root/MultiplayerGameManager")
	if mp_manager:
		if mp_manager.has_signal("player_spawned"):
			mp_manager.player_spawned.connect(_on_player_spawned)

func _process(_delta: float) -> void:
	# Toggle chat with 'T' key
	if Input.is_action_just_pressed("ui_text_submit"):
		if not is_active:
			_activate_input()
		else:
			_deactivate_input()

func _activate_input() -> void:
	is_active = true
	input_box.visible = true
	input_box.grab_focus()

func _deactivate_input() -> void:
	is_active = false
	input_box.visible = false
	input_box.text = ""

func _on_input_submitted(_new_text: String) -> void:
	var message = input_box.text.strip_edges()
	if not message.is_empty():
		var game_data = get_node_or_null("/root/GameData")
		var player_name = game_data.player_name if game_data else "Player"
		broadcast_message(player_name, message)
		input_box.text = ""
	_deactivate_input()

func _on_input_focus_lost() -> void:
	if is_active:
		_deactivate_input()

func add_system_message(text: String) -> void:
	"""Add a system message to chat"""
	var msg = MESSAGE_SCENE.instantiate()
	msg.setup_system(text)
	messages_box.add_child(msg)
	messages.append({"type": "system", "text": text})

	if messages.size() > MAX_MESSAGES:
		messages.pop_front()
		messages_box.get_child(0).queue_free()

func add_player_message(player_name: String, text: String) -> void:
	"""Add a player message to chat"""
	var msg = MESSAGE_SCENE.instantiate()
	msg.setup_player(player_name, text)
	messages_box.add_child(msg)
	messages.append({"type": "player", "name": player_name, "text": text})

	if messages.size() > MAX_MESSAGES:
		messages.pop_front()
		messages_box.get_child(0).queue_free()

@rpc("any_peer")
func broadcast_message(player_name: String, text: String) -> void:
	"""Broadcast a message to all clients via RPC"""
	add_player_message(player_name, text)

func _on_player_spawned(player_id: int, _player_node: Node) -> void:
	"""Called when a player spawns"""
	if player_id == 1:
		var game_data = get_node_or_null("/root/GameData")
		var player_name = game_data.player_name if game_data else "Player"
		add_system_message("[WELCOME] %s joined the game" % player_name)

func show_system_info(game_mode: String, port: int = 0) -> void:
	"""Show game mode info"""
	match game_mode:
		"singleplayer":
			add_system_message("[MODE] Singleplayer mode")
		"multiplayer_server":
			add_system_message("[MODE] Server mode - waiting for clients")
			if port > 0:
				add_system_message("[PORT] Listening on port %d" % port)
				add_system_message("[ID] Your ID is 1")
		"multiplayer_client":
			add_system_message("[MODE] Connected to multiplayer server")

func clear_messages() -> void:
	"""Clear all messages from chat"""
	for child in messages_box.get_children():
		child.queue_free()
	messages.clear()
