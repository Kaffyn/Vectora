extends Resource
## Attack data resource for weapon/ability configuration

class_name AttackData

@export var id: String = ""
@export var name: String = ""
@export var description: String = ""
@export var damage: float = 10.0
@export var damage_type: StringName = &"physical"  # physical, elemental, magic
@export var cooldown_seconds: float = 1.0
@export var stamina_cost: float = 10.0
@export var knockback_force: float = 100.0
@export var range_meters: float = 2.0
@export var animation_name: String = "attack_light"
@export var attack_tags: PackedStringArray = []  # e.g. ["slash", "melee", "combo_starter"]

# Ability integration
@export var ability_id: String = ""
@export var ability_level: int = 1

func _init(
	p_id: String = "",
	p_name: String = "",
	p_damage: float = 10.0,
	p_cooldown: float = 1.0,
	p_stamina: float = 10.0
) -> void:
	id = p_id
	name = p_name
	damage = p_damage
	cooldown_seconds = p_cooldown
	stamina_cost = p_stamina

func get_display_name() -> String:
	return name if name else id

func has_tag(tag: StringName) -> bool:
	"""Check if attack has specific tag"""
	return tag in attack_tags

func add_tag(tag: StringName) -> void:
	"""Add tag to attack"""
	if tag not in attack_tags:
		attack_tags.append(tag)

func remove_tag(tag: StringName) -> void:
	"""Remove tag from attack"""
	attack_tags.erase(tag)

func is_melee() -> bool:
	"""Check if attack is melee range"""
	return range_meters <= 3.0

func is_ranged() -> bool:
	"""Check if attack is ranged"""
	return range_meters > 3.0

func can_execute(stamina: float) -> bool:
	"""Check if player has enough stamina"""
	return stamina >= stamina_cost
