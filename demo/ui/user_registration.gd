extends Control
## User registration screen - gets player name before main menu

@onready var name_input: LineEdit = $VBoxContainer/NameInput
@onready var continue_button: Button = $VBoxContainer/ContinueButton
@onready var error_label: Label = $VBoxContainer/ErrorLabel

func _ready() -> void:
	continue_button.pressed.connect(_on_continue_pressed)
	name_input.text_submitted.connect(_on_name_submitted)
	name_input.grab_focus()
	error_label.text = ""

func _on_continue_pressed() -> void:
	_validate_and_continue()

func _on_name_submitted(_new_text: String) -> void:
	_validate_and_continue()

func _validate_and_continue() -> void:
	var player_name = name_input.text.strip_edges()

	if player_name.is_empty():
		error_label.text = "Name cannot be empty"
		return

	if player_name.length() > 20:
		error_label.text = "Name too long (max 20 characters)"
		return

	# Save player name to GameData
	var game_data = get_node_or_null("/root/GameData")
	if game_data:
		game_data.player_name = player_name
	print("[UserReg] Player name set to: %s" % player_name)

	# Go to main menu
	get_tree().change_scene_to_file("res://ui/main_menu.tscn")
