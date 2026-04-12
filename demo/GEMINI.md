# Godot MBA: O Grimório do Arquiteto

> **Contexto:** Este documento é a "Bíblia" técnica para o desenvolvimento do curso **Godot MBA**. Ele define os pilares da nossa arquitetura e serve como referência absoluta para Agentes de IA e Alunos.

---

## 1. Paradigmas de Desenvolvimento

Para construir sistemas complexos, precisamos entender as ferramentas conceituais à nossa disposição.

### 1.1. Orientação a Objetos (OOP)

**"O Comportamento define o Objeto."**

Na OOP clássica, focamos em **Classes** que encapsulam dados e métodos.

- **Herança:** "Um `Guerreiro` **é um** `Personagem`".
- **Encapsulamento:** O `Guerreiro` protege sua vida (`private health`) e oferece métodos (`take_damage()`).
- **Polimorfismo:** Tratamos `Guerreiro` e `Mago` como `Personagem`, chamando `atacar()` sem saber qual classe específica é.

**Na Godot:**

- Todo Node (`CharacterBody2D`, `Sprite2D`) é um Objeto.
- Scripts `.gd` são Classes.
- Usamos OOP para a **Lógica de Controle** (Controllers, Managers, Machines).

### 1.2. Orientação a Dados (DOP)

**"A Memória define a Performance."**

Na DOP, focamos em como os dados são organizados na memória para otimizar o acesso da CPU (Cache Locality). Em vez de `Array[Objeto]`, preferimos `Array[Int]`, `Array[Vector2]`.

- **ECS (Entity Component System):** Separação total entre Dados (Componentes) e Lógica (Sistemas).

**Na Godot:**

- Usamos `Servers` (`PhysicsServer2D`, `RenderingServer`) para alta performance.
- Usamos `PackedFloat32Array` para grandes volumes de dados numéricos.

### 1.3. Orientação a Resources (ROP)

**"A Kaffyn Way: Onde Dados viram Comportamento."**

O **Resource-Oriented Programming (ROP)** é o "pulo do gato" da Godot. É um híbrido poderoso entre OOP e DOP.

- **Definição:** Resources são objetos de dados serializáveis (`.tres`) que podem conter lógica pura (helper functions).
- **Compartilhamento:** Se 1000 Goblins usam o mesmo `goblin_stats.tres`, eles compartilham a mesma instância na memória RAM.
- **Injeção:** Nodes (Comportamento) recebem Resources (Configuração) para saber o que fazer.

**Diferença Chave:**

- **OOP:** O `Guerreiro` tem `var damage = 10` hardcoded no script.
- **ROP:** O `Guerreiro` tem `var stats: CharacterStats`. O valor `10` vive num arquivo `.tres` que pode ser trocado em tempo real.

---

## 2. Fundamentos da Godot

### 2.1. Herança e `extends`

Em GDScript, a herança é a base da reutilização de código de _comportamento_.

```gdscript
# entity.gd
class_name Entity extends CharacterBody2D

func take_damage(amount: int):
    print("Entity took damage")

# player.gd
class_name Player extends Entity

# Sobrescreve o comportamento do pai
func take_damage(amount: int):
    super(amount) # Chama a lógica do pai
    HUD.shake_screen() # Adiciona comportamento específico
```

### 2.2. Autoloads (Singletons)

São Nodes que a Godot carrega automaticamente na raiz (`/root/`) ao iniciar o jogo. Eles persistem entre trocas de cena.

**Quando usar:**

- Gerenciadores Globais (`SoundManager`, `SaveSystem`, `GameSettings`).
- Dados de Sessão (`SessionData`).

**Quando NÃO usar:**

- Para passar dados entre Player e Inimigo (Use Sinais ou Resources).
- Para lógica de gameplay específica de uma cena.

---

## 3. O Plugin: StateEngineering

O **StateEngineering** é o framework proprietário da Kaffyn para criar Máquinas de Estado Híbridas (Nodes + Resources). Ele elimina o "Spaghetti Code" de `if/else` e `match state`.

### 3.1. Arquitetura do Plugin

A estrutura de arquivos reflete a separação de responsabilidades:

