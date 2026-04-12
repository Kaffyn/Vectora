# Gameplay Implementation Plan

Guia passo-a-passo para implementar o sistema de gameplay completo.

---

## 📂 Estrutura de Arquivos

```
demo/
├── gameplay/
│   ├── core/
│   │   ├── GameplayManager.gd        # Coordenador central
│   │   ├── DamageCalculator.gd
│   │   ├── HitDetector.gd
│   │   └── StatusEffect.gd
│   │
│   ├── progression/
│   │   ├── ProgressionSystem.gd      # XP, Levels, Attributes
│   │   ├── SkillTree.gd
│   │   ├── SkillTreeNode.gd
│   │   ├── RewardManager.gd
│   │   └── resources/
│   │       └── skill_definitions.tres
│   │
│   ├── combat/
│   │   ├── CombatState.gd
│   │   ├── Ability.gd
│   │   ├── AbilityEffect.gd
│   │   └── resources/
│   │       ├── abilities.tres
│   │       └── effects.tres
│   │
│   ├── weapons/
│   │   ├── Weapon.gd               # Base class
│   │   ├── Sword.gd
│   │   ├── Staff.gd
│   │   ├── Axe.gd
│   │   ├── Dagger.gd
│   │   ├── WeaponManager.gd
│   │   ├── Sword.tscn
│   │   ├── Staff.tscn
│   │   ├── Axe.tscn
│   │   └── Dagger.tscn
│   │
│   ├── dummies/
│   │   ├── Dummy.gd                 # Base dummy
│   │   ├── DummyFactory.gd
│   │   ├── DummySpawner.gd
│   │   ├── Dummy.tscn
│   │   ├── TrainingDummy.gd
│   │   ├── AggressiveDummy.gd
│   │   └── EliteDummy.gd
│   │
│   ├── enemies/
│   │   ├── Enemy.gd                 # Base enemy
│   │   ├── EnemySpawner.gd
│   │   ├── EnemyAI.gd
│   │   ├── Enemy.tscn
│   │   ├── Zombie.gd
│   │   ├── Skeleton.gd
│   │   └── Boss.gd
│   │
│   ├── projectiles/
│   │   ├── Projectile.gd
│   │   ├── Projectile.tscn
│   │   ├── ProjectileManager.gd
│   │   └── Projectile types...
│   │
│   ├── level/
│   │   ├── Level.gd                 # Main level manager
│   │   ├── SpawnManager.gd
│   │   ├── ZoneManager.gd
│   │   ├── Level.tscn
│   │   ├── zones/
│   │   │   ├── CombatArena.tscn
│   │   │   ├── TrainingGround.tscn
│   │   │   ├── ChaosZone.tscn
│   │   │   ├── NatureZone.tscn
│   │   │   ├── CorruptedRealm.tscn
│   │   │   └── Sanctuary.tscn
│   │   └── hazards/
│   │       ├── Hazard.gd
│   │       ├── HazardSpawner.gd
│   │       └── TrapSystem.gd
│   │
│   ├── environment/
│   │   ├── PlatformGenerator.gd
│   │   ├── ObstacleGenerator.gd
│   │   └── EnvironmentalEffect.gd
│   │
│   ├── crafting/
│   │   ├── CraftingSystem.gd
│   │   ├── Recipe.gd
│   │   ├── Inventory.gd
│   │   └── resources/
│   │       └── recipes.tres
│   │
│   └── particle_effects/
│       ├── ParticleEmitter.gd
│       ├── ImpactEffect.gd
│       ├── HealEffect.gd
│       └── StatusEffectVisuals.gd
│
├── audio/
│   ├── ProceduralMusic.gd          # Música procedural
│   ├── ProceduralSound.gd          # Sons procedurais
│   ├── AudioCues.gd                # AudioCue resources
│   ├── SoundEffectLibrary.gd
│   └── resources/
│       ├── music_settings.tres
│       ├── sound_effects/
│       └── audio_cues/
│
├── ui/
│   ├── GameHUD.gd                  # Main HUD
│   ├── SkillTreeUI.gd              # Skill tree panel
│   ├── CombatFeedback.gd           # Damage numbers, etc
│   ├── StatusDisplay.gd            # HP bars, etc
│   ├── DialogSystem.gd             # NPC dialog
│   ├── menus/
│   │   ├── MainGameMenu.tscn
│   │   ├── PauseMenu.gd
│   │   ├── SkillTreeMenu.tscn
│   │   └── CraftingMenu.tscn
│   └── hud/
│       ├── HealthBar.tscn
│       ├── ManaBar.tscn
│       ├── XPBar.tscn
│       └── SkillDisplay.tscn
│
├── player/
│   ├── Player.gd                   # MODIFICAR EXISTENTE
│   ├── PlayerController.gd         # Lógica de input
│   ├── PlayerState.gd              # Estado do player
│   ├── PlayerAbilities.gd          # Gerenciador de abilities
│   └── ASComponentPlayer.gd        # Integração com AS
│
└── resources/
    ├── abilities/
    │   ├── basic_attack.tres
    │   ├── power_slash.tres
    │   ├── dash.tres
    │   └── ... (todas as abilities)
    │
    ├── effects/
    │   ├── stun.tres
    │   ├── slow.tres
    │   ├── burn.tres
    │   └── ... (todos os efeitos)
    │
    ├── progression/
    │   ├── skills.tres
    │   ├── experience_table.tres
    │   └── attribute_growth.tres
    │
    └── enemies/
        ├── zombie.tres
        ├── skeleton.tres
        ├── boss_template.tres
        └── ...
```

