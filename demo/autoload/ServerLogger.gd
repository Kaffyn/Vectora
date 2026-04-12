extends Node
## Centralized logging system for all server events

enum LogLevel { DEBUG, INFO, WARN, ERROR }

const ENABLE_FILE_LOGGING = false  # Ativar para persistência
const LOG_FILE_PATH = "user://server.log"

var _log_buffer: PackedStringArray = []
var _start_time_ms: int = 0

func _ready() -> void:
	_start_time_ms = Time.get_ticks_msec()
	log_message(LogLevel.INFO, "ServerLogger", "Initialized")

## Log com nível automático
func log_message(level: LogLevel, source: String, message: String) -> void:
	var timestamp = _get_timestamp()
	var level_str = _level_to_string(level)
	var formatted = "[%s] [%s] %s: %s" % [timestamp, level_str, source, message]

	_log_buffer.append(formatted)
	print(formatted)

	if ENABLE_FILE_LOGGING:
		_write_to_file(formatted)

## Log de conexão de peer
func log_peer_connected(peer_id: int, player_name: String) -> void:
	log_message(LogLevel.INFO, "CONNECT", "Peer %d connected: %s" % [peer_id, player_name])

## Log de desconexão de peer
func log_peer_disconnected(peer_id: int, player_name: String) -> void:
	log_message(LogLevel.INFO, "DISCONNECT", "Peer %d disconnected: %s" % [peer_id, player_name])

## Log de comando executado
func log_command(executor_id: int, command: String, success: bool) -> void:
	var status = "SUCCESS" if success else "FAILED"
	log_message(LogLevel.INFO, "COMMAND", "Peer %d executed '%s': %s" % [executor_id, command, status])

## Log de chat
func log_chat(peer_id: int, player_name: String, message: String) -> void:
	log_message(LogLevel.INFO, "CHAT", "[%s] %s: %s" % [peer_id, player_name, message])

## Log de erro
func log_error(source: String, error_msg: String) -> void:
	log_message(LogLevel.ERROR, source, error_msg)

## Log de status periódico
func log_server_status(connected_players: int, max_players: int, uptime_sec: float) -> void:
	var status_msg = "Status: %d/%d players | Uptime: %.1fs" % [connected_players, max_players, uptime_sec]
	log_message(LogLevel.INFO, "SERVER", status_msg)

## Retorna logs recentes
func get_recent_logs(count: int = 50) -> PackedStringArray:
	var start = maxi(0, _log_buffer.size() - count)
	return _log_buffer.slice(start)

## Limpa buffer de logs
func clear_buffer() -> void:
	_log_buffer.clear()

# ============= HELPERS =============

func _get_timestamp() -> String:
	return Time.get_time_string_from_system()

func _level_to_string(level: LogLevel) -> String:
	match level:
		LogLevel.DEBUG:
			return "DEBUG"
		LogLevel.INFO:
			return "INFO"
		LogLevel.WARN:
			return "WARN"
		LogLevel.ERROR:
			return "ERROR"
		_:
			return "UNKNOWN"

func _write_to_file(message: String) -> void:
	# Append to log file
	var file = FileAccess.open(LOG_FILE_PATH, FileAccess.READ_WRITE)
	if file:
		file.seek_end()
		file.store_line(message)
