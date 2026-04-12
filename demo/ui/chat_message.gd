extends Container
## Individual chat message display

@onready var label: Label = $Label

func setup_system(text: String) -> void:
	label.text = text
	label.add_theme_color_override("font_color", Color(1, 1, 0.5, 1))

func setup_player(player_name: String, text: String) -> void:
	label.text = "%s: %s" % [player_name, text]
	label.add_theme_color_override("font_color", Color(0.8, 0.8, 0.8, 1))
