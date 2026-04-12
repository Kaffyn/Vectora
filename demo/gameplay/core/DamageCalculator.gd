extends Node
## Handles all damage calculations with multipliers and critical hits
## Integrates with Ability System for dynamic damage

class_name DamageCalculator

class DamageResult:
	var base_damage: float
	var multiplier: float
	var final_damage: float
	var critical: bool
	var hit_type: String  # "normal", "critical", "miss"

@export var critical_chance: float = 0.1
@export var critical_multiplier: float = 2.0
@export var miss_chance: float = 0.0

var ability_multipliers: Dictionary = {
	"basic_attack": 1.0,
	"power_slash": 1.5,
	"super_strike": 2.5,
	"dash_attack": 1.8,
	"whirlwind": 0.8,
	"multi_hit": 1.0,
	"vortex": 1.2,
	"annihilate": 4.0,
	"blademaster": 1.3,
	"singularity": 2.0,
	"sword_attack": 1.0,
	"staff_cast": 1.0,
	"axe_swing": 1.0,
}

func _ready() -> void:
	print("[DamageCalculator] Initialized with crit chance: %.1f%%" % (critical_chance * 100))

func calculate_damage(attacker: Node, defender: Node, ability: String) -> DamageResult:
	"""Main damage calculation function"""
	var result = DamageResult.new()

	# Check for miss
	if randf() < miss_chance:
		result.hit_type = "miss"
		result.final_damage = 0.0
		return result

	# Get base damage
	var attacker_power = _get_attacker_power(attacker)
	result.base_damage = attacker_power

	# Get ability multiplier
	var ability_mult = ability_multipliers.get(ability, 1.0)
	result.multiplier = ability_mult

	# Check for critical hit
	result.critical = randf() < critical_chance
	if result.critical:
		result.multiplier *= critical_multiplier
		result.hit_type = "critical"
	else:
		result.hit_type = "normal"

	# Calculate final damage
	result.final_damage = result.base_damage * result.multiplier

	# Apply defender defense reduction
	var defender_defense = _get_defender_defense(defender)
	result.final_damage = max(1.0, result.final_damage - (defender_defense * 0.5))

	return result

func _get_attacker_power(attacker: Node) -> float:
	"""Get attacker's attack power from ASComponent"""
	if attacker == null:
		return 10.0

	if attacker.has_node("ASComponent"):
		var as_comp = attacker.get_node("ASComponent") as ASComponent
		var power = as_comp.get_attribute_value_by_tag(&"attack_power")
		if power > 0:
			return power

	# Fallback: try to get from node itself if it has an asc property
	if attacker.has_meta("as_component"):
		var as_comp = attacker.get_meta("as_component") as ASComponent
		var power = as_comp.get_attribute_value_by_tag(&"attack_power")
		if power > 0:
			return power

	return 10.0

func _get_defender_defense(defender: Node) -> float:
	"""Get defender's defense from ASComponent"""
	if defender == null:
		return 0.0

	if defender.has_node("ASComponent"):
		var as_comp = defender.get_node("ASComponent") as ASComponent
		var defense = as_comp.get_attribute_value_by_tag(&"defense")
		if defense > 0:
			return defense

	return 0.0

func add_ability_multiplier(ability_id: String, multiplier: float) -> void:
	"""Dynamically add or override an ability multiplier"""
	ability_multipliers[ability_id] = multiplier
	print("[DamageCalculator] Registered ability multiplier: %s = %.1f" % [ability_id, multiplier])

func get_ability_multiplier(ability_id: String) -> float:
	"""Get the multiplier for an ability"""
	return ability_multipliers.get(ability_id, 1.0)
