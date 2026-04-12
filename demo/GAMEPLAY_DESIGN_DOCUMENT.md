# Ability System Demo - Gameplay Design Document

**Demonstração Multiplayer do Ability System com sistemas de combate, progressão, skill tree e mais.**

## 📋 Visão Geral

Este é um **protótipo de demonstração** do Ability System em um ambiente multiplayer. O foco é:
- ✅ Demonstrar o Ability System em ação
- ✅ Múltiplos jogadores interagindo
- ✅ Combate emergente baseado em abilities
- ✅ Progressão e customização
- ✅ Tudo procedural (sem imagens)

**Nota**: Não estamos fazendo um "jogo completo" - é uma demonstração. As mecânicas não precisam ser realistas, apenas demonstrarem o poder do framework.

---

## 1. LEVEL DESIGN

### 1.1 Conceito Geral

**Mapa Procedural Expandido**: Arena vasta com múltiplas zonas temáticas.

```
┌─────────────────────────────────────────────────┐
│                     2000 x 2000                 │
├─────────────────────────────────────────────────┤
│  Zone 1         │     Zone 2      │   Zone 3    │
│  (Combat)       │   (Training)    │  (Chaos)    │
│                 │                 │             │
│  Spawn Dummies  │  Training Area  │ Auto-Hazards│
│  Open Arena     │  Targets        │ Projectiles │
│                 │                 │             │
├─────────────────────────────────────────────────┤
│  Zone 4         │     Zone 5      │   Zone 6    │
│  (Nature)       │   (Corrupted)   │ (Sanctuary) │
│                 │                 │             │
│  Obstacles      │  Hostile NPCs   │ Safe Zone   │
│  Platforms      │  Traps          │ Vendor      │
│                 │                 │             │
└─────────────────────────────────────────────────┘
```

### 1.2 Zonas Detalhadas

#### ZONA 1: Combat Arena (500x500 @ 0,0)
- **Propósito**: Teste livre de combate
- **Features**:
  - Spawn points para 4 jogadores
  - 10-15 dummies treináveis
  - Paredes de contenção
  - Sem hazards
- **Nodes**:
  - `SpawnPoint` (Node2D com trigger para respawn)
  - `CombatDummy` (ASComponent com AI básica)
  - `Arena Walls` (StaticBody2D procedural)

#### ZONA 2: Training Grounds (500x500 @ 500,0)
- **Propósito**: Desenvolvimento de skills
- **Features**:
  - Targets estacionários (não contam)
  - Dummies passivos (não atacam)
  - Blocos de prática (destruíveis)
  - Áreas de efeito para teste
- **Nodes**:
  - `TrainingTarget` (visual feedback, não colisão)
  - `PracticeBlock` (destruído com X damage)
  - `PassiveDummy` (só toma hit, não revida)

#### ZONA 3: Chaos Zone (500x500 @ 1000,0)
- **Propósito**: Teste de resistência
- **Features**:
  - Projectiles aleatórios
  - Hazards procedurais
  - Enemies agressivos
  - Arenas variáveis
- **Nodes**:
  - `HazardSpawner` (emite obstacles)
  - `ProjectileEmitter` (atira aleatoriamente)
  - `AggressiveEnemy` (AI combativa)

#### ZONA 4: Natural Environment (500x500 @ 0,500)
- **Propósito**: Navegação e obstáculos
- **Features**:
  - Plataformas em altura
  - Gaps e pontes
  - Obstacles de movimento
  - Dummies para praticar mobilidade
- **Nodes**:
  - `Platform` (procedurally generated)
  - `MovingObstacle` (patrulha)
  - `Teleporter` (volta na zona)

#### ZONA 5: Corrupted Realm (500x500 @ 500,500)
- **Propósito**: Combate desafiador
- **Features**:
  - Múltiplos NPCs agressivos
  - Status effects no ambiente
  - Traps
  - Boss arena opcional
- **Nodes**:
  - `EnvironmentalEffect` (aplica status)
  - `Trap` (trigger damage)
  - `EliteEnemy` (AI sofisticada)
  - `BossArena` (arena maior, fechada)

#### ZONA 6: Sanctuary (500x500 @ 1000,500)
- **Propósito**: Hub de descanso e customização
- **Features**:
  - Sem combate
  - NPCs vendor
  - Skill tree UI
  - Crafting area
- **Nodes**:
  - `Vendor` (node com dialog)
  - `SkillTreeUI` (painel de skills)
  - `CraftingStation` (node de crafting)
  - Healing zones

### 1.3 Sistema de Spawn

```gdscript
# demo/gameplay/SpawnManager.gd
extends Node

@export var spawn_points: Array[Vector2] = [
	Vector2(100, 100),    # Zone 1 - Spawn 1
	Vector2(200, 100),    # Zone 1 - Spawn 2
	Vector2(100, 200),    # Zone 1 - Spawn 3
	Vector2(200, 200),    # Zone 1 - Spawn 4
]

@export var dummy_spawn_points: Array[Vector2] = [
	Vector2(150, 300),
	Vector2(250, 300),
	Vector2(350, 300),
	# ... 10+ pontos
]

func get_spawn_for_player(player_id: int) -> Vector2:
	return spawn_points[player_id % spawn_points.size()]

func spawn_dummy_at_zone(zone: int, dummy_type: String = "basic") -> Node:
	var dummy = preload("res://demo/gameplay/dummies/Dummy.tscn").instantiate()
	dummy.global_position = _get_random_dummy_spawn(zone)
	dummy.difficulty = dummy_type
	return dummy
```

---

