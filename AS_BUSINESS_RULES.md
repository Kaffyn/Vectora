# BUSINESS RULES: ABILITY SYSTEM - GOVERNANCE CONTRACT

> [!TIP]
> **Read this in other languages / Leia isto em outros idiomas:**
> [**English**](BUSINESS_RULES.md) | [**Português**](BUSINESS_RULES.pt.md)

This document establishes the architectural boundaries and mandatory business rules. Any implementation violating these limits must be refactored immediately.

---

## 1. PHILOSOPHY AND RIGOROUS ENGINEERING

The project rejects **"Vibe-Coding"** (programming by intuition or luck). Every line of business logic is treated as an industrial engineering commitment.

### 1.1 Pair Programming and Governance

- **Radical Code Detachment:** If code fails, the flaw lies in communication or architecture. Correction is made via dialogue and documentation adjustment, never manual patches.
- **SSOT (Single Source of Truth):** This file is the Iron Law. Before any complex change, the rule must be documented here.
- **Language:** Code and technical documentation in **English**. Dialogue and pair programming pitch in **Portuguese**.

### 1.2 TDD Protocol (Red-Green-Refactor)

No business logic exists without a test justifying it.

1. **RED:** Write the failing test, defining the contract.
2. **GREEN:** Implement the minimal code to pass.
3. **REFACTOR:** Optimize while maintaining the passing status.

---

## 2. THE IDENTITY MATRIX: TAGS (THE "STATE")

Tags are not classes; they are **Superpowered Hierarchical Identifiers** based on `StringName`. They represent the absolute truth about an Actor's present state.

### 2.1 Tags Golden Rules (v0.2.0 Security)

- **CONDITIONAL Type (Logic Requirement):** Only tags marked as `CONDITIONAL` are valid in requirement lists (`Required/Blocked`) for Abilities, Effects, or Cues. Attempting to use a Name or Event tag for logical blocking is a security violation.
- **NAME Type (Resource Identity):** Only `NAME` tags are valid for identifying and naming `ASAbility` or `ASEffect` resources. This ensures a state (e.g., `state.stunned`) is never confused with an ability (e.g., `ability.warrior.powerhit`).
- **EVENT Type (Frame Occurrence):** `EVENT` tags are transient and used exclusively for dispatching occurrences via `ASTagUtils`. They do not persist in the Singleton's RefCount.

### 2.2 The 3 Canonical Tag Types (Tag Types / ASTagTypes)

| Type          | Semantic Role                            | Usage Restriction                                                    |
| ------------- | ---------------------------------------- | -------------------------------------------------------------------- |
| `NAME`        | Static identity (Who I Am/What this is)  | **EXCLUSIVE** for Resource Identification (`ASAbility`/`ASEffect`).  |
| `CONDITIONAL` | Persistent gameplay state (How I Am)     | **EXCLUSIVE** for Logical Requirements (`Required`/`Blocked`).       |
| `EVENT`       | Instantaneous occurrence (What Happened) | **EXCLUSIVE** for event dispatching via `ASTagUtils` and `ASBridge`. |

---

## 3. THE SINGLETON: ABILITY SYSTEM & ASUTILS

- **Role:** The **Global Configuration API** and repository for atomic static logic.
- **ASUtils:** C++ namespace centralizing high-performance utility functions for tag manipulation and attribute calculations.
- **Business Rules:**
  - The Singleton is the only entity authorized to persist the tag list in `project.godot`.
  - Acts as a **Typing Validator**: Prevents wrong tag types from being used in restricted fields at edit-time (via Inspector).

### 2.3 Tag Type Implementation (Semantic Overview)

The unified tag system provides type-safe manipulation across three primary categories. Detailed implementation resides in `as_tag_types.h`.

| Category             | Purpose                                        | Typical Lifetime                     |
| :------------------- | :--------------------------------------------- | :----------------------------------- |
| **ASNameTag**        | Persistent identity (Class, Team, Traits).     | Indefinite (Until manual removal).   |
| **ASConditionalTag** | Logic gates (Immunity, Prerequisites, States). | Transient or Linked to Effects.      |
| **ASEventTag**       | Instantaneous occurrences (Hit, Death, Cast).  | 1 Tick (Historical record persists). |

### 2.4 Historical Tracking Infrastructure

The `ASComponent` maintains high-performance **128-entry circular buffers** for each tag category to support reactivity and multiplayer reconciliation.

- **NAME History**: Tracks when persistent states were added/removed.
- **CONDITIONAL History**: Tracks changes in permissions and immunities.
- **EVENT History**: Stores full `ASEventTagData` (Instigator, Magnitude, Tick) for recent occurrences.

#### Historical Query APIs

The `ASTagUtils` namespace provides optimized methods to query these buffers:

