extends Node
## Central audio manager for gameplay sounds
## Handles SFX playback with optional procedural generation and variation

# Sound effect dictionary - lazy loaded from file system
var sound_effects: Dictionary = {}

# Audio players pool
var audio_players: Array[AudioStreamPlayer] = []
var max_simultaneous_sounds: int = 16

func _ready() -> void:
	print("[AudioManager] Initializing...")
	# Create audio stream players pool
	for i in range(max_simultaneous_sounds):
		var player = AudioStreamPlayer.new()
		player.bus = &"Master"
		add_child(player)
		audio_players.append(player)
	print("[AudioManager] Initialized with %d audio players" % max_simultaneous_sounds)

## Play a sound effect by name with optional pitch variation
func play_sound(sound_name: String, pitch_variation: float = 0.0) -> void:
	"""Play a sound effect with optional pitch variation"""
	if sound_name not in sound_effects:
		print("[AudioManager] WARNING: Sound '%s' not found" % sound_name)
		return

	var audio_stream = sound_effects[sound_name]
	if not audio_stream:
		# Try to generate procedurally or use fallback
		_play_procedural_fallback(sound_name)
		return

	_play_audio_stream(audio_stream, pitch_variation)

## Play audio stream with variation
func _play_audio_stream(stream: AudioStream, pitch_variation: float = 0.0) -> void:
	"""Find free audio player and play stream"""
	var player = _get_free_audio_player()
	if not player:
		print("[AudioManager] WARNING: No free audio players available")
		return

	player.stream = stream
	if pitch_variation > 0:
		player.pitch_scale = randf_range(1.0 - pitch_variation, 1.0 + pitch_variation)
	else:
		player.pitch_scale = 1.0
	player.play()

## Get free audio player from pool
func _get_free_audio_player() -> AudioStreamPlayer:
	"""Get the next free audio player"""
	for player in audio_players:
		if not player.playing:
			return player
	# If all busy, return the oldest one
	return audio_players[0] if audio_players.size() > 0 else null

## Procedural fallback for missing sound files
func _play_procedural_fallback(sound_name: String) -> void:
	"""Generate sound procedurally if audio file doesn't exist"""
	print("[AudioManager] Using procedural fallback for '%s'" % sound_name)
	# In a real implementation, this would use ProceduralSound or similar
	# For now, just log that we tried

## Stop all sounds
func stop_all_sounds() -> void:
	"""Stop all playing sounds"""
	for player in audio_players:
		player.stop()

## Set master volume
func set_master_volume(volume_db: float) -> void:
	"""Set master audio bus volume"""
	var master_bus = AudioServer.get_bus_index(&"Master")
	if master_bus >= 0:
		AudioServer.set_bus_volume_db(master_bus, volume_db)

## Get master volume
func get_master_volume() -> float:
	"""Get current master volume"""
	var master_bus = AudioServer.get_bus_index(&"Master")
	if master_bus >= 0:
		return AudioServer.get_bus_volume_db(master_bus)
	return 0.0

## Play sound with position (for 3D spatial audio)
func play_sound_at_position(sound_name: String, position: Vector3 = Vector3.ZERO, pitch_variation: float = 0.0) -> void:
	"""Play sound at specific position (for 3D audio)"""
	# For 2D games, we can ignore position for now
	play_sound(sound_name, pitch_variation)

## Play sound with variation (pitch and volume)
func play_sound_with_variation(sound_name: String, pitch_min: float = 0.9, pitch_max: float = 1.1, volume_min: float = 0.8, volume_max: float = 1.0) -> void:
	"""Play sound with pitch and volume variation"""
	if sound_name not in sound_effects:
		print("[AudioManager] WARNING: Sound '%s' not found" % sound_name)
		return

	var audio_stream = sound_effects[sound_name]
	if not audio_stream:
		_play_procedural_fallback(sound_name)
		return

	var player = _get_free_audio_player()
	if not player:
		print("[AudioManager] WARNING: No free audio players available")
		return

	player.stream = audio_stream
	player.pitch_scale = randf_range(pitch_min, pitch_max)
	player.volume_db = linear_to_db(randf_range(volume_min, volume_max))
	player.play()
