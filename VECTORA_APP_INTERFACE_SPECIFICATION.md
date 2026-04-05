# VECTORA APP INTERFACE SPECIFICATION

**Status:** Especificação Completa — Pronto para Implementação
**Versão:** 1.0
**Data:** 2026-04-05
**Idioma:** Português (PT-BR)
**Escopo:** Interface Web/Desktop da Aplicação Vectora com 4 Abas Principais

---

## ÍNDICE

1. [Visão Geral](#1-visão-geral)
2. [Arquitetura da Aplicação](#2-arquitetura-da-aplicação)
3. [Abas Principais](#3-abas-principais)
4. [State Management](#4-state-management)
5. [Integração IPC com Daemon](#5-integração-ipc-com-daemon)
6. [Componentes Reutilizáveis](#6-componentes-reutilizáveis)
7. [Design System](#7-design-system)
8. [Tratamento de Erros](#8-tratamento-de-erros)
9. [Performance e Otimizações](#9-performance-e-otimizações)
10. [Testes](#10-testes)

---

## 1. VISÃO GERAL

### 1.1 Objetivo

A Interface Vectora App é uma aplicação desktop moderna (Wails + Next.js) que fornece acesso visual a todas as capacidades do Vectora:
- **Chat com IA** baseado em RAG (Retrieval-Augmented Generation)
- **Explorador de Código** com editor integrado e acesso a arquivos
- **Gerenciador de Index/Workspaces** com visualização de datasets
- **Painel de Gerenciamento** de pacotes (MPM/LPM) e configurações

### 1.2 Premissas de Design

```
RN-APP-01: Abas mutualmente exclusivas — 1 aba ativa por sessão
RN-APP-02: Responsividade: Desktop only (1280x800 minimum)
RN-APP-03: Dark mode obrigatório (Kaffyn Zinc palette)
RN-APP-04: Latência < 100ms para UI feedback
RN-APP-05: Zero dados sensíveis em localStorage (apenas IDs e preferences)
RN-APP-06: Todos os dados persistem via daemon (bbolt)
RN-APP-07: Streaming de respostas LLM (não aguardar conclusão)
RN-APP-08: Undo/Redo de operações de arquivo via GitBridge
```

### 1.3 Stack Tecnológico

| Camada | Tecnologia | Versão | Justificativa |
|--------|-----------|--------|--------------|
| Framework Desktop | Wails | v3 | Lightweight, native performance |
| Frontend Framework | Next.js | 14 | SSG, routing, convenções |
| Styling | TailwindCSS | 3.4 | Utility-first, dark mode |
| Componentes | Shadcn/UI | Latest | Acessibilidade, theming |
| State Management | Zustand | 4.4 | Minimalista, sem boilerplate |
| HTTP Client | TanStack Query | 5.0 | Caching, refetching automático |
| Animações | Framer Motion | 10.x | Micro-interações fluidas |
| Code Editor | Monaco | Latest | Syntax highlighting, completude |
| Icons | Lucide React | Latest | Design system integrado |
| Markdown | react-markdown | 9.0 | Renderização de respostas LLM |
| Escrita de Arquivos | editor-js | Latest | Suporte a blocos (Future) |

---

## 2. ARQUITETURA DA APLICAÇÃO

### 2.1 Estrutura de Diretórios

```
internal/app/
├── app/
│   ├── (main)/
│   │   ├── layout.tsx              # Root layout, sidebar, tabs
│   │   ├── page.tsx                # Aba ativa padrão (chat)
│   │   └── error.tsx               # Fallback de erro global
│   ├── chat/
│   │   ├── page.tsx                # Tela da aba Chat
│   │   └── layout.tsx
│   ├── codigo/
│   │   ├── page.tsx                # Tela da aba Código
│   │   └── layout.tsx
│   ├── index/
│   │   ├── page.tsx                # Tela da aba Index
│   │   └── layout.tsx
│   └── manager/
│       ├── page.tsx                # Tela da aba Manager
│       └── layout.tsx
├── components/
│   ├── Chat/
│   │   ├── ChatFeed.tsx            # Histórico de mensagens
│   │   ├── InputArea.tsx           # Textarea + envio
│   │   ├── MessageBubble.tsx       # Mensagem individual
│   │   ├── ToolCallVisualizer.tsx  # Mostrar tool calls
│   │   └── SourceBadge.tsx         # Badge de fonte (RAG)
│   ├── Codigo/
│   │   ├── FileTree.tsx            # Árvore de diretórios
│   │   ├── FileViewer.tsx          # Visualizador de arquivo
│   │   ├── CodeEditor.tsx          # Editor Monaco
│   │   ├── Terminal.tsx            # Execução de shell
│   │   └── DiffViewer.tsx          # Visualizador de diff
│   ├── Index/
│   │   ├── WorkspaceList.tsx       # Lista de workspaces
│   │   ├── WorkspaceDetail.tsx     # Detalhe + chunks
│   │   ├── WorkspaceUploader.tsx   # Upload de arquivos
│   │   └── DatasetBrowser.tsx      # Browse do Index
│   ├── Manager/
│   │   ├── PackageManagerTabs.tsx  # Abas LPM/MPM
│   │   ├── LPMPanel.tsx            # Gerenciador de builds
│   │   ├── MPMPanel.tsx            # Gerenciador de modelos
│   │   ├── ProgressMonitor.tsx     # Progresso de downloads
│   │   └── ConfigurationPanel.tsx  # Configurações globais
│   ├── Common/
│   │   ├── Sidebar.tsx             # Navegação de abas
│   │   ├── Header.tsx              # Título + info
│   │   ├── Modal.tsx               # Dialog wrapper
│   │   ├── Toast.tsx               # Notificações
│   │   ├── SkeletonLoader.tsx      # Placeholder de loading
│   │   └── ErrorBoundary.tsx       # Catch de erros
│   └── UI/
│       ├── Button.tsx
│       ├── Input.tsx
│       ├── Select.tsx
│       ├── Progress.tsx
│       ├── Tabs.tsx
│       └── ... (primitivos Shadcn)
├── hooks/
│   ├── useIPC.ts                   # Cliente IPC para daemon
│   ├── useChatHistory.ts           # Carrega histórico de chat
│   ├── useWorkspace.ts             # CRUD de workspaces
│   ├── useFileTree.ts              # Árvore de diretórios
│   ├── useCodeEditor.ts            # Estado do editor
│   ├── usePackageManager.ts        # LPM/MPM via IPC
│   ├── useAsyncOperation.ts        # Loading/Error states
│   └── useStreamingResponse.ts     # Streaming de LLM
├── store/
│   ├── appStore.ts                 # Zustand root store
│   ├── chatStore.ts                # Estado da aba Chat
│   ├── codigoStore.ts              # Estado da aba Código
│   ├── indexStore.ts               # Estado da aba Index
│   ├── managerStore.ts             # Estado da aba Manager
│   └── uiStore.ts                  # Modal abertos, temas, etc
├── services/
│   ├── ipc/
│   │   ├── client.ts               # IPC client com retry
│   │   ├── types.ts                # Tipos IPC
│   │   └── eventEmitter.ts         # Subscriptions a eventos
│   ├── workspace/
│   │   ├── api.ts                  # Chamadas workspace.* IPC
│   │   └── types.ts
│   ├── code/
│   │   ├── api.ts                  # Chamadas tool.* IPC
│   │   └── types.ts
│   └── package/
│       ├── api.ts                  # Chamadas package.* IPC
│       └── types.ts
├── utils/
│   ├── formatters.ts               # formatDate, formatBytes, etc
│   ├── markdown.ts                 # Parse markdown seguro
│   ├── syntax.ts                   # Detectar linguagem de código
│   ├── validators.ts               # Validação de input
│   └── constants.ts                # Limites, timeouts, etc
├── styles/
│   ├── globals.css                 # TailwindCSS imports
│   ├── variables.css               # CSS custom properties (cores, fonts)
│   └── animations.css              # Keyframes custom
├── public/
│   ├── logo.svg                    # Logo Vectora
│   └── favicons/
├── next.config.js
├── tailwind.config.js
├── tsconfig.json
└── package.json
```

### 2.2 Layout Principal (Root Layout)

**Arquivo:** `app/(main)/layout.tsx`

```typescript
'use client';

import { ReactNode } from 'react';
import Sidebar from '@/components/Common/Sidebar';
import Header from '@/components/Common/Header';
import { useUIStore } from '@/store/uiStore';
import { Toaster } from '@/components/Common/Toast';

interface MainLayoutProps {
  children: ReactNode;
}

export default function MainLayout({ children }: MainLayoutProps) {
  const { activeTab } = useUIStore();

  return (
    <div className="flex h-screen bg-zinc-950 text-zinc-50">
      {/* Sidebar com abas */}
      <Sidebar />

      {/* Área principal */}
      <div className="flex-1 flex flex-col">
        {/* Header com contexto */}
        <Header />

        {/* Conteúdo da aba ativa */}
        <main className="flex-1 overflow-hidden">
          {children}
        </main>
      </div>

      {/* Notificações Toast */}
      <Toaster />
    </div>
  );
}
```

### 2.3 Sidebar Navigation

**Arquivo:** `components/Common/Sidebar.tsx`

```typescript
'use client';

import { useUIStore } from '@/store/uiStore';
import { useRouter } from 'next/navigation';
import { MessageSquare, Code2, Database, Settings } from 'lucide-react';

type TabType = 'chat' | 'codigo' | 'index' | 'manager';

const TABS = [
  { id: 'chat', label: 'Chat', icon: MessageSquare, href: '/chat' },
  { id: 'codigo', label: 'Código', icon: Code2, href: '/codigo' },
  { id: 'index', label: 'Index', icon: Database, href: '/index' },
  { id: 'manager', label: 'Manager', icon: Settings, href: '/manager' },
];

export default function Sidebar() {
  const { activeTab, setActiveTab } = useUIStore();
  const router = useRouter();

  const handleTabClick = (tabId: TabType) => {
    setActiveTab(tabId);
    const tab = TABS.find(t => t.id === tabId);
    router.push(tab?.href || '/chat');
  };

  return (
    <aside className="w-64 bg-zinc-900 border-r border-zinc-800 flex flex-col">
      {/* Logo */}
      <div className="p-4 border-b border-zinc-800">
        <h1 className="text-xl font-bold text-emerald-500">Vectora</h1>
      </div>

      {/* Navegação */}
      <nav className="flex-1 p-4 space-y-2">
        {TABS.map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;

          return (
            <button
              key={tab.id}
              onClick={() => handleTabClick(tab.id as TabType)}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                isActive
                  ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/50'
                  : 'text-zinc-400 hover:text-zinc-300 hover:bg-zinc-800'
              }`}
            >
              <Icon className="w-5 h-5" />
              <span className="font-medium">{tab.label}</span>
            </button>
          );
        })}
      </nav>

      {/* Footer com status */}
      <div className="border-t border-zinc-800 p-4 space-y-2">
        <div className="flex items-center gap-2 text-xs text-zinc-500">
          <div className="w-2 h-2 bg-emerald-500 rounded-full" />
          <span>Daemon Conectado</span>
        </div>
      </div>
    </aside>
  );
}
```

---

## 3. ABAS PRINCIPAIS

### 3.1 ABA CHAT (Melhorada)

#### 3.1.1 Visão Geral

A aba Chat é a interface principal para interação com o Vectora. Mantém as funcionalidades existentes e adiciona melhorias de UX:

- **Histórico Persistente:** Salvo no daemon via bbolt
- **Seleção de Workspace:** Dropdown para escolher contexto
- **Streaming de Respostas:** Resposta aparece em tempo real
- **Tool Calls Visualizados:** Mostra ferramentas que o LLM executou
- **Fontes de Resposta:** Badges indicando chunks RAG usados

#### 3.1.2 Componente Principal

**Arquivo:** `app/chat/page.tsx`

```typescript
'use client';

import { useState } from 'react';
import ChatFeed from '@/components/Chat/ChatFeed';
import InputArea from '@/components/Chat/InputArea';
import { useWorkspace } from '@/hooks/useWorkspace';
import { useChat } from '@/hooks/useChatHistory';
import { useStreamingResponse } from '@/hooks/useStreamingResponse';
import { useChatStore } from '@/store/chatStore';

export default function ChatPage() {
  const { workspaces, activeWorkspace, setActiveWorkspace } = useWorkspace();
  const { messages, addMessage, loading } = useChat(activeWorkspace?.id);
  const { streamMessage, isStreaming } = useStreamingResponse();
  const { setMessages } = useChatStore();

  const handleSendQuery = async (query: string) => {
    if (!activeWorkspace) {
      alert('Selecione um workspace');
      return;
    }

    // Adiciona mensagem do usuário
    addMessage({
      role: 'user',
      content: query,
      id: crypto.randomUUID(),
      timestamp: new Date(),
    });

    try {
      // Stream da resposta
      await streamMessage(
        activeWorkspace.id,
        query,
        (chunk: string) => {
          // Atualizar último message do assistant
          setMessages((msgs) => {
            const last = msgs[msgs.length - 1];
            if (last?.role === 'assistant') {
              last.content += chunk;
            }
            return [...msgs];
          });
        },
        (sources: SourceReference[]) => {
          // Salvar fontes do mensaje final
        }
      );
    } catch (error) {
      addMessage({
        role: 'system',
        content: `Erro: ${error instanceof Error ? error.message : 'Desconhecido'}`,
        id: crypto.randomUUID(),
        timestamp: new Date(),
      });
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header com seleção de workspace */}
      <div className="px-6 py-4 border-b border-zinc-800 flex items-center justify-between">
        <h2 className="text-lg font-semibold">Chat</h2>
        <select
          value={activeWorkspace?.id || ''}
          onChange={(e) => {
            const ws = workspaces.find(w => w.id === e.target.value);
            if (ws) setActiveWorkspace(ws);
          }}
          className="px-3 py-2 rounded bg-zinc-800 text-zinc-50 text-sm border border-zinc-700"
        >
          <option value="">-- Selecione um workspace --</option>
          {workspaces.map((ws) => (
            <option key={ws.id} value={ws.id}>
              {ws.name} ({ws.chunkCount} chunks)
            </option>
          ))}
        </select>
      </div>

      {/* Feed de mensagens */}
      <ChatFeed messages={messages} isLoading={isStreaming} />

      {/* Área de input */}
      <InputArea
        onSend={handleSendQuery}
        disabled={!activeWorkspace || isStreaming}
        placeholder="Faça uma pergunta sobre seu workspace..."
      />
    </div>
  );
}
```

#### 3.1.3 ChatFeed Component

**Arquivo:** `components/Chat/ChatFeed.tsx`

```typescript
'use client';

import { useEffect, useRef } from 'react';
import { MessageBubble } from './MessageBubble';
import { SkeletonLoader } from '@/components/Common/SkeletonLoader';
import type { Message } from '@/services/ipc/types';

interface ChatFeedProps {
  messages: Message[];
  isLoading: boolean;
}

export default function ChatFeed({ messages, isLoading }: ChatFeedProps) {
  const scrollRef = useRef<HTMLDivElement>(null);

  // Auto-scroll para o final
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  return (
    <div
      ref={scrollRef}
      className="flex-1 overflow-y-auto p-6 space-y-4"
    >
      {messages.length === 0 ? (
        <div className="flex items-center justify-center h-full text-zinc-500">
          <p>Nenhuma mensagem ainda. Comece digitando.</p>
        </div>
      ) : (
        <>
          {messages.map((msg) => (
            <MessageBubble key={msg.id} message={msg} />
          ))}
          {isLoading && (
            <SkeletonLoader type="message" count={1} />
          )}
        </>
      )}
    </div>
  );
}
```

#### 3.1.4 MessageBubble com Markdown e Tool Calls

**Arquivo:** `components/Chat/MessageBubble.tsx`

```typescript
'use client';

import ReactMarkdown from 'react-markdown';
import { ToolCallVisualizer } from './ToolCallVisualizer';
import { SourceBadge } from './SourceBadge';
import type { Message } from '@/services/ipc/types';

interface MessageBubbleProps {
  message: Message;
}

export function MessageBubble({ message }: MessageBubbleProps) {
  const isUser = message.role === 'user';
  const isSystem = message.role === 'system';

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'}`}>
      <div
        className={`max-w-2xl rounded-lg px-4 py-3 ${
          isUser
            ? 'bg-emerald-600 text-white'
            : isSystem
            ? 'bg-red-900/50 text-red-200 border border-red-800'
            : 'bg-zinc-800 text-zinc-50'
        }`}
      >
        {/* Conteúdo Markdown */}
        <ReactMarkdown
          components={{
            code: ({ node, inline, className, children, ...props }) => (
              <code
                className={
                  inline
                    ? 'bg-zinc-700 px-2 py-1 rounded text-sm'
                    : 'block bg-zinc-900 p-3 rounded text-sm overflow-x-auto'
                }
                {...props}
              >
                {children}
              </code>
            ),
            a: ({ node, ...props }) => (
              <a
                className="text-emerald-400 underline hover:text-emerald-300"
                {...props}
              />
            ),
          }}
        >
          {message.content}
        </ReactMarkdown>

        {/* Tool Calls (se houver) */}
        {message.metadata?.toolCalls?.length > 0 && (
          <ToolCallVisualizer toolCalls={message.metadata.toolCalls} />
        )}

        {/* Fontes RAG (se houver) */}
        {message.metadata?.sources?.length > 0 && (
          <div className="mt-3 flex flex-wrap gap-2">
            {message.metadata.sources.map((source) => (
              <SourceBadge key={source.chunkId} source={source} />
            ))}
          </div>
        )}

        {/* Timestamp */}
        <div className="text-xs opacity-60 mt-2">
          {new Date(message.timestamp).toLocaleTimeString('pt-BR')}
        </div>
      </div>
    </div>
  );
}
```

#### 3.1.5 InputArea com Auto-grow

**Arquivo:** `components/Chat/InputArea.tsx`

```typescript
'use client';

import { useRef, useState } from 'react';
import { Send } from 'lucide-react';

interface InputAreaProps {
  onSend: (query: string) => void;
  disabled?: boolean;
  placeholder?: string;
}

export default function InputArea({
  onSend,
  disabled = false,
  placeholder = 'Escreva sua pergunta...',
}: InputAreaProps) {
  const [value, setValue] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Shift+Enter = nova linha, Enter = enviar
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleSend = () => {
    if (!value.trim() || disabled) return;
    onSend(value);
    setValue('');
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setValue(e.target.value);
    // Auto-grow textarea
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`;
    }
  };

  return (
    <div className="border-t border-zinc-800 p-6 bg-zinc-900">
      <div className="flex gap-3 items-end">
        <textarea
          ref={textareaRef}
          value={value}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          disabled={disabled}
          className="flex-1 px-4 py-3 rounded-lg bg-zinc-800 text-zinc-50 placeholder-zinc-500 border border-zinc-700 resize-none focus:outline-none focus:ring-2 focus:ring-emerald-500"
          rows={1}
        />
        <button
          onClick={handleSend}
          disabled={!value.trim() || disabled}
          className="px-4 py-3 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <Send className="w-5 h-5" />
        </button>
      </div>
      <p className="text-xs text-zinc-500 mt-2">
        Dica: Shift+Enter para nova linha, Enter para enviar
      </p>
    </div>
  );
}
```

---

### 3.2 ABA CÓDIGO (Novo)

#### 3.2.1 Visão Geral

A aba Código fornece acesso direto ao sistema de arquivos do usuário com:

- **File Tree:** Explorador de diretórios com ícones por tipo
- **Code Editor:** Monaco com syntax highlighting
- **Terminal Integrado:** Executar comandos shell
- **Diff Viewer:** Visualizar mudanças antes/depois
- **GitBridge Undo:** Reverter alterações de arquivo

#### 3.2.2 Página Principal

**Arquivo:** `app/codigo/page.tsx`

```typescript
'use client';

import { useState } from 'react';
import FileTree from '@/components/Codigo/FileTree';
import FileViewer from '@/components/Codigo/FileViewer';
import CodeEditor from '@/components/Codigo/CodeEditor';
import Terminal from '@/components/Codigo/Terminal';
import DiffViewer from '@/components/Codigo/DiffViewer';
import { useCodeEditor } from '@/hooks/useCodeEditor';
import { useCodigoStore } from '@/store/codigoStore';

export default function CodigoPage() {
  const {
    selectedFile,
    setSelectedFile,
    openFiles,
    closeFile,
    isDirty,
    content,
    setContent,
  } = useCodeEditor();

  const {
    showTerminal,
    setShowTerminal,
    showDiff,
    setShowDiff,
    diffContent,
  } = useCodigoStore();

  const handleSaveFile = async () => {
    if (!selectedFile) return;
    try {
      await window.go.main.App.CallIPC('tool.execute', JSON.stringify({
        tool: 'write_file',
        args: {
          path: selectedFile.path,
          content,
        },
      }));
      setContent(content, false); // Mark as saved
    } catch (error) {
      console.error('Erro ao salvar arquivo:', error);
    }
  };

  return (
    <div className="flex h-full gap-0">
      {/* Painel esquerdo: File Tree */}
      <div className="w-64 border-r border-zinc-800 overflow-y-auto bg-zinc-900">
        <FileTree onSelectFile={setSelectedFile} />
      </div>

      {/* Painel central: Editor */}
      <div className="flex-1 flex flex-col">
        {/* Abas de arquivos abertos */}
        {openFiles.length > 0 && (
          <div className="flex border-b border-zinc-800 bg-zinc-900 overflow-x-auto">
            {openFiles.map((file) => (
              <div
                key={file.id}
                onClick={() => setSelectedFile(file)}
                className={`px-4 py-2 cursor-pointer flex items-center gap-2 border-r border-zinc-800 ${
                  selectedFile?.id === file.id
                    ? 'bg-zinc-800 border-b-2 border-emerald-500'
                    : 'bg-zinc-900 hover:bg-zinc-800'
                }`}
              >
                <span className="text-sm">{file.name}</span>
                {isDirty && selectedFile?.id === file.id && (
                  <span className="text-emerald-500">●</span>
                )}
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    closeFile(file.id);
                  }}
                  className="text-zinc-500 hover:text-zinc-300"
                >
                  ✕
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Editor ou mensagem vazia */}
        {selectedFile ? (
          <>
            <div className="flex-1">
              <CodeEditor
                file={selectedFile}
                content={content}
                onChange={setContent}
              />
            </div>

            {/* Toolbar com botões */}
            <div className="border-t border-zinc-800 p-4 flex gap-2 bg-zinc-900">
              <button
                onClick={handleSaveFile}
                disabled={!isDirty}
                className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded disabled:opacity-50 text-sm"
              >
                {isDirty ? 'Salvar' : 'Salvo'}
              </button>
              <button
                onClick={() => setShowDiff(!showDiff)}
                className="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 text-white rounded text-sm"
              >
                {showDiff ? 'Ocultar' : 'Ver'} Diff
              </button>
              <button
                onClick={() => setShowTerminal(!showTerminal)}
                className="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 text-white rounded text-sm"
              >
                {showTerminal ? 'Ocultar' : 'Mostrar'} Terminal
              </button>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-zinc-500">
            <p>Selecione um arquivo</p>
          </div>
        )}
      </div>

      {/* Painel direito: Diff ou Terminal */}
      {showDiff && diffContent && (
        <div className="w-96 border-l border-zinc-800 overflow-y-auto bg-zinc-900">
          <div className="p-4 border-b border-zinc-800">
            <h3 className="text-sm font-semibold">Alterações</h3>
          </div>
          <DiffViewer
            originalContent={diffContent.before}
            newContent={diffContent.after}
          />
        </div>
      )}

      {showTerminal && (
        <div className="w-96 border-l border-zinc-800 overflow-y-auto bg-zinc-900">
          <Terminal />
        </div>
      )}
    </div>
  );
}
```

#### 3.2.3 FileTree Component

**Arquivo:** `components/Codigo/FileTree.tsx`

```typescript
'use client';

import { useState, useEffect } from 'react';
import { ChevronRight, ChevronDown, File, Folder } from 'lucide-react';
import { useAsyncOperation } from '@/hooks/useAsyncOperation';

interface FileTreeNode {
  id: string;
  name: string;
  path: string;
  isDirectory: boolean;
  children?: FileTreeNode[];
}

interface FileTreeProps {
  onSelectFile: (file: FileTreeNode) => void;
}

export default function FileTree({ onSelectFile }: FileTreeProps) {
  const [tree, setTree] = useState<FileTreeNode | null>(null);
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(new Set());
  const { execute: loadTree, loading } = useAsyncOperation();

  useEffect(() => {
    loadTree(async () => {
      const response = await window.go.main.App.CallIPC(
        'tool.execute',
        JSON.stringify({
          tool: 'read_folder',
          args: { path: process.env.HOME || '.' },
        })
      );
      const data = JSON.parse(response);
      setTree(data);
    });
  }, [loadTree]);

  const toggleDir = (dirId: string) => {
    setExpandedDirs((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(dirId)) {
        newSet.delete(dirId);
      } else {
        newSet.add(dirId);
      }
      return newSet;
    });
  };

  const renderNode = (node: FileTreeNode, depth = 0) => {
    const isExpanded = expandedDirs.has(node.id);

    return (
      <div key={node.id}>
        <div
          className="flex items-center gap-2 px-2 py-1 hover:bg-zinc-800 cursor-pointer"
          style={{ paddingLeft: `${depth * 16 + 8}px` }}
          onClick={() => {
            if (node.isDirectory) {
              toggleDir(node.id);
            } else {
              onSelectFile(node);
            }
          }}
        >
          {node.isDirectory ? (
            <>
              {isExpanded ? (
                <ChevronDown className="w-4 h-4" />
              ) : (
                <ChevronRight className="w-4 h-4" />
              )}
              <Folder className="w-4 h-4 text-yellow-500" />
            </>
          ) : (
            <>
              <div className="w-4" />
              <File className="w-4 h-4 text-blue-500" />
            </>
          )}
          <span className="text-sm truncate">{node.name}</span>
        </div>

        {node.isDirectory && isExpanded && node.children && (
          <>
            {node.children.map((child) => renderNode(child, depth + 1))}
          </>
        )}
      </div>
    );
  };

  return (
    <div className="p-2">
      {loading ? (
        <p className="text-xs text-zinc-500">Carregando...</p>
      ) : tree ? (
        renderNode(tree)
      ) : (
        <p className="text-xs text-zinc-500">Erro ao carregar arquivos</p>
      )}
    </div>
  );
}
```

#### 3.2.4 CodeEditor (Monaco)

**Arquivo:** `components/Codigo/CodeEditor.tsx`

```typescript
'use client';

import { useEffect, useRef } from 'react';
import * as monaco from 'monaco-editor';
import { detectLanguage } from '@/utils/syntax';

interface CodeEditorProps {
  file: { path: string; name: string };
  content: string;
  onChange: (content: string, isDirty: boolean) => void;
}

export default function CodeEditor({
  file,
  content,
  onChange,
}: CodeEditorProps) {
  const editorRef = useRef<HTMLDivElement>(null);
  const monacoEditorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);

  useEffect(() => {
    if (!editorRef.current) return;

    const language = detectLanguage(file.name);

    const editor = monaco.editor.create(editorRef.current, {
      value: content,
      language,
      theme: 'vs-dark',
      fontSize: 14,
      minimap: { enabled: false },
      wordWrap: 'on',
      formatOnPaste: true,
      autoClosingBrackets: 'always',
    });

    monacoEditorRef.current = editor;

    const changeDisposable = editor.onDidChangeModelContent(() => {
      const newContent = editor.getValue();
      onChange(newContent, newContent !== content);
    });

    return () => {
      changeDisposable.dispose();
      editor.dispose();
    };
  }, [file, onChange, content]);

  return <div ref={editorRef} className="w-full h-full" />;
}
```

#### 3.2.5 Terminal Integrado

**Arquivo:** `components/Codigo/Terminal.tsx`

```typescript
'use client';

import { useState, useRef, useEffect } from 'react';
import { useAsyncOperation } from '@/hooks/useAsyncOperation';

interface TerminalLine {
  type: 'input' | 'output' | 'error';
  content: string;
  timestamp: Date;
}

export default function Terminal() {
  const [lines, setLines] = useState<TerminalLine[]>([]);
  const [currentCommand, setCurrentCommand] = useState('');
  const [cwd, setCwd] = useState(process.env.HOME || '.');
  const terminalRef = useRef<HTMLDivElement>(null);
  const { execute: runCommand, loading } = useAsyncOperation();

  // Auto-scroll
  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [lines]);

  const handleExecuteCommand = async () => {
    if (!currentCommand.trim()) return;

    setLines((prev) => [
      ...prev,
      {
        type: 'input',
        content: `$ ${currentCommand}`,
        timestamp: new Date(),
      },
    ]);

    await runCommand(async () => {
      try {
        const response = await window.go.main.App.CallIPC(
          'tool.execute',
          JSON.stringify({
            tool: 'shell_command',
            args: {
              command: currentCommand,
              cwd,
              timeout_sec: 30,
            },
          })
        );
        const data = JSON.parse(response);
        setLines((prev) => [
          ...prev,
          {
            type: data.result.exit_code === 0 ? 'output' : 'error',
            content: data.result.stdout || data.result.stderr,
            timestamp: new Date(),
          },
        ]);
      } catch (error) {
        setLines((prev) => [
          ...prev,
          {
            type: 'error',
            content: `Erro: ${error instanceof Error ? error.message : 'Desconhecido'}`,
            timestamp: new Date(),
          },
        ]);
      }
    });

    setCurrentCommand('');
  };

  return (
    <div className="flex flex-col h-full bg-zinc-950">
      {/* Output */}
      <div
        ref={terminalRef}
        className="flex-1 overflow-y-auto p-4 font-mono text-sm space-y-1"
      >
        {lines.map((line, idx) => (
          <div
            key={idx}
            className={
              line.type === 'input'
                ? 'text-emerald-400'
                : line.type === 'error'
                ? 'text-red-400'
                : 'text-zinc-300'
            }
          >
            {line.content}
          </div>
        ))}
      </div>

      {/* Input */}
      <div className="border-t border-zinc-800 p-4 flex gap-2">
        <span className="text-emerald-400 font-mono">$</span>
        <input
          type="text"
          value={currentCommand}
          onChange={(e) => setCurrentCommand(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleExecuteCommand();
          }}
          disabled={loading}
          className="flex-1 bg-transparent text-zinc-50 outline-none font-mono text-sm"
          placeholder="Digite um comando..."
        />
      </div>
    </div>
  );
}
```

---

### 3.3 ABA INDEX (Novo)

#### 3.3.1 Visão Geral

A aba Index fornece gerenciamento completo de workspaces e datasets:

- **Lista de Workspaces:** Todos os workspaces criados localmente
- **Detalhe de Workspace:** Visualizar chunks, histórico de indexação
- **Upload de Arquivos:** Indexar novos documentos
- **Browse do Vectora Index:** Explorar e baixar datasets pré-feitos

#### 3.3.2 Página Principal

**Arquivo:** `app/index/page.tsx`

```typescript
'use client';

import { useState } from 'react';
import WorkspaceList from '@/components/Index/WorkspaceList';
import WorkspaceDetail from '@/components/Index/WorkspaceDetail';
import WorkspaceUploader from '@/components/Index/WorkspaceUploader';
import DatasetBrowser from '@/components/Index/DatasetBrowser';
import { useWorkspace } from '@/hooks/useWorkspace';
import { useIndexStore } from '@/store/indexStore';
import { Plus, Download } from 'lucide-react';

export default function IndexPage() {
  const { workspaces, activeWorkspace, setActiveWorkspace, createWorkspace } = useWorkspace();
  const { showUploader, setShowUploader, showBrowser, setShowBrowser } = useIndexStore();
  const [tab, setTab] = useState<'local' | 'browse'>('local');

  return (
    <div className="flex h-full gap-4 p-6">
      {/* Painel esquerdo: Lista de workspaces */}
      <div className="w-64 flex flex-col border border-zinc-800 rounded-lg overflow-hidden bg-zinc-900">
        <div className="p-4 border-b border-zinc-800 flex justify-between items-center">
          <h3 className="font-semibold text-sm">Workspaces</h3>
          <button
            onClick={() => setShowUploader(true)}
            className="p-1 hover:bg-zinc-800 rounded"
            title="Criar novo workspace"
          >
            <Plus className="w-4 h-4" />
          </button>
        </div>
        <WorkspaceList
          workspaces={workspaces}
          activeWorkspace={activeWorkspace}
          onSelectWorkspace={setActiveWorkspace}
        />
      </div>

      {/* Painel central/direito: Detalhe ou Browse */}
      {tab === 'local' && activeWorkspace ? (
        <div className="flex-1 border border-zinc-800 rounded-lg overflow-hidden flex flex-col bg-zinc-900">
          <div className="p-4 border-b border-zinc-800">
            <h2 className="text-lg font-semibold">{activeWorkspace.name}</h2>
            <p className="text-xs text-zinc-500">{activeWorkspace.description}</p>
          </div>
          <WorkspaceDetail workspace={activeWorkspace} />
        </div>
      ) : (
        <div className="flex-1 border border-zinc-800 rounded-lg overflow-hidden flex flex-col bg-zinc-900">
          <div className="p-4 border-b border-zinc-800 flex gap-2">
            <button
              onClick={() => setTab('local')}
              className={`px-4 py-2 text-sm rounded ${
                tab === 'local' ? 'bg-emerald-600' : 'bg-zinc-800'
              }`}
            >
              Meus Workspaces
            </button>
            <button
              onClick={() => setTab('browse')}
              className={`px-4 py-2 text-sm rounded flex items-center gap-2 ${
                tab === 'browse' ? 'bg-emerald-600' : 'bg-zinc-800'
              }`}
            >
              <Download className="w-4 h-4" />
              Explorar Datasets
            </button>
          </div>
          {tab === 'browse' && <DatasetBrowser />}
        </div>
      )}

      {/* Modais */}
      {showUploader && (
        <WorkspaceUploader
          onClose={() => setShowUploader(false)}
          onSuccess={() => setShowUploader(false)}
        />
      )}
    </div>
  );
}
```

#### 3.3.3 WorkspaceList Component

**Arquivo:** `components/Index/WorkspaceList.tsx`

```typescript
'use client';

import { Workspace } from '@/services/ipc/types';
import { Trash2, Zap } from 'lucide-react';
import { useAsyncOperation } from '@/hooks/useAsyncOperation';

interface WorkspaceListProps {
  workspaces: Workspace[];
  activeWorkspace: Workspace | null;
  onSelectWorkspace: (ws: Workspace) => void;
}

export default function WorkspaceList({
  workspaces,
  activeWorkspace,
  onSelectWorkspace,
}: WorkspaceListProps) {
  const { execute: deleteWorkspace } = useAsyncOperation();

  const handleDelete = async (wsId: string) => {
    if (!confirm('Deletar este workspace?')) return;
    await deleteWorkspace(async () => {
      await window.go.main.App.CallIPC(
        'workspace.delete',
        JSON.stringify({ ws_id: wsId })
      );
    });
  };

  return (
    <div className="flex-1 overflow-y-auto">
      {workspaces.length === 0 ? (
        <div className="p-4 text-center text-zinc-500 text-sm">
          Nenhum workspace
        </div>
      ) : (
        <div className="space-y-2 p-2">
          {workspaces.map((ws) => (
            <div
              key={ws.id}
              onClick={() => onSelectWorkspace(ws)}
              className={`p-3 rounded cursor-pointer transition-colors ${
                activeWorkspace?.id === ws.id
                  ? 'bg-emerald-600/20 border border-emerald-500'
                  : 'bg-zinc-800 hover:bg-zinc-700'
              }`}
            >
              <div className="flex justify-between items-start">
                <div className="flex-1 min-w-0">
                  <p className="font-semibold text-sm truncate">{ws.name}</p>
                  <p className="text-xs text-zinc-400 truncate">{ws.description}</p>
                  <div className="text-xs text-zinc-500 mt-2">
                    {ws.chunkCount} chunks • {ws.indexedAt
                      ? new Date(ws.indexedAt).toLocaleDateString('pt-BR')
                      : 'Não indexado'}
                  </div>
                </div>
                {ws.status === 'indexing' && (
                  <Zap className="w-4 h-4 text-yellow-500 animate-pulse" />
                )}
              </div>
              {ws.status === 'indexing' && (
                <div className="mt-2 w-full bg-zinc-700 rounded-full h-1">
                  <div
                    className="bg-emerald-500 h-1 rounded-full transition-all"
                    style={{ width: `${ws.indexProgress || 0}%` }}
                  />
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
```

#### 3.3.4 WorkspaceUploader Component

**Arquivo:** `components/Index/WorkspaceUploader.tsx`

```typescript
'use client';

import { useState } from 'react';
import { useAsyncOperation } from '@/hooks/useAsyncOperation';

interface WorkspaceUploaderProps {
  onClose: () => void;
  onSuccess: () => void;
}

export default function WorkspaceUploader({
  onClose,
  onSuccess,
}: WorkspaceUploaderProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [files, setFiles] = useState<File[]>([]);
  const { execute: createWorkspace, loading } = useAsyncOperation();

  const handleCreateWorkspace = async () => {
    if (!name.trim()) {
      alert('Nome do workspace é obrigatório');
      return;
    }

    await createWorkspace(async () => {
      // 1. Criar workspace
      const createResponse = await window.go.main.App.CallIPC(
        'workspace.create',
        JSON.stringify({
          name,
          description,
        })
      );
      const wsData = JSON.parse(createResponse);
      const wsId = wsData.ws_id;

      // 2. Indexar arquivos (se houver)
      if (files.length > 0) {
        // Implementar upload de arquivos
        // Por enquanto, apenas criar workspace vazio
      }

      onSuccess();
    });
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-zinc-900 rounded-lg p-6 w-full max-w-md border border-zinc-800">
        <h2 className="text-lg font-semibold mb-4">Novo Workspace</h2>

        <div className="space-y-4">
          <div>
            <label className="text-sm font-medium text-zinc-300">Nome</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full mt-1 px-3 py-2 bg-zinc-800 text-zinc-50 border border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-emerald-500"
              placeholder="Ex: Documentação Godot"
            />
          </div>

          <div>
            <label className="text-sm font-medium text-zinc-300">Descrição</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full mt-1 px-3 py-2 bg-zinc-800 text-zinc-50 border border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-emerald-500 resize-none"
              placeholder="Descrição opcional..."
              rows={3}
            />
          </div>

          {/* Upload de arquivos será implementado depois */}
        </div>

        <div className="flex gap-2 mt-6 justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 text-zinc-300 hover:bg-zinc-800 rounded transition-colors"
          >
            Cancelar
          </button>
          <button
            onClick={handleCreateWorkspace}
            disabled={loading || !name.trim()}
            className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded disabled:opacity-50 transition-colors"
          >
            {loading ? 'Criando...' : 'Criar'}
          </button>
        </div>
      </div>
    </div>
  );
}
```

---

### 3.4 ABA MANAGER (Novo - Controle de Pacotes)

#### 3.4.1 Visão Geral

A aba Manager fornece interface para controlar MPM (Model Package Manager) e LPM (Llama Package Manager):

- **LPM Panel:** Gerenciar builds do llama.cpp
- **MPM Panel:** Gerenciar modelos Qwen3
- **Configurações Globais:** API keys, limites de RAM
- **Monitoramento:** Status de daemon, versões instaladas

#### 3.4.2 Página Principal

**Arquivo:** `app/manager/page.tsx`

```typescript
'use client';

import { useState } from 'react';
import PackageManagerTabs from '@/components/Manager/PackageManagerTabs';
import ConfigurationPanel from '@/components/Manager/ConfigurationPanel';
import { useManagerStore } from '@/store/managerStore';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/UI/Tabs';

export default function ManagerPage() {
  const { activeTab, setActiveTab } = useManagerStore();

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="px-6 py-4 border-b border-zinc-800">
        <h2 className="text-lg font-semibold">Gerenciador de Pacotes</h2>
        <p className="text-sm text-zinc-500">
          Controle builds (llama.cpp) e modelos (Qwen3)
        </p>
      </div>

      {/* Conteúdo */}
      <div className="flex-1 overflow-y-auto">
        <Tabs defaultValue={activeTab} onValueChange={setActiveTab}>
          <TabsList className="w-full justify-start bg-zinc-900 border-b border-zinc-800 rounded-none px-6">
            <TabsTrigger value="packages">Pacotes</TabsTrigger>
            <TabsTrigger value="configuration">Configurações</TabsTrigger>
          </TabsList>

          <TabsContent value="packages" className="p-6">
            <PackageManagerTabs />
          </TabsContent>

          <TabsContent value="configuration" className="p-6">
            <ConfigurationPanel />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
```

#### 3.4.3 PackageManagerTabs Component

**Arquivo:** `components/Manager/PackageManagerTabs.tsx`

```typescript
'use client';

import { useState } from 'react';
import LPMPanel from './LPMPanel';
import MPMPanel from './MPMPanel';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/UI/Tabs';

export default function PackageManagerTabs() {
  return (
    <Tabs defaultValue="lpm">
      <TabsList className="mb-4">
        <TabsTrigger value="lpm">Llama.cpp Builds (LPM)</TabsTrigger>
        <TabsTrigger value="mpm">Modelos Qwen3 (MPM)</TabsTrigger>
      </TabsList>

      <TabsContent value="lpm">
        <LPMPanel />
      </TabsContent>

      <TabsContent value="mpm">
        <MPMPanel />
      </TabsContent>
    </Tabs>
  );
}
```

#### 3.4.4 LPMPanel Component

**Arquivo:** `components/Manager/LPMPanel.tsx`

```typescript
'use client';

import { useState, useEffect } from 'react';
import { useAsyncOperation } from '@/hooks/useAsyncOperation';
import { usePackageManager } from '@/hooks/usePackageManager';
import ProgressMonitor from './ProgressMonitor';
import { Download, Trash2, Check } from 'lucide-react';

interface LPMBuild {
  id: string;
  name: string;
  version: string;
  size: number;
  installed: boolean;
  active: boolean;
  gpu: string;
}

export default function LPMPanel() {
  const [builds, setBuilds] = useState<LPMBuild[]>([]);
  const [installing, setInstalling] = useState<string | null>(null);
  const { execute: listBuilds, loading: listLoading } = useAsyncOperation();
  const { installBuild, uninstallBuild } = usePackageManager();

  useEffect(() => {
    loadBuilds();
  }, []);

  const loadBuilds = async () => {
    await listBuilds(async () => {
      try {
        const response = await window.go.main.App.CallIPC(
          'package.lpm_list',
          JSON.stringify({})
        );
        const data = JSON.parse(response);
        setBuilds(data.builds);
      } catch (error) {
        console.error('Erro ao listar builds:', error);
      }
    });
  };

  const handleInstall = async (buildId: string) => {
    setInstalling(buildId);
    try {
      await installBuild(buildId, (progress) => {
        // Atualizar barra de progresso
      });
      await loadBuilds();
    } finally {
      setInstalling(null);
    }
  };

  const handleUninstall = async (buildId: string) => {
    if (!confirm('Remover este build?')) return;
    try {
      await uninstallBuild(buildId);
      await loadBuilds();
    } catch (error) {
      console.error('Erro ao desinstalar:', error);
    }
  };

  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-sm font-semibold mb-3">Builds Disponíveis</h3>
        {listLoading ? (
          <p className="text-zinc-500">Carregando...</p>
        ) : (
          <div className="space-y-3">
            {builds.map((build) => (
              <div
                key={build.id}
                className="border border-zinc-800 rounded-lg p-4 bg-zinc-900"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h4 className="font-semibold">{build.name}</h4>
                      {build.active && (
                        <span className="text-xs px-2 py-1 bg-emerald-600/20 text-emerald-400 border border-emerald-500 rounded">
                          Ativo
                        </span>
                      )}
                      {build.installed && (
                        <Check className="w-4 h-4 text-emerald-500" />
                      )}
                    </div>
                    <p className="text-sm text-zinc-500">
                      v{build.version} • {build.gpu}
                    </p>
                    <p className="text-xs text-zinc-600 mt-1">
                      Tamanho: {(build.size / 1024 / 1024).toFixed(2)} MB
                    </p>
                  </div>

                  <div className="flex gap-2">
                    {!build.installed ? (
                      <button
                        onClick={() => handleInstall(build.id)}
                        disabled={installing === build.id}
                        className="px-3 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded text-sm flex items-center gap-1 disabled:opacity-50"
                      >
                        <Download className="w-4 h-4" />
                        {installing === build.id ? 'Instalando...' : 'Instalar'}
                      </button>
                    ) : (
                      <button
                        onClick={() => handleUninstall(build.id)}
                        className="px-3 py-2 bg-red-600/20 hover:bg-red-600/30 text-red-400 rounded text-sm flex items-center gap-1"
                      >
                        <Trash2 className="w-4 h-4" />
                        Remover
                      </button>
                    )}
                  </div>
                </div>

                {installing === build.id && <ProgressMonitor buildId={build.id} />}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
```

#### 3.4.5 MPMPanel Component

**Arquivo:** `components/Manager/MPMPanel.tsx`

```typescript
'use client';

import { useState, useEffect } from 'react';
import { useAsyncOperation } from '@/hooks/useAsyncOperation';
import { usePackageManager } from '@/hooks/usePackageManager';
import ProgressMonitor from './ProgressMonitor';
import { Download, Trash2, Check } from 'lucide-react';

interface MPMModel {
  id: string;
  name: string;
  family: string;
  size: number;
  quantization: string;
  installed: boolean;
  active: boolean;
  embedding?: boolean;
}

export default function MPMPanel() {
  const [models, setModels] = useState<MPMModel[]>([]);
  const [installing, setInstalling] = useState<string | null>(null);
  const [filterFamily, setFilterFamily] = useState<string>('qwen3');
  const { execute: listModels, loading: listLoading } = useAsyncOperation();
  const { installModel, uninstallModel } = usePackageManager();

  useEffect(() => {
    loadModels();
  }, []);

  const loadModels = async () => {
    await listModels(async () => {
      try {
        const response = await window.go.main.App.CallIPC(
          'package.mpm_list',
          JSON.stringify({})
        );
        const data = JSON.parse(response);
        setModels(data.models);
      } catch (error) {
        console.error('Erro ao listar modelos:', error);
      }
    });
  };

  const handleInstall = async (modelId: string) => {
    setInstalling(modelId);
    try {
      await installModel(modelId, (progress) => {
        // Atualizar barra de progresso
      });
      await loadModels();
    } finally {
      setInstalling(null);
    }
  };

  const handleUninstall = async (modelId: string) => {
    if (!confirm('Remover este modelo?')) return;
    try {
      await uninstallModel(modelId);
      await loadModels();
    } catch (error) {
      console.error('Erro ao desinstalar:', error);
    }
  };

  const filteredModels = models.filter(
    (m) => !filterFamily || m.family === filterFamily
  );

  return (
    <div className="space-y-4">
      {/* Filtros */}
      <div>
        <label className="text-sm font-medium text-zinc-300">Família</label>
        <select
          value={filterFamily}
          onChange={(e) => setFilterFamily(e.target.value)}
          className="w-full mt-2 px-3 py-2 bg-zinc-800 text-zinc-50 border border-zinc-700 rounded"
        >
          <option value="">Todas</option>
          <option value="qwen3">Qwen3</option>
          <option value="qwen-embedding">Qwen Embedding</option>
        </select>
      </div>

      {/* Lista de modelos */}
      <div>
        <h3 className="text-sm font-semibold mb-3">Modelos Disponíveis</h3>
        {listLoading ? (
          <p className="text-zinc-500">Carregando...</p>
        ) : (
          <div className="space-y-3">
            {filteredModels.map((model) => (
              <div
                key={model.id}
                className="border border-zinc-800 rounded-lg p-4 bg-zinc-900"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h4 className="font-semibold">{model.name}</h4>
                      {model.active && (
                        <span className="text-xs px-2 py-1 bg-emerald-600/20 text-emerald-400 border border-emerald-500 rounded">
                          Ativo
                        </span>
                      )}
                      {model.embedding && (
                        <span className="text-xs px-2 py-1 bg-blue-600/20 text-blue-400 border border-blue-500 rounded">
                          Embedding
                        </span>
                      )}
                      {model.installed && (
                        <Check className="w-4 h-4 text-emerald-500" />
                      )}
                    </div>
                    <p className="text-sm text-zinc-500">
                      {model.quantization} • {(model.size / 1024 / 1024 / 1024).toFixed(2)} GB
                    </p>
                  </div>

                  <div className="flex gap-2">
                    {!model.installed ? (
                      <button
                        onClick={() => handleInstall(model.id)}
                        disabled={installing === model.id}
                        className="px-3 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded text-sm flex items-center gap-1 disabled:opacity-50"
                      >
                        <Download className="w-4 h-4" />
                        {installing === model.id ? 'Instalando...' : 'Instalar'}
                      </button>
                    ) : (
                      <button
                        onClick={() => handleUninstall(model.id)}
                        className="px-3 py-2 bg-red-600/20 hover:bg-red-600/30 text-red-400 rounded text-sm flex items-center gap-1"
                      >
                        <Trash2 className="w-4 h-4" />
                        Remover
                      </button>
                    )}
                  </div>
                </div>

                {installing === model.id && <ProgressMonitor modelId={model.id} />}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
```

#### 3.4.6 ConfigurationPanel Component

**Arquivo:** `components/Manager/ConfigurationPanel.tsx`

```typescript
'use client';

import { useState, useEffect } from 'react';
import { useAsyncOperation } from '@/hooks/useAsyncOperation';

interface Configuration {
  geminiApiKey: string;
  maxRamDaemon: number;
  maxRamIndexing: number;
  preferredLLMProvider: 'gemini' | 'qwen_local';
  logLevel: string;
}

export default function ConfigurationPanel() {
  const [config, setConfig] = useState<Configuration | null>(null);
  const [isDirty, setIsDirty] = useState(false);
  const { execute: loadConfig, loading: loadLoading } = useAsyncOperation();
  const { execute: saveConfig, loading: saveLoading } = useAsyncOperation();

  useEffect(() => {
    loadConfig(async () => {
      try {
        const response = await window.go.main.App.CallIPC(
          'config.get',
          JSON.stringify({})
        );
        const data = JSON.parse(response);
        setConfig(data);
      } catch (error) {
        console.error('Erro ao carregar configurações:', error);
      }
    });
  }, [loadConfig]);

  const handleSave = async () => {
    if (!config) return;
    await saveConfig(async () => {
      try {
        await window.go.main.App.CallIPC(
          'config.set',
          JSON.stringify(config)
        );
        setIsDirty(false);
      } catch (error) {
        console.error('Erro ao salvar configurações:', error);
      }
    });
  };

  if (loadLoading || !config) {
    return <p className="text-zinc-500">Carregando configurações...</p>;
  }

  return (
    <div className="max-w-2xl space-y-6">
      {/* Gemini API Key */}
      <div>
        <label className="text-sm font-semibold text-zinc-300">Gemini API Key</label>
        <p className="text-xs text-zinc-500 mb-2">
          Deixe em branco para usar apenas Qwen local
        </p>
        <input
          type="password"
          value={config.geminiApiKey}
          onChange={(e) => {
            setConfig({ ...config, geminiApiKey: e.target.value });
            setIsDirty(true);
          }}
          className="w-full px-3 py-2 bg-zinc-800 text-zinc-50 border border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-emerald-500"
          placeholder="sk-..."
        />
      </div>

      {/* Preferred Provider */}
      <div>
        <label className="text-sm font-semibold text-zinc-300">
          Provedor de IA Preferido
        </label>
        <div className="mt-2 space-y-2">
          {['qwen_local', 'gemini'].map((provider) => (
            <label key={provider} className="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                name="provider"
                value={provider}
                checked={config.preferredLLMProvider === provider}
                onChange={(e) => {
                  setConfig({
                    ...config,
                    preferredLLMProvider: e.target.value as any,
                  });
                  setIsDirty(true);
                }}
                className="rounded"
              />
              <span className="text-sm">
                {provider === 'qwen_local' ? 'Qwen Local' : 'Gemini (Cloud)'}
              </span>
            </label>
          ))}
        </div>
      </div>

      {/* Limites de RAM */}
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="text-sm font-semibold text-zinc-300">
            Max RAM Daemon (GB)
          </label>
          <input
            type="number"
            min={1}
            max={32}
            value={Math.round(config.maxRamDaemon / 1024 / 1024 / 1024)}
            onChange={(e) => {
              const val = parseInt(e.target.value) * 1024 * 1024 * 1024;
              setConfig({ ...config, maxRamDaemon: val });
              setIsDirty(true);
            }}
            className="w-full mt-2 px-3 py-2 bg-zinc-800 text-zinc-50 border border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-emerald-500"
          />
        </div>

        <div>
          <label className="text-sm font-semibold text-zinc-300">
            Max RAM Indexing (MB)
          </label>
          <input
            type="number"
            min={128}
            max={2048}
            value={Math.round(config.maxRamIndexing / 1024 / 1024)}
            onChange={(e) => {
              const val = parseInt(e.target.value) * 1024 * 1024;
              setConfig({ ...config, maxRamIndexing: val });
              setIsDirty(true);
            }}
            className="w-full mt-2 px-3 py-2 bg-zinc-800 text-zinc-50 border border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-emerald-500"
          />
        </div>
      </div>

      {/* Botões */}
      <div className="flex gap-2 pt-4 border-t border-zinc-800">
        <button
          onClick={handleSave}
          disabled={!isDirty || saveLoading}
          className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded disabled:opacity-50 transition-colors"
        >
          {saveLoading ? 'Salvando...' : 'Salvar Configurações'}
        </button>
        {isDirty && (
          <p className="text-xs text-yellow-500 self-center">
            Há alterações não salvas
          </p>
        )}
      </div>
    </div>
  );
}
```

---

## 4. STATE MANAGEMENT

### 4.1 Zustand Stores

**Arquivo:** `store/appStore.ts`

```typescript
import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';

type TabType = 'chat' | 'codigo' | 'index' | 'manager';

interface UIState {
  activeTab: TabType;
  setActiveTab: (tab: TabType) => void;
  modalsOpen: Record<string, boolean>;
  openModal: (modalId: string) => void;
  closeModal: (modalId: string) => void;
  theme: 'dark' | 'light';
  setTheme: (theme: 'dark' | 'light') => void;
}

export const useUIStore = create<UIState>()(
  subscribeWithSelector((set) => ({
    activeTab: 'chat',
    setActiveTab: (tab) => set({ activeTab: tab }),
    modalsOpen: {},
    openModal: (id) =>
      set((state) => ({
        modalsOpen: { ...state.modalsOpen, [id]: true },
      })),
    closeModal: (id) =>
      set((state) => ({
        modalsOpen: { ...state.modalsOpen, [id]: false },
      })),
    theme: 'dark',
    setTheme: (theme) => set({ theme }),
  }))
);
```

**Arquivo:** `store/chatStore.ts`

```typescript
import { create } from 'zustand';
import type { Message } from '@/services/ipc/types';

interface ChatState {
  messages: Message[];
  setMessages: (messages: Message[] | ((prev: Message[]) => Message[])) => void;
  addMessage: (message: Message) => void;
  clearMessages: () => void;
}

export const useChatStore = create<ChatState>((set) => ({
  messages: [],
  setMessages: (messages) =>
    set((state) => ({
      messages: typeof messages === 'function' ? messages(state.messages) : messages,
    })),
  addMessage: (message) =>
    set((state) => ({
      messages: [...state.messages, message],
    })),
  clearMessages: () => set({ messages: [] }),
}));
```

Similar para `codigoStore.ts`, `indexStore.ts`, `managerStore.ts`.

---

## 5. INTEGRAÇÃO IPC COM DAEMON

### 5.1 Cliente IPC

**Arquivo:** `services/ipc/client.ts`

```typescript
import { EventEmitter } from 'events';

interface IPCMessage {
  id: string;
  type: 'request' | 'response' | 'event';
  method: string;
  payload: Record<string, any>;
  error?: { code: string; message: string };
}

export class IPCClient extends EventEmitter {
  private requestMap: Map<string, { resolve: Function; reject: Function; timeout: NodeJS.Timeout }> = new Map();
  private requestIdCounter = 0;

  async request(method: string, payload: Record<string, any>): Promise<any> {
    const requestId = `req-${++this.requestIdCounter}-${Date.now()}`;
    const message: IPCMessage = {
      id: requestId,
      type: 'request',
      method,
      payload,
    };

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.requestMap.delete(requestId);
        reject(new Error(`Timeout na chamada IPC: ${method}`));
      }, 30000); // 30s timeout

      this.requestMap.set(requestId, { resolve, reject, timeout });

      // Enviar para daemon via Wails
      if (typeof window !== 'undefined' && window.go?.main?.App) {
        window.go.main.App.CallIPC(method, JSON.stringify(payload))
          .then((response: string) => {
            const data = JSON.parse(response);
            clearTimeout(timeout);
            this.requestMap.delete(requestId);
            if (data.error) {
              reject(new Error(data.error.message));
            } else {
              resolve(data.payload || data);
            }
          })
          .catch((error: Error) => {
            clearTimeout(timeout);
            this.requestMap.delete(requestId);
            reject(error);
          });
      }
    });
  }

  subscribe(eventName: string, callback: (data: any) => void) {
    this.on(eventName, callback);
    return () => this.off(eventName, callback);
  }
}

export const ipcClient = new IPCClient();
```

### 5.2 Hook useIPC

**Arquivo:** `hooks/useIPC.ts`

```typescript
import { useCallback } from 'react';
import { ipcClient } from '@/services/ipc/client';

export function useIPC(method: string) {
  return useCallback(
    async (payload: Record<string, any>) => {
      try {
        return await ipcClient.request(method, payload);
      } catch (error) {
        console.error(`Erro na chamada IPC ${method}:`, error);
        throw error;
      }
    },
    [method]
  );
}
```

### 5.3 Hook usePackageManager

**Arquivo:** `hooks/usePackageManager.ts`

```typescript
import { useAsyncOperation } from './useAsyncOperation';

export function usePackageManager() {
  const { execute } = useAsyncOperation();

  const installBuild = async (
    buildId: string,
    onProgress?: (progress: { current: number; total: number; percent: number }) => void
  ) => {
    return execute(async () => {
      // Chamada IPC para instalar build
      const response = await window.go.main.App.CallIPC(
        'package.lpm_install',
        JSON.stringify({ build_id: buildId })
      );
      return JSON.parse(response);
    });
  };

  const installModel = async (
    modelId: string,
    onProgress?: (progress: { current: number; total: number; percent: number }) => void
  ) => {
    return execute(async () => {
      const response = await window.go.main.App.CallIPC(
        'package.mpm_install',
        JSON.stringify({ model_id: modelId })
      );
      return JSON.parse(response);
    });
  };

  const uninstallBuild = async (buildId: string) => {
    return execute(async () => {
      const response = await window.go.main.App.CallIPC(
        'package.lpm_uninstall',
        JSON.stringify({ build_id: buildId })
      );
      return JSON.parse(response);
    });
  };

  const uninstallModel = async (modelId: string) => {
    return execute(async () => {
      const response = await window.go.main.App.CallIPC(
        'package.mpm_uninstall',
        JSON.stringify({ model_id: modelId })
      );
      return JSON.parse(response);
    });
  };

  return {
    installBuild,
    installModel,
    uninstallBuild,
    uninstallModel,
  };
}
```

---

## 6. COMPONENTES REUTILIZÁVEIS

### 6.1 Modal Wrapper

**Arquivo:** `components/Common/Modal.tsx`

```typescript
'use client';

import { ReactNode } from 'react';
import { X } from 'lucide-react';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}

export function Modal({ isOpen, onClose, title, children }: ModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-zinc-900 rounded-lg border border-zinc-800 max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between p-6 border-b border-zinc-800">
          <h2 className="text-lg font-semibold">{title}</h2>
          <button onClick={onClose} className="text-zinc-500 hover:text-zinc-300">
            <X className="w-5 h-5" />
          </button>
        </div>
        <div className="p-6">{children}</div>
      </div>
    </div>
  );
}
```

### 6.2 SkeletonLoader

**Arquivo:** `components/Common/SkeletonLoader.tsx`

```typescript
'use client';

interface SkeletonLoaderProps {
  type: 'message' | 'file' | 'workspace';
  count?: number;
}

export function SkeletonLoader({ type, count = 1 }: SkeletonLoaderProps) {
  return (
    <>
      {Array.from({ length: count }).map((_, idx) => (
        <div key={idx} className="animate-pulse">
          {type === 'message' && (
            <div className="flex justify-start gap-2 mb-4">
              <div className="w-8 h-8 bg-zinc-800 rounded-full" />
              <div className="flex-1">
                <div className="h-4 bg-zinc-800 rounded w-3/4 mb-2" />
                <div className="h-4 bg-zinc-800 rounded w-1/2" />
              </div>
            </div>
          )}
          {type === 'file' && (
            <div className="h-6 bg-zinc-800 rounded mb-2" />
          )}
          {type === 'workspace' && (
            <div className="border border-zinc-800 rounded p-4 mb-2">
              <div className="h-5 bg-zinc-800 rounded w-2/3 mb-2" />
              <div className="h-4 bg-zinc-800 rounded w-1/2" />
            </div>
          )}
        </div>
      ))}
    </>
  );
}
```

### 6.3 Toast/Notificações

**Arquivo:** `components/Common/Toast.tsx`

```typescript
'use client';

import { useEffect, useState } from 'react';
import { AlertCircle, CheckCircle, Info, AlertTriangle, X } from 'lucide-react';

type ToastType = 'success' | 'error' | 'info' | 'warning';

interface ToastMessage {
  id: string;
  type: ToastType;
  message: string;
  duration?: number;
}

const TOAST_STORE = {
  toasts: [] as ToastMessage[],
  listeners: [] as Array<(toasts: ToastMessage[]) => void>,

  add(message: string, type: ToastType = 'info', duration = 5000) {
    const id = crypto.randomUUID();
    const toast = { id, type, message, duration };
    this.toasts.push(toast);
    this.notify();
    if (duration > 0) {
      setTimeout(() => this.remove(id), duration);
    }
  },

  remove(id: string) {
    this.toasts = this.toasts.filter((t) => t.id !== id);
    this.notify();
  },

  notify() {
    this.listeners.forEach((listener) => listener(this.toasts));
  },

  subscribe(listener: (toasts: ToastMessage[]) => void) {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter((l) => l !== listener);
    };
  },
};

export function Toaster() {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  useEffect(() => {
    return TOAST_STORE.subscribe(setToasts);
  }, []);

  return (
    <div className="fixed bottom-6 right-6 space-y-3 pointer-events-none">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={`flex items-start gap-3 p-4 rounded-lg border pointer-events-auto ${
            toast.type === 'success'
              ? 'bg-emerald-950 border-emerald-800 text-emerald-200'
              : toast.type === 'error'
              ? 'bg-red-950 border-red-800 text-red-200'
              : toast.type === 'warning'
              ? 'bg-yellow-950 border-yellow-800 text-yellow-200'
              : 'bg-blue-950 border-blue-800 text-blue-200'
          }`}
        >
          {toast.type === 'success' && <CheckCircle className="w-5 h-5 flex-shrink-0" />}
          {toast.type === 'error' && <AlertCircle className="w-5 h-5 flex-shrink-0" />}
          {toast.type === 'warning' && <AlertTriangle className="w-5 h-5 flex-shrink-0" />}
          {toast.type === 'info' && <Info className="w-5 h-5 flex-shrink-0" />}
          <div className="flex-1">
            <p className="text-sm">{toast.message}</p>
          </div>
          <button
            onClick={() => TOAST_STORE.remove(toast.id)}
            className="text-current hover:opacity-70"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      ))}
    </div>
  );
}

export function useToast() {
  return {
    success: (message: string, duration?: number) =>
      TOAST_STORE.add(message, 'success', duration),
    error: (message: string, duration?: number) =>
      TOAST_STORE.add(message, 'error', duration),
    warning: (message: string, duration?: number) =>
      TOAST_STORE.add(message, 'warning', duration),
    info: (message: string, duration?: number) =>
      TOAST_STORE.add(message, 'info', duration),
  };
}
```

---

## 7. DESIGN SYSTEM

### 7.1 Paleta de Cores (Kaffyn Dark)

**Arquivo:** `styles/variables.css`

```css
:root {
  /* Base - Zinc */
  --color-base-950: #030712;
  --color-base-900: #09090b;
  --color-base-800: #18181b;
  --color-base-700: #27272a;
  --color-base-600: #3f3f46;
  --color-base-500: #52525b;
  --color-base-400: #71717a;
  --color-base-300: #a1a1aa;
  --color-base-200: #e4e4e7;
  --color-base-100: #f4f4f5;
  --color-base-50: #fafafa;

  /* Accent - Emerald */
  --color-accent: #10b981;
  --color-accent-light: #34d399;
  --color-accent-dark: #059669;

  /* Semantic */
  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-error: #ef4444;
  --color-info: #3b82f6;

  /* Typography */
  --font-family-sans: 'Inter', system-ui, sans-serif;
  --font-family-mono: 'Menlo', 'Monaco', 'Courier', monospace;

  /* Spacing */
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 2rem;
}
```

### 7.2 Tailwind Config

**Arquivo:** `tailwind.config.js`

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './app/**/*.{js,ts,jsx,tsx}',
    './components/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        zinc: {
          950: '#030712',
          900: '#09090b',
          // ... rest of zinc palette
        },
        emerald: {
          500: '#10b981',
          600: '#059669',
          // ... rest of emerald palette
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['Menlo', 'Monaco', 'Courier', 'monospace'],
      },
      animation: {
        'pulse': 'pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'bounce': 'bounce 1s infinite',
      },
    },
  },
  plugins: [],
};
```

---

## 8. TRATAMENTO DE ERROS

### 8.1 Error Boundary

**Arquivo:** `components/Common/ErrorBoundary.tsx`

```typescript
'use client';

