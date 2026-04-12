extends Area2D
## Detects hits and applies damage to overlapping targets
## Prevents hitting same target multiple times per attack

class_name HitDetector

@export var damage: float = 10.0
@export var knockback: float = 200.0
@export var hit_effects: Array[String] = []
@export var cooldown_time: float = 0.1

var owner_player: Node
var hit_targets: Dictionary = {}  # target: bool (hit this attack or not)
var active: bool = false

signal hit_target(target: Node, damage: float)
signal activated
signal deactivated

func _ready() -> void:
	area_entered.connect(_on_area_entered)

func activate() -> void:
	"""Activate hit detection"""
	if active:
		return
	active = true
	hit_targets.clear()
	activated.emit()

func deactivate() -> void:
	"""Deactivate hit detection"""
	active = false
	hit_targets.clear()
	deactivated.emit()

func _on_area_entered(area: Area2D) -> void:
	"""Handle area collision"""
	if not active or not owner_player:
		return

	var target = area.get_parent() if area.get_parent() else area

	# Prevent hitting same target twice
	if target in hit_targets:
		return

	# Don't hit self
	if target == owner_player:
		return

	hit_targets[target] = true

	# Apply damage
	if target.has_method("take_damage"):
		target.take_damage(damage)

	# Apply knockback
	if target.has_method("apply_knockback"):
		target.apply_knockback(global_position, knockback)

	# Apply effects
	for effect in hit_effects:
		if target.has_method("apply_effect"):
			target.apply_effect(effect)

	# Play sound
	if AudioManager:
		AudioManager.play_sound("hit_impact")

	# Emit signal
	hit_target.emit(target, damage)

	# Schedule cooldown
	await get_tree().create_timer(cooldown_time).timeout
	hit_targets.erase(target)

func get_hit_count() -> int:
	"""Get number of targets hit this activation"""
	return hit_targets.size()

func reset() -> void:
	"""Reset hit detection state"""
	hit_targets.clear()
	deactivate()