---

## 🔧 Implementação Sequencial

### FASE 0: Setup (3 horas)

#### 0.1 Criar pastas e arquivos base

```bash
# Criar estrutura
mkdir -p demo/gameplay/{core,progression,combat,weapons,dummies,enemies,projectiles,level,environment,crafting,particle_effects}
mkdir -p demo/audio/resources/{sound_effects,audio_cues}
mkdir -p demo/ui/{menus,hud}
mkdir -p demo/resources/{abilities,effects,progression,enemies}
```

#### 0.2 Criar GameplayManager central

```gdscript
# demo/gameplay/core/GameplayManager.gd
extends Node
class_name GameplayManager

# Core systems references
var progression_system: Node
var damage_calculator: Node
var weapon_manager: Node
var skill_tree: Node
var crafting_system: Node
var audio_manager: Node

var level: Node

func _ready():
	_init_systems()
	level = get_parent()

func _init_systems():
	progression_system = ProgressionSystem.new()
	damage_calculator = DamageCalculator.new()
	weapon_manager = WeaponManager.new()
	skill_tree = SkillTree.new()
	crafting_system = CraftingSystem.new()
	audio_manager = AudioManager

func register_player(player_id: int, player_node: Node):
	progression_system.register_player(player_id)
	player_node.gameplay_manager = self

func unregister_player(player_id: int):
	progression_system.unregister_player(player_id)

func apply_damage(attacker: Node, defender: Node, ability: String) -> float:
	var result = damage_calculator.calculate_damage(attacker, defender, ability)
	if audio_manager:
		audio_manager.play_sound("hit_" + ability)
	return result.final_damage
```

---

### FASE 1: Core Systems (8 horas)

#### 1.1 Implement DamageCalculator

```gdscript
# demo/gameplay/core/DamageCalculator.gd
extends Node
class_name DamageCalculator

class DamageResult:
	var base_damage: float
	var multiplier: float
	var final_damage: float
	var critical: bool
	var hit_type: String  # "normal", "critical", "miss"

var ability_multipliers: Dictionary = {
	"basic_attack": 1.0,
	"power_slash": 1.5,
	"super_strike": 2.5,
	"dash_attack": 1.8,
	"whirlwind": 0.8,
	"annihilate": 4.0,
}

var critical_chance: float = 0.1
var critical_multiplier: float = 2.0

func calculate_damage(attacker: Node, defender: Node, ability: String) -> DamageResult:
	var result = DamageResult.new()

	# Get attacker power
	var attacker_power = _get_attacker_power(attacker)
	result.base_damage = attacker_power

	# Get ability multiplier
	var ability_mult = ability_multipliers.get(ability, 1.0)
	result.multiplier = ability_mult

	# Critical hit check
	result.critical = randf() < critical_chance
	if result.critical:
		result.multiplier *= critical_multiplier

	# Final damage
	result.final_damage = result.base_damage * result.multiplier

	# Apply defense (if applicable)
	var defender_defense = _get_defender_defense(defender)
	result.final_damage = max(1.0, result.final_damage - defender_defense)

	result.hit_type = "critical" if result.critical else "normal"

	return result

func _get_attacker_power(attacker: Node) -> float:
	if attacker.has_node("ASComponent"):
		return attacker.get_node("ASComponent").get_attribute("attack_power", 10.0)
	return 10.0

func _get_defender_defense(defender: Node) -> float:
	if defender.has_node("ASComponent"):
		return defender.get_node("ASComponent").get_attribute("defense", 0.0)
	return 0.0
```

