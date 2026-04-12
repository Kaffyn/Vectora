# Ability System - Demo Scene

## Visão Geral

Esta cena de demonstração funcional exemplifica como utilizar o Ability System v4.6 para:

- **Movimentação do Player**: Idle, Walk, Run, Jump, Fall, Dash
- **Sistema de Habilidades**: Combo Attack, Heavy Attack, Dash Attack
- **Rastreamento de Estado**: Via tags NAME, CONDITIONAL e EVENT
- **UI Responsiva**: Atualização reativa com sinais `tag_changed`

## Arquitetura

### Player Component (`demo/player/`)

Componentes responsáveis pela lógica e visual do jogador:

- **`player.tscn`** - Cena principal com CharacterBody2D
- **`player.gd`** - Script de lógica do player com ASComponent integrado
- **`player_animations.tres`** - Biblioteca centralizada de animações
- **`resources/`** - Assets e configurações específicas do player

### UI Component (`demo/ui/`)

Interface de usuário reativa:

- **`hud.tscn`** - Cena principal da HUD
- **`hud.gd`** - Rastreamento de estado e atributos
  - Conectado ao sinal `ASComponent.tag_changed`
  - Exibição de barras de health/stamina
  - Estado atual visualmente identificado
  - Histórico de ativações de habilidades

### Resources (`demo/resources/`)

Dados de configuração do Ability System:

- **`abilities/`** - Definições de habilidades (ability_*.tres)
- **`attributes/`** - Definições de atributos (attr_*.tres)
- **`effects/`** - Definições de efeitos (effect_*.tres)
- **`cues/`** - Definições de sinais visuais/sonoros (cue_*.tres)
- **`packages/`** - Pacotes de entrega de efeitos (package_*.tres)
- **`container/`** - Contêiner com todas as habilidades do player
- **`as_state_snapshot.tres`** - Snapshot de estado (optional)

## Conceitos Chave

### Sistema de Tags

O Ability System utiliza 3 tipos de tags para rastreamento de estado:

- **NAME**: Tags persistentes (ex: `weapon.unarmed`, `weapon.sword`)
- **CONDITIONAL**: Tags condicionais removíveis (ex: `state.stunned`, `state.invulnerable`)
- **EVENT**: Tags de eventos únicos (ex: `action.attacked`, `action.dashed`)

### Fluxo de Habilidades

```
1. Player Input
   ↓
2. ASComponent.try_activate_ability_by_tag()
   ↓
3. Habilidade executa suas Fases
   ↓
4. Efeitos são disparados (ASDelivery, ASPackage)
   ↓
5. Sinais emitidos (ability_activated, tag_changed)
   ↓
6. UI atualiza de forma Reativa
```

### Vinculação com UI (Reactive Binding)

A HUD se conecta via signal ao ASComponent:

```gdscript
var asc = $Player.get_node("ASComponent")
asc.tag_changed.connect(_on_asc_tag_changed)
```

Desta forma, **não há polling** - a UI atualiza apenas quando há mudanças reais de estado.

## Como Executar a Demo

1. **Abra a cena principal**: `demo/demo.tscn`
2. **Pressione Play** (F5)
3. **Controles**:
   - `WASD` - Movimento (Idle → Walk → Run)
   - `Space` - Pular (Jump → Fall)
   - `Mouse Click` - Atacar (Combo → Heavy Attack)
   - `Shift + Click` - Dash Attack

## Debugging

Logs de debug estão habilitados em:

- **`demo/ui/hud.gd`**:
  - `[HUD] tag_changed` - Quando tags mudam
  - `[HUD] State updated` - Quando estado do player atualiza
  - `[HUD] Weapon updated` - Quando arma ativa muda

- **`demo/player/player.gd`**:
  - `[State] Activating motion.*` - Quando estado de movimento ativa
  - `[Tag Color]` - Quando cor do player muda baseado em estado
  - `[ERROR]` - Alertas de problemas

**Verifique a console** (Output do editor) para ver estes logs durante execução.

## Estrutura de Arquivos

```
demo/
├── ABILITY_SYSTEM_DEMO.md          ← Este arquivo
├── README.md                        ← Overview geral
├── demo.tscn                        ← Cena principal
│
├── player/
│   ├── player.tscn                 ← Cena do player
│   ├── player.gd                   ← Lógica do player
│   ├── animations/                 ← Animações
│   └── resources/                  ← Items, hotbar, etc
│
├── ui/
│   ├── hud.tscn                    ← Interface
│   └── hud.gd                      ← Lógica da HUD
│
├── resources/
│   ├── abilities/                  ← Habilidades
│   ├── attributes/                 ← Atributos
│   ├── effects/                     ← Efeitos
│   ├── cues/                        ← Sinais visuais/sonoros
│   ├── packages/                    ← Pacotes de entrega
│   └── container/                  ← Contêiner do player
│
├── autoload/
│   └── GameMachine.gd              ← Gerenciador global
│
└── scenes/
    └── level.tscn                  ← Cena do nível
```

## Próximas Etapas

Para expandir essa demo:

1. **Adicionar novos Estados**: Crie abilities em `resources/abilities/`
2. **Adicionar Efeitos**: Crie effects em `resources/effects/`
3. **Adicionar UI**: Expanda `hud.gd` com novos painel
4. **Adicionar Inimigos**: Use ASComponent para IA também
5. **Adicionar Persistência**: Use `ASStateSnapshot` para save/load

## Referências

- [Ability System Framework Documentation](../src/doc_classes/)
- [Business Rules (Português)](../BUSINESS_RULES.pt.md)
- [Implementation Plan](../IMPLEMENTATION_PLAN.md)
