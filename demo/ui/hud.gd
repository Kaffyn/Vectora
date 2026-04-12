extends Control

var asc: ASComponent
var player_ref: Player
var time_elapsed: float = 0.0

@onready var label_state: Label = $Panel/MarginContainer/VBoxContainer/LabelState
@onready var label_context: Label = $Panel/MarginContainer/VBoxContainer/LabelContext

@onready var label_item: Label = $Panel2/MarginContainer/VBoxContainer/LabelItem
@onready var label_compose: RichTextLabel = $Panel2/MarginContainer/VBoxContainer/LabelCompose

var _current_state: StringName = &""
var _current_weapon: StringName = &""

const TAG_COLORS := {
	&"motion.idle": Color(0.785, 0.786, 0.761), # Light Grey
	&"motion.walk": Color(0.5, 0.5, 0.5),      # Grey
	&"motion.run": Color(0.2, 0.2, 0.9),       # Dark Blue
	&"motion.dash": Color(0.4, 0.4, 1.0),      # Medium Blue
	&"state.hyperdash": Color(0.6, 0.6, 1.0),  # Light Blue
	&"jump.high": Color(0.0, 0.5, 1.0),        # Sky Blue
	&"jump.fall": Color(1.0, 1.0, 0.0),        # Yellow
	&"status.hurt": Color(1.0, 0.5, 0.0),      # Orange
	&"status.dead": Color(1.0, 0.0, 0.0),      # Red
	&"attack.fast": Color(1.0, 0.2, 0.2),      # Soft Red (Attacks)
	&"attack.normal": Color(1.0, 0.27, 0.0),   # Orange-Red
	&"attack.special": Color(1.0, 0.0, 0.27),  # Pink-Red
	&"attack.charged": Color(0.8, 0.0, 0.0),   # Dark Red
	&"attack.dash_attack": Color(0.53, 0.0, 0.8), # Purple
}

const LISTED_ABILITIES := [
	{"name": "Idle", "tag": &"motion.idle"},
	{"name": "Walk", "tag": &"motion.walk"},
	{"name": "Run", "tag": &"motion.run"},
	{"name": "Jump", "tag": &"jump.high"},
	{"name": "Fall", "tag": &"jump.fall"},
	{"name": "Dash", "tag": &"motion.dash"},
	{"name": "HyperDash", "tag": &"state.hyperdash"},
	{"name": "Hurt", "tag": &"status.hurt"},
	{"name": "Die", "tag": &"status.dead"},
	{"name": "Combo 1", "tag": &"attack.fast", "phase": 0},
	{"name": "Combo 2", "tag": &"attack.fast", "phase": 1},
	{"name": "Combo 3", "tag": &"attack.fast", "phase": 2},
	{"name": "Heavy", "tag": &"attack.charged"},
	{"name": "Dash Attack", "tag": &"attack.dash_attack"},
]

var _activation_counts := {}
var _last_combo_phase := -1

func _on_ability_activated(spec: ASAbilitySpec) -> void:
	if not is_instance_valid(spec): return
	var ability = spec.get_ability()
	if not ability: return

	var tag = ability.get_ability_tag()
	if tag in TAG_COLORS:
		_current_state = tag

	# Incrementa contador de ativação
	for item in LISTED_ABILITIES:
		if not item.has("phase") and item.tag == tag:
			_activation_counts[tag] = _activation_counts.get(tag, 0) + 1
			break

	_update_labels()

func _ready() -> void:
	# Inicializa contadores
	for item in LISTED_ABILITIES:
		var key = String(item.tag) + (str(item.phase) if item.has("phase") else "")
		_activation_counts[key] = 0

	var level = get_parent()
	_find_asc()

func _process(delta: float) -> void:
	time_elapsed += delta
	if not is_instance_valid(asc) or not is_instance_valid(player_ref):
		_find_asc()

	_track_combo_phases()
	_update_bars()
	_update_timer()
	_update_status_list()

func _track_combo_phases() -> void:
	if not is_instance_valid(asc): return

	var active_specs = asc.get_active_abilities()
	var current_combo_phase = -1

	for spec in active_specs:
		var ability = spec.get_ability()
		if ability and ability.get_ability_tag() == &"attack.fast":
			current_combo_phase = spec.get_current_phase()
			break

	# Apenas rastreia a mudança, não incrementa (incremento já acontece em _on_ability_activated)
	if current_combo_phase != _last_combo_phase:
		_last_combo_phase = current_combo_phase

func _update_status_list() -> void:
	if not is_instance_valid(asc) or not label_compose: return

	var selected_item = player_ref.hotbar.get_selected_item() if player_ref and player_ref.hotbar else null
	var item_name = selected_item.get_display_name() if selected_item else "None"
	var final_text = "[color=#ffffff]Item: %s[/color]\n" % item_name
	final_text += "[color=#aaaaaa]States:[/color]\n"

	for item in LISTED_ABILITIES:
		var tag = item.tag
		var name = item.name
		var key = String(tag) + (str(item.phase) if item.has("phase") else "")
		var count = _activation_counts.get(key, 0)

		var is_active = false
		if item.has("phase"):
			is_active = (_last_combo_phase == item.phase)
		else:
			is_active = asc.has_tag(tag)

		var cooldown = asc.get_cooldown_remaining(tag)

		var color_hex = "666666" # Disabled color
		if is_active:
			if tag in TAG_COLORS:
				color_hex = TAG_COLORS[tag].to_html(false)
			else:
				color_hex = "ffffff"

		var line = "[color=#%s]%s (%d)[/color]" % [color_hex, name, count]
		if not item.has("phase") and cooldown > 0:
			line += " [color=#ff5555]%.1fs[/color]" % cooldown

		final_text += line + "\n"

	label_compose.text = final_text

