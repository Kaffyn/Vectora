extends Camera2D

@export_category("Settings")
@export var fast_smoothing: float = 10.0
@export var slow_smoothing: float = 5.0

var _target: Node2D

func _ready() -> void:
	# Tenta achar o spawn point inicialmente
	_target = get_tree().get_first_node_in_group("PlayerSpawn")

func _physics_process(_delta: float) -> void:
	# 1. Tenta encontrar o Player
	var player = get_tree().get_first_node_in_group("Player")

	if player and is_instance_valid(player):
		_target = player

		# Acelera a câmera na queda (se o target for o player)
		if "velocity" in player:
			if player.velocity.y > 0: # Caindo
				position_smoothing_speed = fast_smoothing
			else:
				position_smoothing_speed = slow_smoothing
	else:
		# Se não tem player, foca no Spawn
		var spawn = get_tree().get_first_node_in_group("PlayerSpawn")
		if spawn:
			_target = spawn
			position_smoothing_speed = slow_smoothing # Suave ao voltar pro spawn

	# 2. Segue o Target
	if _target and is_instance_valid(_target):
		global_position = _target.global_position