#### 1.2 Implement HitDetector

```gdscript
# demo/gameplay/core/HitDetector.gd
extends Area2D
class_name HitDetector

@export var damage: float = 10.0
@export var knockback: float = 200.0
@export var hit_effects: Array[String] = ["stun"]

var owner_player: Node
var hit_targets: Dictionary = {}

func _ready():
	area_entered.connect(_on_area_entered)

func _on_area_entered(area: Area2D):
	if not owner_player:
		return

	var target = area.get_parent() if area.get_parent() else area

	# Prevent multiple hits same frame
	if target in hit_targets:
		return

	hit_targets[target] = true

	# Apply damage
	if target.has_method("take_damage"):
		target.take_damage(damage)

	# Apply knockback
	if target.has_method("apply_knockback"):
		target.apply_knockback(global_position, knockback)

	# Play sound
	if AudioManager:
		AudioManager.play_sound("hit_impact")

	# Remove hit after cooldown
	await get_tree().create_timer(0.1).timeout
	hit_targets.erase(target)
```

#### 1.3 Implement ProgressionSystem

```gdscript
# demo/gameplay/progression/ProgressionSystem.gd
extends Node
class_name ProgressionSystem

class PlayerProgress:
	var level: int = 1
	var experience: float = 0.0
	var experience_to_next: float = 100.0
	var skill_points: int = 0
	var unlocked_skills: Array[String] = []
	var attributes: Dictionary = {
		"max_health": 100.0,
		"max_mana": 50.0,
		"attack_power": 10.0,
		"defense": 5.0,
		"movement_speed": 200.0,
	}

var player_progress: Dictionary = {}

signal level_up(player_id: int, new_level: int)
signal skill_point_gained(player_id: int)
signal xp_gained(player_id: int, amount: float)

func register_player(player_id: int):
	player_progress[player_id] = PlayerProgress.new()

func add_experience(player_id: int, amount: float):
	if player_id not in player_progress:
		register_player(player_id)

	var progress = player_progress[player_id]
	progress.experience += amount
	xp_gained.emit(player_id, amount)

	if progress.experience >= progress.experience_to_next:
		_level_up(player_id)

func _level_up(player_id: int):
	var progress = player_progress[player_id]
	progress.level += 1
	progress.experience = 0.0
	progress.experience_to_next = 100.0 * (progress.level * 1.2)
	progress.skill_points += 1

	# Increase stats
	for stat in progress.attributes.keys():
		progress.attributes[stat] *= 1.1

	# Notify
	AudioManager.play_sound("level_up")
	level_up.emit(player_id, progress.level)

func get_progress(player_id: int) -> PlayerProgress:
	return player_progress.get(player_id, PlayerProgress.new())

func unlock_skill(player_id: int, skill_id: String) -> bool:
	var progress = get_progress(player_id)

	if progress.skill_points <= 0:
		return false

	progress.unlocked_skills.append(skill_id)
	progress.skill_points -= 1
	skill_point_gained.emit(player_id)

	return true
```

---

### FASE 2: Weapons (6 horas)

#### 2.1 Base Weapon Class

