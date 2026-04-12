# Arquitetura de Query Hash Map (O(1))

> **Status:** Implementado
> **Versão:** 1.0

O sistema de **State Engineering** utiliza uma arquitetura de indexação invertida para garantir que a seleção de estados ocorra em tempo constante (`O(1)`), independentemente do número de estados (habilidades) que o personagem possua.

---

## 1. Estrutura de Dados

### 1.1. O Índice (`Compose.gd`)

Cada recurso `Compose` (o "Deck" de habilidades) mantém três dicionários de regras que mapeiam uma **Chave de Contexto** para uma **Lista de Candidatos**.

```gdscript
# Compose.gd
@export var move_rules : Dictionary = {}   # { Motion.RUN: [RunState, ...], ... }
@export var attack_rules : Dictionary = {} # { Attack.FAST: [Slash1, ...], ... }
@export var interactive_rules : Dictionary = {}
```

Estes dicionários são construídos automaticamente no Editor (via `@tool`) e salvos no disco, eliminando o custo de processamento durante o carregamento do jogo.

### 1.2. A Chave de Indexação (`State.gd`)

Cada estado define sua própria chave de indexação através do método `get_lookup_key()`. Esta chave determina em qual "balde" o estado será armazenado.

```gdscript
# State.gd
func get_lookup_key() -> int:
    if self is MoveData:
        var m = requirements.get("motion", 0)
        # Tratamento Especial: EXCEPT_DASH (5) é um filtro negativo,
        # então ele deve ser indexado no balde genérico (0/ANY) para ser testado sempre.
        if m == 5: return 0
        return m
    if self is AttackData:
        return requirements.get("attack", 0)
    return 0 # ANY
```

---

## 2. Fluxo de Execução (Runtime)

### 2.1. O Processo de Query (`PlayerMachine.gd`)

Quando a Machine precisa decidir o próximo estado, ela não itera sobre todos os estados. Ela consulta diretamente o balde relevante para o contexto atual.

**Exemplo: Movimento**

1. **Leitura do Contexto:** A Machine verifica o contexto atual de movimento.

   ```gdscript
   var key = context.get("Motion", BehaviorStates.Motion.IDLE)
   ```

2. **Acesso Direto (O(1)):** Acessa o dicionário `move_rules` usando a chave.

   ```gdscript
   var candidates = []
   # Busca no Item Atual
   candidates.append_array(current_item.compose.move_rules.get(key, []))
   # Busca também no balde genérico (ANY)
   if key != 0:
       candidates.append_array(current_item.compose.move_rules.get(0, []))
   ```

3. **Fallback:** Se necessário, repete o processo para o `fallback_item` (Desarmado).

### 2.2. Validação e Pontuação (`find_best_match`)

Com a lista reduzida de candidatos (geralmente 1 a 3 estados), a Machine executa o algoritmo de **Scoring**:

1.  **Filtro Rígido:** Se um estado requer `Physics: AIR` e estamos no `GROUND`, ele é descartado imediatamente.
2.  **Pontuação de Especificidade:**
    - Se o estado requer `Weapon: KATANA` e temos uma Katana, ele ganha **+1 Ponto**.
    - Se o estado requer `Weapon: ANY`, ele ganha **0 Pontos**.
3.  **Seleção:** O estado com a maior pontuação vence.

Isso permite que um "Ataque Genérico" (Score 0) seja substituído automaticamente por um "Ataque de Katana" (Score 1) quando a arma é equipada, sem necessidade de `if/else`.

---

## 3. Ferramentas de Editor (`@tool`)

O script `Compose.gd` opera em modo `@tool`.

- **Auto-Build:** Ao carregar ou editar um `Compose` no editor, o método `_ready()` (ou `initialize()`) é chamado.
- **Persistência:** Ele reconstrói os dicionários `_rules` baseados nos arrays de estados (`move_states`, etc.) e **salva o recurso no disco**.
- **Benefício:** O jogo final carrega arquivos `.tres` que já contêm os mapas de hash prontos, sem overhead de CPU na inicialização.

---

## 4. Migração de Dados

Os estados utilizam um Dicionário `requirements` para armazenar seus filtros.

```gdscript
@export var requirements: Dictionary = {
    "motion": BehaviorStates.Motion.ANY,
    "jump": BehaviorStates.Jump.ANY,
    ...
}
```

> **Nota:** Scripts de migração (`MigrateStates.gd`) foram utilizados para converter propriedades antigas para este formato de dicionário, garantindo que dados de projetos legados não fossem perdidos.
