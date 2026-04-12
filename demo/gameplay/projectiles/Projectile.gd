extends Node2D
## Projectile fired by ranged weapons (e.g., staff)
## Travels in a direction and applies damage on impact

class_name Projectile

@export var speed: float = 300.0
@export var lifetime: float = 10.0
@export var damage: float = 8.0
@export var knockback: float = 150.0

var direction: Vector2 = Vector2.RIGHT
var owner_id: int = 0
var hit_targets: Array = []

var time_alive: float = 0.0

func _ready() -> void:
	# Connect area detection
	var areas = get_children().filter(func(c): return c is Area2D)
	for area in areas:
		area.area_entered.connect(_on_area_entered)

func _process(delta: float) -> void:
	# Move projectile
	global_position += direction * speed * delta

	# Update lifetime
	time_alive += delta
	if time_alive >= lifetime:
		queue_free()

func _on_area_entered(area: Area2D) -> void:
	"""Handle collision with target"""
	var target = area.get_parent() if area.get_parent() else area

	# Don't hit owner
	if target.network_id == owner_id if target.has_meta("network_id") else false:
		return

	# Don't hit same target twice
	if target in hit_targets:
		return

	hit_targets.append(target)

	# Apply damage
	if target.has_method("take_damage"):
		target.take_damage(damage)

	# Apply knockback
	if target.has_method("apply_knockback"):
		target.apply_knockback(global_position, knockback)

	# Play sound
	if AudioManager:
		AudioManager.play_sound("hit_impact")

	# Destroy projectile
	queue_free()
