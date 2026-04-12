# StateIndex (Legacy Analysis)

> **Objetivo:** Indexação exata e detalhada de todos os componentes do plugin legado `state_engineering` para referência durante a migração.

## 1. Core (A Base)

### [StateMachine.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/core/StateMachine.gd) (Autoload)

**Caminho:** [addons/state_engineering/core/StateMachine.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/core/StateMachine.gd)
**Extends:** `Node`

#### Constantes

- `ANY = 0`

#### Enums (Vocabulário Global)

| Enum                | Valores                                                                        |
| :------------------ | :----------------------------------------------------------------------------- |
| **Motion**          | `ANY`, `IDLE`, `WALK`, `RUN`, `DASH`, `EXCEPT_DASH`                            |
| **Jump**            | `ANY`, `NONE`, `LOW`, `HIGH`, `FALL`                                           |
| **Attack**          | `ANY`, `NONE`, `FAST`, `NORMAL`, `CHARGED`, `SPECIAL`                          |
| **Physics**         | `ANY`, `GROUND`, `AIR`, `WATER`, `EXCEPT_GROUND`, `EXCEPT_AIR`, `EXCEPT_WATER` |
| **Effect**          | `ANY`, `NONE`, `FIRE`, `ICE`, `POISON`, `ELECTRIC`                             |
| **Weapon**          | `ANY`, `NONE`, `KATANA`, `BOW`, `EXCEPT_NONE`                                  |
| **Armor**           | `ANY`, `NONE`, `IRON`, `STEEL`, `GOLD`, `DIAMOND`                              |
| **Stance**          | `ANY`, `STAND`, `CROUCH`, `BLOCK`, `CLIMB`, `COVER`                            |
| **Tier**            | `ANY`, `BASE`, `UPGRADED`, `MASTER`, `CORRUPTED`                               |
| **GameState**       | `ANY`, `PLAYING`, `PAUSED`, `CUTSCENE`, `MENU`                                 |
| **StateType**       | `MOVE`, `ATTACK`, `INTERACTIVE`, `GAME`                                        |
| **Status**          | `ANY`, `NORMAL`, `STUNNED`, `INVULNERABLE`, `SUPER_ARMOR`, `DEAD`              |
| **InputSource**     | `ANY`, `PLAYER`, `AI`, `CINEMATIC`, `FORCE`                                    |
| **EnvType**         | `ANY`, `OPEN`, `TIGHT_CORRIDOR`, `LEDGE`, `WATER_SURFACE`                      |
| **Reaction**        | `IGNORE`, `CANCEL`, `ADAPT`, `FINISH`                                          |
| **CostType**        | `NONE`, `STAMINA`, `MANA`, `HEALTH`, `AMMO`                                    |
| **LowResourceRule** | `IGNORE_COMMAND`, `EXECUTE_WEAK`, `CONSUME_HEALTH`                             |
| **ComboStep**       | `NONE`, `STEP_1`, `STEP_2`, `STEP_3`, `STEP_4`, `FINISHER`                     |

#### Funções Estáticas

- `check_match(category: String, req_val: int, ctx_val: int) -> bool`: Verifica compatibilidade de contexto (suporta `ANY` e `EXCEPT_*`).

---

### [State.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/core/State.gd) (Resource)

**Caminho:** [addons/state_engineering/core/State.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/core/State.gd)
**Extends:** `Resource`
**Class Name:** `State`

#### Propriedades Exportadas

**Identity & Visuals**

- `name: String` (Default: "State")
- `texture: Texture2D`
- `hframes: int` (Default: 1)
- `vframes: int` (Default: 1)
- `animation_res: Animation`
- `loop: bool` (Default: false)
- `speed: float` (Default: 1.0)
- `sound: AudioStream`
- `debug_color: Color` (Default: White)

**Movement Logic**

- `speed_multiplier: float` (Default: 1.0)
- `duration: float` (Default: 0.0)
- `lock_movement: bool` (Default: false)
- `cancel_on_wall: bool` (Default: false)
- `ignore_gravity: bool` (Default: false)

**Physics Parameters**

- `acceleration: float` (Default: 0.0)
- `friction: float` (Default: 0.0)
- `air_resistance: float` (Default: 0.0)
- `jump_force: float` (Default: 0.0)

**Combat (Melee)**

- `damage: int` (Default: 0)
- `cooldown: float` (Default: 0.0)
- `area_pivot: Vector2` (Default: ZERO)
- `area_size: Vector2` (Default: ZERO)
- `preserve_momentum: bool` (Default: false)

**Combat (Ranged)**

- `projectile_scene: PackedScene`
- `projectile_speed: float` (Default: 0.0)
- `projectile_count: int` (Default: 0)
- `projectile_spread: float` (Default: 0.0)
- `spawn_offset: Vector2` (Default: ZERO)

**Combo System**

