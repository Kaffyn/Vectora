class_name Player
extends CharacterBody2D

signal died

@export var asc: ASComponent
@export var hotbar: Hotbar

# Referências internas
@onready var _sprite: Node = $ColorRect
@onready var _anim_player: AnimationPlayer = $AnimationPlayer
@onready var _hitbox: Area2D = $Hitbox
@onready var _delivery: ASDelivery = $Hitbox/ASDelivery

var is_stunned: bool = false
var player_name: String = "Player"

# Timers internos
var _coyote_timer: float = 0.0
var _jump_buffer_timer: float = 0.0
var _dash_coyote_timer: float = 0.0
var _hurt_timer: float = 0.0
var _stabilization_timer: float = 0.2

# Combat timers
var _attack_timer: float = 0.0
var _is_attacking: bool = false

# Logic State
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

var facing_direction: float = 1.0:
	set(value):
		if value != 0:
			facing_direction = value
			if scale.x != 0:
				scale.x = abs(scale.x) * sign(value)

var _is_blocked: bool:
	get: return is_stunned or current_health <= 0 or (asc and asc.has_tag(&"status.dead"))

# Dash physics params
const DASH_SPEED_MULT := 3.0
const DASH_ACCEL := 5000.0
const DASH_DURATION := 0.5
const HYPERDASH_SPEED_MULT := 3.5
const HYPERDASH_DURATION := 0.6

# Attack durations
const COMBO_DURATION := 0.4
const HEAVY_DURATION := 0.6
const DASH_ATTACK_DURATION := 0.5

# --- Attribute Accessors (lê do ASAttributeSet via ASComponent) ---

## Retorna o valor atual de um atributo AS.
func _attr(attr_name: StringName) -> float:
	if asc:
		return asc.get_attribute_value_by_tag(attr_name)
	return 0.0

## Define o valor base de um atributo AS.
func _set_attr(attr_name: StringName, value: float) -> void:
	if asc:
		asc.set_attribute_base_value_by_tag(attr_name, value)

## Atalhos para atributos frequentes.
var current_health: float:
	get: return _attr(&"health")
	set(v): _set_attr(&"health", v)

var current_stamina: float:
	get: return _attr(&"stamina")
	set(v): _set_attr(&"stamina", v)

var max_health: float:
	get: return 100.0

var max_stamina: float:
	get: return 100.0

var max_speed: float:
	get: return _attr(&"max_speed")

var default_acceleration: float:
	get: return _attr(&"acceleration")

var default_friction: float:
	get: return _attr(&"friction")

var default_air_resistance: float:
	get: return _attr(&"air_resistance")

var jump_force: float:
	get: return _attr(&"jump_force")

var gravity_scale: float:
	get: return _attr(&"gravity_scale")

var stamina_regen_rate: float:
	get: return _attr(&"stamina_regen")

var coyote_time: float:
	get: return _attr(&"coyote_time")

var jump_buffer_time: float:
	get: return _attr(&"jump_buffer_time")

func _ready() -> void:
	# Load player name from GameData
	var game_data = get_node_or_null("/root/GameData")
	if game_data:
		player_name = game_data.player_name

	if not asc:
		asc = find_child("ASComponent") as ASComponent

	# Encontra nós visuais (Fallback se necessário, mas @onready é preferível)
	if not _sprite: _sprite = find_child("ColorRect")
	if not _sprite: _sprite = find_child("Sprite2D")

	# Conecta sinais do ASComponent
	if asc:
		if asc.container:
			asc.apply_container(asc.container)
			# Destrava o catálogo de habilidades do container para o player
			for ability in asc.container.get_abilities():
				if ability:
					asc.unlock_ability_by_resource(ability)

		if not asc.is_connected("tag_changed", Callable(self, "_on_tag_changed")):
			asc.tag_changed.connect(_on_tag_changed)
			print("[Player] tag_changed signal connected")
		if _sprite:
			asc.register_node(&"ColorRect", _sprite)
		if _hitbox:
			asc.register_node(&"Hitbox", _hitbox)
		if _delivery:
			_delivery.source_component = asc

	# Inicializa Hotbar
	if not hotbar:
		# Create hotbar if not assigned in inspector
		hotbar = Hotbar.new()
		add_child(hotbar)

	if hotbar:
		if not hotbar.is_connected("selection_changed", _on_hotbar_selection_changed):
			hotbar.connect("selection_changed", _on_hotbar_selection_changed)
		# Initialize hotbar with inventory reference
		# TODO: Get inventory from player or parent

	call_deferred("_activate_initial_state")

