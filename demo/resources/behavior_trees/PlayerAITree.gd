extends Node
## LimboAI Behavior Tree for Multiplayer Player AI
## Fully deterministic, seed-based autonomous behaviors
## Behaviors: Patrol, Combat, Ability Use

class_name PlayerAITree

var player: Node
var blackboard: Dictionary = {}
var rng: RandomNumberGenerator

# AI states
enum AIState { PATROL, COMBAT, EVADE, SPECIAL_ABILITY }

func _ready() -> void:
	player = get_parent()
	if not player or not player.is_in_group("Player"):
		queue_free()
		return

	# Initialize deterministic RNG
	var player_id = player.network_id if "network_id" in player else -1
	if player_id < 1:
		queue_free()
		return

	rng = RandomNumberGenerator.new()
	rng.seed = player_id * 12345

	# Initialize blackboard
	_init_blackboard(player_id)

	print("[PlayerAITree] Initialized for player %d" % player_id)

func _init_blackboard(player_id: int) -> void:
	"""Initialize AI state variables"""
	blackboard["player_id"] = player_id
	blackboard["ai_state"] = AIState.PATROL
	blackboard["facing_direction"] = 1.0
	blackboard["patrol_timer"] = 0.0
	blackboard["patrol_duration"] = 3.0
	blackboard["last_attack_time"] = 0.0
	blackboard["attack_cooldown"] = 1.0
	blackboard["detected_enemies"] = []
	blackboard["target_position"] = player.global_position

func _process(delta: float) -> void:
	"""Main AI loop - execute behavior tree"""
	if not player or (player.network_id if "network_id" in player else -1) < 1:
		return

	# Update blackboard timers
	blackboard["patrol_timer"] += delta
	blackboard["last_attack_time"] += delta

	# Main selector: Choose behavior based on conditions
	var chosen_behavior = _select_behavior()

	match chosen_behavior:
		AIState.COMBAT:
			_execute_combat(delta)
		AIState.EVADE:
			_execute_evade(delta)
		AIState.SPECIAL_ABILITY:
			_execute_special_ability(delta)
		AIState.PATROL:
			_execute_patrol(delta)

func _select_behavior() -> int:
	"""Selector: Choose highest priority behavior"""
	# Check health first - evade if low
	if player.current_health < 30.0:
		return AIState.EVADE

	# Check for nearby enemies
	var enemies = _detect_nearby_enemies()
	if enemies.size() > 0:
		blackboard["detected_enemies"] = enemies
		return AIState.COMBAT

	# Check if stamina high enough for ability
	if player.current_stamina > 70.0 and _should_use_ability():
		return AIState.SPECIAL_ABILITY

	# Default: patrol
	return AIState.PATROL

func _execute_patrol(_delta: float) -> void:
	"""Patrol behavior: Walk back and forth"""
	blackboard["ai_state"] = AIState.PATROL

	# Change direction every patrol_duration seconds
	if blackboard["patrol_timer"] > blackboard["patrol_duration"]:
		blackboard["facing_direction"] = -blackboard["facing_direction"]
		blackboard["patrol_timer"] = 0.0
		player.facing_direction = blackboard["facing_direction"]

	# Apply movement
	if player.has_method("set_input_axis"):
		player.set_input_axis(blackboard["facing_direction"] * 0.5)

	# Occasionally jump
	if rng.randf() < 0.02:  # 2% per frame
		_try_jump()

	# Occasionally attack
	if rng.randf() < 0.01:  # 1% per frame
		_try_attack()

	# Occasionally dash
	if rng.randf() < 0.005 and blackboard["last_attack_time"] > blackboard["attack_cooldown"]:
		_try_dash()

func _execute_combat(_delta: float) -> void:
	"""Combat behavior: Move towards target, attack"""
	blackboard["ai_state"] = AIState.COMBAT

	var enemies = blackboard.get("detected_enemies", [])
	if enemies.is_empty():
		blackboard["ai_state"] = AIState.PATROL
		return

	var target = enemies[0] as Node2D
	if not target:
		return

	# Move towards target
	var direction = sign(target.global_position.x - player.global_position.x)
	if direction != 0:
		blackboard["facing_direction"] = direction
		player.facing_direction = direction
		if player.has_method("set_input_axis"):
			player.set_input_axis(direction * 0.7)

	# Attack behavior
	var attack_selection = rng.randi_range(0, 100)
	if blackboard["last_attack_time"] > blackboard["attack_cooldown"]:
		if attack_selection < 60:  # 60% light attack
			_try_attack()
		elif attack_selection < 85:  # 25% dash attack
			_try_dash()
		else:  # 15% jump away
			_try_jump()

func _execute_evade(_delta: float) -> void:
	"""Evade behavior: Jump away and dash"""
	blackboard["ai_state"] = AIState.EVADE

	# Dash away from nearest enemy
	var enemies = _detect_nearby_enemies()
	if enemies.size() > 0:
		var target = enemies[0] as Node2D
		if target:
			var away_direction = sign(player.global_position.x - target.global_position.x)
			player.facing_direction = away_direction
			if player.has_method("set_input_axis"):
				player.set_input_axis(away_direction)

	# Jump first
	_try_jump()

	# Then dash
	if blackboard["last_attack_time"] > blackboard["attack_cooldown"]:
		_try_dash()

func _execute_special_ability(_delta: float) -> void:
	"""Special ability behavior: Use dash/hyperdash/charged attack"""
	blackboard["ai_state"] = AIState.SPECIAL_ABILITY

	var ability_selection = rng.randi_range(0, 100)

	if ability_selection < 40:  # 40% dash
		_try_dash()
	elif ability_selection < 65:  # 25% hyper dash
		if player.asc and player.asc.has_tag(&"state.hyperdash"):
			player.asc.add_tag(&"state.hyperdash")
	elif ability_selection < 85:  # 20% charged attack
		if player.asc:
			player.asc.try_activate_ability_by_tag(&"attack.charged")
	# else: 15% cooldown

	blackboard["last_attack_time"] = 0.0

func _detect_nearby_enemies() -> Array:
	"""Raycast/proximity check for nearby players"""
	var enemies: Array = []
	var detection_range = 200.0

	# Get all players
	var all_players = get_tree().get_nodes_in_group("Player")
	for p in all_players:
		if p == player:
			continue  # Skip self

		var distance = player.global_position.distance_to(p.global_position)
		if distance < detection_range:
			enemies.append(p)

	return enemies

func _should_use_ability() -> bool:
	"""Check if it's a good time to use ability"""
	return rng.randf() < 0.1  # 10% chance per frame

func _try_jump() -> void:
	"""Attempt jump"""
	if player and player.has_method("try_jump"):
		player.try_jump()

func _try_dash() -> void:
	"""Attempt dash"""
	if player and player.has_method("try_dash"):
		if player.try_dash():
			blackboard["last_attack_time"] = 0.0

func _try_attack() -> void:
	"""Attempt light attack"""
	if player and player.has_method("attack_light"):
		if player.attack_light():
			blackboard["last_attack_time"] = 0.0

func get_ai_state_name() -> String:
	"""For debugging: return current AI state as string"""
	var state = blackboard.get("ai_state", AIState.PATROL)
	match state:
		AIState.PATROL:
			return "PATROL"
		AIState.COMBAT:
			return "COMBAT"
		AIState.EVADE:
			return "EVADE"
		AIState.SPECIAL_ABILITY:
			return "SPECIAL_ABILITY"
		_:
			return "UNKNOWN"
