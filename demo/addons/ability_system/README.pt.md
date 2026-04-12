<p align="center">
  <img src="doc/images/logo.svg" width="600" alt="Ability System Logo">
</p>
<br/>

# Ability System - Lógica Orientada a Dados para Godot 4.x

[![Godot Engine](https://img.shields.io/badge/Godot-4.6-blue?logo=godot-engine&logoColor=white)](https://godotengine.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> [!TIP]
> **Leia isto em outros idiomas / Read this in other languages:**
> [**Português**](README.pt.md) | [**English**](README.md)

O **Ability System** (AS) é um framework poderoso para criação de combate, habilidades e atributos modulares. Projetado para escalar desde mecânicas simples até sistemas complexos de RPG — tudo com alta performance em C++ e arquitetura modular.

---

## 📦 Instalação

1. Baixe o `ability-system-plugin.zip` mais recente em nossas [Releases](https://github.com/MachiTwo/AbilitySystem/releases/download/v0.1.0/ability-system+v0.1.0-dev.gdextension.zip).
2. Extraia e copie a pasta `addons/ability_system` para o diretório `addons/` do seu projeto.
3. Reinicie a Godot e vá em **Projeto > Configurações do Projeto > Plugins** para ativar o plugin "Ability System". Isso ativará o Editor de Tags e as funcionalidades customizadas do Inspetor.

---

## 🧬 Sistemas de Alta Escalabilidade

Este framework introduz conceitos avançados para desacoplar a lógica de jogo:

- **Ability Phases**: Divida suas habilidades em estágios como `Windup`, `Execution` e `Recovery`. Gerencie durações e efeitos específicos para cada fase de forma nativa.
- **AS Events**: Gatilhos transitórios que carregam dados (`ASEventTag`). Perfeitos para comunicar impactos, interações de UI ou gatilhos de animação sem poluir o estado do personagem.
- **Events Historical**: Memória de curto prazo para eventos que ocorreram recentemente, permitindo que habilidades e efeitos consultem o passado imediato (ex: "Se bloqueou um ataque nos últimos 0.5s").

---

## 🛠️ Arquitetura de Build Dual

Este projeto foi projetado exclusivamente para suportar **Compilação Dual**, atendendo tanto ao desenvolvimento do núcleo da engine quanto ao ecossistema de plugins:

- **Zyris Engine (Nativo):** Este framework é um componente nativo do núcleo do **Zyris Engine**. O desenvolvimento neste repositório permite automação centralizada e validação rigorosa. Após estabilizado, o código é oficialmente integrado à branch `master` do Zyris Engine.
- **GDExtension (Plugin):** Uma versão em biblioteca dinâmica para projetos Godot 4.x padrões. Oferece 100% de paridade lógica sem exigir a recompilação do motor, ideal para projetos que utilizam a distribuição oficial da Godot.

Uma base de código C++ unificada alimenta ambas as versões, utilizando um sistema robusto de pré-processamento (`#ifdef ABILITY_SYSTEM_GDEXTENSION`) para gerenciar as diferenças de integração, mantendo uma única fonte de verdade para a lógica.

---

## 🏗️ Guia de Início Rápido

Transforme sua lógica de jogo em dados com estes passos fundamentais. Vamos construir um sistema onde interações sociais e combate coexistem.

### 1. Acesse o Gerenciador de Tags

O Gerenciador de Tags é o coração do vocabulário do seu projeto. Sua localização depende da versão utilizada:

- **GDExtension (Plugin):** Procure pela aba **Ability System Tags** no **Painel Inferior** (ao lado de Output/Debugger).
- **Zyris Engine (Nativo):** Vá em **Projeto > Configurações do Projeto** e procure pela aba **Ability System Tags** logo após o Input Map (Mapa de Entrada).

### 2. Defina o Vocabulário (Tags)

Defina identificadores hierárquicos. Pontos criam ramos que a lógica pode consultar:

- `ability.social.talk`: A base para todas as conversas.
- `state.emotional.angry`: Um estado que pode bloquear interações sociais.
- `state.stun.frozen`: Um estado de combate (congelado).

> [!TIP]
> **Lógica Abrangente:** Verificar por `state.emotional` retornará verdadeiro se o ator tiver `state.emotional.angry` ou `state.emotional.happy`.

### 3. Crie o Esquema de Atributos (AttributeSet)

Crie um recurso **AttributeSet** (ex: `AtributosRPG.tres`).

- Adicione stats de combate: `Saúde`, `Mana`.
- Adicione stats sociais: `Carisma`, `Paciência`.
- Defina valores Mín (0), Máx (100) e Base (100).
- O `ASComponent` (ASC) irá **duplicar profundamente** este set ao spawnar, garantindo pools de saúde/paciência únicos para cada NPC.

### 4. Projete Habilidades Sociais (Diálogos)

Crie uma **Ability** (ex: `HabilidadeConversar.tres`):

- **Tag da Habilidade:** `ability.social.talk`.
- **Tags de Ativação Bloqueadas:** Adicione `state.emotional.angry`. Um personagem furioso não consegue iniciar uma conversa educada!
- **Tags de Posse (Owned Tags):** Adicione `state.busy.talking`. Esta tag é concedida _enquanto_ a habilidade está ativa e pode ser usada para bloquear outras ações como "Correr".

### 5. Crie Efeitos Emocionais

Crie um **Effect** (ex: `EfeitoRaiva.tres`):

- **Política de Duração:** `Infinite` (Passivo) ou `Duration` (Temporário).
- **Modificadores:** Adicione um modificador para `Carisma` (Multiplicar por 0.5) e `Ataque` (Somar 20).
- **Tags Concedidas:** `state.emotional.angry`.

### 6. Construa o Arquétipo (ASContainer)

Crie um **ASContainer** (ex: `ArquetipoAldeao.tres`). Este é o seu "Blueprint de NPC":

- Atribua o `AttributeSet`.
- Adicione a habilidade `Conversar` ao catálogo.
- Adicione efeitos padrão (ex: um estado "Neutro").

### 7. Inicialize o Ator (ASC)

Adicione o nó `ASComponent` ao seu CharacterBody e atribua o **ASContainer** no Inspetor.
Registre seus nós de feedback (AnimationPlayer/Audio) via script para que o sistema possa disparar **Cues** automaticamente.

### 8. Combate vs Interação Social

O sistema lida com ambos de forma transparente. Aplicar um efeito "Congelado" irá bloquear automaticamente a habilidade "Conversar" se você adicionar `state.stun` à lista de tags bloqueadas da habilidade.

```gdscript

func interagir_com_npc(npc: ASComponent):
 if npc.can_activate_ability_by_tag(&"ability.social.talk"):
  npc.try_activate_ability_by_tag(&"ability.social.talk")
 else:
  print("O NPC está muito furioso ou atordoado para conversar!")
```

### 9. Reagindo a Mudanças de Estado (Sinais)

O ASC notifica sua lógica de jogo quando mudanças significativas ocorrem:

```gdscript

func _ready():
 asc.tag_changed.connect(_on_tag_changed)

func _on_tag_changed(tag: StringName, is_present: bool):
 if tag == &"state.emotional.angry" and is_present:
  $Sprite.modulate = Color.RED # Feedback visual para mudança de emoção
```

### 10. O Payload: ASPackage

O nó `ASDelivery` é a ponte entre suas habilidades e o impacto delas. Ele precisa de um "payload" (carga). Crie um recurso **ASPackage** (ex: `ExplosaoPayload.tres`).

- Adicione **Effects**: Dano instantâneo, queima por tempo, etc.
- Adicione **Cues**: Som de explosão, efeitos de partículas.

### 11. Entregas de Combate (3 Formas)

#### A. Ataques Melee (Area2D)

Tipicamente um Area2D anexado ao jogador que fica ativo apenas durante frames de impacto.

1. Adicione um `Area2D` ao seu Jogador (ex: `AreaEspada`).
2. Adicione um `ASDelivery` como filho.
3. **Configure**: `Package` = `DanoMelee.tres`, `Auto Connect` = **Ligado**.
4. **Dispare**:

   ```gdscript
   func _on_attack_animation_hit_frame():
       $AreaEspada/ASDelivery.activate(0.1) # Ativo por 100ms
   ```

#### B. Projéteis (Balas)

Um projétil carrega a intenção através da tela.

1. Crie uma cena de Bala (`Area2D` + `ASDelivery`).
2. **Configure**: `Auto Connect` = **Ligado**, `One Shot` = **Ligado**.
3. **Spawne**:

   ```gdscript
   func disparar():
       var bala = CenaBala.instantiate()
       get_parent().add_child(bala)
       # Diga à entrega quem a enviou para lógica de atributos (Força/Atacante)
       bala.get_node("ASDelivery").set_source_component(self.asc)
       bala.get_node("ASDelivery").activate()
   ```

#### C. Armadilhas e Perigos

Entregas passivas esperando por uma vítima.

1. Crie uma cena de Armadilha (`Area2D` + `ASDelivery`).
2. **Configure**: `Auto Connect` = **Ligado**, `Active` = **Ligado**, `Life Span` = **0**.
3. **Resultado**: Nenhum script necessário! Qualquer `ASComponent` que entrar na área recebe o pacote.

---

## 🌐 Guia de Multiplayer (Predição & Rollback)

O Ability System foi projetado para alta performance em rede utilizando uma arquitetura de **Rollback Baseado em Estado**. Siga estes passos para implementar habilidades sem latência perceptível.

### 1. Ative os Snapshots

Para personagens jogáveis, atribua um recurso `ASStateSnapshot` à propriedade `snapshot_state` do `ASComponent`. Isso serve como uma ponte persistente para o `ASStateCache` interno de alta velocidade.

### 2. O Loop de Predição (Cliente)

Os clientes devem prever a ativação de habilidades para eliminar a latência. Antes de executar uma ação preditiva, capture o estado atual.

```gdscript
func _physics_process(delta):
    if is_local_authority():
        # 1. Captura o estado do tick atual antes da predição
        asc.capture_snapshot()

        # 2. Ativa localmente de forma preditiva
        if Input.is_action_just_pressed("fire"):
            asc.try_activate_ability_by_tag(&"ability.bola_fogo")

            # 3. Informa o servidor sobre sua intenção
            asc.request_activate_ability.rpc_id(1, &"ability.bola_fogo")
```

### 3. Validação e Confirmação (Servidor)

O servidor atua como a autoridade. Ele recebe a requisição, valida custos/requisitos e notifica o cliente para confirmar ou corrigir o estado.

```gdscript
@rpc("any_peer", "call_local")
func request_activate_ability(tag: StringName):
    if is_server():
        if asc.can_activate_ability_by_tag(tag):
            asc.try_activate_ability_by_tag(tag)
            # Confirma ao cliente que a predição estava correta
            asc.confirm_ability_activation.rpc(tag)
        else:
            # Se for inválido, o servidor poderá enviar uma correção forçada
            pass
```

### 4. Gerenciamento de Rollback (Automático)

Se ocorrer um erro de predição (ex: o servidor diz que você estava atordoado quando tentou disparar), use `apply_snapshot(tick)` para reverter os atributos, tags e efeitos do ator para um estado válido anterior.

- **ASEffectState:** Rastreia internamente a duração e stacks de cada efeito.
- **ASStateCache:** Mantém automaticamente um buffer circular dos últimos 128 ticks para rollback quase instantâneo sem custo de alocação de memória.

### 5. Efeitos Determinísticos

Para garantir que servidor e cliente calculem o dano de forma idêntica:

- Use **Attributes** para todos os cálculos de gameplay.
- Evite usar `randf()` diretamente na lógica da habilidade; utilize um sistema de sementes (seed) ou deixe que o servidor dite resultados aleatórios via níveis de `ASPackage`.

## 📚 Referências

### 🧠 Componentes Principais

| Componente        | Propósito                  | Destaques                                                  |
| :---------------- | :------------------------- | :--------------------------------------------------------- |
| **AbilitySystem** | Coordenador Global.        | Registro central de Tags, integração com Project Settings. |
| **ASComponent**   | Processador do Ator (ASC). | Concede/Ativa habilidades, gerencia efeitos, dispara Cues. |
| **ASDelivery**    | Injetor de Física.         | Propaga cargas (ASPackage) via Area2D/Area3D naturalmente. |

### ⚙️ Recursos (Blueprints)

| Recurso             | Propósito               | Destaques                                                  |
| :------------------ | :---------------------- | :--------------------------------------------------------- |
| **ASAbility**       | Lógica de uma ação.     | Custos, Cooldowns e Tags de Ativação nativos.              |
| **ASAbilityPhase**  | Definição de Fase.      | Define duração e efeitos para estágios específicos.        |
| **ASEffect**        | Pacote de alteração.    | Dano instantâneo, buffs temporários ou passivas.           |
| **ASAttributeSet**  | Container de stats.     | Gerencia coleções de atributos. Único por ator.            |
| **ASAttribute**     | Definição de stat.      | Esquema individual de HP, Mana com limites (clamping).     |
| **ASContainer**     | Arquetipo / Blueprint.  | Catálogo de habilidades e efeitos permitidos.              |
| **ASPackage**       | Payload de Dados.       | Coleção de efeitos e cues para sistemas de entrega.        |
| **ASCue**           | Base de Feedback.       | Lógica de ativação de eventos audiovisuais.                |
| **ASCueAnimation**  | Feedback de Animação.   | Especializado em tocar montagens (montages) nos atores.    |
| **ASCueAudio**      | Feedback de Áudio.      | Especializado em tocar sons espaciais ou globais.          |
| **ASStateSnapshot** | Persistência de Estado. | Captura Atributos/Tags para Rollback Multiplayer ou Saves. |
| **ASEventTag**      | Definição de Evento.    | Define eventos customizados para interação entre sistemas. |

### 🚀 Objetos de Runtime (Specs)

| Spec              | Propósito                | Dados Chave                                                |
| :---------------- | :----------------------- | :--------------------------------------------------------- |
| **ASAbilitySpec** | Instância de habilidade. | Nível, estado ativo, consultas de cooldown/custo.          |
| **ASEffectSpec**  | Instância de efeito.     | Duração restante, stacks, posição (Variant).               |
| **ASCueSpec**     | Contexto de feedback.    | Magnitude, referências de origem/alvo, dados de impacto.   |
| **ASTagSpec**     | Gestão de Tags.          | Container de alta performance para tags de estado do ator. |

### 🛠️ Padrões de API e Acesso Seguro

A interface do Ability System prioriza segurança e previsibilidade através de convenções rigorosas de nomenclatura:

- **Instanciação Dual (`by_tag` vs `by_resource`)**: Cada método possui duas assinaturas. Use `by_tag` para ações dinâmicas ou `by_resource` para referências diretas.
- **Execução Segura (`try_activate`)**: O uso obrigatório de `try_...` garante que os custos e requisitos do Blueprint sejam respeitados antes da execução.
- **Autorização (`can_`)**: Avalia se uma ação é teoricamente permitida sem iniciá-la.
- **Consulta de Status (`is_` / `has_`)**: Verificações limpas para validar a posse de Tags ou estados operacionais.
- **Interrupção (`cancel_`)**: Encerra ativamente fluxos contínuos de habilidades ou efeitos.
- **Intenção em Rede (`request_`)**: Metodologia formal para solicitar execuções via RPC (Autoridade do Servidor).
- **Extração Segura (`get_`)**: Recupera valores e instâncias de forma limpa, sem expor mutabilidade desnecessária.
- **Gestão de Inventário (`unlock_` / `lock_`)**: Define o catálogo ativo de habilidades que podem ser equipadas pelo ator.

---

Desenvolvido com ❤️ por **MachiTwo**