- `was_tag_added/removed(tag, lookback_sec)`: Checks for state transitions.
- `did_occur(event_tag, lookback_sec)`: Checks for recent combat/gameplay events.
- `get_last_magnitude/instigator(event_tag)`: Retrieves specific data from the most recent occurrence.
  );

````

### 2.5 Event System Integration

**Events are now a specialized tag type** managed through the unified tag system. Resource ASEvent was deprecated and replaced by ASTagUtils-based event handling.

#### 2.5.1 Event Tag Data Structure

Events use the `ASEventTagData` struct in `ASUtils` for full payload information:

- `event_tag`: Exact occurrence identifier (ex: `event.interrupt`)
- `instigator`: The Node that caused it (offender)
- `target`: The affected Node (victim)
- `magnitude`: Base event intensity (impact power)
- `custom_payload`: GDScript/Variant dictionary for metadata transport
- `timestamp`: Natively registered in precise milliseconds
- `tick_id`: Tick identifier for multiplayer synchronization

#### 2.5.2 Event Dispatch API

```gdscript
# Signature: dispatch_event(event_tag, instigator, magnitude, custom_payload)
var asc: ASComponent = target.get_node("ASComponent")
asc.dispatch_event(&"event.weapon.hit", self, 35.0, {})
````

#### 2.5.3 Short-Term Memory (Events Historical)

Events die in 1 tick, but their _"memory"_ persists in a lightweight manner:

- The `ASComponent` mantém um `_event_history` (buffer circular C++ super otimizado de até 128 entradas).
- **Practical use:** Reactive components (como _Parry_ ou _Counter-Attack_) não precisam estar no estado eterno de "parrying". Podem checar o passado recente: `has_event_occurred(&"event.damage.block", 0.4)`. Se o bloqueio aconteceu no último 0.4s, autorize a habilidade de contra-ataque.
- **Rule:** Nunca usar este cache para persistir história (missões, quests). Use exclusivamente para frames de reatividade temporal.

#### 2.5.4 Individual Historical Tracking

All three tag types maintain individual historical buffers in `ASComponent`:

- **NAME History**: Tracks persistent state changes (`state.stunned`, `class.warrior`)
- **CONDITIONAL History**: Tracks permission/immunity changes (`can.parried`, `immune.fire`)
- **EVENT History**: Tracks event occurrences with full payload data

Cada buffer mantém 128 entradas com gerenciamento automático de overflow. As APIs históricas fornecem consultas especializadas:

```cpp
// Consultas de tags NAME
ASTagUtils::name_was_tag_added("state.stunned", target, 2.0f);
ASTagUtils::name_count_additions("state.stunned", target, 10.0f);

// Consultas de tags CONDITIONAL
ASTagUtils::cond_had_tag("immune.fire", enemy, 5.0f);

// Consultas de tags EVENT
ASTagUtils::event_did_occur("event.damage", target, 1.0f);
ASTagUtils::event_get_last_magnitude("event.damage", target);