```text
addons/state_engineering/
├── core/                   # O Coração (Classes Abstratas)
│   ├── State.gd            # A "Ficha" de dados de um estado (Resource)
│   ├── Machine.gd          # O "Cérebro" que processa os estados (Node)
│   ├── Player.gd     # Controller base para Players
│   └── EnemyBase2D.gd      # Controller base para Inimigos
├── resources/              # Definições de Dados Específicos
│   ├── MoveData.gd         # Dados de Movimento (Velocidade, Fricção)
│   ├── AttackData.gd       # Dados de Ataque (Dano, Hitbox)
│   └── ItemData.gd         # Dados de Itens (Inventário)
└── machines/               # Implementações Concretas
    ├── PlayerMachine.gd    # Lógica específica do Player
    └── EnemyMachine.gd     # Lógica específica de Inimigos
```

### 3.2. Fluxo de Execução

```mermaid
graph TD
    Input[Player Input] -->|1. Contexto| Machine[PlayerMachine (Node)]
    Machine -->|2. Avalia| StateRes[Current State (Resource)]

    subgraph "Ciclo de Decisão"
        Machine -- Lê Contexto --> Context{Contexto?}
        Context -- Motion: RUN --> RunState[.tres]
        Context -- Attack: FAST --> AttackState[.tres]
    end

    Machine -->|3. Aplica| Player[Player]

    RunState -- Speed Mult --> Player
    AttackState -- Hitbox/Anim --> Player
```

### 3.3. Como Usar (Passo a Passo)

#### 1. Criar os Resources (Os Estados)

No FileSystem, crie arquivos `.tres` para cada ação possível:

- `Idle.tres` (MoveData)
- `Run.tres` (MoveData)
- `Attack_Slash.tres` (AttackData)

Exemplo de `Run.tres`:

- **Speed Multiplier:** 1.5
- **Animation:** "Run_Loop"
- **Cost Type:** Stamina (10/sec)

#### 2. Configurar a Machine (O Cérebro)

No nó `PlayerMachine` (filho do Player):

1. Arraste os `.tres` criados para as listas correspondentes (`Move States`, `Attack States`).
2. A Machine indexa esses estados automaticamente.

#### 3. O Código do Player (O Corpo)

O `Player` não sabe o que é "Correr". Ele apenas obedece à Machine.

```gdscript
# Player.gd (Simplificado)

func _physics_process(delta):
    # 1. Atualiza o Contexto para a Machine decidir
    if Input.is_action_pressed("run"):
        machine.set_context("Motion", BehaviorStates.Motion.RUN)
    else:
        machine.set_context("Motion", BehaviorStates.Motion.WALK)

    # 2. A Machine muda o estado interno e emite 'state_changed'

    # 3. O Player reage ao estado ATUAL (seja ele qual for)
    var current_state = machine.current_state_res

    # Aplica multiplicadores do estado atual
    var final_speed = base_speed * current_state.speed_multiplier
    move_and_slide()
```

### 3.4. Diferenciais do StateEngineering

1. **Duck Typing de Estados:** Um estado pode ser `MoveData` e `AttackData` ao mesmo tempo (ex: Dash Attack). O sistema verifica `if state is AttackData` e aplica a lógica.
2. **Context-Driven:** Não fazemos `machine.change_state("Run")`. Nós dizemos `machine.set_context("Motion", RUN)` e a Machine decide qual o melhor estado de "Run" para o momento (ex: se estiver machucado, usa "Run_Limping").
3. **Separação Total:** O Game Designer pode criar um combo inteiro de 10 hits apenas criando arquivos `.tres` e linkando-os via `next_combo_state`, sem pedir uma linha de código ao programador.

### 3.5. Otimização O(1) (HashMap)

Para evitar loops lineares (`O(N)`) a cada frame, o sistema utiliza **Indexação Invertida**:

1. **Build Time:** O `Compose.gd` (`@tool`) varre os estados e cria dicionários de busca (`move_rules`, `attack_rules`).
2. **Runtime:** A Machine consulta o dicionário usando uma chave numérica (`get_lookup_key()`).
    - Ex: `move_rules[Motion.RUN]` retorna `[RunState, RunLimpingState]`.
3. **Resultado:** A busca é instantânea, independente de ter 10 ou 1000 estados.

### 3.6. Sistema de Scoring

Se múltiplos estados passarem pelo filtro inicial (ex: dois ataques aéreos), o sistema usa um **Score System**:

- **Match Exato:** +1 Ponto (ex: `req_weapon: KATANA` == `Context: KATANA`).
- **Match Genérico:** 0 Pontos (ex: `req_weapon: ANY`).

O estado com maior pontuação vence. Isso permite especialização automática sem `if/else`.
