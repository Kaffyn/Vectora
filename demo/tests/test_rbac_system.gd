extends Node
## Test suite for RBAC and AdminCommand system

class_name TestRBACSystem

# Test results
var tests_passed: int = 0
var tests_failed: int = 0
var test_results: Array = []

func _ready() -> void:
	var separator = "============================================================"
	print("\n" + separator)
	print("🧪 RBAC & AdminCommand System Tests")
	print(separator)

	# Run all tests
	test_rbac_initialization()
	test_role_permissions()
	test_player_registration()
	test_ban_system()
	test_admin_commands()

	# Print summary
	_print_summary()

func test_rbac_initialization() -> void:
	"""Test: RBAC system initializes correctly"""
	print("\n[TEST] RBAC Initialization")

	if RBAC == null:
		_fail("RBAC autoload not found")
		return

	# Check that server is registered as admin on peer_id 1
	if multiplayer.is_server():
		var role = RBAC.get_player_role(1)
		if role == "admin":
			_pass("Server registered as admin")
		else:
			_fail("Server not registered as admin (got: %s)" % role)
	else:
		_pass("Client instance (RBAC will be populated by server)")

func test_role_permissions() -> void:
	"""Test: Role-based permissions work correctly"""
	print("\n[TEST] Role Permissions")

	if RBAC == null:
		_fail("RBAC not available")
		return

	# Test admin permissions
	RBAC.peer_roles[999] = "admin"
	if RBAC.can_execute(999, "kick_player"):
		_pass("Admin can execute kick_player")
	else:
		_fail("Admin cannot execute kick_player")

	if RBAC.can_execute(999, "ban_player"):
		_pass("Admin can execute ban_player")
	else:
		_fail("Admin cannot execute ban_player")

	# Test user permissions
	RBAC.peer_roles[888] = "user"
	if not RBAC.can_execute(888, "kick_player"):
		_pass("User cannot execute kick_player")
	else:
		_fail("User can execute kick_player (should not)")

	if RBAC.can_execute(888, "chat"):
		_pass("User can execute chat")
	else:
		_fail("User cannot execute chat")

	# Cleanup
	RBAC.peer_roles.erase(999)
	RBAC.peer_roles.erase(888)

func test_player_registration() -> void:
	"""Test: Player registration and tracking"""
	print("\n[TEST] Player Registration")

	if RBAC == null:
		_fail("RBAC not available")
		return

	# Register a test player
	RBAC.register_peer(100, "TestPlayer")

	var role = RBAC.get_player_role(100)
	if role == "user":
		_pass("New player registered with user role")
	else:
		_fail("New player registered with wrong role: %s" % role)

	var data = RBAC.get_peer_data(100)
	if data.get("name") == "TestPlayer":
		_pass("Player name stored correctly")
	else:
		_fail("Player name not stored correctly")

	# Test unregistration
	RBAC.unregister_peer(100)
	var data_after = RBAC.get_peer_data(100)
	if data_after.is_empty():
		_pass("Player unregistered successfully")
	else:
		_fail("Player not unregistered")

func test_ban_system() -> void:
	"""Test: Ban system with expiration"""
	print("\n[TEST] Ban System")

	if RBAC == null:
		_fail("RBAC not available")
		return

	# Ban a player
	var player_id = 200
	var success = RBAC.ban_peer(player_id, 1.0, "Test ban")  # 1 hour ban

	if success:
		_pass("Player banned successfully")
	else:
		_fail("Failed to ban player")

	# Check if banned
	if RBAC.is_banned(player_id):
		_pass("Banned player detected")
	else:
		_fail("Banned player not detected")

	# Check ban info
	var ban_info = RBAC.get_ban_info(player_id)
	if ban_info.get("banned", false):
		_pass("Ban info retrieved correctly")
	else:
		_fail("Ban info not found")

	# Unban player
	var unban_success = RBAC.unban_peer(player_id)
	if unban_success:
		_pass("Player unbanned successfully")
	else:
		_fail("Failed to unban player")

	if not RBAC.is_banned(player_id):
		_pass("Banned player no longer banned")
	else:
		_fail("Banned player still banned after unban")

func test_admin_commands() -> void:
	"""Test: AdminCommand system"""
	print("\n[TEST] Admin Commands")

	if AdminCommand == null:
		_fail("AdminCommand autoload not found")
		return

	# Check command registration
	var handlers_count = AdminCommand._command_handlers.size()
	if handlers_count >= 6:  # At least 6 commands: kick, ban, shutdown, broadcast, list, mute, unmute, status
		_pass("Command handlers registered (%d commands)" % handlers_count)
	else:
		_fail("Not enough command handlers registered: %d" % handlers_count)

	# Test list_players command (non-destructive)
	var success = AdminCommand._cmd_list_players(1, [])
	if success:
		_pass("list_players command executed")
	else:
		_fail("list_players command failed")

	# Test get_server_status command
	success = AdminCommand._cmd_get_server_status(1, [])
	if success:
		_pass("get_server_status command executed")
	else:
		_fail("get_server_status command failed")

# Helper functions

func _pass(message: String) -> void:
	var test_msg = "✓ PASS: %s" % message
	print("  " + test_msg)
	test_results.append(test_msg)
	tests_passed += 1

func _fail(message: String) -> void:
	var test_msg = "✗ FAIL: %s" % message
	print("  " + test_msg)
	test_results.append(test_msg)
	tests_failed += 1

func _print_summary() -> void:
	print("\n" + "=".repeat(60))
	print("📊 Test Summary")
	print("=".repeat(60))
	print("Total: %d | Passed: %d | Failed: %d" % [tests_passed + tests_failed, tests_passed, tests_failed])

	if tests_failed > 0:
		print("\n❌ Some tests failed. Review above for details.")
	else:
		print("\n✅ All tests passed!")

	print("=".repeat(60) + "\n")
