# Gameplay Systems - Quick Reference

Referência rápida de todas as classes, métodos e sinais do sistema de gameplay.

---

## 📦 Core Classes

### GameplayManager
**Arquivo**: `demo/gameplay/core/GameplayManager.gd`
**Propósito**: Coordenador central de todos os sistemas

```gdscript
func _init_systems()
func register_player(player_id: int, player_node: Node)
func unregister_player(player_id: int)
func apply_damage(attacker: Node, defender: Node, ability: String) -> float
```

**Variáveis Públicas**:
- `progression_system: ProgressionSystem`
- `damage_calculator: DamageCalculator`
- `weapon_manager: WeaponManager`
- `skill_tree: SkillTree`
- `crafting_system: CraftingSystem`
- `audio_manager: AudioManager`

---

### DamageCalculator
**Arquivo**: `demo/gameplay/core/DamageCalculator.gd`
**Propósito**: Cálculo de dano com multipliers e críticos

```gdscript
func calculate_damage(attacker: Node, defender: Node, ability: String) -> DamageResult

class DamageResult:
	var base_damage: float
	var multiplier: float
	var final_damage: float
	var critical: bool
	var hit_type: String  # "normal", "critical", "miss"
```

**Exportados**:
- `critical_chance: float = 0.1` (10%)
- `critical_multiplier: float = 2.0`

**Ability Multipliers**:
- basic_attack: 1.0
- power_slash: 1.5
- super_strike: 2.5
- dash_attack: 1.8
- whirlwind: 0.8
- annihilate: 4.0

---

### HitDetector
**Arquivo**: `demo/gameplay/core/HitDetector.gd`
**Propósito**: Detectar hits e aplicar dano

```gdscript
@export var damage: float
@export var knockback: float
@export var hit_effects: Array[String]

func _on_area_entered(area: Area2D)
func _apply_hit(target: Node, direction: Vector2)
```

**Sinais**: Nenhum (usar área2D signals)

---

## 📊 Progression System

### ProgressionSystem
**Arquivo**: `demo/gameplay/progression/ProgressionSystem.gd`
**Propósito**: Gerenciar XP, levels e skills

```gdscript
class PlayerProgress:
	var level: int = 1
	var experience: float = 0.0
	var experience_to_next: float = 100.0
	var skill_points: int = 0
	var unlocked_skills: Array[String]
	var attributes: Dictionary

func register_player(player_id: int)
func add_experience(player_id: int, amount: float)
func _level_up(player_id: int)
func get_progress(player_id: int) -> PlayerProgress
func unlock_skill(player_id: int, skill_id: String) -> bool
```

**Sinais**:
```gdscript
signal level_up(player_id: int, new_level: int)
signal skill_point_gained(player_id: int)
signal xp_gained(player_id: int, amount: float)
```

**Attributes Padrão**:
- max_health: 100.0
- max_mana: 50.0
- attack_power: 10.0
- defense: 5.0
- movement_speed: 200.0

---

### SkillTreeNode
**Arquivo**: `demo/gameplay/progression/SkillTreeNode.gd`
**Propósito**: Nó individual da skill tree

```gdscript
@export var skill_id: String
@export var tier: int
@export var cost: int = 1
@export var requirements: Array[String]
@export var skill_data: Dictionary

func is_unlocked(player_id: int) -> bool
func can_unlock(player_id: int) -> bool
func unlock(player_id: int) -> bool
```

**Grupos**: `skill_tier_{tier}`

**Exemplo de skill_data**:
```gdscript
{
	"damage_multiplier": 1.5,
	"cooldown": 1.0,
	"duration": 3.0,
	"area_radius": 100.0,
}
```

---

## ⚔️ Combat System

### Weapon (Base)
**Arquivo**: `demo/gameplay/weapons/Weapon.gd`
**Propósito**: Classe base para armas

```gdscript
@export var weapon_name: String
@export var damage: float
@export var attack_speed: float
@export var range_: float
@export var knockback: float
@export var attack_duration: float

func attack(direction: Vector2) -> bool
func _perform_attack(direction: Vector2)
func get_damage() -> float
```

