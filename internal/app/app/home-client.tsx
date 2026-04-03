"use client";

import React from "react";
import { Sidebar } from "../components/Sidebar";
import { ChatInput } from "../components/ChatInput";
import { Message } from "../components/Message";
import { Zap, Brain, Sparkles, Database, ChevronRight } from "lucide-react";
import type { Locale, Translations } from "../lib/i18n-types";
import { LanguageSelector } from "../components/LanguageSelector";
import { SuggestionCard } from "../components/SuggestionCard";
import { callVectora } from "../lib/ipc-bridge";

interface ChatMessage {
  id: string;
  role: "user" | "ai";
  content: string;
  sources?: Array<{
    id: string;
    name: string;
    path: string;
    relevance: number;
  }>;
  isThinking?: boolean;
}

interface ChatSession {
  id: string;
  title: string;
  last: string;
  messages: ChatMessage[];
}

interface HomeClientProps {
  initialTranslations: Translations;
  initialLocale: Locale;
}

export default function HomeClient({
  initialTranslations,
  initialLocale,
}: HomeClientProps) {
  const [locale, setLocale] = React.useState<Locale>(initialLocale);
  const [tArr, setTArr] = React.useState<Translations>(initialTranslations);
  const [isSidebarVisible, setIsSidebarVisible] = React.useState(true);
  const [model, setModel] = React.useState("qwen3-0.6b");
  const [provider, setProvider] = React.useState<"qwen" | "gemini">("qwen");
  const [geminiKey, setGeminiKey] = React.useState("");

  // Translation helper
  const t = (key: string, count?: number, vars?: any) => {
    const entry = tArr[key] || { one: key, other: key };
    let res = (count === 1 ? entry.one : entry.other) || entry.one || key;
    if (vars) {
      Object.entries(vars).forEach(([k, v]) => {
        res = res.replace(new RegExp(`{${k}}`, 'g'), String(v));
      });
    }
    if (count !== undefined) res = res.replace('{}', String(count));
    return res;
  };

  const [sessions, setSessions] = React.useState<ChatSession[]>([
    { 
      id: "welcome", 
      title: "history_welcome", 
      last: "time_now",
      messages: [] 
    }
  ]);
  const [currentSessionId, setCurrentSessionId] = React.useState("welcome");
  const [isTyping, setIsTyping] = React.useState(false);
  const [isOnline, setIsOnline] = React.useState(false);
  const [engineVersion, setEngineVersion] = React.useState("v1.0.0");
  const scrollRef = React.useRef<HTMLDivElement>(null);

  const currentSession = sessions.find(s => s.id === currentSessionId) || sessions[0];

  // ── Persistence helpers ─────────────────────────────────────────────────────
  const API = '/api/v1';

  // Load all sessions from Go/SQLite on mount
  React.useEffect(() => {
    const loadSessions = async () => {
      try {
        const data = await callVectora("chat.list");
        if (!data || data.length === 0) return;
        const loaded: ChatSession[] = data.map((c: any) => ({
          id: c.id,
          title: c.title || 'history_welcome',
          last: 'time_now',
          messages: (c.messages || []).map((m: any) => ({
            id: m.timestamp || Date.now().toString(),
            role: m.role as 'user' | 'ai',
            content: m.content,
            sources: m.sources || [],
          }))
        }));
        const firstSession = loaded[0];
        if (firstSession) {
          setSessions(loaded);
          setCurrentSessionId(firstSession.id);
        }
      } catch {
        // Go offline → keep in-memory default
      }
    };
    loadSessions();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Bidirectional Synchronization of the RAG Model on Mount
  React.useEffect(() => {
    callVectora("provider.get")
      .then(data => {
        if (data) {
          const p = (data.active_provider === 'gemini' ? 'gemini' : 'qwen') as "qwen" | "gemini";
          setProvider(p);
          setGeminiKey(data.gemini_api_key || "");
          if (p === 'qwen') {
            setModel(data.local_model || "qwen3-0.6b");
          } else {
            setModel(data.gemini_model || "gemini-1.5-flash");
          }
        }
      })
      .catch((e) => console.log('Settings load fallback', e));
  }, []);

  const saveSettings = async (updates: any) => {
    try {
      // Calculamos o estado unificado para o Go
      const body = {
        active_provider: updates.provider || provider,
        gemini_api_key: updates.geminiKey !== undefined ? updates.geminiKey : geminiKey,
        local_model: (updates.provider || provider) === 'qwen' ? (updates.model || model) : undefined,
        gemini_model: (updates.provider || provider) === 'gemini' ? (updates.model || model) : undefined,
      };

      await callVectora("provider.set", body);
    } catch (e) {
      console.error('Failed to sync settings with Go Core', e);
    }
  };

  const handleProviderChange = (p: "qwen" | "gemini") => {
    setProvider(p);
    const newModel = p === 'qwen' ? "qwen3-0.6b" : "gemini-1.5-flash";
    setModel(newModel);
    saveSettings({ provider: p, model: newModel });
  };

  const handleModelChange = (newModel: string) => {
    setModel(newModel);
    saveSettings({ model: newModel });
  };

  const handleGeminiKeyChange = (key: string) => {
    setGeminiKey(key);
    // We don't save on every keystroke, maybe use a debounce or save button?    // Por enquanto, salvamos pra garantir persistência imediata.
    saveSettings({ geminiKey: key });
  };

  // Health check
  React.useEffect(() => {
    const checkStatus = async () => {
      try {
        const data = await callVectora("app.health", {});
        setIsOnline(!!data);
      } catch {
        setIsOnline(false);
      }
    };
    checkStatus();
    const interval = setInterval(checkStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  // Helper to parse cookies on the client
  const getCookie = (name: string) => {
    if (typeof document === 'undefined') return null;
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop()?.split(';').shift();
    return null;
  };

  // Re-hydration: Detect locale on mount and sync with Go Core
  React.useEffect(() => {
    const hydratateLocale = async () => {
      const saved = getCookie("zyris_locale") as Locale;
      if (saved && saved !== locale) {
        handleLocaleChange(saved);
      } else {
        // Even if using default 'en', ensure we have latest translations from Go
        try {
          const data = await callVectora("i18n.get", { locale: locale });
          if (data) setTArr(data);
        } catch (e) {
          console.error("IPC i18n Handshake Failed", e);
        }
      }
    };
    hydratateLocale();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleLocaleChange = async (newLocale: Locale) => {
    setLocale(newLocale);
    if (typeof document !== 'undefined') {
      document.cookie = `zyris_locale=${newLocale}; Max-Age=31536000; Path=/; SameSite=Lax`;
    }
    try {
      // Request translations from the Go Engine via IPC bridge
      const data = await callVectora("i18n.get", { locale: newLocale });
      if (data) setTArr(data);
    } catch (e) {
      console.error("Failed to load translations from Engine", e);
    }
  };

  React.useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTo({
        top: scrollRef.current.scrollHeight,
        behavior: "smooth",
      });
    }
  }, [currentSession?.messages, isTyping]);

  const handleSendMessage = async (content: string) => {
    const userMsg: ChatMessage = { id: Date.now().toString(), role: "user", content };
    setSessions(prev => prev.map(s => s.id === currentSessionId 
      ? { ...s, messages: [...s.messages, userMsg], last: "time_now" } : s));
    setIsTyping(true);

    // Persist user message to SQLite via Go
    callVectora("message.add", { 
      conversationId: currentSessionId, 
      role: 'user', 
      content 
    }).catch(() => {});

    try {
      const data = await callVectora("workspace.query", { 
        message: content, 
        conversationId: currentSessionId, 
        provider: provider,
        model: model,
        api_key: provider === 'gemini' ? geminiKey : undefined
      });
      
      const aiMsg: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: "ai",
        content: data.reply || "Sem resposta do motor.",
        sources: data.sources || [],
      };
      setSessions(prev => prev.map(s => s.id === currentSessionId 
        ? { ...s, messages: [...s.messages, aiMsg] } : s));
    } catch (err: any) {
      const errorMsg: ChatMessage = {
        id: (Date.now() + 1).toString(), role: "ai",
        content: err.message || "The Go RAG Engine is offline.",
        sources: [],
      };
      setSessions(prev => prev.map(s => s.id === currentSessionId 
        ? { ...s, messages: [...s.messages, errorMsg] } : s));
    } finally {
      setIsTyping(false);
    }
  };

  const handleNewChat = () => {
    const newId = Date.now().toString();
    const title = `Conversation ${sessions.length + 1}`;
    const newChat: ChatSession = { id: newId, title, last: "time_now", messages: [] };
    setSessions([newChat, ...sessions]);
    setCurrentSessionId(newId);
    // Persist to Go SQLite
    callVectora("chat.create", { id: newId, title }).catch(() => {});
  };

  const handleClearSession = (id: string) => {
    setSessions(prev => {
      const remaining = prev.filter(s => s.id !== id);
      if (remaining.length === 0) {
        const reset: ChatSession = { id: Date.now().toString(), title: "history_welcome", last: "time_now", messages: [] };
        setCurrentSessionId(reset.id);
        return [reset];
      }
      if (currentSessionId === id && remaining.length > 0) {
        const first = remaining[0];
        if (first) setCurrentSessionId(first.id);
      }
      return remaining;
    });
    callVectora("chat.delete", { id }).catch(() => {});
  };

  const handleClearAll = () => {
    const reset: ChatSession = { id: Date.now().toString(), title: "history_welcome", last: "time_now", messages: [] };
    // Delete all existing
    sessions.forEach(s => callVectora("chat.delete", { id: s.id }).catch(() => {}));
    setSessions([reset]);
    setCurrentSessionId(reset.id);
  };

  const handleRenameSession = (id: string, newTitle: string) => {
    setSessions(prev => prev.map(s => s.id === id ? { ...s, title: newTitle } : s));
    callVectora("chat.rename", { id, title: newTitle }).catch(() => {});
  };

  return (
    <div className="flex h-screen overflow-hidden bg-[#09090b] text-white selection:bg-indigo-500/30">
      {/* Sidebar Container with correct Collapse Logic */}
      <div className={`transition-all duration-500 ease-in-out overflow-hidden flex-shrink-0 ${isSidebarVisible ? 'w-64 border-r border-white/5' : 'w-0'}`}>
        <div className="w-64 h-full"> 
          <Sidebar 
            t={t} 
            locale={locale} 
            onLocaleChange={handleLocaleChange} 
            sessions={sessions}
            currentSessionId={currentSessionId}
            onSelectSession={setCurrentSessionId}
            onNewChat={handleNewChat}
            onCollapse={() => setIsSidebarVisible(false)}
            onClearSession={handleClearSession}
            onRenameSession={handleRenameSession}
            onClearAll={handleClearAll}
            isOnline={isOnline}
            engineVersion={engineVersion}
            model={model}
            setModel={handleModelChange}
            provider={provider}
            setProvider={handleProviderChange}
            geminiKey={geminiKey}
            setGeminiKey={handleGeminiKeyChange}
          />
        </div>
      </div>

      <main className="flex-1 flex flex-col relative bg-transparent overflow-hidden">
        {!isSidebarVisible && (
          <button 
            onClick={() => setIsSidebarVisible(true)}
            className="absolute top-6 left-6 z-[100] w-10 h-10 rounded-xl bg-zinc-900 border border-white/10 flex items-center justify-center hover:bg-white/5 transition-all shadow-3xl animate-in slide-in-from-left-6"
          >
            <ChevronRight className="w-5 h-5 text-indigo-400" />
          </button>
        )}

        <div className="absolute top-6 right-6 z-50">
          <LanguageSelector t={t} currentLocale={locale} onLocaleChange={handleLocaleChange} />
        </div>

        <div className="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 pointer-events-none" />
        
        <div ref={scrollRef} className="flex-1 overflow-y-auto pt-24 px-6 lg:px-12 pb-40 custom-scrollbar relative z-10">
          <div className="max-w-4xl mx-auto flex flex-col gap-3">
            {!currentSession?.messages?.length && (
              <div className="flex flex-col items-center justify-center pt-20 text-center animate-in fade-in zoom-in duration-1000">
                <div className="w-16 h-16 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center mb-8 shadow-2xl">
                  <Zap className="w-8 h-8 text-indigo-400" />
                </div>
                <h2 className="text-2xl font-bold tracking-tight text-white mb-3">{t("suggestion_help")}</h2>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-16 w-full max-w-4xl mx-auto">
                  {[
                    { key: "vlaeg", icon: Brain, query: "Explain VLAEG protocol." },
                    { key: "arch", icon: Sparkles, query: "Improve RAG architecture." },
                    { key: "meta", icon: Database, query: "Analyze vector metadata." },
                  ].map((item) => (
                    <SuggestionCard key={item.key} icon={item.icon as any} query={item.query} label={t(`suggestion_${item.key}`)} onClick={handleSendMessage} />
                  ))}
                </div>
              </div>
            )}

            {currentSession?.messages?.map((msg) => (
              <Message key={msg.id} role={msg.role} content={msg.content} sources={msg.sources} t={t} />
            ))}

            {isTyping && <Message role="ai" content="" isThinking={true} sources={[]} t={t} />}
          </div>
        </div>

        <div className="absolute bottom-0 inset-x-0 bg-gradient-to-t from-[#09090b] via-[#09090b]/90 to-transparent pt-10 z-20">
          <ChatInput onSend={handleSendMessage} disabled={isTyping} t={t} isCompact={(currentSession?.messages?.length ?? 0) > 0} />
        </div>
      </main>
    </div>
  );
}