## 2. COMBAT SYSTEM

### 2.1 Fluxo de Combate

```
Player Input
    ↓
Ability Manager (via Ability System)
    ↓
Hit Detection (Area2D/RayCast checks)
    ↓
Apply Damage → RBAC permission check
    ↓
Status Effects Applied
    ↓
Update HUD / Sounds
```

### 2.2 Detecção de Hit

```gdscript
# demo/gameplay/HitDetector.gd
extends Area2D
class_name HitDetector

@export var damage: float = 10.0
@export var knockback: float = 200.0
@export var hit_effects: Array[String] = ["stun", "knockback"]

var hit_entities: Dictionary = {}  # Prevent multiple hits same frame

func _on_area_entered(area: Area2D):
	var target = area.get_parent() if area.get_parent() else area

	if target in hit_entities:
		return  # Already hit

	hit_entities[target] = true

	# Apply damage via ASComponent
	if target.has_node("ASComponent"):
		var as_comp = target.get_node("ASComponent")
		as_comp.take_damage(damage)
		as_comp.apply_knockback(global_position, knockback)

	# Play sound
	if AudioManager:
		AudioManager.play_sound("hit_" + hit_effects[0])

	# Remove after brief cooldown
	await get_tree().create_timer(0.1).timeout
	hit_entities.erase(target)
```

### 2.3 Sistema de Dano

```gdscript
# demo/gameplay/DamageCalculator.gd
extends Node

class DamageResult:
	var base_damage: float
	var multiplier: float
	var final_damage: float
	var critical: bool

func calculate_damage(attacker: Node, defender: Node, ability: String) -> DamageResult:
	var result = DamageResult.new()

	# Get attacker attributes via ASComponent
	var attacker_power = 10.0
	if attacker.has_node("ASComponent"):
		attacker_power = attacker.get_node("ASComponent").get_attribute("attack_power", 10.0)

	result.base_damage = attacker_power

	# Apply ability multiplier
	var ability_multiplier = _get_ability_multiplier(ability)
	result.multiplier = ability_multiplier

	# Critical chance
	result.critical = randf() < 0.1  # 10% crit chance
	if result.critical:
		result.multiplier *= 2.0

	result.final_damage = result.base_damage * result.multiplier

	return result

func _get_ability_multiplier(ability: String) -> float:
	var multipliers = {
		"basic_attack": 1.0,
		"power_slash": 1.5,
		"super_strike": 2.5,
		"combo_finisher": 3.0,
	}
	return multipliers.get(ability, 1.0)
```

### 2.4 Sistema de Status Effects

```gdscript
# demo/gameplay/StatusEffect.gd
extends Resource
class_name StatusEffect

enum Type { STUN, KNOCKBACK, BLEED, POISON, SLOW, BURN, FREEZE }

@export var type: Type = Type.STUN
@export var duration: float = 1.0
@export var magnitude: float = 1.0

func apply(target: Node):
	if not target.has_node("ASComponent"):
		return

	var as_comp = target.get_node("ASComponent")

	match type:
		Type.STUN:
			as_comp.set_tag("status.stunned", true)
			await get_tree().create_timer(duration).timeout
			as_comp.remove_tag("status.stunned")

		Type.KNOCKBACK:
			# Applied via velocity
			pass

		Type.SLOW:
			as_comp.apply_multiplier("movement_speed", 0.5, duration)

		Type.BURN:
			_apply_damage_over_time(as_comp, magnitude, duration)

func _apply_damage_over_time(as_comp: Node, damage: float, duration: float):
	var ticks = int(duration / 0.5)
	for i in range(ticks):
		await get_tree().create_timer(0.5).timeout
		as_comp.take_damage(damage)
```

---

## 3. SISTEMA DE PROGRESSÃO

### 3.1 XP e Levels

```gdscript
# demo/gameplay/ProgressionSystem.gd
extends Node

class PlayerProgress:
	var level: int = 1
	var experience: float = 0.0
	var experience_to_next: float = 100.0
	var skill_points: int = 0
	var attributes: Dictionary = {
		"max_health": 100.0,
		"max_mana": 50.0,
		"attack_power": 10.0,
		"defense": 5.0,
		"movement_speed": 200.0,
	}

func add_experience(player_id: int, amount: float):
	var progress = _get_progress(player_id)
	progress.experience += amount

	if progress.experience >= progress.experience_to_next:
		_level_up(player_id)

func _level_up(player_id: int):
	var progress = _get_progress(player_id)
	progress.level += 1
	progress.experience = 0.0
	progress.experience_to_next = 100.0 * (progress.level * 1.2)
	progress.skill_points += 1

	# Apply stat increases
	for stat in progress.attributes.keys():
		progress.attributes[stat] *= 1.1

	# Notify
	if AudioManager:
		AudioManager.play_sound("level_up")

	_broadcast_level_up(player_id, progress.level)

func _broadcast_level_up(player_id: int, new_level: int):
	if AdminCommand:
		AdminCommand.execute_command(1, "broadcast_message",
			["Player %d leveled up to %d!" % [player_id, new_level]])
```

### 3.2 Reward System