// Consultas unificadas
ASTagUtils::history_was_tag_present("state.stunned", target, 2.0f);
```

### 2.6 Tag Groups (Visual Organization)

**Tag Groups are not code entities.** They are an editorial convention emerging automatically from dot (`.`) hierarchy in tag identifiers.

- `ASTagsPanel` renderiza tags como uma **árvore**, usando cada segmento separado por ponto como um nó pai.
- `state.stunned`, `state.dead` agrupam automaticamente sob o nó visual `state`.
- Não existe um objeto `TagGroup` C++ — o "grupo" é apenas o prefixo compartilhado.
- **Mandatory Convention:** O prefixo raiz DEVE refletir seu `Tag Type` (ex: tags `Event.*` são sempre `ASTagType::EVENT`).

### 2.7 Logical Evaluation (Predicates)

O sistema suporta 4 estados lógicos em Blueprints (Ability, Effect, Cue) na hora de avaliar os requisitos e bloqueios de um alvo:

- `Required All` (AND): Sucesso se possuir TODAS.
- `Required Any` (OR): Sucesso se possuir PELO MENOS UMA.
- `Blocked Any` (OR): Falha se possuir QUALQUER UMA.
- `Blocked All` (AND): Falha se possuir TODAS SIMULTANEAMENTE.

> [!NOTE]
> Predicates work exclusively on `CONDITIONAL` tags. Tags do tipo `NAME` e `EVENT` não entram em listas de requisito/bloqueio de blueprints. O Editor enforce isso automaticamente via `ASInspectorPlugin`.

### 2.8 Split Registry Pattern

Event identifiers (ex: `event.weapon.hit`) **are registered in the Singleton** like any other tag — to work in editor autocomplete, `ASTagsPanel`, and prevent designer typos. The difference lies in the type: they are registered as `Tag Type = EVENT`.

What **never** goes up to the Singleton is the **data instance** — the `ASEventTagData` struct. Essa separação configura o padrão de **Registro Dividido**:

- **Singleton (Registry):** Conhece o _nome_ `event.weapon.hit`. Garante que existe, que tem o tipo certo e aparece no autocomplete.
- **ASComponent (Occurrence):** Conhece o _acontecimento_. Sabe quem bateu, em quem, com qual força e em qual tick. O Singleton não precisa — nem deve — saber disso.

> [!IMPORTANT]
> **Golden Rule:** Never call `dispatch_event` with an `event_tag` not registered in the Singleton with `Tag Type = EVENT`. Isso seria equivalente a usar um `StringName` digitado à mão sem validação — o campo exato de erros que o sistema foi projetado para eliminar.

---

## 3. ASUtils Structures Registry (Semantic Overview)

All internal data structures are centralized in `as_utils.h` with dedicated serialization and validation support.

| Structure            | Purpose                                    | Primary Data                                          |
| :------------------- | :----------------------------------------- | :---------------------------------------------------- |
| **ASStateCache**     | High-perf 128-tick circular buffer.        | Collection of `ASStateCacheEntry`.                    |
| **ASComponentState** | Full state representation for Persistence. | Attributes, Tags, Active Effects, Cooldowns, History. |
| **ASAttributeValue** | Atomic stat container.                     | Base value, Current value, Modifiers.                 |
| **ASEffectState**    | Runtime effect instance data.              | Tag, Remaining time, Stacks, Level.                   |
| **ASEventTagData**   | Event payload for dispatching.             | Tag, Instigator, Magnitude, Custom Payload, Tick.     |
| **AS\*Historical**   | Individual historical log entries.         | Tag, Context, Timestamp, Tick ID.                     |

### 3.1 Standardized Features

Every structure in `ASUtils` follows the **Universal Implementation Pattern**:

- **Serialization**: `to_dict()` and `from_dict()` for native Godot storage.
- **Validation**: `is_valid()` checks for architectural integrity.
- **Tick Awareness**: Native support for `tick_id` for multiplayer synchronization.
- **API Consistency**: Symmetric getters and setters for all internal fields.

---

## 4. THE SINGLETON: ABILITY SYSTEM (PROJECT INTERFACE)

- **Role:** The **Global Configuration API** e a ponte com o `ProjectSettings`.
- **Business Rules:**
  - É o único responsável por salvar e carregar a lista global de tags nas configurações do projeto (`project.godot`).
  - Atua como um **Registro Central de Nomes** para garantir que recursos duplicados não entrem em conflito.
- **Limit:** Não deve armazenar estado de nenhum Actor. Se uma informação pertence a uma instância de personagem, ela **não** deve estar aqui.

---

## 5. TOOLS LAYER: EDITORS

Interface entre o Humano e os Resources.

### 5.1 ASEditorPlugin

- **Role:** **Bootloader**.
- **Rule:** Registro de tipos, ícones e inicialização de outros sub-editores. Proibido conter lógica de jogo.

### 5.2 ASTagsPanel

- **Role:** Interface visual para o **Registro Global**.
- **Rule:** Manipula exclusivamente o dicionário de tags do `AbilitySystem` Singleton.

### 5.3 ASInspectorPlugin (and Property Selectors)

- **Role:** Contextualization.
- **Rule:** Deve fornecer seletores inteligentes (dropdowns de tags, busca de atributos) para facilitar a configuração de Resources e Components no Inspetor.

### 5.4 The Compat Layer (Camada de Compatibilidade)

Localizada em `src/compat/`.
O projeto é arquitetado estritamente sob a **Estratégia de Compilação Dual**, suportando compilação tanto como Module nativo da Godot quanto como GDExtension.

- **Role:** Blindar o núcleo do framework das divergências entre as APIs internas do Module e da GDExtension.
- **Contour Rule:** Toda lógica central (`src/core`, `src/resources`, `src/scene`) DEVE ser desenhada de forma agnóstica. Qualquer diferença estrutural necessária para interagir com a Engine (especialmente funcionalidades de GUI do Editor) deve ser isolada nesta pasta, resolvendo a compatibilidade via `#ifdef ABILITY_SYSTEM_GDEXTENSION`.
- **Exclusivity:** Core files nunca devem usar diretamente classes exclusivas de GDExtension que quebrem compilação de Module em C++; eles devem invocar o Wrapper no `compat/`.

---

## 6. THE BLUEPRINTS: RESOURCES (THE "WHAT")

Localizados em `src/resources/`. São as **Definições de Dados**.

- **Resources (Blueprints):** São objetos estáticos (`.tres`) que definem o "DNA" de uma habilidade ou efeito. **Regra:** Não devem ser modificados em runtime (exceto `ASStateSnapshot`). Eles são compartilhados entre centenas de instâncias.
- **Specs (Runtime Instances):** São instâncias leves criadas a partir de Resources que mantêm o estado dinâmico (cooldowns, stacks, duração).
- **Snapshot Exception:** O `ASStateSnapshot` é grafado como `Resource` mas é populado em runtime. Ele rompe a regra de imutabilidade para permitir persistência nativa (Save/Load) e cache de rede via sistema de arquivos/recursos da Godot.

