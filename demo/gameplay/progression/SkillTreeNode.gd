extends Node
## Individual skill tree node
## Represents a single skill with requirements and progression data

class_name SkillTreeNode

@export var skill_id: String = ""
@export var display_name: String = ""
@export var description: String = ""
@export var tier: int = 0
@export var cost: int = 1  # Skill points required
@export var requirements: Array[String] = []  # Required skills to unlock

@export var skill_data: Dictionary = {
	"damage_multiplier": 1.0,
	"cooldown": 0.0,
	"duration": 0.0,
	"area_radius": 0.0,
	"projectile_speed": 0.0,
	"knockback_force": 0.0,
}

var progression_system: ProgressionSystem

func _ready() -> void:
	# Add to tier group for UI organization
	add_to_group("skill_tier_%d" % tier)
	add_to_group("all_skills")

	# Get progression system reference
	progression_system = get_tree().root.get_node_or_null("ProgressionSystem")

	if not skill_id:
		push_error("[SkillTreeNode] skill_id not set!")

	if not display_name:
		display_name = skill_id.to_upper()

func is_unlocked(player_id: int) -> bool:
	"""Check if this skill is already unlocked for a player"""
	if not progression_system:
		return false

	var progress = progression_system.get_progress(player_id)
	return skill_id in progress.unlocked_skills

func can_unlock(player_id: int) -> bool:
	"""Check if a player can unlock this skill"""
	if not progression_system:
		return false

	# Already unlocked
	if is_unlocked(player_id):
		return false

	var progress = progression_system.get_progress(player_id)

	# Check skill points
	if progress.skill_points < cost:
		return false

	# Check requirements
	for req in requirements:
		if not (req in progress.unlocked_skills):
			return false

	return true

func unlock(player_id: int) -> bool:
	"""Attempt to unlock this skill for a player"""
	if not can_unlock(player_id):
		return false

	return progression_system.unlock_skill(player_id, skill_id)

func get_required_skills() -> Array[String]:
	"""Get list of required skills"""
	return requirements.duplicate()

func has_requirement(skill_id_to_check: String) -> bool:
	"""Check if this skill requires another skill"""
	return skill_id_to_check in requirements

func get_skill_data(key: String, default = null):
	"""Get a value from skill_data"""
	return skill_data.get(key, default)

func get_full_data() -> Dictionary:
	"""Get all skill information as a dictionary"""
	return {
		"id": skill_id,
		"name": display_name,
		"description": description,
		"tier": tier,
		"cost": cost,
		"requirements": requirements.duplicate(),
		"data": skill_data.duplicate(),
	}
