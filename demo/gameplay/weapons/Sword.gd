extends Weapon
## Sword weapon - melee, single target, balanced

class_name Sword

func _ready() -> void:
	weapon_name = "Iron Sword"
	weapon_type = "sword"
	damage = 10.0
	attack_speed = 1.0
	range_ = 60.0
	knockback = 200.0
	attack_duration = 0.3

	super._ready()

func _create_visual() -> void:
	"""Create sword visual as procedural shapes"""
	# Blade
	var blade = ColorRect.new()
	blade.size = Vector2(15, 60)
	blade.color = Color(0.8, 0.8, 0.8)  # Silver
	blade.offset = Vector2(-7.5, -30)
	add_child(blade)

	# Handle
	var handle = ColorRect.new()
	handle.size = Vector2(15, 20)
	handle.color = Color(0.4, 0.2, 0)  # Brown
	handle.offset = Vector2(-7.5, -10)
	add_child(handle)

	# Guard
	var guard = ColorRect.new()
	guard.size = Vector2(25, 5)
	guard.color = Color(1, 0.84, 0)  # Gold
	guard.offset = Vector2(-12.5, -5)
	add_child(guard)

	# Point (tip)
	var point = Polygon2D.new()
	point.polygon = PackedVector2Array([
		Vector2(0, -35),
		Vector2(5, -30),
		Vector2(-5, -30)
	])
	point.color = Color.WHITE
	add_child(point)

func _perform_attack(direction: Vector2) -> void:
	"""Sword attack - linear slash"""
	super._perform_attack(direction)

	# Rotate to face direction
	look_at(owner_player.global_position + direction.normalized() * 100)

	# Play slash sound with variation
	if AudioManager:
		AudioManager.play_sound_with_variation("weapon_slash", 0.1)