### Restrictions of use and performance

> [!WARNING]
> **O `ASStateSnapshot` é um recurso pesado.** Devido à natureza da captura completa de estado, ele consome CPU e memória significativos se usado em larga escala.

1. **Uso Exclusivo para Players:** O uso de Snapshots deve ser restrito a entidades controladas por jogadores (Playable Characters), onde o determinismo e o rollback são críticos para o multiplayer online.
2. **NPCs e Inimigos:** Entidades não-jogáveis **NÃO** devem usar `ASStateSnapshot`. Para persistência de NPCs, utilize mecanismos mais leves como o `SaveServer` da Zyris Engine ou sistemas personalizados de dicionários em Godot.
3. **Regra de Ouro:** Se é jogável, use `ASStateSnapshot`. Se não for, descarte-o.
4. **Independência de Referência:** O snapshot armazena valores primitivos e nomes de tags, não ponteiros para objetos, garantindo que possa ser serializado com segurança.

- **What lives here:** Valores base, ícones, nomes, tags de requisito e listas de modificadores brutos.

### 6.1 ASAbility & ASEffect (Actions and Modifiers)

- **ASAbility - Role:** Definir a lógica de uma ação (Custos, Cooldown, Triggers).
- **ASAbility - Rule:** Único Resource capaz de gerenciar requisitos de ativação e custos de atributos através de especificação. Suporta **Ability Phases** para execuções complexas.
- **ASEffect - Role:** Modificador de estado (Buffs, Debuffs, Dano).
- **ASEffect - Rule:** Define políticas de empilhamento (Stacking) e magnitudes de mudança nos atributos.

### 6.2 ASAttribute & ASAttributeSet (The Attribute System)

- **ASAttribute - Role:** Define os metadados (limites min/max) de uma única estatística.
- **ASAttributeSet - Role:** Agrupa as estatísticas e define o estado inicial de um personagem. Detém a lógica de modificação de atributos.
- **ASAttributeSet - Regra (Attribute Drivers):** Permite derivar o valor base de um atributo a partir de outro (ex: 2 \* Força = 1 Ataque). O recalculo é automático em mudanças de valor base.
- **ASAttributeSet - Regra (Prioridade):** Modificadores (Flat Add, Multiplier) são aplicados _após_ o cálculo dos Drivers.

### 6.3 ASContainer & ASPackage (Archetypes and Payloads)

- **ASContainer - Role:** Arquétipo completo (Dicionário de Identidade do Ator).
- **ASContainer - Regra:** Atua como o "Template de Fábrica" para inicialização total do `ASComponent`.
- **ASPackage - Role:** Agrupador de transporte (Envelope de Dados).
- **ASPackage - Regra:** Deve ser usado exclusivamente para transmitir coleções de efeitos e cues via `ASDelivery`.

### 6.4 ASCue (Visual Feedbacks)

- **Role:** Feedback audiovisual puro (Animação, Som, Partículas).
- **Rule:** Proibido alterar qualquer dado de gameplay. Deve ser disparado reativamente.

### 6.5 ASAbilityPhase (O Ciclo de Vida Complexo)

A funcionalidade mais poderosa para designers em termos de "Máquinas de Estado" embutidas (Hierarchical Abilities).

- **Role:** Fragmentar a execução engessada de uma habilidade em estágios granulares e altamente configuráveis (ex: `Windup`, `Execution`, `Recovery`).
- **Natureza:** Se uma habilidade padrão age como uma "pistola" (inicia, aplica e termina num click), uma habilidade com Fases atua como um "ritual" com várias etapas no tempo.
- **Regras Críticas:**
  - **Temporário & Específico:** Cada Fase pode aplicar e remover seus próprios `ASEffects` transitórios que duram apenas enquanto aquela fase estiver ativa.
  - **Duração ou Evento:** Uma Fase avança para a próxima de duas formas: (a) Expirou o tempo de duração da fase; (b) Ocorreu o _Transition Trigger Event_ (a habilidade estava aguardando o Node de Animação enviar um Evento `.Hit` para avançar).
  - **Avisos Autônomos:** A transição entre Fases sempre dispara um AS Event automático da própria framework para permitir fluidez e resposta de UI.

---

## 7. THE EXECUTORS: SPECS (THE "HOW")

Localizados em `src/core/`. Onde o estado e a lógica de execução residem.

- **Role:** Representar a **Instância Ativa**. É o dono do **"Agora"**.
- **Golden Rule: STATE SOVEREIGNTY.**
- **O que deve viver aqui (e não no Component):**
  - `duration_remaining`: O timer individual de cada instância.
  - `stack_count`: Quantas vezes este efeito específico está acumulado.
  - `calculate_...`: Lógica de cálculo que depende de atributos variáveis (ex: dano baseado em força atual).