func _on_hotbar_selection_changed(slot: int) -> void:
	if not asc or not hotbar: return

	var item = hotbar.get_item(slot)

	# 1. Limpa todas as tags de arma atuais (Lowercase weapon.*)
	var current_tags = asc.get_tags()
	for tag in current_tags:
		if String(tag).begins_with("weapon."):
			asc.remove_tag(tag)

	# 2. Concede a nova tag baseada no item
	if item and "weapon_tag" in item and not item.weapon_tag.is_empty():
		var w_tag = String(item.weapon_tag).to_lower()
		asc.add_tag(StringName(w_tag))
		print("[Player] Weapon Swapped: ", w_tag)

	# 3. Unlock de habilidade única (opcional)
	if item and "ability" in item and item.ability is ASAbility:
		asc.unlock_ability_by_resource(item.ability)

func _activate_initial_state() -> void:
	if asc:
		print("[Player] Attempting to activate motion.idle...")
		var success = asc.try_activate_ability_by_tag(&"motion.idle")
		print("[Player] Activation result: ", success)
		print("[Player] Current tags: ", asc.get_tags())
		# Garante que player está colorido com o estado inicial
		call_deferred("_on_tag_changed", &"", false)

func _physics_process(delta: float) -> void:
	if Engine.is_editor_hint(): return

	# Estabilização de Física (Frame 0 Fix)
	if _stabilization_timer > 0:
		_stabilization_timer -= delta
		move_and_slide()
		velocity = Vector2.ZERO
		return

	# Atualiza Timers
	var ct = coyote_time if coyote_time > 0 else 0.15
	if is_on_floor():
		_coyote_timer = ct
	else:
		_coyote_timer -= delta

	if _dash_coyote_timer > 0:
		_dash_coyote_timer -= delta

	if _hurt_timer > 0:
		_hurt_timer -= delta

	var jbt = jump_buffer_time if jump_buffer_time > 0 else 0.1
	if Input.is_action_just_pressed("jump"):
		_jump_buffer_timer = jbt
	else:
		_jump_buffer_timer -= delta

	# Sync internal attacking state from ASC
	_is_attacking = asc.is_ability_active(&"attack.fast") or asc.is_ability_active(&"attack.normal") or asc.is_ability_active(&"attack.special") or asc.is_ability_active(&"attack.charged")

	_handle_gravity(delta)
	_handle_input()
	_handle_movement(delta)
	move_and_slide()
	_update_state_from_physics()
	_update_context_tags()

func _handle_gravity(delta: float) -> void:
	if asc and asc.has_tag(&"motion.dash"):
		velocity.y = 0
		return

	if not is_on_floor():
		var gs = gravity_scale if gravity_scale > 0 else 1.0
		var g = 980.0 * gs
		velocity.y += g * delta

func _handle_input() -> void:
	if current_health <= 0: return

	# Dash Input
	if Input.is_action_just_pressed("dash") and current_stamina >= 10:
		if asc.try_activate_ability_by_tag(&"motion.dash"):
			asc.apply_effect_by_tag(&"effect.dash_cost")
			_start_dash()
			return

	# Jump Input
	if _jump_buffer_timer > 0 and _coyote_timer > 0 and not _is_attacking:
		if asc.try_activate_ability_by_tag(&"jump.high"):
			var jf = jump_force if jump_force > 0 else 500.0
			velocity.y = -jf
			_jump_buffer_timer = 0
			_coyote_timer = 0
			return

func _unhandled_input(event: InputEvent) -> void:
	if not asc: return
	if current_health <= 0: return

	# Weapon Switching via Hotbar
	if event.is_action_pressed("hotbar_1"):
		hotbar.select_slot(0)
	elif event.is_action_pressed("hotbar_2"):
		hotbar.select_slot(1)
	elif event.is_action_pressed("hotbar_3"):
		hotbar.select_slot(2)

	# Light Attack Input (Phased Combo)
	if Input.is_action_just_pressed("attack_light") and not _is_blocked:
		asc.try_activate_ability_by_tag(&"attack.fast")

	# Heavy Attack Input
	if Input.is_action_just_pressed("attack_heavy") and not _is_blocked:
		asc.try_activate_ability_by_tag(&"attack.charged")