```gdscript
# demo/gameplay/weapons/Weapon.gd
extends Node2D
class_name Weapon

@export var weapon_name: String = "Sword"
@export var damage: float = 10.0
@export var attack_speed: float = 1.0
@export var range_: float = 50.0
@export var knockback: float = 200.0
@export var attack_duration: float = 0.3

var owner_player: Node
var cooldown: float = 0.0
var is_attacking: bool = false
var attack_timer: float = 0.0

signal weapon_attacked
signal weapon_hit(target: Node)

func _ready():
	_create_visual()

func _create_visual():
	# Override in subclasses
	pass

func _process(delta):
	if cooldown > 0:
		cooldown -= delta

	if is_attacking:
		attack_timer -= delta

func attack(direction: Vector2) -> bool:
	if cooldown > 0 or is_attacking:
		return false

	is_attacking = true
	attack_timer = attack_duration
	cooldown = 1.0 / attack_speed

	_perform_attack(direction)

	return true

func _perform_attack(direction: Vector2):
	# Override in subclasses
	pass

func get_damage() -> float:
	var multiplier = 1.0
	if owner_player and owner_player.has_node("ASComponent"):
		multiplier = owner_player.get_node("ASComponent").get_attribute("attack_power", 10.0) / 10.0

	return damage * multiplier
```

#### 2.2 Sword Implementation

```gdscript
# demo/gameplay/weapons/Sword.gd
extends Weapon

func _create_visual():
	# Blade
	var blade = ColorRect.new()
	blade.size = Vector2(15, 60)
	blade.color = Color(0.8, 0.8, 0.8)
	blade.offset = Vector2(-7.5, -30)
	add_child(blade)

	# Handle
	var handle = ColorRect.new()
	handle.size = Vector2(15, 20)
	handle.color = Color(0.4, 0.2, 0)
	handle.offset = Vector2(-7.5, -10)
	add_child(handle)

	# Collision
	var collision = CollisionShape2D.new()
	var shape = CapsuleShape2D.new()
	shape.height = 60
	shape.radius = 8
	collision.shape = shape
	add_child(collision)

func _perform_attack(direction: Vector2):
	# Create hit detector
	var hit_area = Area2D.new()
	var shape = CapsuleShape2D.new()
	shape.height = range_
	shape.radius = 15
	var collision = CollisionShape2D.new()
	collision.shape = shape
	hit_area.add_child(collision)

	# Position
	hit_area.global_position = owner_player.global_position + direction.normalized() * (range_ / 2)

	# Get overlapping areas
	var bodies = hit_area.get_overlapping_areas()

	for body in bodies:
		var target = body.get_parent()
		if target != owner_player and not (target in _get_hit_this_attack()):
			_apply_hit(target, direction)

	hit_area.queue_free()

	# Sound
	if AudioManager:
		AudioManager.play_sound("weapon_slash")

	weapon_attacked.emit()

func _apply_hit(target: Node, direction: Vector2):
	var damage = get_damage()

	if target.has_method("take_damage"):
		target.take_damage(damage)

	if target.has_method("apply_knockback"):
		target.apply_knockback(global_position, knockback)

	weapon_hit.emit(target)

func _get_hit_this_attack() -> Array:
	# To prevent hitting same target twice
	return []
```

#### 2.3 Create Weapon Scenes

```
demo/gameplay/weapons/Sword.tscn:
[Node2D]
script = res://demo/gameplay/weapons/Sword.gd

demo/gameplay/weapons/Staff.tscn:
[Node2D]
script = res://demo/gameplay/weapons/Staff.gd

demo/gameplay/weapons/Axe.tscn:
[Node2D]
script = res://demo/gameplay/weapons/Axe.gd

demo/gameplay/weapons/Dagger.tscn:
[Node2D]
script = res://demo/gameplay/weapons/Dagger.gd
```

---

### FASE 3: Dummies (6 horas)

#### 3.1 Dummy Base Class