- **Responsabilidade:** O Spec deve saber se "terminou" ou não. O Component apenas pergunta a ele.

### 7.1 ASAbilitySpec & ASEffectSpec (Execution)

- **ASAbilitySpec - Role:** Habilidade em execução ativa ou equipada.
- **ASAbilitySpec - Regra:** Gerencia o cooldown individual e o estado de ativação.
- **ASEffectSpec - Role:** Instância ativa de um modificador.
- **ASEffectSpec - Regra:** Detém a soberania sobre o tempo restante (`duration`) e pilhas (`stacks`).

### 7.2 ASCueSpec & ASTagSpec (Feedback and Identity)

- **ASCueSpec - Role:** Gerenciador do ciclo de vida de um feedback na cena.
- **ASCueSpec - Regra:** Garante a limpeza (Queue Free) do Node instanciado após o término.
- **ASTagSpec - Role:** Contador de referências (Refcount) para Tags.
- **ASTagSpec - Regra:** Garante que uma Tag só saia do ator quando todos os seus Specs de origem expirarem.

### 7.3 ASAbilitySpec (Gestão de Fases)

- **Role:** Gerencia o índice da fase atual e a progressão temporal entre os estágios definidos no `ASAbility`.
- **Regra:** Deve ser capaz de avançar para a próxima fase via tempo ou via recebimento de um `ASEventTag` específico.

---

## 8. THE ORCHESTRATOR: COMPONENT (THE HUB)

O `ASComponent` (ASC).

- **Role:** **Gestor de Coleções** e **Roteador de Sinais**.
- **Business Rules:**
  - Não deve gerenciar timers individuais de efeitos (isso é do Spec).
  - Responsável por manter a lista de `active_specs` e `unlocked_specs`.
  - Atua como o **Dono dos Atributos** (via `AttributeSet`).
  - É o único que pode adicionar/remover tags do Actor.
- **Garantir Determinismo:** O ASC deve ser capaz de retroceder e avançar seu estado (Rollback/Prediction) via `ASStateSnapshot`.
- **Cache de Estado:** Manter um buffer interno de snapshots para sincronização de rede.
- **Limite:** O ASC não deve saber os detalhes internos de como uma habilidade funciona. Ele apenas diz: `spec->_activate()`, `spec->tick(delta)`, `spec->_deactivate()`.
- **Node Registry:** O Componente deve manter um registro de aliases de nós (ex: "Muzzle") para que Cues saibam onde instanciar efeitos visuais sem dependências de caminhos de cena.
- **ASEventTagTagHistorical (Memória de Eventos):**
  - **Role:** O ASC mantém um buffer circular de eventos recentes para permitir lógica condicional baseada no passado imediato.
  - **Golden Rule:** Não deve ser usado para persistência de longo prazo. É um "cache de reatividade".
  - **Queries:** Permite perguntar: "Ocorreu o evento X nos últimos Y ticks?".
  - **Sincronia:** Assim como o `ASStateCache`, o histórico de eventos deve ser sensível ao `tick` temporal para garantir consistência em situações de Rollback.

---

## 9. DELIVERY SYSTEMS AND REACTIVITY

### 9.1 ASDelivery (Payload Injections)

- **Role:** Desacoplar o emissor do alvo em interações espaciais (projéteis, AoEs).
- **Rule:** Transporta um `ASPackage` e injeta o conteúdo ao colidir com um ASC.

### 9.2 Ability Triggers (Reactive Automation)

- **Role:** Permitir ativação automática de habilidades baseada em eventos de estado (Tags) ou eventos transitórios (AS Events).
- **Rule:** Ativação baseada exclusivamente em `ON_TAG_ADDED`, `ON_TAG_REMOVED` ou `ON_EVENT_RECEIVED`.

---

## 10. MULTIPLAYER ARCHITECTURE: PREDICTION & ROLLBACK

The Ability System is built from the ground up for authoritative multiplayer using a **Client-Side Prediction (CSP)** and **Server Reconciliation** model.

### 10.1 Temporal State Awareness (The 128-Tick Buffer)

Every `ASComponent` maintains an **`ASStateCache`**—a lightweight, high-performance circular buffer storing the state of the last 128 ticks.

- **Capture**: Every `physics_tick` (60Hz), the client and server record a snapshot of attributes, tags, and active effects into the cache.
- **Tick Affinity**: Every event, effect application, and historical entry is tagged with a `tick_id`, allowing precise temporal cross-referencing.

### 10.2 Client-Side Prediction (Latent zero feel)

When a player initiates an action (e.g., `try_activate_ability`):

1. **Predictive Execution**: The client immediately applies the logic locally (reduces attributes, plays cues, adds predicted tags).
2. **Buffering**: The result is stored in the local `ASStateCache`.
3. **Transmission**: The intention is sent to the server via `request_activate_ability`.