# [ABILITY PHASES] The C++ core handles sequential execution of attack.fast phases.
# No manual queueing or time tracking needed here anymore.
func _finish_attack() -> void:
	# Forcefully cancel all attack abilities
	for tag in [&"attack.fast", &"attack.normal", &"attack.special", &"attack.charged"]:
		asc.cancel_ability_by_tag(tag)
	_update_state_from_physics()

func _start_dash() -> void:
	_dash_coyote_timer = DASH_DURATION
	var dash_dir = facing_direction
	var input_dir = Input.get_axis("ui_left", "ui_right")
	if input_dir: dash_dir = sign(input_dir)
	facing_direction = dash_dir

func _handle_movement(delta: float) -> void:
	var spd = max_speed if max_speed > 0 else 230.0
	var accel = default_acceleration if default_acceleration > 0 else 1200.0
	var fric = default_friction if default_friction > 0 else 1000.0
	var air_res = default_air_resistance if default_air_resistance > 0 else 200.0
	var can_control = true

	# Dash: locked movement
	if asc and asc.has_tag(&"motion.dash"):
		can_control = false
		var target_speed = spd * DASH_SPEED_MULT * facing_direction
		velocity.x = move_toward(velocity.x, target_speed, DASH_ACCEL * delta)
		if is_on_wall():
			asc.cancel_ability_by_tag(&"motion.dash")
			velocity.x = 0
			_dash_coyote_timer = 0
		return


	# Attacking: preserve momentum
	if _is_attacking:
		can_control = false
		if not is_on_floor():
			velocity.x = move_toward(velocity.x, 0, air_res * delta)
		else:
			velocity.x = move_toward(velocity.x, 0, fric * delta)
		return

	# Run modifier
	var current_max_speed = spd
	if Input.is_action_pressed("run") and current_stamina > 0 and velocity.x != 0:
		current_max_speed *= 1.5
		current_stamina -= 10.0 * delta

	# Input Direction
	var direction := Input.get_axis("ui_left", "ui_right")
	if direction and can_control:
		facing_direction = sign(direction)

	# Physics
	if can_control:
		if direction:
			velocity.x = move_toward(velocity.x, direction * current_max_speed, accel * delta)
		else:
			if is_on_floor():
				velocity.x = move_toward(velocity.x, 0, fric * delta)
			else:
				velocity.x = move_toward(velocity.x, 0, air_res * delta)

func _update_state_from_physics() -> void:
	if not asc: return
	if current_health <= 0: return
	if _is_attacking: return

	# Dash in progress
	if asc and asc.has_tag(&"motion.dash"):
		if _dash_coyote_timer <= 0:
			asc.cancel_ability_by_tag(&"motion.dash")
		else:
			return

	# Hurt
	if _hurt_timer > 0:
		asc.try_activate_ability_by_tag(&"status.hurt")
		return

	# Airborne
	if not is_on_floor():
		if velocity.y < 0:
			if not asc.is_ability_active(&"jump.high"):
				asc.try_activate_ability_by_tag(&"jump.high")
		else:
			if not asc.is_ability_active(&"jump.fall"):
				asc.try_activate_ability_by_tag(&"jump.fall")
		return

	# Ground
	var direction := Input.get_axis("ui_left", "ui_right")
	if direction != 0:
		if Input.is_action_pressed("run") and current_stamina > 0:
			if not asc.is_ability_active(&"motion.run"):
				asc.try_activate_ability_by_tag(&"motion.run")
		else:
			if not asc.is_ability_active(&"motion.walk"):
				asc.try_activate_ability_by_tag(&"motion.walk")
	else:
		if not asc.is_ability_active(&"motion.idle"):
			asc.try_activate_ability_by_tag(&"motion.idle")

func take_damage(amount: int) -> void:
	if current_health <= 0: return
	current_health -= amount
	_hurt_timer = 0.1
	if current_health <= 0:
		die()

func die() -> void:
	asc.try_activate_ability_by_tag(&"status.dead")
	emit_signal("died")
	queue_free()

func _process(delta: float) -> void:
	if Engine.is_editor_hint(): return

	# Regeneração de Stamina
	var is_consuming = false
	if (asc and (asc.has_tag(&"motion.run") or asc.has_tag(&"motion.dash"))) or _is_attacking:
		is_consuming = true

	if not is_consuming and current_stamina < max_stamina:
		var regen = stamina_regen_rate if stamina_regen_rate > 0 else 12.0
		current_stamina += regen * delta

