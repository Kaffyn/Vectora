# Vectora Frontend - Quick Start

## 1 Minuto Setup

```bash
# 1. Instalar dependências
npm install

# 2. Rodar dev server
npm run dev

# 3. Abrir no browser
# http://localhost:3000
```

## Estrutura Rápida

```
app/                  - Next.js pages
  layout.tsx         - Root layout
  page.tsx           - Home
  chat/page.tsx      - Chat

components/
  Common/            - Sidebar, Header
  Chat/              - ChatFeed, InputArea, MessageBubble

store/               - Zustand stores
  uiStore.ts        - Tab state
  chatStore.ts      - Messages state
```

## Componentes Principais

- **Sidebar** - 4 abas (Chat, Código, Index, Manager)
- **Header** - Título dinâmico
- **ChatFeed** - Feed de mensagens com auto-scroll
- **InputArea** - Textarea com auto-grow
- **MessageBubble** - Styling de mensagens

## Customizações Rápidas

### Adicionar Aba
1. `store/uiStore.ts` - Adicionar tipo
2. `components/Common/Sidebar.tsx` - Adicionar ao TABS
3. `components/Common/Header.tsx` - Adicionar título
4. `app/[nome]/page.tsx` - Criar página

### Mudar Tema
Editar `tailwind.config.js`:
```javascript
colors: {
  primary: '#...',
  // etc
}
```

### Integrar API
Em `app/chat/page.tsx`:
```typescript
const response = await fetch('/api/chat', {
  method: 'POST',
  body: JSON.stringify({ message })
})
```

## Arquivos Importantes

- `tsconfig.json` - Path aliases (@/*)
- `tailwind.config.js` - Tema dark
- `next.config.js` - Output export
- `package.json` - Dependencies

## Commands

```
npm run dev       - Dev server
npm run build     - Build
npm run start     - Production
npm run lint      - Lint
npm test          - Tests
```

## Documentação Completa

- `README_SETUP.md` - Guia detalhado
- `IMPLEMENTATION_STATUS.md` - Status implementação

---

Mais informações: veja `README_SETUP.md`

