extends Resource
## Movement data resource for locomotion configuration

class_name MoveData

@export var id: String = ""
@export var name: String = ""
@export var description: String = ""

# Movement parameters
@export var max_speed: float = 10.0
@export var acceleration: float = 30.0
@export var friction: float = 20.0
@export var rotation_speed: float = 8.0

# Jump/aerial parameters
@export var jump_force: float = 20.0
@export var jump_stamina_cost: float = 15.0
@export var air_control: float = 0.5  # 0.0 = none, 1.0 = full
@export var gravity_scale: float = 1.0

# Dash parameters
@export var dash_speed: float = 50.0
@export var dash_duration_ms: int = 300
@export var dash_stamina_cost: float = 25.0
@export var dash_cooldown_seconds: float = 2.0

# Fall/landing parameters
@export var max_fall_speed: float = 50.0
@export var fall_damage_threshold: float = 30.0  # Velocity to take damage

# Tags
@export var move_tags: PackedStringArray = []  # e.g. ["sprint", "dash", "climb"]

func _init(
	p_id: String = "",
	p_name: String = "",
	p_max_speed: float = 10.0,
	p_jump_force: float = 20.0
) -> void:
	id = p_id
	name = p_name
	max_speed = p_max_speed
	jump_force = p_jump_force

func get_display_name() -> String:
	return name if name else id

func has_tag(tag: StringName) -> bool:
	"""Check if movement has specific tag"""
	return tag in move_tags

func add_tag(tag: StringName) -> void:
	"""Add tag to movement"""
	if tag not in move_tags:
		move_tags.append(tag)

func remove_tag(tag: StringName) -> void:
	"""Remove tag from movement"""
	move_tags.erase(tag)

func can_jump(stamina: float) -> bool:
	"""Check if player has enough stamina for jump"""
	return stamina >= jump_stamina_cost

func can_dash(stamina: float) -> bool:
	"""Check if player has enough stamina for dash"""
	return stamina >= dash_stamina_cost

func is_fast_movement() -> bool:
	"""Check if movement is high-speed"""
	return max_speed > 15.0 or dash_speed > 40.0