- `combo_step: BehaviorStates.ComboStep` (Default: NONE)
- `next_combo_state: Resource`
- `combo_window_start: float` (Default: 0.0)

**Charged Attack**

- `is_charged: bool` (Default: false)
- `min_charge_time: float` (Default: 0.0)
- `max_charge_time: float` (Default: 0.0)
- `fully_charged_damage_multiplier: float` (Default: 1.0)

**Buffs & Debuffs**

- `buffs: Array[Resource]`

**Filters (Requirements)**
> **Nota:** Agora armazenados em um `Dictionary` chamado `requirements`.

- `motion`: `BehaviorStates.Motion` (Default: ANY)
- `jump`: `BehaviorStates.Jump` (Default: ANY)
- `attack`: `BehaviorStates.Attack` (Default: ANY)
- `weapon`: `BehaviorStates.Weapon` (Default: ANY)
- `physics`: `BehaviorStates.Physics` (Default: ANY)
- `status`: `BehaviorStates.Status` (Default: ANY)
- `effect`: `BehaviorStates.Effect` (Default: ANY)

**Indexação**

- `get_lookup_key() -> int`: Retorna a chave primária para o HashMap (ex: `req_motion` para `MoveData`). Trata `EXCEPT_DASH` como `ANY` (0).

**Timing & Windows**

- `cancel_min_time: float` (Default: 0.0)
- `enable_buffering: bool` (Default: false)
- `buffer_window_start: float` (Default: 0.0)

**Costs**

- `cost_type: CostType` (Default: NONE)
- `cost_amount: int` (Default: 0)
- `on_insufficient_resource: LowResourceRule` (Default: IGNORE_COMMAND)

**Reaction Rules**

- `on_physics_change: BehaviorStates.Reaction` (Default: ADAPT)
- `on_weapon_change: BehaviorStates.Reaction` (Default: IGNORE)
- `on_motion_change: BehaviorStates.Reaction` (Default: IGNORE)
- `on_take_damage: BehaviorStates.Reaction` (Default: IGNORE)

**Context Cooldowns**

- `context_cooldown_filter: BehaviorStates.ContextFilter` (Default: NONE)
- `context_cooldown_time: float` (Default: 0.0)

---

### [ItemData.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/resources/ItemData.gd) (Resource)

**Caminho:** [addons/state_engineering/resources/ItemData.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/resources/ItemData.gd)
**Extends:** `Resource`
**Class Name:** `ItemData`

#### Propriedades

- `display_name: String` (Default: "Item")
- `icon: Texture2D`
- `context_modifiers: Dictionary` (Default: `{"Weapon": 0}`)

#### Funções

- `get_context_modifier(category: String) -> int`: Retorna valor do modificador ou -1.

---

### [MoveData.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/resources/MoveData.gd) (Resource)

**Caminho:** [addons/state_engineering/resources/MoveData.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/resources/MoveData.gd)
**Extends:** `State`
**Class Name:** `MoveData`
_Herda tudo de State. Usado para diferenciação de tipo._

---

### [AttackData.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/resources/AttackData.gd) (Resource)

**Caminho:** [addons/state_engineering/resources/AttackData.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/resources/AttackData.gd)
**Extends:** `State`
**Class Name:** `AttackData`
_Herda tudo de State. Usado para diferenciação de tipo._

---

### [Player.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/core/Player.gd) (Node)

**Caminho:** [addons/state_engineering/core/Player.gd](file:///c:/Users/bruno/Desktop/Games/002/addons/state_engineering/core/Player.gd)
**Extends:** `CharacterBody2D`
**Class Name:** `Player`

#### Dependências Exportadas

- `machine: PlayerMachine`
- `stats: CharacterStats`
- `hotbar: Hotbar`

#### Propriedades de Game Feel

- `coyote_time: float` (Default: 0.15)
- `jump_buffer_time: float` (Default: 0.1)

#### Variáveis de Estado

- `current_stamina: int`
- `current_health: int`
- `is_stunned: bool`
- `facing_direction: float`

#### Sinais

- `died`

#### Funções Principais

- `_physics_process`: Gerencia gravidade, movimento, input e atualiza contexto da Machine.
- `_handle_gravity`: Aplica gravidade (considera `ignore_gravity` do estado).
- `_handle_movement`: Aplica velocidade baseada em `MoveData` (speed_multiplier, acceleration, friction) ou defaults.
- `_update_machine_context`: Traduz inputs/velocidade para `Motion`, `Physics`, `Jump`, `Status`.
- `_unhandled_input`: Gerencia troca de arma (Hotbar keys) e input de ataque (Light/Heavy).
- `_on_state_changed`: Aplica custos, força de pulo, dash instantâneo, animação e cor de debug.
- `_on_attack_executed`: Posiciona Hitbox ou spawna Projétil.
- `consume_stamina`: Reduz stamina.
- `take_damage`: Reduz vida e chama `die()`.
