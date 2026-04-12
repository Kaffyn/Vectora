extends Node
## Manages spawn points for players and dummies

class_name SpawnManager

@export var player_spawn_points: Array[Vector2] = [
	Vector2(100, 100),
	Vector2(200, 100),
	Vector2(100, 200),
	Vector2(200, 200),
]

@export var dummy_spawn_points: Array[Vector2] = [
	Vector2(150, 300),
	Vector2(250, 300),
	Vector2(350, 300),
	Vector2(150, 400),
	Vector2(250, 400),
	Vector2(350, 400),
	Vector2(150, 500),
	Vector2(250, 500),
	Vector2(350, 500),
]

@export var zone_spawn_points: Dictionary = {
	1: [Vector2(100, 300), Vector2(200, 300), Vector2(300, 300)],  # Zone 1
	2: [Vector2(600, 300), Vector2(700, 300), Vector2(800, 300)],  # Zone 2
	3: [Vector2(1100, 300), Vector2(1200, 300), Vector2(1300, 300)],  # Zone 3
	4: [Vector2(100, 800), Vector2(200, 800), Vector2(300, 800)],  # Zone 4
	5: [Vector2(600, 800), Vector2(700, 800), Vector2(800, 800)],  # Zone 5
	6: [Vector2(1100, 800), Vector2(1200, 800), Vector2(1300, 800)],  # Zone 6
}

func _ready() -> void:
	print("[SpawnManager] Initialized with %d player spawns, %d dummy spawns" % [
		player_spawn_points.size(),
		dummy_spawn_points.size()
	])

func get_player_spawn(player_id: int) -> Vector2:
	"""Get spawn point for a player"""
	if player_spawn_points.is_empty():
		return Vector2.ZERO
	return player_spawn_points[player_id % player_spawn_points.size()]

func get_random_player_spawn() -> Vector2:
	"""Get a random player spawn point"""
	if player_spawn_points.is_empty():
		return Vector2.ZERO
	return player_spawn_points[randi() % player_spawn_points.size()]

func get_dummy_spawn(dummy_index: int = -1) -> Vector2:
	"""Get a spawn point for a dummy"""
	if dummy_spawn_points.is_empty():
		return Vector2.ZERO
	if dummy_index < 0:
		return dummy_spawn_points[randi() % dummy_spawn_points.size()]
	return dummy_spawn_points[dummy_index % dummy_spawn_points.size()]

func get_zone_spawn(zone: int) -> Vector2:
	"""Get a random spawn point in a zone"""
	if zone not in zone_spawn_points:
		return Vector2.ZERO
	var spawns = zone_spawn_points[zone]
	if spawns.is_empty():
		return Vector2.ZERO
	return spawns[randi() % spawns.size()]

func add_player_spawn(position: Vector2) -> void:
	"""Add a new player spawn point"""
	player_spawn_points.append(position)

func add_dummy_spawn(position: Vector2) -> void:
	"""Add a new dummy spawn point"""
	dummy_spawn_points.append(position)

func add_zone_spawn(zone: int, position: Vector2) -> void:
	"""Add a spawn point to a zone"""
	if zone not in zone_spawn_points:
		zone_spawn_points[zone] = []
	zone_spawn_points[zone].append(position)

func get_all_spawn_points() -> Dictionary:
	"""Get all spawn points"""
	return {
		"player": player_spawn_points.duplicate(),
		"dummy": dummy_spawn_points.duplicate(),
		"zones": zone_spawn_points.duplicate(),
	}
