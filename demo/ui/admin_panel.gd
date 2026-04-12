extends PanelContainer
## Admin Panel para Host Multiplayer
## Permite gerenciar jogadores e servidor quando em modo LAN

@onready var players_list = $MarginContainer/VBoxContainer/PlayersSection/PlayersMargin/PlayersVBox/PlayersList
@onready var broadcast_input = $MarginContainer/VBoxContainer/BroadcastSection/BroadcastMargin/BroadcastVBox/BroadcastInput
@onready var broadcast_button = $MarginContainer/VBoxContainer/BroadcastSection/BroadcastMargin/BroadcastVBox/BroadcastButton
@onready var close_lan_button = $MarginContainer/VBoxContainer/ServerSection/ServerMargin/ServerVBox/ServerHBox/CloseToLANButton
@onready var shutdown_button = $MarginContainer/VBoxContainer/ServerSection/ServerMargin/ServerVBox/ServerHBox/ShutdownButton

# Referência para guardar player items para controle
var player_items: Dictionary = {}  # {peer_id: HBoxContainer node}

signal close_to_lan_requested
signal shutdown_requested

func _ready() -> void:
	# Esconde por padrão
	visible = false

	# Conecta sinais
	broadcast_button.pressed.connect(_on_broadcast_pressed)
	close_lan_button.pressed.connect(_on_close_to_lan_pressed)
	shutdown_button.pressed.connect(_on_shutdown_pressed)

	# Conecta a atualizações de jogadores
	if multiplayer.is_server():
		multiplayer.peer_connected.connect(_on_peer_connected)
		multiplayer.peer_disconnected.connect(_on_peer_disconnected)

	# Carrega jogadores iniciais
	_refresh_players_list()

## Mostra o painel de admin
func show_admin_panel() -> void:
	if not multiplayer.is_server():
		print("[AdminPanel] Only server (host) can open admin panel")
		return

	visible = true
	_refresh_players_list()
	print("[AdminPanel] Admin panel opened")

## Esconde o painel de admin
func hide_admin_panel() -> void:
	visible = false
	print("[AdminPanel] Admin panel closed")

## Atualiza lista de jogadores
func _refresh_players_list() -> void:
	# Limpa lista anterior
	for child in players_list.get_children():
		child.queue_free()
	player_items.clear()

	var peers = RBAC.get_all_peers()

	if peers.is_empty():
		var label = Label.new()
		label.text = "Nenhum jogador conectado"
		label.add_theme_color_override("font_color", Color(0.7, 0.7, 0.7, 1))
		players_list.add_child(label)
		return

	for peer_info in peers:
		var peer_id = peer_info["peer_id"]
		var player_name = peer_info["name"]
		var role = peer_info["role"]

		# Cria item do jogador
		var item = _create_player_item(peer_id, player_name, role)
		players_list.add_child(item)
		player_items[peer_id] = item

## Cria um item de jogador com botões de ação
func _create_player_item(peer_id: int, player_name: String, role: String) -> HBoxContainer:
	var container = HBoxContainer.new()
	container.add_theme_constant_override("separation", 5)

	# Label com informações do jogador
	var info_label = Label.new()
	info_label.text = "[%d] %s (%s)" % [peer_id, player_name, role]
	info_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	container.add_child(info_label)

	# Botão Kick (exceto para servidor)
	if peer_id != 1:
		var kick_button = Button.new()
		kick_button.text = "Kick"
		kick_button.custom_minimum_size = Vector2(50, 25)
		kick_button.pressed.connect(func(): _on_kick_player(peer_id, player_name))
		container.add_child(kick_button)

		# Botão Ban (exceto para servidor)
		var ban_button = Button.new()
		ban_button.text = "Ban"
		ban_button.custom_minimum_size = Vector2(50, 25)
		ban_button.add_theme_color_override("font_color", Color(1, 0.3, 0.3, 1))
		ban_button.pressed.connect(func(): _on_ban_player(peer_id, player_name))
		container.add_child(ban_button)

	return container

## Chamado quando um peer se conecta
func _on_peer_connected(peer_id: int) -> void:
	_refresh_players_list()

## Chamado quando um peer se desconecta
func _on_peer_disconnected(peer_id: int) -> void:
	_refresh_players_list()

## Handler: Broadcast de mensagem
func _on_broadcast_pressed() -> void:
	var message = broadcast_input.text.strip_edges()

	if message.is_empty():
		print("[AdminPanel] Empty message")
		return

	# Executa comando de broadcast via AdminCommand
	AdminCommand.execute_command(1, "broadcast_message", [message])

	# Limpa input
	broadcast_input.text = ""
	broadcast_input.grab_focus()

	print("[AdminPanel] Broadcast sent: %s" % message)

## Handler: Fechar LAN
func _on_close_to_lan_pressed() -> void:
	print("[AdminPanel] Close to LAN requested")
	close_to_lan_requested.emit()

	# Chama MultiplayerGameManager para fechar o servidor
	var game_mgr = get_node_or_null("/root/MultiplayerGameManager")
	if game_mgr and game_mgr.has_method("close_to_lan"):
		game_mgr.close_to_lan()

	hide_admin_panel()

## Handler: Desligar servidor
func _on_shutdown_pressed() -> void:
	print("[AdminPanel] Shutdown requested")
	shutdown_requested.emit()

	# Executa comando de shutdown
	AdminCommand.execute_command(1, "shutdown_server", ["Host closed server"])

## Handler: Kick de jogador
func _on_kick_player(peer_id: int, player_name: String) -> void:
	var reason = "Kicked by host"
	print("[AdminPanel] Kicking player %d (%s): %s" % [peer_id, player_name, reason])

	# Executa comando de kick
	AdminCommand.execute_command(1, "kick_player", [peer_id, reason])

	# Aguarda um frame e atualiza lista
	await get_tree().process_frame
	_refresh_players_list()

## Handler: Ban de jogador
func _on_ban_player(peer_id: int, player_name: String) -> void:
	var reason = "Banned by host"
	print("[AdminPanel] Banning player %d (%s): %s" % [peer_id, player_name, reason])

	# Executa comando de ban (24 horas por padrão)
	AdminCommand.execute_command(1, "ban_player", [peer_id, 24.0, reason])

	# Aguarda um frame e atualiza lista
	await get_tree().process_frame
	_refresh_players_list()

## Atualiza informações de status
func refresh() -> void:
	_refresh_players_list()

## Debug: Imprime estado
func debug_print_state() -> void:
	print("\n=== AdminPanel State ===")
	print("Visible: %s" % visible)
	print("Player items: %d" % player_items.size())
	print("========================\n")