### 10.3 Server Reconciliation & Rollback

The server processes the request and sends back the **Authoritative State** for that specific tick.

1. **Comparison**: The client compares the incoming server state with its locally cached state for the same `tick_id`.
2. **Detection**: If a divergence is detected (e.g., predicted damage was mitigated differently on server), a **Rollback** is triggered.
3. **Reversion**: The client uses `ASComponentState::from_dict()` to instantly overwrite its current state with the server's authoritative data.
4. **Re-prediction**: The client re-simulates all local inputs from the authoritative tick up to the current frame to maintain visual continuity.

### 10.4 Deterministic Gameplay Rules

To minimize rollbacks, all simulation logic MUST be deterministic:

- **Physics Only**: All gameplay logic MUST reside in `_physics_process`. Accessing `_process` for gameplay mutation is a violation of the network protocol.
- **Attribute Priorities**: Deterministic calculation order (Drivers -> Base -> Modifiers) ensures identical results across machine architectures.
- **Historical Recalculation**: During rollback, the historical logs (`ASAttributeHistorical`, etc.) are critical for re-syncing visual feedback (Cues) that might have been skipped during the reset.

---

## 11. REACTIVITY PROTOCOL (THE TRICHROMATIC MATRIX)

Para evitar que a arquitetura decline e se torne um emaranhado caótico onde todos os sistemas interferem uns nos outros, estabelecemos o _Protocolo de Reatividade_. Esta é a doutrina de como os 3 Pilares operam em uníssono orquestrado.

### 11.1 A Ordem Natural

1. **INPUT/AÇÃO:** Uma interação, término temporizador ou impacto físico emite um **AS Event** (`event.damage`).
2. **ESCUTA/PROCESSAMENTO:** Uma Entidade escuta via Triggers (`ON_EVENT`).
3. **MUTAÇÃO:** A habilidade reativa acerta os requisitos, invoca e aplica o mutador (`ASEffect`).
4. **ESTADO (Fim do Ciclo):** O Effect mutou os atributos ou adicionou permanentemente a **AS Tag** (`state.stunned`).

> [!CAUTION]
> **Fatal Errors punished with deep refactor:**
>
> - Esperar que uma habilidade inicie baseada em "perda de tag". (Isso é sintoma de acoplamento de estado; dispare um Evento avisando o fim).
> - Se uma Habilidade falhar num requisito de Tag ou Custo, NUNCA gerencie estado (aplicar tags temporárias). Emita o gatilho `event.ability.failed` relatando por qual motivo (Dicionário Payload), para loggers ou UI reagirem.

### 11.2 Hybrid Triggers

A ativação de Habilidades através do _Ability Triggers_ no Spec passou por revisão de hierarquia. Pode-se construir automação pura através deles:

- `TRIGGER_ON_EVENT`: É o padrão ouro para responsidade de combate (Reagir instantaneamente no impacto).
- `TRIGGER_ON_TAG_ADDED` / `REMOVED`: É o padrão para automação de ambiente (Ligar aura de lentidão quando entrar na água).

### 11.3 O Pacto do ASDelivery

O componente `ASDelivery` (ex: um míssil ou aura rastreado) carrega o envelope mortífero `ASPackage`.

- **Regra e Obrigação:** Todo ASDelivery, ao concluir rota e aplicar pacote a um ASC alvo, DEVE obrigatoriamente emitir o disparo de Evento invocando `target_asc->dispatch_event(package_tag)`.
- Isso assegura que "tomar uma fireball na cara" automaticamente preencha a memória temporal do ASC alvo (`ASEventTagHistorical`), habilitando bloqueios ou triggers reativos purificados.

---

---

## 12. ATOMIC AI INTEGRATION (ASBridge & HSM)

To interact with the LimboAI framework atomically, the Ability System provides **ASBridge** — not as a class or Singleton, but as an architectural native integration layer (located at `src/bridge/`).

- **Status v0.2.0:** The AI infrastructure is part of the same binary as the Ability System, guaranteeing zero latency. Support for the obsolete `ASBridge` Singleton has been completely decommissioned.
- **The Integrator Layer Role:** `ASBridge` acts simply as the collection of Behavior Tree Tasks and States (`BTAction`, `BTCondition`, `LimboState`) that connect LimboAI directly to the `ASComponent`.
- **Centralized Resolution:** The global `AbilitySystem` Singleton is the absolute authority responsible for `resolve_component()`. When an AI agent needs to find its `ASComponent`, the Bridge Tasks query `AbilitySystem::get_singleton()->resolve_component(agent, path)`. The Singleton acts as the **Reference Judge**, maintaining a secure node search and resolving string aliases efficiently so AI nodes do not need to hardcode absolute scene paths.
- **HSM Synergy:** `ASComponent` is natively compatible with **LimboHSM**. The character's hierarchical state machine is driven by Ability System's Tag identifiers and Events.
- **Rigorous Decoupling:** Even though they are in the same binary, BT/HSM tasks never call ability cores directly. They use the Bridge layer to preserve the **Reactivity Protocol**.

