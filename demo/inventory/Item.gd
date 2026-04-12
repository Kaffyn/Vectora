extends Resource
## Represents a single inventory item
## Can be equipment, consumable, or weapon

class_name Item

@export var id: String = ""
@export var name: String = ""
@export var description: String = ""
@export var icon: Texture2D
@export var max_stack: int = 1
@export var quantity: int = 1

# Equipment properties
@export var weapon_tag: StringName = &""
@export var armor_rating: float = 0.0
@export var stat_modifiers: Dictionary = {}  # {"health": 10, "stamina": 5}

# Ability association
@export var ability: Resource = null

func _init(p_id: String = "", p_name: String = "", p_quantity: int = 1) -> void:
	id = p_id
	name = p_name
	quantity = p_quantity

func get_display_name() -> String:
	if quantity > 1 and max_stack > 1:
		return "%s x%d" % [name, quantity]
	return name

func can_stack_with(other: Item) -> bool:
	if id != other.id:
		return false
	if max_stack <= 1:
		return false
	return quantity + other.quantity <= max_stack

func add_quantity(amount: int) -> int:
	"""Add quantity and return overflow"""
	var new_total = quantity + amount
	if new_total > max_stack:
		quantity = max_stack
		return new_total - max_stack
	quantity = new_total
	return 0

func is_equipment() -> bool:
	return weapon_tag != &"" or armor_rating > 0.0

func is_consumable() -> bool:
	return max_stack > 1 and !is_equipment()