```gdscript
# demo/gameplay/RewardManager.gd
extends Node

enum RewardType { XP, GOLD, LOOT, SKILL_UNLOCK }

class Reward:
	var type: RewardType
	var amount: float
	var data: Dictionary

func grant_reward(player_id: int, reward: Reward):
	match reward.type:
		RewardType.XP:
			ProgressionSystem.add_experience(player_id, reward.amount)

		RewardType.GOLD:
			var progress = ProgressionSystem._get_progress(player_id)
			progress.gold = progress.get("gold", 0) + reward.amount

		RewardType.SKILL_UNLOCK:
			ProgressionSystem.unlock_skill(player_id, reward.data["skill_id"])

		RewardType.LOOT:
			ProgressionSystem.add_to_inventory(player_id, reward.data["item_id"])

func create_reward_from_kill(enemy_level: int) -> Reward:
	var reward = Reward.new()
	reward.type = RewardType.XP
	reward.amount = enemy_level * 25.0
	return reward
```

---

## 4. SKILL TREE

### 4.1 Estrutura do Skill Tree

```
                        ULTIMATE SKILLS
                              ↑
                    ┌─────────┴─────────┐
                    │                   │
                  TIER 3            TIER 3
               Combat Skills      Support Skills
                    ↑                   ↑
                ┌───┴───┐           ┌───┴───┐
                │       │           │       │
              TIER 2  TIER 2       TIER 2  TIER 2
              Skills Skills       Skills  Skills
                │       │           │       │
                └───┬───┘           └───┬───┘
                    │                   │
                  TIER 1          TIER 1
              Basic Skills     Utility Skills
                    │                   │
                    └───────┬───────────┘
                            │
                        START (Tier 0)
```

### 4.2 Skills Definidos

#### TIER 0: Basic Skills (Gratuitos)

| Skill | Requirement | Effect |
|-------|------------|--------|
| Basic Attack | - | Ataque simples, 1x damage |
| Sprint | - | Aumenta velocidade 2x por 3s |
| Block | - | Reduz damage em 50% por 2s |

#### TIER 1: Combat Skills (Custa 1 Skill Point)

| Skill | Requirement | Effect |
|-------|------------|--------|
| Power Slash | Basic Attack | Ataque com knockback, 1.5x damage |
| Dash Attack | Sprint | Dash + ataque, 1.8x damage |
| Whirlwind | Basic Attack | Área circular, hits múltiplos, 0.8x damage cada |
| Counter | Block | Bloqueia e revida, 1.2x damage |

#### TIER 2: Advanced Skills (Custa 2 Skill Points)

| Skill | Requirement | Effect |
|-------|------------|--------|
| Super Strike | Power Slash (Tier 1) | AOE ground slam, 2.5x damage |
| Multi-Hit | Dash Attack (Tier 1) | 3 hits rápidos, 1.0x damage cada |
| Vortex | Whirlwind (Tier 1) | Puxar inimigos, stun 2s |
| Parry | Counter (Tier 1) | Parry com invulnerabilidade 0.5s |

#### TIER 3: Ultimate Skills (Custa 3 Skill Points)

| Skill | Requirement | Effect |
|-------|------------|--------|
| Annihilate | Super Strike (Tier 2) | HUGE AOE, 4.0x damage, stun 3s |
| Blademaster | Multi-Hit (Tier 2) | 5 hits, cada hit melhor que o anterior |
| Singularity | Vortex (Tier 2) | Puxar tudo, stun, então explodir |

#### Support Skills

| Skill | Tier | Effect |
|-------|------|--------|
| Heal | 1 | Cure 30% HP |
| Shield | 2 | Damage reduction 30% x 5s |
| Revive | 3 | Ressuscita aliado com 50% HP |
| Berserk | 2 | +100% damage, -50% defense 8s |

### 4.3 Skill Tree Node

```gdscript
# demo/gameplay/SkillTreeNode.gd
extends Node
class_name SkillTreeNode

@export var skill_id: String = ""
@export var tier: int = 0
@export var cost: int = 1
@export var requirements: Array[String] = []
@export var position_in_tree: Vector2 = Vector2.ZERO

@export var skill_data: Dictionary = {
	"damage_multiplier": 1.0,
	"cooldown": 0.0,
	"duration": 0.0,
	"area_radius": 0.0,
}

func is_unlocked(player_id: int) -> bool:
	var progress = ProgressionSystem._get_progress(player_id)
	return skill_id in progress.get("unlocked_skills", [])

func can_unlock(player_id: int) -> bool:
	var progress = ProgressionSystem._get_progress(player_id)

	# Check skill points
	if progress.skill_points < cost:
		return false

	# Check requirements
	for req in requirements:
		if not (req in progress.get("unlocked_skills", [])):
			return false

	return true

func unlock(player_id: int) -> bool:
	if not can_unlock(player_id):
		return false

	var progress = ProgressionSystem._get_progress(player_id)
	progress.unlocked_skills.append(skill_id)
	progress.skill_points -= cost

	return true
```

### 4.4 Skill Tree UI

```gdscript
# demo/ui/SkillTreeUI.gd
extends Control

@onready var tree_container = $ScrollContainer/VBoxContainer
var skill_buttons: Dictionary = {}

func _ready():
	_build_skill_tree()

func _build_skill_tree():
	# Tier 0
	_add_tier_section("Basic Skills", 0)

	# Tier 1
	_add_tier_section("Combat Skills", 1)

	# Tier 2
	_add_tier_section("Advanced Skills", 2)

	# Tier 3
	_add_tier_section("Ultimate Skills", 3)

func _add_tier_section(title: String, tier: int):
	var label = Label.new()
	label.text = title
	label.add_theme_font_size_override("font_size", 20)
	tree_container.add_child(label)

	var hbox = HBoxContainer.new()
	tree_container.add_child(hbox)

	# Get all skills for this tier
	for skill_node in get_tree().get_nodes_in_group("skill_tier_%d" % tier):
		var button = _create_skill_button(skill_node)
		hbox.add_child(button)

func _create_skill_button(skill_node: SkillTreeNode) -> Button:
	var button = Button.new()
	button.text = skill_node.skill_id.to_upper()
	button.custom_minimum_size = Vector2(150, 80)

	var player_id = multiplayer.get_unique_id()

	if skill_node.is_unlocked(player_id):
		button.modulate = Color.GREEN
		button.disabled = true
		button.text += "\n(LEARNED)"
	elif skill_node.can_unlock(player_id):
		button.modulate = Color.YELLOW
		button.pressed.connect(func(): _unlock_skill(skill_node, player_id))
	else:
		button.modulate = Color.GRAY
		button.disabled = true
		button.text += "\n(Locked)"

	return button

func _unlock_skill(skill_node: SkillTreeNode, player_id: int):
	if skill_node.unlock(player_id):
		AudioManager.play_sound("skill_unlock")
		_build_skill_tree()  # Refresh
```

