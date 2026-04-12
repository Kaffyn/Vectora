extends CharacterBody2D
## Training dummy NPC
## Can be attacked, provides XP when killed

class_name Dummy

@export var difficulty: String = "basic"  # basic, advanced, elite
@export var max_health: float = 50.0
@export var experience_reward: float = 25.0

var health: float
var as_component: Node
var ai_behavior: String = "passive"  # passive, aggressive
var target: Node = null
var is_dead: bool = false

signal dummy_died(dummy: Dummy)

func _ready() -> void:
	health = max_health

	add_to_group("dummies")
	add_to_group("dummy_%s" % difficulty)

	# Create AS component
	as_component = preload("res://gameplay/ASComponentDummy.tscn").instantiate()
	add_child(as_component)

	# Visual
	_create_visual()

	# Setup AI
	if difficulty in ["advanced", "elite"]:
		ai_behavior = "aggressive"

	print("[Dummy] Created %s dummy at %s" % [difficulty, global_position])

func _create_visual() -> void:
	"""Create visual representation"""
	var color = _get_difficulty_color()

	var body = ColorRect.new()
	body.size = Vector2(30, 50)
	body.color = color
	body.position = -body.size / 2
	add_child(body)

	# Health bar background
	var health_bar_bg = ColorRect.new()
	health_bar_bg.size = Vector2(30, 5)
	health_bar_bg.color = Color.BLACK
	health_bar_bg.position = Vector2(-15, -60)
	add_child(health_bar_bg)

	# Health bar
	var health_bar = ColorRect.new()
	health_bar.size = Vector2(30, 5)
	health_bar.color = Color.GREEN
	health_bar.position = Vector2(-15, -60)
	health_bar.name = "HealthBar"
	add_child(health_bar)

	# Collision
	var collision = CollisionShape2D.new()
	var shape = CapsuleShape2D.new()
	shape.height = 50
	shape.radius = 15
	collision.shape = shape
	add_child(collision)

func _get_difficulty_color() -> Color:
	"""Get color based on difficulty"""
	match difficulty:
		"basic":
			return Color.LIGHT_BLUE
		"advanced":
			return Color.YELLOW
		"elite":
			return Color.RED
		_:
			return Color.WHITE

func _process(delta: float) -> void:
	if is_dead:
		return

	_update_health_bar()

	if ai_behavior == "aggressive":
		_update_ai(delta)

func _update_health_bar() -> void:
	"""Update health bar visual"""
	if has_node("HealthBar"):
		var health_bar = get_node("HealthBar")
		var percent = clamp(health / max_health, 0.0, 1.0)
		health_bar.size.x = 30.0 * percent

func _update_ai(delta: float) -> void:
	"""Simple AI behavior"""
	var players = get_tree().get_nodes_in_group("Player")
	if players.is_empty():
		return

	target = _find_nearest_player(players)
	if not target:
		return

	# Move towards player
	var direction = (target.global_position - global_position).normalized()
	velocity = direction * 100.0
	move_and_slide()

	# Attack if close
	if global_position.distance_to(target.global_position) < 60:
		_attack(target)

func _find_nearest_player(players: Array) -> Node:
	"""Find nearest player"""
	var nearest = players[0]
	var nearest_dist = global_position.distance_to(nearest.global_position)

	for player in players.slice(1):
		var dist = global_position.distance_to(player.global_position)
		if dist < nearest_dist:
			nearest = player
			nearest_dist = dist

	return nearest

func _attack(target: Node) -> void:
	"""Attack a target"""
	if as_component:
		as_component.perform_ability("basic_attack")

	if target.has_method("take_damage"):
		target.take_damage(5.0)

func take_damage(amount: float) -> void:
	"""Take damage"""
	if is_dead:
		return

	health -= amount

	if as_component:
		as_component.take_damage(amount)

	if health <= 0:
		_die()

func _die() -> void:
	"""Handle dummy death"""
	if is_dead:
		return

	is_dead = true

	# Grant XP to nearby players
	var players = get_tree().get_nodes_in_group("Player")
	for player in players:
		if global_position.distance_to(player.global_position) < 500:
			var gameplay_manager = _get_gameplay_manager()
			if gameplay_manager:
				gameplay_manager.grant_xp(player.network_id, experience_reward)

	# Sound
	if AudioManager:
		AudioManager.play_sound("enemy_killed")

	# Emit signal
	dummy_died.emit(self)

	# Remove after delay
	await get_tree().create_timer(0.5).timeout
	queue_free()

func apply_knockback(source: Vector2, force: float) -> void:
	"""Apply knockback effect"""
	var direction = (global_position - source).normalized()
	velocity = direction * force

func apply_effect(effect: String) -> void:
	"""Apply a status effect"""
	if as_component:
		as_component.apply_effect(effect)

func _get_gameplay_manager() -> Node:
	"""Get GameplayManager instance from level"""
	var level = get_tree().current_scene
	while level and not level.has_meta("gameplay_manager"):
		level = level.get_parent()
	return level.get_meta("gameplay_manager") if level else null