```gdscript
# demo/gameplay/dummies/Dummy.gd
extends CharacterBody2D
class_name Dummy

@export var difficulty: String = "basic"
@export var max_health: float = 50.0

var health: float
var as_component: Node
var ai_behavior: String = "passive"

func _ready():
	health = max_health

	# Create AS Component
	as_component = preload("res://demo/gameplay/ASComponentDummy.tscn").instantiate()
	add_child(as_component)

	# Visual
	_create_visual()

	# AI
	if difficulty in ["advanced", "elite"]:
		ai_behavior = "aggressive"

func _create_visual():
	var color = _get_difficulty_color()

	var body = ColorRect.new()
	body.size = Vector2(30, 50)
	body.color = color
	body.position = -body.size / 2
	add_child(body)

	# Health bar above
	var health_bar_bg = ColorRect.new()
	health_bar_bg.size = Vector2(30, 5)
	health_bar_bg.color = Color.BLACK
	health_bar_bg.position = Vector2(-15, -60)
	add_child(health_bar_bg)

	var health_bar = ColorRect.new()
	health_bar.size = Vector2(30, 5)
	health_bar.color = Color.GREEN
	health_bar.position = Vector2(-15, -60)
	health_bar.name = "HealthBar"
	add_child(health_bar)

func _get_difficulty_color() -> Color:
	match difficulty:
		"basic":
			return Color.LIGHT_BLUE
		"advanced":
			return Color.YELLOW
		"elite":
			return Color.RED
		_:
			return Color.WHITE

func _process(delta):
	_update_health_bar()

	if ai_behavior == "aggressive":
		_update_ai(delta)

func _update_health_bar():
	if has_node("HealthBar"):
		var health_bar = get_node("HealthBar")
		health_bar.size.x = 30.0 * (health / max_health)

func _update_ai(delta):
	# Simple AI: find and attack nearest player
	var players = get_tree().get_nodes_in_group("Player")
	if players.is_empty():
		return

	var target = _find_nearest_player(players)
	if not target:
		return

	var direction = (target.global_position - global_position).normalized()
	velocity = direction * 100.0
	move_and_slide()

	if global_position.distance_to(target.global_position) < 60:
		_attack(target)

func _find_nearest_player(players: Array) -> Node:
	var nearest = players[0]
	var nearest_dist = global_position.distance_to(nearest.global_position)

	for player in players.slice(1):
		var dist = global_position.distance_to(player.global_position)
		if dist < nearest_dist:
			nearest = player
			nearest_dist = dist

	return nearest

func _attack(target: Node):
	if as_component:
		as_component.perform_ability("basic_attack")

	if target.has_method("take_damage"):
		target.take_damage(5.0)

func take_damage(amount: float):
	health -= amount

	if as_component:
		as_component.take_damage(amount)

	if health <= 0:
		_die()

func _die():
	# Grant XP to nearby players
	var players = get_tree().get_nodes_in_group("Player")
	for player in players:
		if global_position.distance_to(player.global_position) < 500:
			ProgressionSystem.add_experience(player.network_id, 25.0)

	AudioManager.play_sound("enemy_killed")
	queue_free()
```

#### 3.2 DummySpawner

```gdscript
# demo/gameplay/dummies/DummySpawner.gd
extends Node
class_name DummySpawner

@export var spawn_interval: float = 3.0
@export var max_dummies: int = 10
@export var difficulty: String = "basic"
@export var spawn_radius: float = 150.0

var spawn_timer: float = 0.0
var spawned_dummies: Array = []

func _ready():
	add_to_group("dummy_spawners")

func _process(delta):
	spawn_timer += delta

	# Clean dead dummies
	spawned_dummies = spawned_dummies.filter(func(d): return is_instance_valid(d))

	if spawn_timer >= spawn_interval and spawned_dummies.size() < max_dummies:
		spawn_timer = 0.0
		_spawn_dummy()

func _spawn_dummy():
	var dummy = DummyFactory.create_dummy(difficulty)
	var spawn_pos = global_position + Vector2(
		randf_range(-spawn_radius, spawn_radius),
		randf_range(-spawn_radius, spawn_radius)
	)
	dummy.global_position = spawn_pos
	get_parent().add_child(dummy)
	spawned_dummies.append(dummy)
```

---

### FASE 4: Skill Tree (6 horas)

#### 4.1 SkillTreeNode

