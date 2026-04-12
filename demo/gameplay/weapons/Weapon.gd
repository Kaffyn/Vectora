extends Node2D
## Base weapon class
## Extended by Sword, Staff, Axe, Dagger for specific behavior

class_name Weapon

@export var weapon_name: String = "Weapon"
@export var weapon_type: String = "sword"  # sword, staff, axe, dagger
@export var damage: float = 10.0
@export var attack_speed: float = 1.0  # attacks per second
@export var range_: float = 50.0
@export var knockback: float = 200.0
@export var attack_duration: float = 0.3

var owner_player: Node = null
var cooldown: float = 0.0
var is_attacking: bool = false
var attack_timer: float = 0.0
var hit_detector: HitDetector = null

signal weapon_attacked(weapon: Weapon)
signal weapon_hit(target: Node, damage: float)
signal weapon_ready

func _ready() -> void:
	_create_visual()
	_setup_hit_detector()

func _create_visual() -> void:
	"""Create visual representation - override in subclasses"""
	# Default: simple rectangle
	var rect = ColorRect.new()
	rect.size = Vector2(20, 40)
	rect.color = Color(0.8, 0.8, 0.8)
	rect.position = -rect.size / 2
	add_child(rect)

func _setup_hit_detector() -> void:
	"""Setup hit detection area"""
	hit_detector = HitDetector.new()
	hit_detector.damage = damage
	hit_detector.knockback = knockback
	hit_detector.owner_player = owner_player
	add_child(hit_detector)

	# Create collision shape
	var collision = CollisionShape2D.new()
	var shape = CapsuleShape2D.new()
	shape.height = range_
	shape.radius = 10
	collision.shape = shape
	hit_detector.add_child(collision)

	hit_detector.hit_target.connect(func(target, dmg): weapon_hit.emit(target, dmg))

func _process(delta: float) -> void:
	if cooldown > 0:
		cooldown -= delta

	if is_attacking:
		attack_timer -= delta
		if attack_timer <= 0:
			is_attacking = false

	# Update hit detector position
	if hit_detector:
		hit_detector.owner_player = owner_player

func can_attack() -> bool:
	"""Check if weapon can attack right now"""
	return cooldown <= 0 and not is_attacking

func attack(direction: Vector2) -> bool:
	"""Perform an attack in the given direction"""
	if not can_attack():
		return false

	is_attacking = true
	attack_timer = attack_duration
	cooldown = 1.0 / attack_speed

	# Activate hit detector
	if hit_detector:
		hit_detector.activate()

	_perform_attack(direction)

	weapon_attacked.emit(self)

	# Sound
	if AudioManager:
		AudioManager.play_sound("weapon_slash")

	return true

func _perform_attack(direction: Vector2) -> void:
	"""Perform attack - override in subclasses"""
	# Position hit detector in front of player
	if hit_detector and owner_player:
		hit_detector.global_position = owner_player.global_position + direction.normalized() * (range_ / 2)

func get_damage() -> float:
	"""Get modified damage based on attacker's stats"""
	var multiplier = 1.0

	if owner_player:
		# Try to get attack power from ASComponent
		var asc = owner_player.get_node_or_null("ASComponent") as ASComponent
		if asc:
			var power = asc.get_attribute_value_by_tag(&"attack_power")
			if power > 0:
				multiplier = power / 10.0

	return damage * multiplier

func set_owner_player(player: Node) -> void:
	"""Set the player that owns this weapon"""
	owner_player = player
	if hit_detector:
		hit_detector.owner_player = player

func get_weapon_info() -> Dictionary:
	"""Get weapon information"""
	return {
		"name": weapon_name,
		"type": weapon_type,
		"damage": get_damage(),
		"attack_speed": attack_speed,
		"range": range_,
		"cooldown": cooldown,
		"ready": can_attack(),
	}