**Sinais**:
```gdscript
signal weapon_attacked
signal weapon_hit(target: Node)
```

**Variáveis Internas**:
- `cooldown: float`
- `is_attacking: bool`
- `attack_timer: float`

---

### Sword
**Arquivo**: `demo/gameplay/weapons/Sword.gd`

```gdscript
@export var weapon_name: String = "Iron Sword"
@export var damage: float = 10.0
@export var attack_speed: float = 1.0
@export var range_: float = 60.0
@export var knockback: float = 200.0

func _perform_attack(direction: Vector2)
func _apply_hit(target: Node, direction: Vector2)
```

---

### Staff
**Arquivo**: `demo/gameplay/weapons/Staff.gd`

```gdscript
@export var weapon_name: String = "Magic Staff"
@export var damage: float = 8.0
@export var attack_speed: float = 1.5
@export var range_: float = 300.0  # Projectile range

func _perform_attack(direction: Vector2)
# Cria projectile ao invés de hit direto
```

---

### Axe
**Arquivo**: `demo/gameplay/weapons/Axe.gd`

```gdscript
@export var weapon_name: String = "Axe"
@export var damage: float = 15.0
@export var attack_speed: float = 0.7
@export var range_: float = 80.0  # AOE radius
@export var knockback: float = 250.0

func _perform_attack(direction: Vector2)
# AOE circular ao invés de hit único
```

---

### Dagger
**Arquivo**: `demo/gameplay/weapons/Dagger.gd`

```gdscript
@export var weapon_name: String = "Dagger"
@export var damage: float = 7.0
@export var attack_speed: float = 2.0
@export var range_: float = 40.0
@export var knockback: float = 150.0

# Quick hits, low range, high speed
```

---

### WeaponManager
**Arquivo**: `demo/gameplay/weapons/WeaponManager.gd`
**Propósito**: Gerenciar criação e stats de armas

```gdscript
func create_weapon(weapon_type: String) -> Weapon
func get_weapon_stats(weapon_type: String) -> Dictionary
```

**Tipos de Arma Suportados**:
- "sword"
- "staff"
- "axe"
- "dagger"

---

## 👹 Dummy System

### Dummy
**Arquivo**: `demo/gameplay/dummies/Dummy.gd`
**Propósito**: NPC treináveis para práticar combate

```gdscript
@export var difficulty: String  # "basic", "advanced", "elite"
@export var max_health: float = 50.0

var health: float
var as_component: Node
var ai_behavior: String

func take_damage(amount: float)
func _die()
func _attack(target: Node)
```

**Dificuldades**:
- **basic**: Pasivo, sem IA, HP 50
- **advanced**: IA agressiva, HP 100
- **elite**: IA inteligente, HP 200

---

### DummyFactory
**Arquivo**: `demo/gameplay/dummies/DummyFactory.gd`
**Propósito**: Factory para criar dummies

```gdscript
static func create_dummy(type: String = "basic", position: Vector2 = Vector2.ZERO) -> Dummy
static func create_combat_dummy() -> Dummy
static func create_training_dummy() -> Dummy
static func create_elite_dummy() -> Dummy
```

---

### DummySpawner
**Arquivo**: `demo/gameplay/dummies/DummySpawner.gd`
**Propósito**: Spawn automático de dummies

```gdscript
@export var spawn_interval: float = 3.0
@export var max_dummies: int = 10
@export var difficulty: String = "basic"
@export var spawn_radius: float = 150.0

func _spawn_dummy()
func change_difficulty(new_difficulty: String)
```

---

## 👾 Enemy System

### Enemy
**Arquivo**: `demo/gameplay/enemies/Enemy.gd`
**Propósito**: Inimigos hostis

```gdscript
@export var level: int = 1
@export var health: float = 50.0
@export var damage: float = 5.0
@export var experience_reward: float = 25.0

var ai_state: String  # "patrol", "chase", "attack"
var target: Node

func take_damage(amount: float)
func _die()
func _patrol(delta)
func _chase(delta)
func _perform_attack(delta)
```