```gdscript
# demo/gameplay/progression/SkillTreeNode.gd
extends Node
class_name SkillTreeNode

@export var skill_id: String = ""
@export var tier: int = 0
@export var cost: int = 1
@export var requirements: Array[String] = []

@export var skill_data: Dictionary = {
	"damage_multiplier": 1.0,
	"cooldown": 0.0,
	"duration": 0.0,
	"area_radius": 0.0,
}

var description: String = ""
var display_name: String = ""

func _ready():
	add_to_group("skill_tier_%d" % tier)

func is_unlocked(player_id: int) -> bool:
	var progress = ProgressionSystem.get_progress(player_id)
	return skill_id in progress.unlocked_skills

func can_unlock(player_id: int) -> bool:
	var progress = ProgressionSystem.get_progress(player_id)

	# Check skill points
	if progress.skill_points < cost:
		return false

	# Check requirements
	for req in requirements:
		if not (req in progress.unlocked_skills):
			return false

	return true

func unlock(player_id: int) -> bool:
	if not can_unlock(player_id):
		return false

	var progress = ProgressionSystem.get_progress(player_id)
	progress.unlocked_skills.append(skill_id)
	progress.skill_points -= cost

	return true
```

#### 4.2 SkillTreeUI

```gdscript
# demo/ui/SkillTreeUI.gd
extends Control
class_name SkillTreeUI

@onready var tree_container = $ScrollContainer/VBoxContainer
var player_id: int
var tier_sections: Dictionary = {}

func _ready():
	player_id = multiplayer.get_unique_id()
	_build_skill_tree()

func _build_skill_tree():
	# Clear
	for child in tree_container.get_children():
		child.queue_free()

	# Build by tier
	for tier in range(4):
		_add_tier_section(tier)

func _add_tier_section(tier: int):
	var section_label = Label.new()
	section_label.text = _get_tier_name(tier)
	section_label.add_theme_font_size_override("font_size", 20)
	section_label.modulate = Color.YELLOW
	tree_container.add_child(section_label)

	var hbox = HBoxContainer.new()
	hbox.alignment = BoxContainer.ALIGNMENT_CENTER
	tree_container.add_child(hbox)

	# Get all skills for this tier
	var skill_nodes = get_tree().get_nodes_in_group("skill_tier_%d" % tier)
	for skill_node in skill_nodes:
		var button = _create_skill_button(skill_node)
		hbox.add_child(button)

func _create_skill_button(skill_node: SkillTreeNode) -> Button:
	var button = Button.new()
	button.text = skill_node.skill_id.to_upper()
	button.custom_minimum_size = Vector2(140, 100)
	button.text_overrun_behavior = TextServer.OVERRUN_TRIM_ELLIPSIS

	if skill_node.is_unlocked(player_id):
		button.modulate = Color.GREEN
		button.disabled = true
		button.text += "\n[LEARNED]"
	elif skill_node.can_unlock(player_id):
		button.modulate = Color.YELLOW
		button.disabled = false
		button.pressed.connect(func(): _unlock_skill(skill_node))
	else:
		button.modulate = Color.GRAY
		button.disabled = true
		button.text += "\n[LOCKED]"

	return button

func _unlock_skill(skill_node: SkillTreeNode):
	if skill_node.unlock(player_id):
		AudioManager.play_sound("skill_unlock")
		_build_skill_tree()

func _get_tier_name(tier: int) -> String:
	match tier:
		0:
			return "TIER 0: Basic Skills"
		1:
			return "TIER 1: Combat Skills"
		2:
			return "TIER 2: Advanced Skills"
		3:
			return "TIER 3: Ultimate Skills"
		_:
			return "Unknown Tier"
```

---

### FASE 5: Level Design (8 horas)

#### 5.1 SpawnManager

```gdscript
# demo/gameplay/level/SpawnManager.gd
extends Node
class_name SpawnManager

@export var player_spawn_points: Array[Vector2] = [
	Vector2(100, 100),
	Vector2(200, 100),
	Vector2(100, 200),
	Vector2(200, 200),
]

@export var dummy_spawn_points: Array[Vector2] = [
	Vector2(150, 300),
	Vector2(250, 300),
	Vector2(350, 300),
	# ... adicionar mais
]

func get_spawn_for_player(player_id: int) -> Vector2:
	return player_spawn_points[player_id % player_spawn_points.size()]

func get_random_dummy_spawn(zone: int = 1) -> Vector2:
	if dummy_spawn_points.is_empty():
		return Vector2.ZERO
	return dummy_spawn_points[randi() % dummy_spawn_points.size()]
```

#### 5.2 Level Main Scene