---

## 5. DUMMY SYSTEM

### 5.1 Dummy Base

```gdscript
# demo/gameplay/dummies/Dummy.gd
extends CharacterBody2D
class_name Dummy

@export var difficulty: String = "basic"  # basic, advanced, elite
@export var health: float = 50.0
@export var max_health: float = 50.0

var as_component: Node
var ai_behavior: String = "passive"
var target: Node = null
var state: String = "idle"

func _ready():
	# Create ASComponent
	as_component = preload("res://demo/gameplay/ASComponentDummy.tscn").instantiate()
	add_child(as_component)
	as_component.max_health = max_health
	as_component.health = health

	# Add visual representation
	_create_visual()

	# Setup AI if aggressive
	if difficulty in ["advanced", "elite"]:
		ai_behavior = "aggressive"
		set_process(true)

func _create_visual():
	# Create dummy as procedural shape
	var body = ColorRect.new()
	body.size = Vector2(30, 50)
	body.color = _get_difficulty_color()
	body.position = -body.size / 2
	add_child(body)

	# Add collision
	var collision = CollisionShape2D.new()
	var shape = CapsuleShape2D.new()
	shape.height = 50
	shape.radius = 15
	collision.shape = shape
	add_child(collision)

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
	if ai_behavior == "aggressive":
		_update_ai(delta)

func _update_ai(delta):
	# Find nearest player
	var players = get_tree().get_nodes_in_group("Player")
	if players.is_empty():
		return

	target = players.reduce(func(closest, p):
		return p if global_position.distance_to(p.global_position) < global_position.distance_to(closest.global_position) else closest
	)

	# Simple AI: move towards player
	var direction = (target.global_position - global_position).normalized()
	velocity = direction * 100.0
	move_and_slide()

	# Attack if close
	if global_position.distance_to(target.global_position) < 60:
		_attack()

func _attack():
	if as_component:
		as_component.perform_ability("basic_attack")

func take_damage(amount: float):
	health -= amount
	if as_component:
		as_component.take_damage(amount)

	if health <= 0:
		_die()

func _die():
	# Grant XP to killers
	var players = get_tree().get_nodes_in_group("Player")
	for player in players:
		if global_position.distance_to(player.global_position) < 500:
			ProgressionSystem.add_experience(player.network_id, 50)

	queue_free()
```

### 5.2 Dummy Factory

```gdscript
# demo/gameplay/DummyFactory.gd
extends Node

static func create_dummy(type: String = "basic", position: Vector2 = Vector2.ZERO) -> Dummy:
	var dummy = preload("res://demo/gameplay/dummies/Dummy.tscn").instantiate()
	dummy.difficulty = type
	dummy.global_position = position
	return dummy

static func create_combat_dummy() -> Dummy:
	return create_dummy("basic")

static func create_training_dummy() -> Dummy:
	var dummy = create_dummy("basic")
	dummy.ai_behavior = "passive"
	return dummy

static func create_elite_dummy() -> Dummy:
	var dummy = create_dummy("elite")
	dummy.max_health = 200
	dummy.health = 200
	return dummy
```

### 5.3 Spawner de Dummies

```gdscript
# demo/gameplay/DummySpawner.gd
extends Node
class_name DummySpawner

@export var spawn_interval: float = 5.0
@export var max_dummies: int = 15
@export var difficulty: String = "basic"
@export var spawn_radius: float = 100.0

var spawn_timer: float = 0.0
var spawned_dummies: Array = []

func _ready():
	add_to_group("dummy_spawners")

func _process(delta):
	spawn_timer += delta

	if spawn_timer >= spawn_interval and spawned_dummies.size() < max_dummies:
		spawn_timer = 0.0
		_spawn_dummy()

func _spawn_dummy():
	var dummy = DummyFactory.create_dummy(difficulty)
	dummy.global_position = global_position + Vector2(randf_range(-spawn_radius, spawn_radius),
													   randf_range(-spawn_radius, spawn_radius))
	get_parent().add_child(dummy)
	spawned_dummies.append(dummy)

	# Remove from list when morre
	dummy.tree_exiting.connect(func(): spawned_dummies.erase(dummy))

func change_difficulty(new_difficulty: String):
	difficulty = new_difficulty
```

---

## 6. WEAPONS SYSTEM

### 6.1 Weapon Base

