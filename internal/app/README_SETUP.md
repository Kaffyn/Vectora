# Vectora Frontend Web UI - Setup & Documentação

## Visão Geral

Frontend web completo do Vectora implementado com **Next.js 14**, **React 18**, **Zustand** e **TailwindCSS**.

Localização: `C:\Users\bruno\Desktop\Vectora\internal\app\`

---

## Instalação Rápida

### 1. Instalar Dependências
```bash
cd C:\Users\bruno\Desktop\Vectora\internal\app
npm install
# ou
yarn install
```

### 2. Rodar Dev Server
```bash
npm run dev
# ou
yarn dev
```

Acesse: http://localhost:3000

### 3. Build para Produção
```bash
npm run build
npm run start
```

---

## Estrutura do Projeto

```
internal/app/
├── app/                          # Next.js App Router
│   ├── layout.tsx               # Root layout com metadata
│   ├── page.tsx                 # Home page com layout principal
│   ├── globals.css              # Estilos globais
│   └── chat/
│       └── page.tsx             # Chat page
│
├── components/                   # Componentes React
│   ├── Common/
│   │   ├── Sidebar.tsx          # Navegação lateral (4 abas)
│   │   └── Header.tsx           # Header dinâmico
│   └── Chat/
│       ├── ChatFeed.tsx         # Feed de mensagens
│       ├── MessageBubble.tsx    # Bolha de mensagem
│       └── InputArea.tsx        # Área de input
│
├── store/                        # Zustand Stores
│   ├── uiStore.ts              # UI state (activeTab, loading)
│   └── chatStore.ts            # Chat state (messages)
│
├── package.json                 # Dependências npm
├── next.config.js              # Configuração Next.js
├── tsconfig.json               # Configuração TypeScript
├── tailwind.config.js          # Tema Tailwind
├── postcss.config.js           # Configuração PostCSS
└── .gitignore                  # Git ignore
```

---

## Componentes Principais

### 1. Sidebar
- Navegação com 4 abas: Chat, Código, Index, Manager
- Indicador de status "Daemon Conectado"
- Integrado com Zustand useUIStore
- Styling com TailwindCSS

### 2. Header
- Título dinâmico baseado na aba ativa
- Versão do app (v2.0)
- Responsive design

### 3. Chat Page
- Feed de mensagens com auto-scroll
- Input de texto com auto-grow de altura
- Simulação de resposta LLM
- Estados: loading, error (preparado)

### 4. Message Bubble
- Diferenciação visual user vs assistant
- Timestamp em português brasileiro
- Estilos customizados (verde para user, cinza para assistant)

---

## State Management (Zustand)

### UIStore
```typescript
interface UIState {
  activeTab: TabType           // 'chat' | 'codigo' | 'index' | 'manager'
  setActiveTab: (tab) => void
  loading: boolean
  setLoading: (loading) => void
}
```

### ChatStore
```typescript
interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
}

interface ChatState {
  messages: Message[]
  addMessage: (message: Message) => void
  clearMessages: () => void
}
```

---

## Desenvolvimento

### Adicionar Nova Aba
1. Atualizar `TabType` em `store/uiStore.ts`
2. Adicionar entrada em `TABS` no `components/Common/Sidebar.tsx`
3. Adicionar título em `TAB_TITLES` em `components/Common/Header.tsx`
4. Criar arquivo de página em `app/[nome]/page.tsx`

### Adicionar Novo Componente
```bash
# Criar em components/[categoria]/NomeComponente.tsx
'use client'

import React from 'react'

interface Props {
  // Define props
}

export default function NomeComponente({ ...props }: Props) {
  return (
    <div>
      {/* JSX */}
    </div>
  )
}
```

### Integrar com API Backend
```typescript
// Em components/Chat/InputArea.tsx
const handleSend = async () => {
  const response = await fetch('/api/chat', {
    method: 'POST',
    body: JSON.stringify({ message: value })
  })
  const data = await response.json()
  // Handle response
}
```

---

## Estilo & Tema

### Cores Implementadas
- **Primary**: Emerald 600 (user messages)
- **Background**: Zinc 950 (main bg)
- **Surface**: Zinc 900 (sidebars)
- **Border**: Zinc 800
- **Text**: Zinc 50 (light text)
- **Status**: Emerald 500 (indicators)

### Customizar Tema
Editar `tailwind.config.js`:
```javascript
theme: {
  extend: {
    colors: {
      // Adicionar cores customizadas
    },
  },
}
```

---

## Comandos Disponíveis

```bash
npm run dev          # Dev server (hot reload)
npm run build        # Build para produção
npm run start        # Rodar server em produção
npm run lint         # ESLint
npm test             # Rodar testes Jest
npm test:watch       # Jest em watch mode
```

---

## Configurações Importantes

### Next.js (`next.config.js`)
- `output: 'export'` - Static export (sem Node.js runtime)
- `reactStrictMode: true` - Detecta problemas em desenvolvimento
- `images.unoptimized: true` - Desativa Image Optimization

### TypeScript (`tsconfig.json`)
- `strict: true` - Modo strict ativado
- Path alias `@/*` para imports absolutos
- Target: ES2020

### Tailwind (`tailwind.config.js`)
- Content paths configuradas para `app/**` e `components/**`
- Cores customizadas em `theme.extend.colors`

---

## Próximas Implementações

### Curto Prazo
- [ ] Implementar páginas das abas (Código, Index, Manager)
- [ ] Conectar com API backend
- [ ] Adicionar autenticação/logout
- [ ] Implementar persistência de chat (localStorage)

### Médio Prazo
- [ ] Rich text editor para código
- [ ] Syntax highlighting com Monaco Editor
- [ ] File upload/management
- [ ] Dark mode toggle
- [ ] Localization (i18n)

### Longo Prazo
- [ ] Real-time updates (WebSocket)
- [ ] Collaborative features
- [ ] Analytics
- [ ] PWA capabilities
- [ ] Offline support

---

## Troubleshooting

### Problema: Module not found
**Solução**: Verifique o `tsconfig.json` e paths aliases (`@/*`)

### Problema: Tailwind styles não aplicando
**Solução**: Verifique `tailwind.config.js` content paths

### Problema: Hydration mismatch
**Solução**: Use `useEffect` e `useState` para componentes interativos

### Problema: Build falha
**Solução**: Verifique `typescript` e `typings` corretos

---

## Referências

- [Next.js 14 Docs](https://nextjs.org/docs)
- [React 18 Docs](https://react.dev)
- [Zustand](https://github.com/pmndrs/zustand)
- [TailwindCSS](https://tailwindcss.com)
- [Lucide Icons](https://lucide.dev)

---

## Licença

Mesmo projeto Vectora (Veja root LICENSE)

---

## Suporte

Para issues ou dúvidas, abra uma issue no repositório principal.

