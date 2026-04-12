extends Weapon
## Axe weapon - melee, AOE, heavy, slow

class_name Axe

func _ready() -> void:
	weapon_name = "Battle Axe"
	weapon_type = "axe"
	damage = 15.0
	attack_speed = 0.7
	range_ = 80.0
	knockback = 250.0
	attack_duration = 0.4

	super._ready()

func _create_visual() -> void:
	"""Create axe visual"""
	# Blade head
	var blade = Polygon2D.new()
	blade.polygon = PackedVector2Array([
		Vector2(-20, -40),
		Vector2(20, -40),
		Vector2(25, -30),
		Vector2(20, -20),
		Vector2(-20, -20),
		Vector2(-25, -30),
	])
	blade.color = Color(0.8, 0.8, 0.8)
	add_child(blade)

	# Handle
	var handle = ColorRect.new()
	handle.size = Vector2(6, 50)
	handle.color = Color(0.4, 0.2, 0)
	handle.offset = Vector2(-3, -20)
	add_child(handle)

	# Bottom weight
	var weight = ColorRect.new()
	weight.size = Vector2(10, 8)
	weight.color = Color(0.3, 0.3, 0.3)
	weight.offset = Vector2(-5, -8)
	add_child(weight)

func _perform_attack(direction: Vector2) -> void:
	"""Axe attack - AOE swing"""
	super._perform_attack(direction)

	# Create AOE detection
	var aoe_area = Area2D.new()
	var circle_shape = CircleShape2D.new()
	circle_shape.radius = range_
	var collision = CollisionShape2D.new()
	collision.shape = circle_shape
	aoe_area.add_child(collision)

	aoe_area.global_position = owner_player.global_position

	# Get all targets in AOE
	var targets = aoe_area.get_overlapping_areas()
	var hit_count = 0

	for target_area in targets:
		var target = target_area.get_parent()
		if target != owner_player and target not in hit_detector.hit_targets:
			if target.has_method("take_damage"):
				target.take_damage(get_damage())
				if target.has_method("apply_knockback"):
					target.apply_knockback(global_position, knockback)
				hit_count += 1
				hit_detector.hit_targets[target] = true

	aoe_area.queue_free()

	if AudioManager:
		AudioManager.play_sound("weapon_heavy_swing")