func _draw() -> void:
	"""Draw player name above the player, never flipped"""
	if not player_name or player_name.is_empty():
		return

	# Draw name at the top of the player, always upright
	var font = ThemeDB.fallback_font
	var text_pos = Vector2(-20, -35)  # Position above player
	draw_string(font, text_pos, player_name, HORIZONTAL_ALIGNMENT_LEFT, -1, 14, Color.WHITE)

func reset_dash_timer() -> void:
	_dash_coyote_timer = 0.0

var _physics_context: int = 2 # Default to AIR (BehaviorStates.Physics.AIR = 2)

func set_physics_context(p_context: int) -> void:
	_physics_context = p_context

func _update_context_tags() -> void:
	if not asc: return

	# Determine effective context: Floor overrides Air, but Water (3) is authoritative from Level
	var ctx = _physics_context
	if is_on_floor() and ctx != 3: # 3 = WATER
		ctx = 1 # 1 = GROUND
	elif not is_on_floor() and ctx == 1:
		ctx = 2 # 2 = AIR

	# Mutually exclusive tags: ensure only one is present
	match ctx:
		1: # GROUND
			asc.remove_tag(&"physics.air")
			asc.remove_tag(&"physics.water")
			asc.add_tag(&"physics.ground")
		3: # WATER
			asc.remove_tag(&"physics.ground")
			asc.remove_tag(&"physics.air")
			asc.add_tag(&"physics.water")
		2, _: # AIR
			asc.remove_tag(&"physics.ground")
			asc.remove_tag(&"physics.water")
			asc.add_tag(&"physics.air")

func _on_tag_changed(_tag: StringName, _added: bool) -> void:
	if not asc:
		print("[ERROR] _on_tag_changed: ASComponent é null!")
		return

	var current_tags = asc.get_tags()

	var best_color = TAG_COLORS.get(&"motion.idle", Color.WHITE)

	# Encontra a tag ativa com maior prioridade (última na lista de TAG_COLORS)
	for t in current_tags:
		if t in TAG_COLORS:
			best_color = TAG_COLORS[t]

	if _sprite:
		_sprite.modulate = best_color
	else:
		print("[ERROR] _sprite é null! Node path: %s" % [$ColorRect])
		# Tenta encontrar o node novamente
		if has_node("ColorRect"):
			_sprite = $ColorRect
			_sprite.modulate = best_color
			print("[SUCCESS] ColorRect encontrado e colorido!")

# ============= MULTIPLAYER SUPPORT =============

var network_id: int = -1
var ai_controller: Node = null
var input_axis: float = 0.0

func setup_network_sync(player_id: int) -> void:
	"""Called by MultiplayerGameManager to setup this player for multiplayer"""
	network_id = player_id
	print("[Player %d] Setting up network synchronization..." % player_id)

	# Add AI controller for automatic behavior
	if player_id >= 1:  # Only AI for non-player characters
		# Try to use advanced PlayerAITree first, fallback to simple PlayerAI
		var ai_script = preload("res://resources/behavior_trees/PlayerAITree.gd")
		if ai_script:
			ai_controller = ai_script.new()
			add_child(ai_controller)
			print("[Player %d] Advanced AI tree attached" % player_id)
		else:
			# Fallback to simple AI
			var simple_ai = preload("res://player/player_ai.gd")
			ai_controller = simple_ai.new()
			add_child(ai_controller)
			if ai_controller.has_method("setup_for_network"):
				ai_controller.setup_for_network(player_id)
			print("[Player %d] Simple AI controller attached" % player_id)

	# TODO: Add MultiplayerSynchronizer for position/state sync

func set_input_axis(axis: float, _vertical: float = 0.0) -> void:
	"""Called by AI controller to set movement input"""
	input_axis = clamp(axis, -1.0, 1.0)

func try_jump() -> bool:
	"""Try to jump (called by AI)"""
	if _jump_buffer_timer > 0 and _coyote_timer > 0 and not _is_attacking:
		if asc and asc.try_activate_ability_by_tag(&"jump.high"):
			var jf = jump_force if jump_force > 0 else 500.0
			velocity.y = -jf
			_jump_buffer_timer = 0
			_coyote_timer = 0
			return true
	return false

func try_dash() -> bool:
	"""Try to dash (called by AI)"""
	if current_stamina >= 10:
		if asc and asc.try_activate_ability_by_tag(&"motion.dash"):
			asc.apply_effect_by_tag(&"effect.dash_cost")
			_start_dash()
			return true
	return false

func attack_light() -> bool:
	"""Light attack (called by AI)"""
	if not _is_blocked and asc:
		return asc.try_activate_ability_by_tag(&"attack.fast")
	return false