```gdscript
# demo/scenes/gameplay_level.gd
extends Node2D
class_name GameplayLevel

var gameplay_manager: GameplayManager
var spawn_manager: SpawnManager
var zone_managers: Dictionary = {}

func _ready():
	# Initialize managers
	gameplay_manager = GameplayManager.new()
	add_child(gameplay_manager)

	spawn_manager = SpawnManager.new()
	add_child(spawn_manager)

	# Create zones
	_create_zones()

	# Spawn dummies
	_setup_dummy_spawners()

func _create_zones():
	# Zone 1: Combat Arena
	var arena = Node2D.new()
	arena.name = "CombatArena"
	arena.position = Vector2(0, 0)
	add_child(arena)

	# Add spawners
	var spawner1 = DummySpawner.new()
	spawner1.max_dummies = 15
	spawner1.difficulty = "basic"
	spawner1.position = Vector2(100, 150)
	arena.add_child(spawner1)

	# ... criar outras zonas

func _setup_dummy_spawners():
	# Zone 2: Training ground
	var spawner2 = DummySpawner.new()
	spawner2.max_dummies = 20
	spawner2.difficulty = "basic"
	add_child(spawner2)

	# ... adicionar mais spawners em pontos diferentes
```

---

## 📋 Checklist de Implementação

### FASE 1 (Core)
- [ ] GameplayManager criado
- [ ] DamageCalculator funcionando
- [ ] HitDetector detectando hits
- [ ] ProgressionSystem XP/Levels

### FASE 2 (Weapons)
- [ ] Weapon base class
- [ ] Sword implementada
- [ ] Staff implementada
- [ ] Axe implementada
- [ ] WeaponManager funcionando

### FASE 3 (Dummies)
- [ ] Dummy base class
- [ ] DummySpawner funcionando
- [ ] Dummies respondendo a ataques
- [ ] 15+ dummies simultâneos

### FASE 4 (Skills)
- [ ] SkillTreeNode funcionando
- [ ] 15 skills definidas
- [ ] Árvore de requerimentos
- [ ] SkillTreeUI funcionando

### FASE 5 (Level)
- [ ] SpawnManager funcionando
- [ ] 6 zonas criadas
- [ ] Procedural hazards
- [ ] Environment effects

### FASE 6 (Polish)
- [ ] HUD completo
- [ ] Audio integrado
- [ ] Feedback visual
- [ ] Balanceamento

---

## 🔗 Dependências de Implementação

```
Core Systems
├── DamageCalculator
├── HitDetector
└── GameplayManager
    ├── ProgressionSystem
    ├── SkillTree
    ├── WeaponManager
    └── CraftingSystem

Player Integration
├── PlayerController
├── PlayerAbilities
└── ASComponentPlayer

Combat & Damage
├── Weapon (base)
├── Sword, Staff, Axe, Dagger
├── Projectiles
└── StatusEffects

Progression
├── ProgressionSystem
├── SkillTree
└── RewardManager

Level Design
├── SpawnManager
├── DummySpawner
├── EnemySpawner
└── HazardSpawner

UI
├── GameHUD
├── SkillTreeUI
├── CombatFeedback
└── HealthBars

Audio
├── ProceduralMusic
├── ProceduralSound
└── AudioManager integration
```

---

## 💡 Tips de Desenvolvimento

1. **Comece com lo-fi**: Crie tudo com ColorRect e shapes básicas
2. **Teste isolado**: Cada sistema deve funcionar sozinho
3. **Integre gradualmente**: Sistema por sistema
4. **Use grupos**: `add_to_group()` para queries rápidas
5. **Signals**: Aproveite para comunicação entre sistemas
6. **RPC**: Certifique-se que damage é sincronizado
7. **Audio**: Integre sons procedurais progressivamente

---

## 🎯 Goals de Cada Fase

- **Fase 1**: Core systems funcionando isoladamente
- **Fase 2**: Armas causando damage real
- **Fase 3**: Inimigos para treinar
- **Fase 4**: Progressão funcional
- **Fase 5**: Mundo coeso e expanso
- **Fase 6**: Balanceado e polido

---

**Próximos Passos**:
1. Criar estrutura de pastas
2. Começar Fase 1 (GameplayManager + DamageCalculator)
3. Tester integração com Player existente
4. Ir para Fase 2 (Weapons)

Quer que eu comece a implementação de um desses sistemas?
