extends Node
## Simple deterministic AI for multiplayer players
## Handles: walking, jumping, dashing, attacking
## Seed-based random for reproducibility

class_name PlayerAI

var player: Player
var network_id: int = -1

# AI state machine
var current_behavior: String = "patrol"  # patrol, combat, evade
var behavior_timer: float = 0.0
var direction: float = 1.0  # 1.0 = right, -1.0 = left
var rng: RandomNumberGenerator

# Behavior timings
var patrol_duration: float = 3.0
var combat_check_range: float = 200.0
var combat_cooldown: float = 1.0
var attack_cooldown_timer: float = 0.0

func _ready() -> void:
	player = get_parent()
	if not player or not player is Player:
		queue_free()
		return

	# Detect if this is a multiplayer player
	network_id = player.network_id if "network_id" in player else -1
	if network_id < 1:
		# Singleplayer, don't use AI
		queue_free()
		return

	# Seed RNG for deterministic behavior
	rng = RandomNumberGenerator.new()
	rng.seed = network_id * 12345  # Unique seed per player

	print("[Player AI %d] Initialized with seed %d" % [network_id, rng.seed])

func _process(delta: float) -> void:
	if network_id < 1:
		return

	behavior_timer += delta
	attack_cooldown_timer += delta

	# Main AI loop
	match current_behavior:
		"patrol":
			_patrol(delta)
		"combat":
			_combat(delta)

func _patrol(delta: float) -> void:
	"""Simple patrol behavior: walk back and forth"""
	# Change direction every patrol_duration seconds
	if behavior_timer > patrol_duration:
		direction = -direction
		behavior_timer = 0.0
		player.facing_direction = direction

	# Walk
	_apply_movement_input(direction * 0.5)

	# Occasionally jump
	if rng.randf() < 0.02:  # 2% chance per frame
		_try_jump()

	# Occasionally attack
	if rng.randf() < 0.01:  # 1% chance per frame
		_try_attack()

	# Occasionally dash
	if rng.randf() < 0.005 and attack_cooldown_timer > combat_cooldown:  # 0.5% chance
		_try_dash()

func _combat(delta: float) -> void:
	"""Combat behavior (placeholder)"""
	# TODO: Implement combat logic
	# For now, just patrol
	_patrol(delta)

func _apply_movement_input(direction: float) -> void:
	"""Apply movement input to player"""
	if player and player.has_method("set_input_axis"):
		player.set_input_axis(direction, 0.0)

func _try_jump() -> void:
	"""Try to jump if on ground"""
	if not player:
		return

	# Check if player is on ground (has "physics.ground" tag)
	if player.asc and player.asc.has_tag(&"physics.ground"):
		if player.has_method("try_jump"):
			player.try_jump()
		elif player.has_method("jump"):
			player.jump()

func _try_dash() -> void:
	"""Try to dash if stamina sufficient"""
	if not player or attack_cooldown_timer < combat_cooldown:
		return

	# Check stamina
	if player.current_stamina < 20.0:
		return

	if player.has_method("try_dash"):
		player.try_dash()
		attack_cooldown_timer = 0.0
	elif player.asc:
		# Try to trigger dash via ASComponent
		player.asc.add_tag(&"motion.dash")
		attack_cooldown_timer = 0.0

func _try_attack() -> void:
	"""Try to attack"""
	if not player or attack_cooldown_timer < combat_cooldown:
		return

	if player.has_method("attack_light"):
		player.attack_light()
		attack_cooldown_timer = 0.0
	elif player.asc:
		# Try to trigger attack via ASComponent
		player.asc.add_tag(&"attack.fast")
		attack_cooldown_timer = 0.0

func setup_for_network(pid: int) -> void:
	"""Called by MultiplayerGameManager to setup network"""
	network_id = pid
	if rng:
		rng.seed = network_id * 12345