```gdscript
# demo/gameplay/weapons/Weapon.gd
extends Node2D
class_name Weapon

@export var weapon_name: String = "Iron Sword"
@export var damage: float = 10.0
@export var attack_speed: float = 1.0
@export var range_: float = 50.0
@export var knockback: float = 200.0

var owner_player: Node
var cooldown: float = 0.0
var is_attacking: bool = false

func _ready():
	_create_visual()

func _create_visual():
	# Será criado em _create_weapon_visual baseado no tipo
	pass

func _process(delta):
	if cooldown > 0:
		cooldown -= delta

	# Animation update
	_update_visual(delta)

func attack(target_direction: Vector2) -> bool:
	if cooldown > 0 or is_attacking:
		return false

	is_attacking = true
	cooldown = 1.0 / attack_speed

	_perform_attack(target_direction)

	is_attacking = false
	return true

func _perform_attack(direction: Vector2):
	# Override em subclasses
	pass

func _update_visual(delta):
	# Override em subclasses
	pass
```

### 6.2 Weapons Específicas

#### Sword

```gdscript
# demo/gameplay/weapons/Sword.gd
extends Weapon

func _create_visual():
	# Blade (parte colorida)
	var blade = ColorRect.new()
	blade.size = Vector2(15, 60)
	blade.color = Color(0.8, 0.8, 0.8)  # Prata
	blade.offset = Vector2(-7.5, -30)
	add_child(blade)

	# Handle
	var handle = ColorRect.new()
	handle.size = Vector2(15, 20)
	handle.color = Color(0.4, 0.2, 0)  # Marrom
	handle.offset = Vector2(-7.5, -10)
	add_child(handle)

	# Point (ponta)
	var point = Polygon2D.new()
	point.polygon = PackedVector2Array([Vector2(0, -35), Vector2(5, -30), Vector2(-5, -30)])
	point.color = Color.WHITE
	add_child(point)

func _perform_attack(direction: Vector2):
	# Create hit area
	var hit_area = Area2D.new()
	var shape = CapsuleShape2D.new()
	shape.height = range_
	shape.radius = 15
	hit_area.add_child(CollisionShape2D.new())
	hit_area.get_child(0).shape = shape

	# Position in front
	hit_area.global_position = owner_player.global_position + direction * range_ / 2

	# Detect hits
	var hit_bodies = hit_area.get_overlapping_areas()
	for body in hit_bodies:
		var target = body.get_parent()
		if target != owner_player and not (target in _get_hit_this_frame()):
			_apply_hit(target, direction)

	hit_area.queue_free()

	# Sound
	if AudioManager:
		AudioManager.play_sound("weapon_slash")

func _apply_hit(target: Node, direction: Vector2):
	var calc = DamageCalculator.new()
	var damage_result = calc.calculate_damage(owner_player, target, "sword_attack")

	if target.has_method("take_damage"):
		target.take_damage(damage_result.final_damage)

	# Knockback
	if target.has_method("apply_knockback"):
		target.apply_knockback(global_position, knockback)
```

#### Staff

```gdscript
# demo/gameplay/weapons/Staff.gd
extends Weapon

var projectile_scene: PackedScene

func _ready():
	super._ready()
	projectile_scene = preload("res://demo/gameplay/projectiles/Projectile.tscn")

func _create_visual():
	# Shaft (cabo)
	var shaft = ColorRect.new()
	shaft.size = Vector2(8, 80)
	shaft.color = Color(0.6, 0.3, 0.1)  # Madeira
	shaft.offset = Vector2(-4, -40)
	add_child(shaft)

	# Orb (esfera de topo)
	var orb = ColorRect.new()
	orb.size = Vector2(25, 25)
	orb.color = Color.PURPLE
	orb.offset = Vector2(-12.5, -50)
	add_child(orb)

func _perform_attack(direction: Vector2):
	# Create projectile
	var projectile = projectile_scene.instantiate()
	projectile.global_position = owner_player.global_position
	projectile.direction = direction
	projectile.damage = damage
	projectile.owner_id = owner_player.network_id

	get_parent().add_child(projectile)

	if AudioManager:
		AudioManager.play_sound("staff_cast")
```

#### Axe

```gdscript
# demo/gameplay/weapons/Axe.gd
extends Weapon

func _create_visual():
	# Head (cabeça)
	var head = Polygon2D.new()
	head.polygon = PackedVector2Array([
		Vector2(-20, -40),
		Vector2(20, -40),
		Vector2(25, -30),
		Vector2(20, -20),
		Vector2(-20, -20),
		Vector2(-25, -30)
	])
	head.color = Color(0.8, 0.8, 0.8)
	add_child(head)

	# Handle
	var handle = ColorRect.new()
	handle.size = Vector2(6, 50)
	handle.color = Color(0.4, 0.2, 0)
	handle.offset = Vector2(-3, -20)
	add_child(handle)

func _perform_attack(direction: Vector2):
	# AOE circular
	var area = Area2D.new()
	var circle = CircleShape2D.new()
	circle.radius = range_
	var collision = CollisionShape2D.new()
	collision.shape = circle
	area.add_child(collision)

	area.global_position = owner_player.global_position

	var targets = area.get_overlapping_areas()
	for target_area in targets:
		var target = target_area.get_parent()
		if target != owner_player:
			var calc = DamageCalculator.new()
			var result = calc.calculate_damage(owner_player, target, "axe_swing")
			if target.has_method("take_damage"):
				target.take_damage(result.final_damage)

	area.queue_free()

	if AudioManager:
		AudioManager.play_sound("weapon_heavy_swing")
```

### 6.3 Weapon Manager