func _find_asc() -> void:
	var player = get_tree().get_first_node_in_group("Player")
	if player and is_instance_valid(player):
		player_ref = player as Player
		asc = player.get_node_or_null("ASComponent")
		if asc:
			if not asc.is_connected("ability_activated", _on_ability_activated):
				asc.connect("ability_activated", _on_ability_activated)
			if not asc.is_connected("tag_changed", _on_asc_tag_changed):
				asc.connect("tag_changed", _on_asc_tag_changed)
				print("[HUD] tag_changed signal connected")
			# Inicializar estado atual
			_initialize_current_state()

func _format_tag_name(tag: StringName) -> String:
	"""Extrai nome legível de uma tag e capitaliza."""
	if tag.is_empty():
		return "None"
	var parts = String(tag).split(".")
	var name_part = parts[-1]
	# Converte snake_case para TitleCase
	var words = name_part.split("_")
	var formatted = ""
	for word in words:
		if not word.is_empty():
			formatted += word[0].to_upper() + word.substr(1).to_lower()
	return formatted

func _initialize_current_state() -> void:
	"""Inicializa estado atual baseado nas tags existentes."""
	if not is_instance_valid(asc): return

	var tags = asc.get_tags()
	print("[HUD] Initialize - Current tags: ", tags)
	for tag in tags:
		if String(tag).begins_with("state.") or String(tag).begins_with("motion."):
			_current_state = tag
			print("[HUD] State initialized to: ", tag)
			break
		if String(tag).begins_with("weapon."):
			_current_weapon = tag

	_update_labels()

func _on_asc_tag_changed(tag: StringName, is_present: bool) -> void:
	"""Rastreia mudanças de estado e arma."""
	print("[HUD] tag_changed: %s (present: %s)" % [tag, is_present])
	if String(tag).begins_with("state.") or String(tag).begins_with("motion."):
		if is_present:
			_current_state = tag
			print("[HUD] State updated to: %s" % tag)
			_update_labels()
		elif tag == _current_state:
			_current_state = &""
			_update_labels()
	elif String(tag).begins_with("weapon."):
		if is_present:
			_current_weapon = tag
			print("[HUD] Weapon updated to: %s" % tag)
			_update_labels()
		elif tag == _current_weapon:
			_current_weapon = &""
			_update_labels()

func _update_labels() -> void:
	if not is_instance_valid(asc) or not label_state: return

	# Usa o estado rastreado via tag_changed signal
	var display_state = _format_tag_name(_current_state) if not _current_state.is_empty() else "None"
	var motion_val = "NONE"
	var jump_val = "NONE"
	var attack_val = "NONE"
	var physics_val = "NONE"
	var status_val = "NORMAL"

	var tags = asc.get_tags()
	var weapon_display = ""

	for tag in tags:
		var s_tag = String(tag)

		if s_tag.begins_with("motion."):
			motion_val = s_tag.replace("motion.", "").to_upper()
		elif s_tag.begins_with("jump."):
			jump_val = s_tag.replace("jump.", "").to_upper()
		elif s_tag.begins_with("attack."):
			attack_val = s_tag.replace("attack.", "").to_upper()
		elif s_tag.begins_with("physics."):
			physics_val = s_tag.replace("physics.", "").to_upper()
		elif s_tag.begins_with("status."):
			status_val = s_tag.replace("status.", "").to_upper()
		elif s_tag.begins_with("weapon."):
			weapon_display = _format_tag_name(tag)

	var left_text = "State: %s\n" % display_state
	left_text += "Context: %s\n" % weapon_display
	left_text += "Motion: %s\n" % motion_val
	left_text += "Jump: %s\n" % jump_val
	left_text += "Attack: %s\n" % attack_val
	left_text += "Physics: %s\n" % physics_val
	left_text += "Status: %s" % status_val

	label_state.text = left_text
	if label_context:
		label_context.visible = false

func _update_bars() -> void:
	if not is_instance_valid(asc): return

	var hp_bar = get_node_or_null("Panel/MarginContainer/VBoxContainer/HealthBar")
	if hp_bar:
		var hp_label = hp_bar.get_node_or_null("Label")
		var cur_health = asc.get_attribute_value_by_tag(&"health")
		hp_bar.value = cur_health
		if hp_label: hp_label.text = "%d / 100" % int(cur_health)

	var st_bar = get_node_or_null("Panel/MarginContainer/VBoxContainer/StaminaBar")
	if st_bar:
		var st_label = st_bar.get_node_or_null("Label")
		var cur_stamina = asc.get_attribute_value_by_tag(&"stamina")
		st_bar.value = cur_stamina
		if st_label: st_label.text = "%d / 100" % int(cur_stamina)

func _update_timer() -> void:
	var timer_label = get_node_or_null("TimerContainer/TimerLabel")
	if not timer_label: return

	var m = int(time_elapsed / 60)
	var s = int(time_elapsed) % 60
	var ms = int((time_elapsed - int(time_elapsed)) * 100)
	timer_label.text = "%02d:%02d:%02d" % [m, s, ms]

	if _current_state in TAG_COLORS:
		timer_label.modulate = TAG_COLORS[_current_state]