import { Component, ReactNode } from 'react';
import { AlertTriangle } from 'lucide-react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <AlertTriangle className="w-12 h-12 text-red-500 mx-auto mb-4" />
            <h2 className="text-lg font-semibold text-zinc-50 mb-2">
              Erro na Aplicação
            </h2>
            <p className="text-sm text-zinc-400 mb-4">
              {this.state.error?.message || 'Um erro inesperado ocorreu'}
            </p>
            <button
              onClick={() => location.reload()}
              className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded text-sm"
            >
              Recarregar Aplicação
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
```

### 8.2 Hook useAsyncOperation

**Arquivo:** `hooks/useAsyncOperation.ts`

```typescript
import { useState, useCallback } from 'react';
import { useToast } from '@/components/Common/Toast';

interface UseAsyncOperationReturn {
  execute: (fn: () => Promise<any>) => Promise<any>;
  loading: boolean;
  error: Error | null;
  data: any;
}

export function useAsyncOperation(): UseAsyncOperationReturn {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState(null);
  const { error: showError } = useToast();

  const execute = useCallback(async (fn: () => Promise<any>) => {
    try {
      setLoading(true);
      setError(null);
      const result = await fn();
      setData(result);
      return result;
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      showError(error.message);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [showError]);

  return { execute, loading, error, data };
}
```

---

## 9. PERFORMANCE E OTIMIZAÇÕES

### 9.1 Lazy Loading de Componentes

```typescript
import dynamic from 'next/dynamic';

const ChatFeed = dynamic(() => import('@/components/Chat/ChatFeed'), {
  loading: () => <SkeletonLoader type="message" count={3} />,
});

const CodeEditor = dynamic(() => import('@/components/Codigo/CodeEditor'), {
  loading: () => <div>Carregando editor...</div>,
});
```

### 9.2 Memoização com React.memo

```typescript
import { memo } from 'react';

export const MessageBubble = memo(({ message }) => {
  return (
    // Component implementation
  );
}, (prevProps, nextProps) => {
  return prevProps.message.id === nextProps.message.id;
});
```

### 9.3 Virtualização de Listas

Para listas grandes (FileTree, WorkspaceList), usar react-window:

```typescript
import { FixedSizeList } from 'react-window';

<FixedSizeList
  height={600}
  itemCount={items.length}
  itemSize={35}
  width="100%"
>
  {({ index, style }) => (
    <div style={style}>
      {items[index].name}
    </div>
  )}
</FixedSizeList>
```

### 9.4 SWR para Cache de Dados

```typescript
import useSWR from 'swr';

function useWorkspaces() {
  const { data, error, isLoading } = useSWR(
    'workspaces',
    async () => {
      const response = await window.go.main.App.CallIPC('workspace.list', '{}');
      return JSON.parse(response).workspaces;
    },
    { revalidateOnFocus: false, dedupingInterval: 30000 }
  );

  return { workspaces: data, loading: isLoading, error };
}
```

---

## 10. TESTES

### 10.1 Testes de Componentes (Vitest + React Testing Library)

**Arquivo:** `components/Chat/__tests__/MessageBubble.test.tsx`

```typescript
import { render, screen } from '@testing-library/react';
import { MessageBubble } from '../MessageBubble';

describe('MessageBubble', () => {
  it('renderiza mensagem do usuário com styling correto', () => {
    const message = {
      id: '1',
      role: 'user' as const,
      content: 'Olá',
      timestamp: new Date(),
      metadata: {},
    };

    render(<MessageBubble message={message} />);
    expect(screen.getByText('Olá')).toHaveClass('bg-emerald-600');
  });

  it('renderiza mensagem do assistente com markdown', () => {
    const message = {
      id: '2',
      role: 'assistant' as const,
      content: '# Olá\nEste é um **teste**',
      timestamp: new Date(),
      metadata: {},
    };

    render(<MessageBubble message={message} />);
    expect(screen.getByRole('heading')).toBeInTheDocument();
  });
});
```

### 10.2 Testes de Integração IPC

**Arquivo:** `__tests__/ipc.integration.test.ts`

```typescript
import { ipcClient } from '@/services/ipc/client';

describe('IPC Integration', () => {
  it('conecta ao daemon e faz requisição', async () => {
    // Mock Wails binding
    global.window = {
      go: {
        main: {
          App: {
            CallIPC: jest.fn().mockResolvedValue(
              JSON.stringify({
                payload: { workspaces: [] },
              })
            ),
          },
        },
      },
    } as any;

    const result = await ipcClient.request('workspace.list', {});
    expect(result).toEqual({ workspaces: [] });
  });

  it('trata timeout de requisição', async () => {
    global.window = {
      go: {
        main: {
          App: {
            CallIPC: jest.fn()
              .mockImplementation(() => new Promise(() => {})), // Never resolves
          },
        },
      },
    } as any;

    await expect(ipcClient.request('workspace.list', {}))
      .rejects
      .toThrow('Timeout');
  });
});
```

### 10.3 Suite de Testes E2E (Playwright)

**Arquivo:** `e2e/chat.spec.ts`

```typescript
import { test, expect } from '@playwright/test';

test('fluxo completo de chat', async ({ page }) => {
  // Navegar para app
  await page.goto('http://localhost:3000');

  // Selecionar workspace
  const workspaceSelect = page.locator('select');
  await workspaceSelect.selectOption('godot');

  // Enviar mensagem
  const input = page.locator('textarea');
  await input.fill('O que é viewport?');
  await page.locator('button:has-text("Enviar")').click();

  // Verificar resposta
  await expect(page.locator('text=viewport')).toBeVisible();
});
```

---

## REGRAS DE NEGÓCIO CONSOLIDADAS (APP)

- **RN-APP-01:** Máximo 1 workspace selecionado por sessão
- **RN-APP-02:** Chat persiste histórico no daemon (bbolt)
- **RN-APP-03:** Código editor integrado (Monaco) com syntax highlighting
- **RN-APP-04:** Terminal executa comandos via tool.shell_command IPC
- **RN-APP-05:** Manager controla MPM/LPM via IPC (não direct subprocess)
- **RN-APP-06:** Todas as operações de arquivo aciona GitBridge (snapshot antes de write)
- **RN-APP-07:** Undo/Redo via snapshots armazenados no daemon
- **RN-APP-08:** Gemini Key armazenada encriptada no daemon (não localStorage)
- **RN-APP-09:** Timeouts: IPC 30s, Download 120s, Query RAG 60s
- **RN-APP-10:** Theme dark obrigatório (Kaffyn Zinc + Emerald)

---

## CHECKLIST DE IMPLEMENTAÇÃO

### Fase 1: Setup Base (1-2 semanas)
- [ ] Estrutura Next.js + TailwindCSS
- [ ] Shadcn/UI components integrados
- [ ] Zustand stores (todos)
- [ ] IPC client com retry/timeout

### Fase 2: Aba Chat (1 semana)
- [ ] ChatFeed com Markdown
- [ ] InputArea com auto-grow
- [ ] MessageBubble com sources
- [ ] Streaming de respostas
- [ ] Tool call visualizer

### Fase 3: Aba Código (1-2 semanas)
- [ ] FileTree com recursão
- [ ] Monaco editor integrado
- [ ] Terminal shell
- [ ] Diff viewer
- [ ] Undo via GitBridge

### Fase 4: Aba Index (1 semana)
- [ ] WorkspaceList com CRUD
- [ ] WorkspaceDetail com chunks
- [ ] Upload/indexing progress
- [ ] DatasetBrowser para Vectora Index

### Fase 5: Aba Manager (1 semana)
- [ ] LPMPanel com list/install/uninstall
- [ ] MPMPanel com models
- [ ] ConfigurationPanel
- [ ] Progress monitoring

### Fase 6: Polish (1 semana)
- [ ] Error handling completo
- [ ] Toast notifications
- [ ] Dark mode refinement
- [ ] Testes E2E

---

**Total Estimado:** 6-8 semanas de desenvolvimento

---

