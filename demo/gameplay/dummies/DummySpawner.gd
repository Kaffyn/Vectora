extends Node2D
## Automatically spawns dummies at regular intervals
## Respects max dummy count

class_name DummySpawner

@export var spawn_interval: float = 3.0
@export var max_dummies: int = 10
@export var difficulty: String = "basic"
@export var spawn_radius: float = 150.0

var spawn_timer: float = 0.0
var spawned_dummies: Array = []

signal dummy_spawned(dummy: Dummy)
signal max_dummies_reached

func _ready() -> void:
	add_to_group("dummy_spawners")
	print("[DummySpawner] Initialized - max_dummies: %d, difficulty: %s" % [max_dummies, difficulty])

func _process(delta: float) -> void:
	# Clean dead dummies
	spawned_dummies = spawned_dummies.filter(func(d): return is_instance_valid(d))

	spawn_timer += delta

	if spawn_timer >= spawn_interval and spawned_dummies.size() < max_dummies:
		spawn_timer = 0.0
		_spawn_dummy()

func _spawn_dummy() -> void:
	"""Spawn a new dummy"""
	var dummy = DummyFactory.create_dummy(difficulty)
	var spawn_pos = global_position + Vector2(
		randf_range(-spawn_radius, spawn_radius),
		randf_range(-spawn_radius, spawn_radius)
	)
	dummy.global_position = spawn_pos
	get_parent().add_child(dummy)
	spawned_dummies.append(dummy)

	# Connect death signal
	dummy.dummy_died.connect(func(_d): spawned_dummies.erase(_d))

	dummy_spawned.emit(dummy)

	print("[DummySpawner] Spawned %s dummy (%d/%d)" % [difficulty, spawned_dummies.size(), max_dummies])

func change_difficulty(new_difficulty: String) -> void:
	"""Change spawn difficulty"""
	difficulty = new_difficulty
	print("[DummySpawner] Difficulty changed to: %s" % difficulty)

func get_active_count() -> int:
	"""Get current number of active dummies"""
	spawned_dummies = spawned_dummies.filter(func(d): return is_instance_valid(d))
	return spawned_dummies.size()

func set_max_dummies(max_count: int) -> void:
	"""Set maximum dummy count"""
	max_dummies = max_count
	print("[DummySpawner] Max dummies set to: %d" % max_dummies)

func pause_spawning() -> void:
	"""Pause dummy spawning"""
	spawn_interval = 999999.0

func resume_spawning(interval: float = 3.0) -> void:
	"""Resume dummy spawning"""
	spawn_interval = interval

func clear_all() -> void:
	"""Clear all active dummies"""
	for dummy in spawned_dummies:
		if is_instance_valid(dummy):
			dummy.queue_free()
	spawned_dummies.clear()