---

## 13. TECHNICAL RIGOR AND TEST QUALITY

### 13.1 Padrão 300% (Iron Law)

Cada funcionalidade deve ser provada por pelo menos **3 variações** no mesmo teste:

1. **Happy Path:** Cenário base ideal.
2. **Negative:** Entrada inválida ou falha esperada.
3. **Edge Case:** Combinações complexas (multi-tags, limites de borda).

### 13.2 Test Suites

- **Core (Unit):** Atômicos, sem efeitos colaterais.
- **Advanced (Integration):** DoT/HoT periódico, fluxos de RPG complexos.
- **Multiplayer (Simulation):** Executado via `utility/multiplayer/runner.py` com latência injetada.

---

## 14. API NAMING STANDARDS (AS Specific)

Para manter a consistência, toda a API pública deve seguir estes padrões próprios do Ability System:

### 14.1 Method Prefixes

Os métodos são categorizados pela sua intenção e camada de acesso:

- **🎮 Camada de Gameplay (Usage for Game Logic)**
  - `try_activate_...`: **Safe execution.** Tenta disparar uma lógica que depende de requisitos prévios. Integra a verificação e a ação. **Uso obrigatório para habilidades e efeitos.**
  - `can_...`: **Pre-authorization.** Avalia se uma ação pode ser executada sem iniciá-la.
  - `is_...`: **Status study.** Verifica condições booleanas de estado ou identidade (ex: `is_ability_active`).
  - `has_...`: **Possession query.** Verifica se o objeto detém uma chave específica (ex: `has_tag`).
  - `get_...`: **Extracting information.** Obtém valores, referências ou metadados de leitura.
  - `cancel_...`: **Interruption.** Encerra voluntariamente um fluxo em execução.
  - `request_...`: **Network intention.** Solicita a execução de uma ação via RPC (Multiplayer).
  - `activation_cancel_tags`: **Interruption Policy.** Tags of active abilities or effects that will be canceled upon triggering this action (Native HSM).
- **🏗️ Camada de Infraestrutura/Interna (Restricted use or Config)**
  - `apply_...`: **Forced application.** Injeta um payload ou container ignorando regras de ativação. Usado em inicialização ou por sistemas de entrega (`ASDelivery`).
  - `add_...` / `remove_...`: **Low-level mutation.** Altera coleções internas. Não deve ser usado como atalho para ativar lógica de jogo (ex: use `try_activate` em vez de tentar "adicionar" um efeito manualmente).
  - `unlock_...` / `lock_...`: **Inventory management.** Altera a disponibilidade de habilidades no catálogo do ator.
  - `register_...` / `unregister_...`: **System bound.** Conecta o componente a nós externos ou gerencia o Singleton global.
  - `rename_...`: **Refactor internals.** Altera identificadores de Tags e propaga a mudança.
  - `set_...`: **Direct mutation.** Define valores brutos (Base Values).
  - `clear_...`: **Reset total.** Limpa estados ou coleções de forma absoluta.
  - `capture_...`: **Persistence.** Congela o estado atual em um snapshot.

> [!IMPORTANT]
> **Não existe `activate` direto na API pública.** O uso de `try_...` é o único caminho seguro para gameplay. Métodos de infraestrutura (`apply_`, `add_`) existem para sistemas de baixo nível e não devem ser usados para contornar verificações de custos e cooldowns.

### 14.2 Suffixes of origin and context

Sufixos são obrigatórios para resolver ambiguidades de entrada, destino ou comportamento:

- **Origem (Input):**
  - `..._by_tag`: Operação baseada na identidade global da tag (ex: `has_tag`, `try_activate_ability_by_tag`).
  - `..._by_resource`: Operação baseada na instância do arquivo de recurso (ex: `try_activate_ability_by_resource`).
- **Destino (Target):**
  - `..._to_self`: A ação é aplicada exclusivamente ao componente que a chamou.
  - `..._to_target`: A ação exige um componente alvo explícito como argumento.
- **Comportamento (Behavior):**
  - `..._debug`: Retorna dados verbosos ou não otimizados para ferramentas de diagnóstico.
  - `..._preview`: Realiza cálculos teóricos (E se...?) sem aplicar efeitos colaterais ou consumir recursos.
  - `..._all`: Operação em massa que afeta toda a coleção relevante (ex: `cancel_all_abilities`).

### 14.3 Arguments and Typing

A nomenclatura de argumentos deve ser autoexplicativa e seguir a hierarquia de tipos do projeto:

**Descritive nomenclature:** Proibido o uso de variáveis de um único caractere (ex: `t`, `a`). Use o nome completo do conceito (`tag`, `ability`, `effect`, `attribute`, `level`, `value`, `target_node`, `data`).

