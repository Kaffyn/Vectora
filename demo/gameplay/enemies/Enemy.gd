extends CharacterBody2D
## Hostile enemy NPC
## Attacks players, drops loot on death

class_name Enemy

@export var level: int = 1
@export var max_health: float = 50.0
@export var damage: float = 5.0
@export var experience_reward: float = 25.0
@export var movement_speed: float = 100.0
@export var attack_range: float = 60.0
@export var attack_cooldown: float = 2.0

var health: float
var as_component: Node
var ai_state: String = "patrol"  # patrol, chase, attack
var target: Node = null
var state_timer: float = 0.0
var attack_timer: float = 0.0
var is_dead: bool = false
var patrol_direction: Vector2 = Vector2.RIGHT

signal enemy_died(enemy: Enemy)

func _ready() -> void:
	health = max_health

	add_to_group("enemies")
	add_to_group("enemy_level_%d" % level)

	# Create AS component
	as_component = preload("res://gameplay/ASComponentEnemy.tscn").instantiate()
	add_child(as_component)

	# Visual
	_create_visual()

	print("[Enemy] Level %d enemy created at %s" % [level, global_position])

func _create_visual() -> void:
	"""Create enemy visual"""
	var color = Color.RED * (0.5 + (level * 0.1))

	var body = ColorRect.new()
	body.size = Vector2(35, 50)
	body.color = color
	body.position = -body.size / 2
	add_child(body)

	# Eyes
	var eye1 = ColorRect.new()
	eye1.size = Vector2(5, 5)
	eye1.color = Color.YELLOW
	eye1.position = Vector2(-8, -15)
	add_child(eye1)

	var eye2 = ColorRect.new()
	eye2.size = Vector2(5, 5)
	eye2.color = Color.YELLOW
	eye2.position = Vector2(3, -15)
	add_child(eye2)

	# Health bar
	var health_bar_bg = ColorRect.new()
	health_bar_bg.size = Vector2(35, 5)
	health_bar_bg.color = Color.BLACK
	health_bar_bg.position = Vector2(-17.5, -60)
	add_child(health_bar_bg)

	var health_bar = ColorRect.new()
	health_bar.size = Vector2(35, 5)
	health_bar.color = Color.GREEN
	health_bar.position = Vector2(-17.5, -60)
	health_bar.name = "HealthBar"
	add_child(health_bar)

	# Collision
	var collision = CollisionShape2D.new()
	var shape = CapsuleShape2D.new()
	shape.height = 50
	shape.radius = 17
	collision.shape = shape
	add_child(collision)

func _process(delta: float) -> void:
	if is_dead:
		return

	_update_health_bar()
	_update_ai(delta)

func _update_health_bar() -> void:
	"""Update health bar visual"""
	if has_node("HealthBar"):
		var health_bar = get_node("HealthBar")
		var percent = clamp(health / max_health, 0.0, 1.0)
		health_bar.size.x = 35.0 * percent

func _update_ai(delta: float) -> void:
	"""Update AI state machine"""
	match ai_state:
		"patrol":
			_patrol(delta)
		"chase":
			_chase(delta)
		"attack":
			_perform_attack(delta)

func _patrol(delta: float) -> void:
	"""Patrol behavior"""
	if state_timer <= 0:
		state_timer = randf_range(2, 5)
		patrol_direction = Vector2(randf_range(-1, 1), randf_range(-1, 1)).normalized()

	state_timer -= delta
	velocity = patrol_direction * movement_speed
	move_and_slide()

	# Look for targets
	_find_target()
	if target and global_position.distance_to(target.global_position) < 300:
		ai_state = "chase"

func _chase(delta: float) -> void:
	"""Chase target"""
	if not target or not is_instance_valid(target):
		ai_state = "patrol"
		target = null
		return

	var direction = (target.global_position - global_position).normalized()
	velocity = direction * movement_speed
	move_and_slide()

	# Check distance
	var dist = global_position.distance_to(target.global_position)
	if dist < attack_range:
		ai_state = "attack"
	elif dist > 400:
		ai_state = "patrol"
		target = null

func _perform_attack(delta: float) -> void:
	"""Attack behavior"""
	if not target or not is_instance_valid(target):
		ai_state = "chase"
		return

	var dist = global_position.distance_to(target.global_position)
	if dist > attack_range:
		ai_state = "chase"
		return

	# Attack cooldown
	attack_timer -= delta
	if attack_timer <= 0:
		attack_timer = attack_cooldown
		_attack_target()

func _find_target() -> void:
	"""Find nearest player"""
	var players = get_tree().get_nodes_in_group("Player")
	if players.is_empty():
		return

	target = players[0]
	for player in players.slice(1):
		if global_position.distance_to(player.global_position) < global_position.distance_to(target.global_position):
			target = player

func _attack_target() -> void:
	"""Attack the target"""
	if target and target.has_method("take_damage"):
		target.take_damage(damage * level)

		if AudioManager:
			AudioManager.play_sound("hit_impact")

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
	"""Handle enemy death"""
	if is_dead:
		return

	is_dead = true

	# Grant XP
	if target and target.network_id in get_tree().get_nodes_in_group("Player"):
		var gameplay_manager = _get_gameplay_manager()
		if gameplay_manager:
			gameplay_manager.grant_xp(target.network_id, experience_reward * level)

	# Sound
	if AudioManager:
		AudioManager.play_sound("enemy_killed")

	enemy_died.emit(self)

	await get_tree().create_timer(0.5).timeout
	queue_free()

func apply_knockback(source: Vector2, force: float) -> void:
	"""Apply knockback"""
	var direction = (global_position - source).normalized()
	velocity = direction * force

func apply_effect(effect: String) -> void:
	"""Apply effect"""
	if as_component:
		as_component.apply_effect(effect)

func _get_gameplay_manager() -> Node:
	"""Get GameplayManager instance from level"""
	var level = get_tree().current_scene
	while level and not level.has_meta("gameplay_manager"):
		level = level.get_parent()
	return level.get_meta("gameplay_manager") if level else null