```gdscript
# demo/gameplay/WeaponManager.gd
extends Node

var weapon_templates: Dictionary = {
	"sword": preload("res://demo/gameplay/weapons/Sword.tscn"),
	"staff": preload("res://demo/gameplay/weapons/Staff.tscn"),
	"axe": preload("res://demo/gameplay/weapons/Axe.tscn"),
	"dagger": preload("res://demo/gameplay/weapons/Dagger.tscn"),
}

func create_weapon(weapon_type: String) -> Weapon:
	if weapon_type not in weapon_templates:
		weapon_type = "sword"

	var weapon = weapon_templates[weapon_type].instantiate()
	return weapon

func get_weapon_stats(weapon_type: String) -> Dictionary:
	var stats = {
		"sword": {"damage": 10, "speed": 1.0, "range": 60},
		"staff": {"damage": 8, "speed": 1.5, "range": 300},
		"axe": {"damage": 15, "speed": 0.7, "range": 80},
		"dagger": {"damage": 7, "speed": 2.0, "range": 40},
	}
	return stats.get(weapon_type, stats["sword"])
```

---

## 7. AUDIO SYSTEM

### 7.1 Integração com Procedural Audio

```gdscript
# demo/autoload/AudioManager.gd
extends Node

var sounds = {}
var music: Node

func _ready():
	add_to_group("audio_manager")
	_init_sounds()
	_init_music()

func _init_music():
	music = preload("res://demo/audio/ProceduralMusic.gd").new()
	add_child(music)

func _init_sounds():
	# Player sounds
	sounds["player_shoot"] = _create_sound(ProceduralSound.WaveType.SQUARE, 1200, 0.1, 0.35, -5000, 0.5)
	sounds["player_hit"] = _create_sound(ProceduralSound.WaveType.SQUARE, 55, 0.2, 0.6, -20, 0.1)
	sounds["level_up"] = _create_sound(ProceduralSound.WaveType.SINE, 110, 0.6, 0.6, 440, 0.05)

	# Weapon sounds
	sounds["weapon_slash"] = _create_sound(ProceduralSound.WaveType.SAW, 200, 0.15, 0.4, -1000, 0.3)
	sounds["weapon_heavy_swing"] = _create_sound(ProceduralSound.WaveType.SQUARE, 150, 0.2, 0.5, -800, 0.2)
	sounds["staff_cast"] = _create_sound(ProceduralSound.WaveType.SINE, 800, 0.3, 0.3, 2000, 0.1)

	# Impact sounds
	sounds["hit_impact"] = _create_sound(ProceduralSound.WaveType.NOISE, 60, 0.25, 0.4, 0, 0.8)
	sounds["critical_hit"] = _create_sound(ProceduralSound.WaveType.SQUARE, 1000, 0.2, 0.5, -3000, 0.4)

	# Environmental
	sounds["skill_unlock"] = _create_sound(ProceduralSound.WaveType.SINE, 440, 0.5, 0.4, 880, 0)
	sounds["enemy_killed"] = _create_sound(ProceduralSound.WaveType.NOISE, 60, 0.5, 0.6, 0, 0.9)

func _create_sound(type, freq, dur, vol, bend, noise = 0.0) -> ProceduralSound:
	var s = ProceduralSound.new()
	s.wave_type = type
	s.frequency = freq
	s.duration = dur
	s.volume = vol
	s.pitch_bend = bend
	s.noise_mix = noise
	return s

func play_sound(sound_name: String):
	if sound_name in sounds:
		sounds[sound_name].play(self)

func play_sound_with_variation(base_sound: String, pitch_variation: float = 0.1):
	if base_sound not in sounds:
		return

	var sound = sounds[base_sound].duplicate()
	sound.frequency *= randf_range(1.0 - pitch_variation, 1.0 + pitch_variation)
	sound.play(self)
```

### 7.2 Cues de Audio para Ability System

```gdscript
# demo/resources/audio/AbilityAudioCue.tres (como Resource)
# Será automaticamente carregado pelo Ability System

# Quando uma ability é ativada, o Ability System procura por um AudioCue
# Exemplo: ability "basic_attack" → toca "weapon_slash" + variação
```

---

## 8. ENEMY AI

### 8.1 Enemy Base

```gdscript
# demo/gameplay/enemies/Enemy.gd
extends CharacterBody2D
class_name Enemy

@export var level: int = 1
@export var health: float = 50.0
@export var max_health: float = 50.0
@export var damage: float = 5.0
@export var experience_reward: float = 25.0

var as_component: Node
var ai_state: String = "patrol"
var target: Node = null
var state_timer: float = 0.0

func _ready():
	as_component = preload("res://demo/gameplay/ASComponentEnemy.tscn").instantiate()
	add_child(as_component)

	_create_visual()
	_init_ai()

func _create_visual():
	# Enemy as colored shape
	var body = ColorRect.new()
	body.size = Vector2(40, 40)
	body.color = Color.RED
	body.position = -body.size / 2
	add_child(body)

func _process(delta):
	_update_ai(delta)

func _update_ai(delta):
	match ai_state:
		"patrol":
			_patrol(delta)
		"chase":
			_chase(delta)
		"attack":
			_perform_attack(delta)

func _patrol(delta):
	# Wander around
	if state_timer <= 0:
		state_timer = randf_range(2, 5)
		velocity = Vector2(randf_range(-1, 1), randf_range(-1, 1)).normalized() * 80

	state_timer -= delta
	move_and_slide()

	# Check for targets
	_find_target()
	if target and global_position.distance_to(target.global_position) < 300:
		ai_state = "chase"

func _chase(delta):
	if not target:
		ai_state = "patrol"
		return

	var direction = (target.global_position - global_position).normalized()
	velocity = direction * 120
	move_and_slide()

	if global_position.distance_to(target.global_position) < 60:
		ai_state = "attack"
	elif global_position.distance_to(target.global_position) > 400:
		ai_state = "patrol"
		target = null

func _perform_attack(delta):
	if not target or global_position.distance_to(target.global_position) > 80:
		ai_state = "chase"
		return

	# Attack every 2 seconds
	state_timer -= delta
	if state_timer <= 0:
		state_timer = 2.0
		_attack_target()

func _attack_target():
	if target and target.has_method("take_damage"):
		target.take_damage(damage)
		if AudioManager:
			AudioManager.play_sound("hit_impact")

func _find_target():
	var players = get_tree().get_nodes_in_group("Player")
	if players.is_empty():
		return

	target = players[0]
	for player in players:
		if global_position.distance_to(player.global_position) < global_position.distance_to(target.global_position):
			target = player

func take_damage(amount: float):
	health -= amount
	if as_component:
		as_component.take_damage(amount)

	if health <= 0:
		_die()

func _die():
	# Grant XP
	if target and target.has_method("add_xp"):
		target.add_xp(experience_reward)

	if AudioManager:
		AudioManager.play_sound("enemy_killed")

	queue_free()
```

