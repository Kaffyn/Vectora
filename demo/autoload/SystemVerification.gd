extends Node
## System Verification - Checks all critical systems are initialized
## Run this autoload to ensure everything is properly connected

var verification_results: Dictionary = {}

func _ready() -> void:
	print("\n" + "=".repeat(70))
	print("🔍 SYSTEM VERIFICATION - Ability System Demo")
	print("=".repeat(70) + "\n")

	_verify_autoloads()
	_verify_scenes()
	_verify_multiplayer()
	_verify_signals()

	_print_results()

func _verify_autoloads() -> void:
	"""Verify all required autoloads are present and initialized"""
	print("[VERIFY] Checking Autoloads...")

	var autoloads = {
		"GameData": GameData,
		"GameMachine": GameMachine,
		"MultiplayerGameManager": MultiplayerGameManager,
		"RBAC": RBAC,
		"AdminCommand": AdminCommand,
		"AdminTerminal": AdminTerminal,
		"ServerLogger": ServerLogger,
	}

	for name_str in autoloads.keys():
		var autoload = autoloads[name_str]
		if autoload != null:
			_pass("Autoload: %s" % name_str)
			verification_results[name_str] = "OK"
		else:
			_fail("Autoload: %s not found" % name_str)
			verification_results[name_str] = "MISSING"

func _verify_scenes() -> void:
	"""Verify all required scenes can be loaded"""
	print("\n[VERIFY] Checking Scene Assets...")

	var scenes = {
		"UserRegistration": "res://ui/user_registration.tscn",
		"MainMenu": "res://ui/main_menu.tscn",
		"ChatSystem": "res://ui/chat_system.tscn",
		"PauseMenu": "res://ui/pause_menu.tscn",
		"AdminPanel": "res://ui/admin_panel.tscn",
		"Level": "res://scenes/level.tscn",
	}

	for scene_name in scenes.keys():
		var scene_path = scenes[scene_name]
		var scene = ResourceLoader.load(scene_path)
		if scene != null:
			_pass("Scene: %s" % scene_name)
			verification_results[scene_name] = "OK"
		else:
			_fail("Scene: %s not found at %s" % [scene_name, scene_path])
			verification_results[scene_name] = "MISSING"

func _verify_multiplayer() -> void:
	"""Verify multiplayer infrastructure"""
	print("\n[VERIFY] Checking Multiplayer Infrastructure...")

	# Check if multiplayer is initialized
	if multiplayer != null:
		_pass("Multiplayer API available")
		verification_results["MultiplayerAPI"] = "OK"
	else:
		_fail("Multiplayer API not available")
		verification_results["MultiplayerAPI"] = "MISSING"

	# Check game data
	if GameData.player_name != "":
		_pass("GameData.player_name initialized")
		verification_results["GameDataPlayer"] = "OK"
	else:
		_pass("GameData.player_name ready for assignment")
		verification_results["GameDataPlayer"] = "READY"

	# Check RBAC
	if RBAC and RBAC.has_method("can_execute"):
		_pass("RBAC system initialized")
		verification_results["RBACSystem"] = "OK"
	else:
		_fail("RBAC system not initialized")
		verification_results["RBACSystem"] = "MISSING"

	# Check AdminCommand
	if AdminCommand and AdminCommand._command_handlers.size() > 0:
		_pass("AdminCommand handlers registered (%d)" % AdminCommand._command_handlers.size())
		verification_results["AdminCommandHandlers"] = "OK"
	else:
		_fail("AdminCommand handlers not registered")
		verification_results["AdminCommandHandlers"] = "MISSING"

	# Check ServerLogger
	if ServerLogger and ServerLogger.has_method("log_message"):
		_pass("ServerLogger system initialized")
		verification_results["ServerLogger"] = "OK"
	else:
		_fail("ServerLogger not initialized")
		verification_results["ServerLogger"] = "MISSING"

func _verify_signals() -> void:
	"""Verify critical signals are properly defined"""
	print("\n[VERIFY] Checking Signal Definitions...")

	# Check MultiplayerGameManager signals
	if MultiplayerGameManager.has_signal("player_spawned"):
		_pass("Signal: MultiplayerGameManager.player_spawned")
		verification_results["SignalPlayerSpawned"] = "OK"
	else:
		_fail("Signal: MultiplayerGameManager.player_spawned not found")
		verification_results["SignalPlayerSpawned"] = "MISSING"

	# Check AdminCommand signals
	if AdminCommand.has_signal("command_executed"):
		_pass("Signal: AdminCommand.command_executed")
		verification_results["SignalCommandExecuted"] = "OK"
	else:
		_fail("Signal: AdminCommand.command_executed not found")
		verification_results["SignalCommandExecuted"] = "MISSING"

	# Check admin_panel signals (when instantiated)
	# These will be checked in-game when admin panel loads

func _pass(message: String) -> void:
	print("  ✓ %s" % message)

func _fail(message: String) -> void:
	print("  ✗ %s" % message)

func _print_results() -> void:
	print("\n" + "=".repeat(70))
	print("📊 VERIFICATION SUMMARY")
	print("=".repeat(70))

	var ok_count = 0
	var fail_count = 0
	var ready_count = 0

	for key in verification_results.keys():
		var status = verification_results[key]
		match status:
			"OK":
				ok_count += 1
			"MISSING":
				fail_count += 1
			"READY":
				ready_count += 1

	print("Status: OK=%d | MISSING=%d | READY=%d" % [ok_count, fail_count, ready_count])

	if fail_count > 0:
		print("\n❌ VERIFICATION FAILED - Some systems are missing!")
		print("Please check the errors above and ensure all files are present.")
	else:
		print("\n✅ VERIFICATION PASSED - All systems are properly configured!")
		print("The demo is ready to run.")

	print("=".repeat(70) + "\n")

func print_verification_details() -> void:
	"""Print detailed verification information"""
	print("\nDetailed Verification Results:")
	for key in verification_results.keys():
		print("  %s: %s" % [key, verification_results[key]])
