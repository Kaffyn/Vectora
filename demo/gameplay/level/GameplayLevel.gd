extends Node2D
## Main gameplay level scene
## Coordinates all game systems and zones

class_name GameplayLevel

var gameplay_manager: GameplayManager
var spawn_manager: SpawnManager
var dummy_spawners: Dictionary = {}
var enemy_spawners: Dictionary = {}

func _ready() -> void:
	var separator = "============================================================"
	print("\n" + separator)
	print("🎮 GAMEPLAY LEVEL INITIALIZED")
	print(separator + "\n")

	# Initialize managers
	_setup_managers()

	# Create zones and spawners
	_create_zones()

	# Setup UI
	_setup_ui()

	print("[GameplayLevel] Ready!")

func _setup_managers() -> void:
	"""Initialize core management systems"""
	gameplay_manager = GameplayManager.new()
	add_child(gameplay_manager)

	spawn_manager = SpawnManager.new()
	add_child(spawn_manager)

	print("[GameplayLevel] Managers initialized")

func _create_zones() -> void:
	"""Create game zones with spawners"""
	print("[GameplayLevel] Creating zones...")

	# Zone 1: Combat Arena
	_create_zone(1, "Combat Arena", 10, 15, Vector2(250, 250))

	# Zone 2: Training Grounds
	_create_zone(2, "Training Grounds", 20, 0, Vector2(750, 250))

	# Zone 3: Chaos Zone
	_create_zone(3, "Chaos Zone", 0, 10, Vector2(1250, 250))

	# Zone 4: Nature Zone
	_create_zone(4, "Nature Zone", 10, 8, Vector2(250, 750))

	# Zone 5: Corrupted Realm
	_create_zone(5, "Corrupted Realm", 0, 12, Vector2(750, 750))

	# Zone 6: Sanctuary
	_create_zone(6, "Sanctuary", 15, 0, Vector2(1250, 750))

	print("[GameplayLevel] Zones created!")

func _create_zone(zone_id: int, zone_name: String, dummy_count: int, enemy_count: int, center: Vector2) -> void:
	"""Create a single zone with spawners"""
	print("  Zone %d: %s" % [zone_id, zone_name])

	# Dummy spawner
	if dummy_count > 0:
		var dummy_spawner = DummySpawner.new()
		dummy_spawner.name = "DummySpawner%d" % zone_id
		dummy_spawner.global_position = center
		dummy_spawner.max_dummies = dummy_count
		match zone_id:
			1:
				dummy_spawner.difficulty = "basic"
			2:
				dummy_spawner.difficulty = "basic"
			3:
				dummy_spawner.difficulty = "advanced"
			4:
				dummy_spawner.difficulty = "basic"
			5:
				dummy_spawner.difficulty = "elite"
			6:
				dummy_spawner.difficulty = "basic"

		add_child(dummy_spawner)
		dummy_spawners[zone_id] = dummy_spawner

	# Enemy spawner
	if enemy_count > 0:
		var enemy_spawner = EnemySpawner.new()
		enemy_spawner.name = "EnemySpawner%d" % zone_id
		enemy_spawner.global_position = center
		enemy_spawner.max_enemies = enemy_count
		enemy_spawner.enemy_level = zone_id  # Enemies get harder by zone

		add_child(enemy_spawner)
		enemy_spawners[zone_id] = enemy_spawner

func _setup_ui() -> void:
	"""Setup user interface"""
	print("[GameplayLevel] Setting up UI...")

	# GameHUD is already created by Level.gd from the existing system
	# We'll just verify it's there
	var hud = get_tree().root.get_node_or_null("Level/HUD")
	if hud:
		print("[GameplayLevel] HUD found and connected")
	else:
		print("[GameplayLevel] WARNING: HUD not found, will be created by Level.gd")

func get_dummy_count(zone: int = -1) -> int:
	"""Get current dummy count in a zone or globally"""
	if zone > 0 and zone in dummy_spawners:
		return dummy_spawners[zone].get_active_count()

	var total = 0
	for spawner in dummy_spawners.values():
		total += spawner.get_active_count()
	return total

func get_enemy_count(zone: int = -1) -> int:
	"""Get current enemy count in a zone or globally"""
	if zone > 0 and zone in enemy_spawners:
		return enemy_spawners[zone].get_active_count()

	var total = 0
	for spawner in enemy_spawners.values():
		total += spawner.get_active_count()
	return total

func register_player(player_id: int, player_node: Node) -> void:
	"""Register a player with gameplay systems"""
	if gameplay_manager:
		gameplay_manager.register_player(player_id, player_node)

		# Spawn player at appropriate location
		var spawn_pos = spawn_manager.get_player_spawn(player_id)
		player_node.global_position = spawn_pos

		print("[GameplayLevel] Player %d spawned at %s" % [player_id, spawn_pos])

func _process(_delta: float) -> void:
	"""Update game state"""
	# Could add periodic checks here
	pass

func get_zone_info(zone: int) -> Dictionary:
	"""Get information about a zone"""
	return {
		"zone": zone,
		"dummies": dummy_spawners[zone].get_active_count() if zone in dummy_spawners else 0,
		"enemies": enemy_spawners[zone].get_active_count() if zone in enemy_spawners else 0,
	}

func get_all_zone_info() -> Array:
	"""Get info for all zones"""
	var info: Array = []
	for zone in range(1, 7):
		info.append(get_zone_info(zone))
	return info
