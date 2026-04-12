<p align="center">
  <img src="doc/images/logo.svg" width="600" alt="Ability System Logo">
</p>
<br/>

# Ability System - Data-Driven Abilities for Godot 4.x

[![Godot Engine](https://img.shields.io/badge/Godot-4.6-blue?logo=godot-engine&logoColor=white)](https://godotengine.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> [!TIP]
> **Leia isto em outros idiomas / Read this in other languages:**
> [**Português**](README.pt.md) | [**English**](README.md)

**Ability System** (AS) is a powerful framework for creating modular combat, abilities, and attributes. Designed to scale from simple mechanics to complex RPG systems — all with high C++ performance and modular architecture.

---

## 📦 Installation

1. Download the latest `ability-system-plugin.zip` from [Releases](https://github.com/MachiTwo/AbilitySystem/releases/download/v0.1.0/ability-system+v0.1.0-dev.gdextension.zip).
2. Extract and copy the `addons/ability_system` folder to your project's `addons/` directory.
3. Restart Godot and go to **Project > Project Settings > Plugins** to enable the "Ability System" plugin. This activates the Tag Editor and custom Inspector features.

---

## 🧬 High Scalability Systems

This framework introduces advanced concepts to decouple game logic:

- **Ability Phases**: Divide your abilities into stages like `Windup`, `Execution`, and `Recovery`. Manage durations and specific effects for each phase natively.
- **AS Events**: Transient triggers that carry data (`ASEventTag`). Perfect for communicating impacts, UI interactions, or animation triggers without polluting the character's state.
- **Events Historical**: Short-term memory for events that occurred recently, allowing abilities and effects to query the immediate past (e.g., "If blocked an attack in the last 0.5s").

---

## 🛠️ Dual Build Architecture

This project is uniquely engineered to support **Dual Compilation**, serving both the core engine development and the plugin ecosystem:

- **Zyris Engine (Native):** This framework is a core native component of the **Zyris Engine**. Development in this repository allows for centralized automation and rigorous validation. Once stabilized, the code is officially integrated into the Zyris Engine's `master` branch.
- **GDExtension (Plugin):** A dynamic library version for standard Godot 4.x projects. It provides 100% logic parity without requiring engine recompilation, ideal for projects using the official Godot distribution.

A unified C++ codebase powers both versions, using a robust preprocessor system (`#ifdef ABILITY_SYSTEM_GDEXTENSION`) to manage integration differences while maintaining a single source of truth for the logic.

---

## 🏗️ Quick Start Guide

Transform your game logic into data with these core steps. We will build a system where social interactions and combat coexist.

### 1. Access the Tag Manager

The Tag Manager is the heart of your project's vocabulary. Its location depends on your version:

- **GDExtension (Plugin):** Look for the **Ability System Tags** tab in the **Bottom Panel** (next to Output/Debugger).
- **Zyris Engine (Native):** Go to **Project > Project Settings** and look for the **Ability System Tags** tab after the Input Map.

### 2. Define the Vocabulary (Tags)

Define hierarchical identifiers. Dots create branches that the logic can query:

- `ability.social.talk`: The base for all conversations.
- `state.emotional.angry`: A state that might block social interactions.
- `state.stun.frozen`: A combat state.

> [!TIP]
> **Broad Logic:** Checking for `state.emotional` will return true if the actor has `state.emotional.angry` or `state.emotional.happy`.

### 3. Create the Stat Schema (AttributeSet)

Create an **AttributeSet** resource (e.g., `RPGAttributes.tres`).

- Add combat stats: `Health`, `Mana`.
- Add social stats: `Charisma`, `Patience`.
- Set Min (0), Max (100), and Base (100) values.
- The `ASComponent` (ASC) will **deep-duplicate** this set on spawn, ensuring unique health/patience pools for every NPC.

### 4. Design Social Abilities (Dialogues)

Create an **Ability** (e.g., `TalkAbility.tres`):

- **Ability Tag:** `ability.social.talk`.
- **Activation Blocked Tags:** Add `state.emotional.angry`. An angry character cannot start a polite conversation!
- **Owned Tags:** Add `state.busy.talking`. This tag is granted _while_ the ability is active and can be used to block other actions like "Run".

### 5. Create Emotional Effects

Create an **Effect** (e.g., `EffectAngry.tres`):

- **Duration Policy:** `Infinite` (Passive) or `Duration` (Temporary).
- **Modifiers:** Add a modifier to `Charisma` (Multiply by 0.5) and `Attack` (Add 20).
- **Granted Tags:** `state.emotional.angry`.

### 6. Build the Archetype (ASContainer)

Create an **ASContainer** (e.g., `VillagerArchetype.tres`). This is your "NPC Blueprint":

- Assign the `AttributeSet`.
- Add the `Talk` ability to the catalog.
- Add default effects (e.g., a "Neutral" state).

### 7. Initialize the Actor (ASC)

Add the `ASComponent` node to your CharacterBody and assign the **ASContainer** in the Inspector.
Register your feedback nodes (AnimationPlayer/Audio) in script so the system can trigger **Cues** automatically.

### 8. Combat vs Social Interaction

The system handles both seamlessly. Applying a "Frozen" effect will automatically block the "Talk" ability if you add `state.stun` to the ability's blocked tags list.

```gdscript

func interact_with_npc(npc: ASComponent):
 if npc.can_activate_ability_by_tag(&"ability.social.talk"):
  npc.try_activate_ability_by_tag(&"ability.social.talk")
 else:
  print("The NPC is too angry or stunned to talk!")
```

### 9. Handling State Changes (Signals)

The ASC notifies your game logic when significant changes occur:

```gdscript

func _ready():
 asc.tag_changed.connect(_on_tag_changed)

func _on_tag_changed(tag: StringName, is_present: bool):
 if tag == &"state.emotional.angry" and is_present:
  $Sprite.modulate = Color.RED # Visual feedback for emotion change
```

### 10. The Payload: ASPackage

The `ASDelivery` node is the bridge between your abilities and their impact. It needs a payload. Create an **ASPackage** resource (e.g., `ExplosionPayload.tres`).

- Add **Effects**: Instant damage, burning over time, etc.
- Add **Cues**: Sound of an explosion, particle effects.

### 11. Combat Deliveries (3 Ways)

#### A. Melee Attacks (Area2D)

Typically an Area2D attached to the player that is only active during hit frames.

1. Add an `Area2D` to your Player (e.g., `SwordArea`).
2. Add an `ASDelivery` as a child.
3. **Set**: `Package` = `MeleeDamage.tres`, `Auto Connect` = **On**.
4. **Trigger**:

   ```gdscript
   func _on_attack_animation_hit_frame():
       $SwordArea/ASDelivery.activate(0.1) # Active for 100ms
   ```

#### B. Projectiles (Bullets)

A projectile carries intent across the screen.

1. Create a Bullet scene (`Area2D` + `ASDelivery`).
2. **Set**: `Auto Connect` = **On**, `One Shot` = **On**.
3. **Spawn**:

   ```gdscript
   func fire():
       var bullet = BulletScene.instantiate()
       get_parent().add_child(bullet)
       # Tell the delivery who sent it for attribute logic (Strength/Atk)
       bullet.get_node("ASDelivery").set_source_component(self.asc)
       bullet.get_node("ASDelivery").activate()
   ```

#### C. Traps and Hazards

Passive deliveries waiting for a victim.

1. Create a SpikeTrap scene (`Area2D` + `ASDelivery`).
2. **Set**: `Auto Connect` = **On**, `Active` = **On**, `Life Span` = **0**.
3. **Result**: No script needed! Any `ASComponent` entering the area receives the package.

---

## 🌐 Multiplayer Guide (Prediction & Rollback)

The Ability System is engineered for high-performance networking using a **State-Based Rollback** architecture. Follow these steps to implement lag-free abilities.

### 1. Enable Snapshots

For playable characters, assign an `ASStateSnapshot` resource to the `snapshot_state` property of the `ASComponent`. This acts as a persistent bridge for the internal high-speed `ASStateCache`.

### 2. The Prediction Loop (Client)

Clients should predict ability activation to eliminate perceived latency. Before executing a predictive action, capture the current state.

```gdscript
func _physics_process(delta):
    if is_local_authority():
        # 1. Capture current tick state before prediction
        asc.capture_snapshot()

        # 2. Predictively activate locally
        if Input.is_action_just_pressed("fire"):
            asc.try_activate_ability_by_tag(&"ability.fireball")

            # 3. Inform the server of your intent
            asc.request_activate_ability.rpc_id(1, &"ability.fireball")
```

### 3. Server Validation & Confirmation

The server acts as the authority. It receives the request, validates costs/requirements, and notifies the client to confirm or correct the state.

```gdscript
@rpc("any_peer", "call_local")
func request_activate_ability(tag: StringName):
    if is_server():
        if asc.can_activate_ability_by_tag(tag):
            asc.try_activate_ability_by_tag(tag)
            # Confirm to the client that the prediction was correct
            asc.confirm_ability_activation.rpc(tag)
        else:
            # If invalid, the server's state naturally overrides the client
            # during the next full state synchronization.
            pass
```

### 4. Handling Rollback (Automatic)

If a prediction error occurs (e.g., the server says you were stunned when you tried to fire), use `apply_snapshot(tick)` to revert the actor's attributes, tags, and effects to a previous valid state.

- **ASEffectState:** Internally tracks the duration and stacks of every effect.
- **ASStateCache:** Automatically keeps a circular buffer of the last 128 ticks for near-instant rollback without memory allocation overhead.

### 5. Deterministic Effects

To ensure the server and client calculate damage identically:

- Use **Attributes** for all gameplay math.
- Avoid using `randf()` directly in ability logic; use a seeded random system or let the server dictate random outcomes via `ASPackage` levels.

---

## 📚 Reference

### 🧠 Core Components

| Component         | Purpose                  | Key Features                                                 |
| :---------------- | :----------------------- | :----------------------------------------------------------- |
| **AbilitySystem** | Global Coordinator.      | Central Tag Registry, Project Settings integration.          |
| **ASComponent**   | Actor's Processor (ASC). | Grants/Activates abilities, manages effects, triggers Cues.  |
| **ASDelivery**    | Physics Injector.        | Propagates payloads (ASPackage) via Area2D/Area3D naturally. |

### ⚙️ Resources (Blueprints)

| Resource            | Purpose               | Key Features                                                  |
| :------------------ | :-------------------- | :------------------------------------------------------------ |
| **ASAbility**       | Logic of an action.   | Costs, Cooldowns and Activation Tags. Supports **Phases**.    |
| **ASAbilityPhase**  | Phase Definition.     | Defines duration and effects for specific ability stages.     |
| **ASEffect**        | Modification package. | Instant damage, timed buffs, or infinite passives.            |
| **ASAttributeSet**  | Stat container.       | Manages collections of attributes. Deep-duplicated per actor. |
| **ASAttribute**     | Stat definition.      | Individual HP, Mana, XP schema with clamping.                 |
| **ASContainer**     | Archetype Blueprint.  | Catalog of allowed abilities, effects, and starting stats.    |
| **ASPackage**       | Data Payload.         | Grouped collection of effects and cues for delivery.          |
| **ASCue**           | Feedback Base.        | Event activation logic for audiovisual effects.               |
| **ASCueAnimation**  | Animation Feedback.   | Specialized for playing montages on actors.                   |
| **ASCueAudio**      | Audio Feedback.       | Specialized for playing spatial or global sound.              |
| **ASStateSnapshot** | State Persistence.    | Captures Attributes/Tags for Multiplayer Rollback or Saving.  |
| **ASEventTag**      | Event Definition.     | Defines custom events for ability system interactions.        |

### 🚀 Runtime Objects (Specs)

| Spec              | Purpose           | Key Data                                                  |
| :---------------- | :---------------- | :-------------------------------------------------------- |
| **ASAbilitySpec** | Ability instance. | Level, active state, runtime query methods.               |
| **ASEffectSpec**  | Effect instance.  | Remaining duration, stack count, hit position (Variant).  |
| **ASCueSpec**     | Feedback context. | Triggering magnitude, source/target references, hit data. |
| **ASTagSpec**     | Tag Management.   | High-performance container for actor state tags.          |

### 🛠️ API Patterns and Safe Access

The Ability System interface prioritizes safety and predictability through strict naming conventions:

- **Dual Instantiation (`by_tag` vs `by_resource`)**: Every method features two signatures. Use `by_tag` for dynamic actions or `by_resource` for hard references.
- **Safe Execution (`try_activate`)**: Mandatory use of `try_...` ensures that Blueprint costs and requirements are respected before execution.
- **Authorization (`can_`)**: Evaluates whether an action is theoretically permitted without starting it.
- **Status Query (`is_` / `has_`)**: Clean checks to validate Tag possession or operational states.
- **Interruption (`cancel_`)**: Actively terminates ongoing flows of abilities or effects.
- **Network Intent (`request_`)**: Formal methodology to request executions via RPC (Server Authority).
- **Safe Extraction (`get_`)**: Retrieves values and instances cleanly without exposing mutability.
- **Inventory Management (`unlock_` / `lock_`)**: Defines the active catalog of equipable abilities.

---

Developed with ❤️ by **MachiTwo**