---

### EnemySpawner
**Arquivo**: `demo/gameplay/enemies/EnemySpawner.gd`

```gdscript
@export var spawn_interval: float = 3.0
@export var max_enemies: int = 10
@export var enemy_level: int = 1

func _spawn_enemy()
```

---

## 🎯 Level Design

### SpawnManager
**Arquivo**: `demo/gameplay/level/SpawnManager.gd`
**Propósito**: Gerenciar spawn points

```gdscript
@export var player_spawn_points: Array[Vector2]
@export var dummy_spawn_points: Array[Vector2]

func get_spawn_for_player(player_id: int) -> Vector2
func get_random_dummy_spawn(zone: int = 1) -> Vector2
```

---

### GameplayLevel
**Arquivo**: `demo/scenes/gameplay_level.gd`
**Propósito**: Scene principal do level

```gdscript
var gameplay_manager: GameplayManager
var spawn_manager: SpawnManager
var zone_managers: Dictionary

func _create_zones()
func _setup_dummy_spawners()
```

---

## 🎨 UI Classes

### GameHUD
**Arquivo**: `demo/ui/GameHUD.gd`
**Propósito**: HUD principal em-game

```gdscript
func _update_health()
func _update_mana()
func _update_level()
func _update_xp()
func _on_tag_changed(tag: String, added: bool)
```

---

### SkillTreeUI
**Arquivo**: `demo/ui/SkillTreeUI.gd`
**Propósito**: Interface da skill tree

```gdscript
func _build_skill_tree()
func _add_tier_section(tier: int)
func _create_skill_button(skill_node: SkillTreeNode) -> Button
func _unlock_skill(skill_node: SkillTreeNode)
```

---

### CombatFeedback
**Arquivo**: `demo/ui/CombatFeedback.gd`
**Propósito**: Números de dano flutuantes

```gdscript
func show_damage(position: Vector2, amount: float, is_critical: bool = false)
```

---

## 🔊 Audio System

### AudioManager
**Arquivo**: `demo/autoload/AudioManager.gd` (modificado)
**Propósito**: Gerenciar sons procedurais

```gdscript
var sounds: Dictionary

func _init_sounds()
func _create_sound(type, freq, dur, vol, bend, noise = 0.0) -> ProceduralSound
func play_sound(sound_name: String)
func play_sound_with_variation(base_sound: String, pitch_variation: float = 0.1)
```

**Sons Principais**:
```
weapon_slash
weapon_heavy_swing
staff_cast
hit_impact
critical_hit
skill_unlock
enemy_killed
level_up
```

---

## 🎮 Integration Points

### Player Integration
O Player precisa ter/fazer:

```gdscript
var gameplay_manager: GameplayManager  # Referência
var current_weapon: Weapon

func _ready():
	# Registrar no GameplayManager
	if gameplay_manager:
		gameplay_manager.register_player(network_id, self)

func attack(direction: Vector2):
	if current_weapon:
		current_weapon.attack(direction)

func take_damage(amount: float):
	if has_node("ASComponent"):
		get_node("ASComponent").take_damage(amount)
```

---

## 📊 Data Structures

### PlayerProgress

```gdscript
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
```

### DamageResult

```gdscript
class DamageResult:
	var base_damage: float
	var multiplier: float
	var final_damage: float
	var critical: bool
	var hit_type: String  # "normal", "critical", "miss"
```

---

## 🔗 RPC Methods

Todos os métodos que precisam sincronização multiplayer:

```gdscript
# Player attacking
@rpc("unreliable")
func perform_attack(weapon_type: String, direction: Vector2)

# Damage received
@rpc("reliable")
func take_damage(amount: float, attacker_id: int)

# Knockback
@rpc("unreliable")
func apply_knockback(knockback_force: Vector2)

# Status effect
@rpc("reliable")
func apply_status_effect(effect_type: String, duration: float)

# Kill notification
@rpc("reliable")
func on_killed(killer_id: int)

# Skill unlock
@rpc("reliable")
func unlock_skill(skill_id: String)
```

