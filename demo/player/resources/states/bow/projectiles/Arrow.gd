extends Area2D

var velocity: Vector2 = Vector2.ZERO
var lifetime: float = 3.0

func _ready() -> void:
	# Destroi após X segundos para não vazar memória
	get_tree().create_timer(lifetime).timeout.connect(queue_free)

	# Conecta sinal de colisão
	body_entered.connect(_on_body_entered)

func _physics_process(delta: float) -> void:
	position += velocity * delta

	# Opcional: Rotacionar na direção do movimento
	if velocity.length_squared() > 0:
		rotation = velocity.angle()

func _on_body_entered(body: Node2D) -> void:
	if body.is_in_group("Player"): return # Não acerta o player que atirou

	# Aqui poderia aplicar dano se o body tiver método 'take_damage'
	print("Arrow hit: ", body.name)
	queue_free()
