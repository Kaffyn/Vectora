# 🚀 Release Notes - v0.2.0 (The Atomic Integration Update)

> [!TIP]
> **Read this in other languages / Leia isto em outros idiomas:**
> [**English**](release_note.md) | [**Português**](release_note.pt.md)

This release marks the **pioneer birth** of the integration between **Ability System** and **LimboAI**. We've built from scratch an atomic architecture where both systems coexist in the same binary, providing unprecedented performance and stability for the Zyris Engine ecosystem.

---

## 🏗️ 1. Orchestrated Integration: LimboAI as a Framework Library

- **The Birth of AS Bridge**: We've introduced the first official bridge between LimboAI and the Ability System. This is a native and atomic integration where Behavior Trees can formally "speak" to the Ability System with zero latency.
- **HSM Synergy**: `ASComponent` is now fully compatible with **LimboHSM**, acting as the gameplay executor for LimboAI's state-driven intelligence.

## 🔗 2. ASBridge: The Native AI Node Suite

v0.2.0 introduces a comprehensive collection of nodes for Behavior Trees and HSM, allowing AI to perceive and interact with the Ability System:

### **BT Actions (Behavioral Actions)**:

- **`ASActivateAbility`**: Allows AI to trigger abilities directly by Tag, adhering to costs and requirements.
- **`ASDispatchEvent`**: Allows behavioral logic to manually dispatch gameplay Event Tags.
- **`ASWaitEvent`**: A reactive state node that waits for a specific Event Tag within a temporal window, leveraging the new tick-based memory of the ASC.

### **BT Conditions (AI Sensors)**:

- **`ASCanActivateAbility`**: Pre-check sensor for readiness, costs, and cooldowns before attempting an action.
- **`ASEventOccurred`**: Checks if a specific event occurred recently (within the last few frames/seconds).
- **`ASHasTag`**: Direct query of the actor's state via `CONDITIONAL` or `NAME` tags.

### **LimboHSM Integration**:

- **`ASBridgeState`**: Specialized state nodes that bidirectionally synchronize the state machine state (e.g., Attacking, Stunned) with the Ability System's internal status.

## 🏷️ 3. ASComponent v0.2.0: The Phase Maestro (Native HSM)

- **Phase-Engine (Ability Phasing)**: The ASC now natively manages the lifecycle of **Multi-Phase Abilities** (e.g., Prelude, Action, Recovery), ensuring deterministic frame-by-frame transitions.
- **Standalone HSM**: Even without LimboAI, the ASC acts as a gameplay state machine for ability logic.
- **Tag Security & Governance**: Strict typing rules to ensure state integrity (`NAME`, `CONDITIONAL`, `EVENT`).

## ⚡ 4. AS Event Tags: High-Performance Temporal Events

- **Frame-Level Transience**: Event Tags (e.g., `Event.Damage.Block`) are instantaneous occurrences that do not persist in the actor's state.
- **Reactive Memory**: `ASComponent` now possesses temporal memory via `ASEventTagHistorical`, allowing actors (and AI via Bridge) to react to past events from previous ticks.

## 🛠️ 5. Mass Refactoring & Modularization

- **`ASUtils` & `ASTagTypes`**: Total reorganization of static logic and type safety into dedicated utility classes.
- **Registry Unification**: Centralized types and validations, making the framework robust and optimized for the Zyris Engine.

## 🎮 6. New Functional Demo

A new, simple but functional demo project has been added:

- **Reference Implementation**: Demonstrates character movement, ability activation, and LimboAI behavior trees operating in a unified atomic environment.
- **Showcase**: See `ASDelivery` and `ASPackage` working together with the new bridge nodes (`AS Bridge`).
