extends Node
## Manages weapon creation and statistics
## Factory pattern for all weapons

class_name WeaponManager

var weapon_templates: Dictionary = {
	"sword": preload("res://gameplay/weapons/Sword.gd"),
	"staff": preload("res://gameplay/weapons/Staff.gd"),
	"axe": preload("res://gameplay/weapons/Axe.gd"),
	"dagger": preload("res://gameplay/weapons/Dagger.gd"),
}

var weapon_stats: Dictionary = {
	"sword": {
		"damage": 10,
		"speed": 1.0,
		"range": 60,
		"knockback": 200,
		"type": "melee",
	},
	"staff": {
		"damage": 8,
		"speed": 1.5,
		"range": 300,
		"knockback": 150,
		"type": "ranged",
	},
	"axe": {
		"damage": 15,
		"speed": 0.7,
		"range": 80,
		"knockback": 250,
		"type": "melee",
	},
	"dagger": {
		"damage": 7,
		"speed": 2.0,
		"range": 40,
		"knockback": 150,
		"type": "melee",
	},
}

func _ready() -> void:
	print("[WeaponManager] Initialized with %d weapons" % weapon_templates.size())

func create_weapon(weapon_type: String) -> Weapon:
	"""Create a weapon instance of the given type"""
	if weapon_type not in weapon_templates:
		print("[WeaponManager] WARNING: Unknown weapon type: %s, defaulting to sword" % weapon_type)
		weapon_type = "sword"

	var weapon_script = weapon_templates[weapon_type]
	var weapon = Node2D.new()
	weapon.set_script(weapon_script)

	return weapon

func get_weapon_stats(weapon_type: String) -> Dictionary:
	"""Get statistics for a weapon type"""
	return weapon_stats.get(weapon_type, weapon_stats["sword"]).duplicate()

func is_valid_weapon_type(weapon_type: String) -> bool:
	"""Check if a weapon type exists"""
	return weapon_type in weapon_templates

func get_all_weapon_types() -> Array[String]:
	"""Get list of all available weapon types"""
	return weapon_templates.keys()

func get_weapons_by_type(weapon_class: String) -> Array[String]:
	"""Get all weapons of a certain type (melee, ranged, etc)"""
	var result: Array[String] = []
	for weapon_type in weapon_stats.keys():
		if weapon_stats[weapon_type].get("type", "") == weapon_class:
			result.append(weapon_type)
	return result
