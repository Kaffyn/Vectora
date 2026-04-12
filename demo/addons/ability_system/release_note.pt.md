# 🚀 Notas de Lançamento - v0.2.0 (A Atualização de Integração Atómica)

> [!TIP]
> **Read this in other languages / Leia isto em outros idiomas:**
> [**English**](release_note.md) | [**Português**](release_note.pt.md)

Este lançamento marca o **nascimento pioneiro** da integração entre o **Ability System** e o **LimboAI**. Construímos do zero uma arquitectura atómica onde ambos os sistemas coexistem no mesmo binário, proporcionando uma performance e estabilidade sem precedentes para o ecossistema Zyris Engine.

---

## 🏗️ 1. Integração Orquestrada: LimboAI como Biblioteca do Framework

- **O Nascimento da AS Bridge**: Introduzimos oficialmente a ponte entre o LimboAI e o Ability System. Não se trata de uma simples ligação; é uma integração nativa e atómica onde as Behavior Trees podem "falar" de forma formal e sem latência com o Ability System.
- **Sinergia HSM**: O `ASComponent` é agora totalmente compatível com **LimboHSM**, agindo como o executor de gameplay para a inteligência de estados do LimboAI.

## 🔗 2. ASBridge: A Suite de Nós de IA Nativa

A v0.2.0 introduz uma coleção completa de nós para Behavior Trees e HSM, permitindo que a IA sinta e manipule o sistema de habilidades:

### **BT Actions (Ações de Comportamento)**:

- **`ASActivateAbility`**: Permite que a IA dispare habilidades diretamente por Tag, respeitando custos e requisitos.
- **`ASDispatchEvent`**: Permite que a lógica da árvore de comportamento dispare Event Tags de gameplay manualmente.
- **`ASWaitEvent`**: Um nó de estado reativo que aguarda a ocorrência de uma Event Tag específica dentro de uma janela temporal, utilizando a nova memória de ticks do ASC.

### **BT Conditions (Sensores de IA)**:

- **`ASCanActivateAbility`**: Sensor de prontidão para verificar custos e cooldowns antes de tentar uma ação.
- **`ASEventOccurred`**: Verifica se um evento específico aconteceu recentemente (nos últimos frames/segundos).
- **`ASHasTag`**: Consulta direta ao estado do ator via Tags `CONDITIONAL` ou `NAME`.

### **LimboHSM Integration**:

- **`ASBridgeState`**: Nós de estado especializados que sincronizam o estado da máquina de estados (ex: Atacando, Atordoado) com o status interno do Ability System de forma bidirecional.

## 🏷️ 3. ASComponent v0.2.0: O Maestro de Fases (HSM Nativa)

- **Motor de Fases (Ability Phasing)**: O ASC agora gere nativamente o ciclo de vida de **Habilidades Multi-Fase** (ex: Prelúdio, Ação, Recuperação), garantindo transições determinísticas tick-a-tick.
- **HSM Standalone**: Mesmo sem o LimboAI, o ASC atua como uma máquina de estados de gameplay autossuficiente para a lógica de habilidades.
- **Segurança de Tags (ASTagTypes)**: Tipagem estrita para garantir a integridade dos estados (`NAME`, `CONDITIONAL`, `EVENT`).

## ⚡ 4. AS Event Tags: O Sistema de Eventos Temporais

- **Frame-Level Transience**: As Event Tags (ex: `Event.Damage.Block`) são ocorrências instantâneas que não poluem o estado persistente do ator.
- **Memória Reativa**: O ASC agora possui memória temporal via `ASEventTagHistorical`, permitindo que o ator (e a IA via Bridge) reaja a eventos passados recentes.

## 🛠️ 5. Refatoração em Massa e Modularização

- **`ASUtils` & `ASTagTypes`**: Reorganização total da lógica estática em classes de utilitário dedicadas para performance industrial.
- **Unificação de Registro**: Centralização de tipos e validações, garantindo um framework robusto e otimizado para a Zyris Engine.

## 🎮 6. Novo Projeto Demo Funcional

Foi adicionado um projeto demo simples, mas funcional:

- **Implementação de Referência**: Demonstra movimento de personagem, ativação de habilidades e behavior trees do LimboAI operando num ambiente atómico unificado.
- **Showcase**: Veja o `ASDelivery` e o `ASPackage` a trabalhar em conjunto com os novos nós de ponte (`AS Bridge`).