---

## 📋 Attributes Management

### Getting Attributes

```gdscript
var power = as_component.get_attribute("attack_power", 10.0)
var health = as_component.get_attribute("health", 100.0)
```

### Setting Attributes

```gdscript
as_component.set_attribute("health", 80.0)
as_component.set_attribute("attack_power", 15.0)
```

### Modifying with Multipliers

```gdscript
as_component.apply_multiplier("movement_speed", 0.5, 5.0)  # 50% por 5s
```

---

## 🎯 Common Patterns

### Causing Damage

```gdscript
var calc = DamageCalculator.new()
var result = calc.calculate_damage(attacker, defender, "power_slash")

if result.final_damage > 0:
	defender.take_damage(result.final_damage)
	if result.critical:
		CombatFeedback.show_damage(defender.position, result.final_damage, true)
```

### Granting XP

```gdscript
ProgressionSystem.add_experience(player_id, 50.0)
# Automatically emits xp_gained signal
# Automatically levels up if needed
```

### Unlocking Skill

```gdscript
var progress = ProgressionSystem.get_progress(player_id)
if progress.skill_points > 0:
	ProgressionSystem.unlock_skill(player_id, "power_slash")
```

### Creating Weapon

```gdscript
var weapon = WeaponManager.create_weapon("sword")
weapon.owner_player = player
weapon.global_position = player.global_position
player.add_child(weapon)
```

### Spawning Dummy

```gdscript
var dummy = DummyFactory.create_dummy("basic", Vector2(100, 100))
level.add_child(dummy)
```

---

## ⚙️ Configuration Constants

### Enemy Spawn Rates
```gdscript
DUMMY_SPAWN_INTERVAL = 3.0
DUMMY_MAX = 15
ENEMY_SPAWN_INTERVAL = 4.0
ENEMY_MAX = 10
```

### Combat Values
```gdscript
CRITICAL_CHANCE = 0.1  # 10%
CRITICAL_MULTIPLIER = 2.0
BASE_ATTACK_POWER = 10.0
BASE_DEFENSE = 5.0
```

### Progression
```gdscript
BASE_XP_TO_LEVEL = 100.0
XP_LEVEL_MULTIPLIER = 1.2
ATTRIBUTE_GROWTH_PER_LEVEL = 1.1  # 10% increase
```

### Map Size
```gdscript
MAP_WIDTH = 2000
MAP_HEIGHT = 2000
ZONE_SIZE = 500  # 500x500 each
```

---

## 🚨 Error Handling

### Common Checks

```gdscript
# Check if player exists
if not is_instance_valid(player):
	return

# Check if has ASComponent
if not player.has_node("ASComponent"):
	return

# Check if can perform action
if not can_unlock(player_id):
	print("Cannot unlock skill")
	return

# Check multiplayer
if multiplayer.is_server():
	# Only execute on server
```

---

## 📚 Related Documentation

- **GAMEPLAY_DESIGN_DOCUMENT.md** - Design completo
- **GAMEPLAY_IMPLEMENTATION_PLAN.md** - Plano passo-a-passo
- **RBAC_SYSTEM_GUIDE.md** - Sistema de admin/permissões
- **Player.gd** - Classe player existente

---

## 🔍 Debugging

### Enable Debug Logging

```gdscript
# Em GameplayManager._init_systems()
var debug = true

if debug:
	print("[GameplayManager] Initializing systems...")
	print("[DamageCalculator] Critical chance: %f" % critical_chance)
```

### Check Progress

```gdscript
var progress = ProgressionSystem.get_progress(player_id)
print("Level: %d, XP: %.0f / %.0f" % [progress.level, progress.experience, progress.experience_to_next])
print("Skills: %s" % progress.unlocked_skills)
```

### Monitor Dummies

```gdscript
var spawner = get_node("DummySpawner1")
print("Active dummies: %d / %d" % [spawner.spawned_dummies.size(), spawner.max_dummies])
```

---

**Última Atualização**: 2026-04-05
**Versão**: 1.0
**Status**: Ready for Implementation
