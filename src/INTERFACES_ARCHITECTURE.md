# 🏗️ Arquitetura de Átomos e Sidecars: Vectora

Este manifesto define a topologia final do ecossistema **Vectora**, onde a performance nativa de Go se une à interface rica de Next.js sob o teto de 2GB de RAM.

---

## 🏛️ 1. O Maestro Core (Go)

A base de toda a inteligência. O Core não possui interface própria, ele orquestra os dados.

- **Persistência:** `bbolt` (KV Store) para memória de curto e longo prazo.
- **RAG Engine:** `chromem-go` (Motor Vetorial Nativo) para busca semântica instantânea.
- **Sidecar Manager:** `llama.cpp` para inferência de alta performance.

---

## 🎨 2. Modo Studio (Wails + Go-app/Wasm)

A interface mestre e o ponto de entrada do usuário.

### **Papel: Gestão de Ativos e Configurações**
- **Tecnologia:** Go compilado para WebAssembly dentro do Wails.
- **Por que:** Baixíssimo consumo de RAM (~40MB), permitindo que ele rode sempre minimizado na bandeja.
- **Responsabilidades:**
    - Gerenciar a **Data Library** e downloads do **Backserver (Hono)**.
    - Instalar/Remover Modelos Qwen3.
    - Monitorar térmica e saúde dos sidecars.

---

## 🌐 3. Modo Chat (Wails Window + Next.js 16)

A interface de diagnóstico profundo e chat multimodal.

### **Papel: Visualização e Interação Rica**
- **Tecnologia:** Next.js servido como um módulo embutido no Wails.
- **Por que:** Para manter a experiência de design Premium (ShadcnUI) e visualizações de grafos complexas que exigem o ecossistema React.
- **Responsabilidades:**
    - Chat multimodal com suporte a Imagens, Código e Docs.
    - Visualização de grafos de arquitetura de nós (Godot/Unreal).
    - Análise de infraestrutura profunda.

---

## 🔌 4. Modo CLI (TUI - Bubbletea)

Interface de baixa latência focada em IA via terminal.

### **Papel: Velocidade e Integração IDE**
- **Tecnologia:** Go Puro com Bubbletea.
- **Por que:** Latência <10ms, ideal para uso dentro de terminais de IDEs sem carregar janelas visuais.

---

## 📦 5. O Backserver (Nuvem Edge - Hono)

O cérebro de metadados na Vercel que alimenta o ecossistema global.

- **V1 Registry:** Manifesto de modelos oficiais.
- **V1 Library:** Catálogo comunitário de pacotes de treinamento (.zpack).

---
_Documento gerado como base técnica para a versão 1.0 industrial._
