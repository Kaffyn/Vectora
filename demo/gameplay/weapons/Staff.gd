extends Weapon
## Staff weapon - ranged, projectile-based, magical

class_name Staff

var projectile_scene: PackedScene = null

func _ready() -> void:
	weapon_name = "Magic Staff"
	weapon_type = "staff"
	damage = 8.0
	attack_speed = 1.5
	range_ = 300.0
	knockback = 150.0
	attack_duration = 0.4

	super._ready()

func _create_visual() -> void:
	"""Create staff visual"""
	# Shaft (staff pole)
	var shaft = ColorRect.new()
	shaft.size = Vector2(8, 80)
	shaft.color = Color(0.6, 0.3, 0.1)  # Brown wood
	shaft.offset = Vector2(-4, -40)
	add_child(shaft)

	# Orb (magical top)
	var orb = ColorRect.new()
	orb.size = Vector2(25, 25)
	orb.color = Color(0.8, 0.2, 0.8)  # Purple
	orb.offset = Vector2(-12.5, -50)
	add_child(orb)

	# Glow
	var glow = Polygon2D.new()
	glow.polygon = PackedVector2Array([
		Vector2(0, -50),
		Vector2(15, -40),
		Vector2(10, -60),
		Vector2(0, -65),
		Vector2(-10, -60),
		Vector2(-15, -40),
	])
	glow.color = Color(1, 0.5, 1)
	add_child(glow)

func _perform_attack(direction: Vector2) -> void:
	"""Staff attack - projectile"""
	super._perform_attack(direction)

	# Spawn projectile
	var projectile = _create_projectile()
	projectile.global_position = owner_player.global_position
	projectile.direction = direction.normalized()
	projectile.damage = get_damage()
	projectile.owner_id = owner_player.network_id if owner_player else 0

	owner_player.get_parent().add_child(projectile)

	if AudioManager:
		AudioManager.play_sound("staff_cast")

func _create_projectile() -> Node2D:
	"""Create a simple projectile"""
	var projectile = Node2D.new()
	projectile.name = "StaffProjectile"

	var sprite = ColorRect.new()
	sprite.size = Vector2(15, 15)
	sprite.color = Color(0.8, 0.2, 0.8)
	sprite.position = -sprite.size / 2
	projectile.add_child(sprite)

	var collision = Area2D.new()
	var shape = CircleShape2D.new()
	shape.radius = 8
	var col_shape = CollisionShape2D.new()
	col_shape.shape = shape
	collision.add_child(col_shape)
	projectile.add_child(collision)

	# Add script
	projectile.set_script(preload("res://gameplay/projectiles/Projectile.gd"))

	return projectile
