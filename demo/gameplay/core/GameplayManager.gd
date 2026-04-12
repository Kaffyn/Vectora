extends Node
## Central gameplay manager - coordinates all game systems
## Manages progression, combat, weapons, and level

class_name GameplayManager

# System references
var progression_system: ProgressionSystem
var damage_calculator: DamageCalculator
var weapon_manager: WeaponManager
var skill_tree: Node
var crafting_system: Node
var audio_manager: Node

var level: Node
var active_players: Dictionary = {}

signal gameplay_started
signal gameplay_ended

func _ready() -> void:
	print("[GameplayManager] Initializing...")
	_init_systems()
	gameplay_started.emit()

func _init_systems() -> void:
	# Create progression system
	progression_system = ProgressionSystem.new()
	add_child(progression_system)

	# Create damage calculator
	damage_calculator = DamageCalculator.new()
	add_child(damage_calculator)

	# Create weapon manager
	weapon_manager = WeaponManager.new()
	add_child(weapon_manager)

	# Get audio manager reference
	audio_manager = get_tree().root.get_node_or_null("AudioManager")
	if not audio_manager:
		audio_manager = AudioManager

	print("[GameplayManager] Systems initialized")

func register_player(player_id: int, player_node: Node) -> void:
	"""Register a player with the gameplay system"""
	if player_id in active_players:
		print("[GameplayManager] WARNING: Player %d already registered" % player_id)
		return

	active_players[player_id] = player_node
	progression_system.register_player(player_id)

	print("[GameplayManager] Player %d registered" % player_id)

func unregister_player(player_id: int) -> void:
	"""Unregister a player from the gameplay system"""
	if player_id not in active_players:
		return

	active_players.erase(player_id)
	progression_system.unregister_player(player_id)

	print("[GameplayManager] Player %d unregistered" % player_id)

func apply_damage(attacker: Node, defender: Node, ability: String) -> float:
	"""Calculate and apply damage with full integration"""
	if not damage_calculator:
		return 0.0

	var result = damage_calculator.calculate_damage(attacker, defender, ability)

	# Apply to defender
	if defender.has_method("take_damage"):
		defender.take_damage(result.final_damage)

	# Apply knockback
	if defender.has_method("apply_knockback"):
		var kb_force = 200.0
		defender.apply_knockback(attacker.global_position, kb_force)

	# Play sound
	if audio_manager:
		if result.critical:
			audio_manager.play_sound("critical_hit")
		else:
			audio_manager.play_sound("hit_impact")

	# Show damage feedback
	_show_damage_feedback(defender, result)

	return result.final_damage

func grant_xp(player_id: int, amount: float) -> void:
	"""Grant experience to a player"""
	if player_id in active_players:
		progression_system.add_experience(player_id, amount)

func grant_skill_unlock(player_id: int, skill_id: String) -> bool:
	"""Attempt to unlock a skill for a player"""
	if player_id in active_players:
		return progression_system.unlock_skill(player_id, skill_id)
	return false

func _show_damage_feedback(target: Node, result: DamageCalculator.DamageResult) -> void:
	"""Show damage number feedback"""
	# Will be handled by CombatFeedback UI
	pass

func get_player_progress(player_id: int) -> ProgressionSystem.PlayerProgress:
	"""Get a player's progression data"""
	return progression_system.get_progress(player_id)

func get_player_node(player_id: int) -> Node:
	"""Get a player node by ID"""
	return active_players.get(player_id, null)

func _process(_delta: float) -> void:
	# Could add periodic tasks here
	pass