### 8.2 Enemy Spawner

```gdscript
# demo/gameplay/EnemySpawner.gd
extends Node
class_name EnemySpawner

@export var spawn_interval: float = 3.0
@export var max_enemies: int = 10
@export var enemy_level: int = 1

var spawn_timer: float = 0.0
var spawned_enemies: Array = []

func _process(delta):
	spawn_timer += delta

	if spawn_timer >= spawn_interval and spawned_enemies.size() < max_enemies:
		spawn_timer = 0.0
		_spawn_enemy()

func _spawn_enemy():
	var enemy = preload("res://demo/gameplay/enemies/Enemy.tscn").instantiate()
	enemy.level = enemy_level
	enemy.global_position = global_position + Vector2(randf_range(-100, 100), randf_range(-100, 100))
	get_parent().add_child(enemy)
	spawned_enemies.append(enemy)

	enemy.tree_exiting.connect(func(): spawned_enemies.erase(enemy))
```

---

## 9. HUD & UI

### 9.1 Main HUD

```gdscript
# demo/ui/GameHUD.gd
extends CanvasLayer

@onready var health_bar = $VBoxContainer/HealthBar
@onready var mana_bar = $VBoxContainer/ManaBar
@onready var level_label = $VBoxContainer/LevelLabel
@onready var xp_label = $VBoxContainer/XPLabel
@onready var skills_display = $VBoxContainer/SkillsDisplay

var player: Node

func _ready():
	player = get_tree().get_first_node_in_group("Player")
	if player:
		if player.has_node("ASComponent"):
			var as_comp = player.get_node("ASComponent")
			as_comp.tag_changed.connect(_on_tag_changed)

func _process(delta):
	if not player:
		return

	_update_health()
	_update_mana()
	_update_level()
	_update_xp()

func _update_health():
	if player.has_node("ASComponent"):
		var as_comp = player.get_node("ASComponent")
		var health = as_comp.get_attribute("health", 100)
		var max_health = as_comp.get_attribute("max_health", 100)

		health_bar.value = (health / max_health) * 100

func _update_mana():
	if player.has_node("ASComponent"):
		var as_comp = player.get_node("ASComponent")
		var mana = as_comp.get_attribute("mana", 50)
		var max_mana = as_comp.get_attribute("max_mana", 50)

		mana_bar.value = (mana / max_mana) * 100

func _update_level():
	var progress = ProgressionSystem._get_progress(player.network_id)
	level_label.text = "Level: %d" % progress.level

func _update_xp():
	var progress = ProgressionSystem._get_progress(player.network_id)
	xp_label.text = "XP: %.0f / %.0f" % [progress.experience, progress.experience_to_next]

func _on_tag_changed(tag: String, added: bool):
	if added and tag.starts_with("motion."):
		skills_display.text = "State: %s" % tag.trim_prefix("motion.")
```

### 9.2 Combat Feedback

```gdscript
# demo/ui/CombatFeedback.gd
extends Control

var damage_numbers: Dictionary = {}

func show_damage(position: Vector2, amount: float, is_critical: bool = false):
	var label = Label.new()
	label.text = "%.0f" % amount
	label.add_theme_font_size_override("font_size", 24 if is_critical else 18)
	label.modulate = Color.RED if is_critical else Color.WHITE
	label.global_position = position
	add_child(label)

	# Animation
	var tween = create_tween()
	tween.tween_property(label, "global_position", position + Vector2(0, -50), 1.0)
	tween.tween_property(label, "modulate:a", 0.0, 0.5)
	tween.tween_callback(label.queue_free)
```

---

## 10. INTEGRAÇÃO COM ABILITY SYSTEM

### 10.1 ASComponent para Players

