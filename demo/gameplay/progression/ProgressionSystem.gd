extends Node
## Manages player progression: XP, levels, attributes, skill points
## Integrates with Ability System for dynamic attribute scaling

class_name ProgressionSystem

class PlayerProgress:
	var level: int = 1
	var experience: float = 0.0
	var experience_to_next: float = 100.0
	var skill_points: int = 0
	var unlocked_skills: Array[String] = []
	var attributes: Dictionary = {
		"max_health": 100.0,
		"max_mana": 50.0,
		"attack_power": 10.0,
		"defense": 5.0,
		"movement_speed": 200.0,
	}

@export var base_xp_to_level: float = 100.0
@export var xp_level_multiplier: float = 1.2
@export var attribute_growth_per_level: float = 1.1

var player_progress: Dictionary = {}

signal level_up(player_id: int, new_level: int)
signal skill_point_gained(player_id: int, total_points: int)
signal xp_gained(player_id: int, amount: float)
signal skill_unlocked(player_id: int, skill_id: String)

func _ready() -> void:
	print("[ProgressionSystem] Initialized")

func register_player(player_id: int) -> void:
	"""Register a new player with default progression data"""
	if player_id in player_progress:
		return

	player_progress[player_id] = PlayerProgress.new()
	print("[ProgressionSystem] Player %d registered" % player_id)

func unregister_player(player_id: int) -> void:
	"""Remove a player's progression data"""
	if player_id in player_progress:
		player_progress.erase(player_id)

func add_experience(player_id: int, amount: float) -> void:
	"""Add experience to a player and handle level ups"""
	if player_id not in player_progress:
		register_player(player_id)

	var progress = player_progress[player_id]
	progress.experience += amount
	xp_gained.emit(player_id, amount)

	# Check for level up
	while progress.experience >= progress.experience_to_next:
		_level_up(player_id)

func _level_up(player_id: int) -> void:
	"""Handle player level up"""
	var progress = player_progress[player_id]

	progress.level += 1
	progress.experience = 0.0
	progress.experience_to_next = base_xp_to_level * pow(xp_level_multiplier, progress.level - 1)
	progress.skill_points += 1

	# Increase all attributes
	for stat in progress.attributes.keys():
		progress.attributes[stat] *= attribute_growth_per_level

	# Play sound
	if AudioManager:
		AudioManager.play_sound("level_up")

	# Broadcast
	level_up.emit(player_id, progress.level)
	skill_point_gained.emit(player_id, progress.skill_points)

	print("[ProgressionSystem] Player %d leveled up to %d" % [player_id, progress.level])

func unlock_skill(player_id: int, skill_id: String) -> bool:
	"""Attempt to unlock a skill for a player"""
	if player_id not in player_progress:
		register_player(player_id)

	var progress = player_progress[player_id]

	# Check if already unlocked
	if skill_id in progress.unlocked_skills:
		print("[ProgressionSystem] WARNING: Skill %s already unlocked for player %d" % [skill_id, player_id])
		return false

	# Check if has skill points
	if progress.skill_points <= 0:
		print("[ProgressionSystem] WARNING: Player %d has no skill points" % player_id)
		return false

	# Unlock skill
	progress.unlocked_skills.append(skill_id)
	progress.skill_points -= 1

	# Emit signals
	skill_unlocked.emit(player_id, skill_id)
	skill_point_gained.emit(player_id, progress.skill_points)

	print("[ProgressionSystem] Player %d unlocked skill: %s" % [player_id, skill_id])
	return true

func get_progress(player_id: int) -> PlayerProgress:
	"""Get a player's progression data"""
	if player_id not in player_progress:
		register_player(player_id)
	return player_progress[player_id]

func set_attribute(player_id: int, attribute: String, value: float) -> void:
	"""Set an attribute value"""
	var progress = get_progress(player_id)
	if attribute in progress.attributes:
		progress.attributes[attribute] = value

func get_attribute(player_id: int, attribute: String, default: float = 0.0) -> float:
	"""Get an attribute value"""
	var progress = get_progress(player_id)
	return progress.attributes.get(attribute, default)

func modify_attribute(player_id: int, attribute: String, delta: float) -> void:
	"""Add or subtract from an attribute"""
	var progress = get_progress(player_id)
	if attribute in progress.attributes:
		progress.attributes[attribute] += delta

func get_all_progress() -> Dictionary:
	"""Get all player progression data"""
	return player_progress.duplicate()