- **Performance Typing:**
  - `StringName`: Para todos os identificadores de tags e nomes de atributos (chave de dicionário).
  - `float`: Para todas as magnitudes, níveis e durações (mesmo que sejam valores inteiros na lógica, o motor de atributos opera em ponto flutuante).
  - `Dictionary`: Para payloads de dados variáveis (ex: no sistema de Cues).
- **Canonical Order:** Quando múltiplos argumentos são necessários, siga a ordem de importância:
  1. **Identificador:** `tag` ou `resource`.
  2. **Magnitude:** `level` ou `value`.
  3. **Target/Context:** `target_node` ou `data`.
- **Valores Default:** Argumentos opcionais devem possuir valores neutros documentados (ex: `level = 1.0`, `target_node = null`).

### 14.4 Signals

Sinais devem comunicar eventos que **já ocorreram**, seguindo o padrão de voz passiva:

- **Formato:** `snake_case` no tempo passado.
- **Exemplos Corretos:** `ability_activated`, `effect_removed`, `attribute_changed`, `tag_event_received`.
- **Exemplos Incorretos:** `on_ability_activate` (prefixo desnecessário), `activate_ability` (confunde com método).

### 14.5 Internal C++ Members

Para garantir segurança e legibilidade no código-fonte GDExtension:

- **Variáveis Privadas/Protegidas:** Devem obrigatoriamente começar com `_` (underscore). Ex: `_attribute_set`, `_is_active`.
- **reactive Getters/Setters:** Sempre que uma mudança de variável exigir uma reavaliação (ex: mudar a tag exige recontar o ASTagSpec), deve-se usar um setter formal em vez de acesso direto.
- **Propriedades Públicas:** Devem espelhar os nomes brutos (sem `_`) para serem expostas corretamente ao Inspetor do Godot.
- **Dual Build:** Todo código de teste deve suportar a compilação via `#ifdef ABILITY_SYSTEM_GDEXTENSION`.

---

## 15. DESIGN PATTERNS (C++/GDEXTENSION)

O sistema utiliza padrões clássicos adaptados para a arquitetura de alta performance da Godot Engine.

### 15.1 Spec Pattern (Resource vs. Instance)

- **O Problema:** Resources Godot são compartilhados. Modificar um afetaria todos os inimigos do mesmo tipo.
- **A Solução:** Separação total entre **Resource** (Dados Imutáveis/DNA) e **Spec** (Estado de Runtime/Instância).
- **Regra:** Toda lógica que exige alteração de valor (timers, stacks) deve residir no Spec. O Resource é apenas um Provedor de Dados.

### 15.2 Flyweight Pattern (optimized memory)

- **Conceito:** Milhares de atores compartilham as mesmas referências de `ASAbility` e `ASEffect`.
- **Implementação:** O `ASComponent` armazena apenas ponteiros (RefCounters) para os Resources. Os dados pesados (ícones, curvas de dano) nunca são duplicados na memória.

### 15.3 Command Pattern (Abilities)

- **Conceito:** Cada habilidade é um comando auto-contido que sabe como iniciar, executar e cancelar.
- **Requisito:** Encapsulamento total. O componente não deve conhecer a lógica interna da habilidade; ele apenas despacha o comando.

### 15.4 Data-Driven Design

- **Regra:** O comportamento deve ser definido em arquivos `.tres` no Editor, não em código rígido (Hardcoded).
- **Vantagem:** Permite que designers alterem o balanceamento sem recompilar o plugin.

---

## 16. TEST PATTERNS (RIGOR AND DETERMINISM)

A confiabilidade do sistema é garantida por padrões de teste industriais.

### 16.1 Deterministic Physics Ticking

- **Regra:** Testes de tempo (Cooldowns/Duração) devem ser validados via `physics_process` em passos fixos (Ticks).
- **Mocking do Tempo:** Em testes unitários, simulamos a passagem do tempo chamando `tick(delta)` manualmente para garantir que 1.0s seja exatamente 1.0s, independente do lag do hardware.

### 16.2 Isolation & Mocking

- **Mock Assets:** Para testar habilidades complexas, criamos Resources temporários em memória via código no `TEST_CASE`.
- **Dummy Actors:** O uso de nodes simples com `ASComponent` é preferível a carregar cenas complexas (`.tscn`) para testes unitários, garantindo velocidade de execução.

### 16.3 State Injection

- **Padrão:** Em vez de esperar 10 segundos para testar o fim de um efeito, o teste deve injetar um Spec com `duration_remaining = 0.1` e validar o próximo tick.

### 16.4 Signal Auditing

- **Padrão:** Todo teste de ativação deve auditar se o sinal correspondente (ex: `ability_activated`) foi emitido com os argumentos corretos, garantindo que o feedback visual (Cues) também funcione.