```gdscript
# demo/gameplay/PlayerASComponent.gd
extends "res://addons/ability_system/ASComponent.gd"

func _ready():
	super._ready()

	# Initialize attributes
	add_attribute("health", 100)
	add_attribute("max_health", 100)
	add_attribute("mana", 50)
	add_attribute("max_mana", 50)
	add_attribute("attack_power", 10)
	add_attribute("defense", 5)
	add_attribute("movement_speed", 200)

	# Setup initial abilities
	_setup_abilities()

func _setup_abilities():
	# Basic ability
	add_ability("basic_attack", _create_ability("basic_attack", 0.5))
	add_ability("power_slash", _create_ability("power_slash", 1.0))
	add_ability("dash", _create_ability("dash", 0.3))

func _create_ability(ability_id: String, cooldown: float) -> Dictionary:
	return {
		"id": ability_id,
		"cooldown": cooldown,
		"damage_multiplier": 1.0,
		"tags": [ability_id],
	}

func perform_ability(ability_id: String) -> bool:
	if not has_ability(ability_id):
		return false

	# Check cooldown via ability system
	if not is_ability_ready(ability_id):
		return false

	# Add tag to activate ability state
	add_tag("ability.%s" % ability_id)

	# Trigger damage
	var damage_calc = DamageCalculator.new()
	var result = damage_calc.calculate_damage(get_parent(), get_parent(), ability_id)

	# Play sound
	AudioManager.play_sound("weapon_slash")

	# Start cooldown
	set_ability_cooldown(ability_id, 0.0)

	return true

func take_damage(amount: float):
	var current_health = get_attribute("health", 100)
	set_attribute("health", max(0, current_health - amount))

	if current_health <= 0:
		add_tag("status.dead")
```

### 10.2 Cues de Audio para Abilities

```gdscript
# Cada ability pode ter seu próprio AudioCue
# O Ability System chamará o AudioCue quando a ability for ativada

# Exemplo: res://demo/resources/abilities/basic_attack.tres
# [sub_resource type="AudioStreamWAV"]
# ... (procedurally generated)
```

---

## 11. SISTEMA DE CRAFTING (Opcional)

```gdscript
# demo/gameplay/CraftingSystem.gd
extends Node

class Recipe:
	var id: String
	var name: String
	var ingredients: Dictionary  # item_id: quantity
	var result: String
	var result_quantity: int = 1

var recipes: Dictionary = {
	"iron_sword": Recipe.new(),
	"health_potion": Recipe.new(),
}

func craft(recipe_id: String, player_id: int) -> bool:
	if recipe_id not in recipes:
		return false

	var recipe = recipes[recipe_id]
	var progress = ProgressionSystem._get_progress(player_id)

	# Check ingredients
	for item_id in recipe.ingredients.keys():
		var needed = recipe.ingredients[item_id]
		var has = progress.inventory.get(item_id, 0)
		if has < needed:
			return false

	# Consume ingredients
	for item_id in recipe.ingredients.keys():
		progress.inventory[item_id] -= recipe.ingredients[item_id]

	# Add result
	progress.inventory[recipe.result] = progress.inventory.get(recipe.result, 0) + recipe.result_quantity

	return true
```

---

## 12. MAPA DE IMPLEMENTAÇÃO

### Fase 1: Core Systems (Semana 1)
- [ ] Level design base (6 zonas)
- [ ] Dummy system com spawners
- [ ] Player ASComponent integrada
- [ ] Combat básico (hit detection)
- [ ] Audio procedural integrado

### Fase 2: Gameplay (Semana 2)
- [ ] Progression system
- [ ] Skill tree (todas as 15 skills)
- [ ] Weapons system (4 armas principais)
- [ ] Enemy AI
- [ ] HUD

### Fase 3: Polish (Semana 3)
- [ ] Status effects completos
- [ ] Crafting system
- [ ] Boss arena
- [ ] Refinements e balancing
- [ ] Documentação

---

## 13. CHECKLIST DE FEATURES

### Core Gameplay
- [ ] Players podem se mover livremente no mapa
- [ ] Dummies spawnam automaticamente
- [ ] Players conseguem atacar dummies
- [ ] Damage é calculado corretamente
- [ ] Players ganham XP ao matar inimigos
- [ ] Level up funciona e aumenta stats

### Combat
- [ ] 4 weapons implementadas com visuais
- [ ] Hit detection com área
- [ ] Critical hits (10% chance)
- [ ] Knockback aplicado corretamente
- [ ] Status effects causam efeitos visuais

### Progression
- [ ] XP system funciona
- [ ] Level up aumenta atributos
- [ ] Skill points são ganhos
- [ ] Skill tree pode ser acessado

### Skills
- [ ] 15 skills diferentes implementadas
- [ ] Árvore de requerimentos funciona
- [ ] Skills podem ser desbloqueadas
- [ ] Skills causam damage diferentes
- [ ] Cooldowns funcionam

### Audio
- [ ] Música procedural toca em loop
- [ ] Sounds de ataque tocam
- [ ] Sons de XP/level up tocam
- [ ] Efeitos sonoros variam

### Multiplayer
- [ ] Multiple players podem jogar juntos
- [ ] Danos são sincronizados via RPC
- [ ] Kills são creditadas corretamente
- [ ] Chat mostra eventos

---

## 14. NOTAS DE DESIGN

### Por que Procedural?
- Sem assets visuais = fácil de iterar
- Tudo é código = fácil de entender
- Demonstra criatividade do framework
- Escalável e customizável

### Filosofia
- **Simplicidade**: Cada sistema faz uma coisa bem
- **Modularidade**: Podem ser testados independentemente
- **Emergência**: Combate é emergente do Ability System
- **Transparência**: Tudo é visível no código

### Não é Realista?
Correto! É uma **demonstração**. O objetivo é:
1. Mostrar poder do Ability System
2. Demonstrar multiplayer com RBAC
3. Servir como referência para desenvolvimento
4. Ser divertido (em um nível abstrato)

---

## 15. PRÓXIMOS PASSOS

1. **Começar implementação** dos systems em ordem de dependência
2. **Testar cada sistema** isoladamente
3. **Integrar com Ability System**
4. **Testar multiplayer**
5. **Balancear dificuldade**
6. **Polir e documentar**

---

**Versão do Documento**: 1.0
**Data**: 2026-04-05
**Status**: Pronto para Implementação
