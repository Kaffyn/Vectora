extends Weapon
## Dagger weapon - melee, fast, low damage, low range

class_name Dagger

func _ready() -> void:
	weapon_name = "Dagger"
	weapon_type = "dagger"
	damage = 7.0
	attack_speed = 2.0  # Fast!
	range_ = 40.0  # Short!
	knockback = 150.0
	attack_duration = 0.2

	super._ready()

func _create_visual() -> void:
	"""Create dagger visual"""
	# Blade
	var blade = Polygon2D.new()
	blade.polygon = PackedVector2Array([
		Vector2(0, -30),
		Vector2(5, -20),
		Vector2(3, 10),
		Vector2(0, 15),
		Vector2(-3, 10),
		Vector2(-5, -20),
	])
	blade.color = Color(0.9, 0.9, 1.0)  # Shiny
	add_child(blade)

	# Handle
	var handle = ColorRect.new()
	handle.size = Vector2(8, 15)
	handle.color = Color(0.5, 0.3, 0.2)
	handle.offset = Vector2(-4, 10)
	add_child(handle)

func _perform_attack(direction: Vector2) -> void:
	"""Dagger attack - quick thrust"""
	super._perform_attack(direction)

	if AudioManager:
		AudioManager.play_sound_with_variation("weapon_slash", 0.15)
