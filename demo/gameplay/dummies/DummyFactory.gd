extends Node
## Factory for creating dummies with different configurations

class_name DummyFactory

static func create_dummy(difficulty: String = "basic", position: Vector2 = Vector2.ZERO):
	"""Create a dummy instance"""
	var dummy_script = preload("res://gameplay/dummies/Dummy.gd")
	var dummy = CharacterBody2D.new()
	dummy.set_script(dummy_script)
	dummy.difficulty = difficulty
	dummy.global_position = position

	# Set stats based on difficulty
	match difficulty:
		"basic":
			dummy.max_health = 50.0
			dummy.experience_reward = 25.0
		"advanced":
			dummy.max_health = 100.0
			dummy.experience_reward = 50.0
		"elite":
			dummy.max_health = 200.0
			dummy.experience_reward = 100.0

	return dummy

static func create_combat_dummy(position: Vector2 = Vector2.ZERO) -> Dummy:
	"""Create a basic combat training dummy"""
	return create_dummy("basic", position)

static func create_training_dummy(position: Vector2 = Vector2.ZERO) -> Dummy:
	"""Create a passive training dummy (no AI)"""
	var dummy = create_dummy("basic", position)
	dummy.ai_behavior = "passive"
	return dummy

static func create_advanced_dummy(position: Vector2 = Vector2.ZERO) -> Dummy:
	"""Create an advanced dummy with AI"""
	return create_dummy("advanced", position)

static func create_elite_dummy(position: Vector2 = Vector2.ZERO) -> Dummy:
	"""Create an elite dummy - very challenging"""
	return create_dummy("elite", position)
