extends Node2D
## Automatically spawns enemies at regular intervals

class_name EnemySpawner

@export var spawn_interval: float = 4.0
@export var max_enemies: int = 10
@export var enemy_level: int = 1
@export var spawn_radius: float = 200.0

var spawn_timer: float = 0.0
var spawned_enemies: Array = []

signal enemy_spawned(enemy: Enemy)

func _ready() -> void:
	add_to_group("enemy_spawners")
	print("[EnemySpawner] Initialized - level: %d, max: %d" % [enemy_level, max_enemies])

func _process(delta: float) -> void:
	# Clean dead enemies
	spawned_enemies = spawned_enemies.filter(func(e): return is_instance_valid(e))

	spawn_timer += delta

	if spawn_timer >= spawn_interval and spawned_enemies.size() < max_enemies:
		spawn_timer = 0.0
		_spawn_enemy()

func _spawn_enemy() -> void:
	"""Spawn a new enemy"""
	var enemy_script = preload("res://gameplay/enemies/Enemy.gd")
	var enemy = CharacterBody2D.new()
	enemy.set_script(enemy_script)
	enemy.level = enemy_level

	# Random spawn position
	var spawn_pos = global_position + Vector2(
		randf_range(-spawn_radius, spawn_radius),
		randf_range(-spawn_radius, spawn_radius)
	)
	enemy.global_position = spawn_pos
	get_parent().add_child(enemy)
	spawned_enemies.append(enemy)

	# Connect death signal
	enemy.enemy_died.connect(func(_e): spawned_enemies.erase(_e))

	enemy_spawned.emit(enemy)

	print("[EnemySpawner] Spawned level %d enemy (%d/%d)" % [enemy_level, spawned_enemies.size(), max_enemies])

func get_active_count() -> int:
	"""Get current number of active enemies"""
	spawned_enemies = spawned_enemies.filter(func(e): return is_instance_valid(e))
	return spawned_enemies.size()

func set_level(new_level: int) -> void:
	"""Set enemy level"""
	enemy_level = new_level
	print("[EnemySpawner] Level set to: %d" % enemy_level)

func clear_all() -> void:
	"""Clear all enemies"""
	for enemy in spawned_enemies:
		if is_instance_valid(enemy):
			enemy.queue_free()
	spawned_enemies.clear()
