import React from "react";
import { Zap, Trash2, Plus, MessageSquare, Database, Layers, ChevronDown, ChevronLeft, FolderOpen, MoreHorizontal, Pencil, Check } from "lucide-react";
import * as Select from "@radix-ui/react-select";
import type { Locale } from "../lib/i18n-types";

interface ChatSession {
  id: string;
  title: string;
  last: string;
}

interface SidebarProps {
  t: (key: string, count?: number, vars?: any) => string;
  locale: Locale;
  onLocaleChange: (l: Locale) => void;
  sessions: ChatSession[];
  currentSessionId: string;
  onSelectSession: (id: string) => void;
  onNewChat: () => void;
  onCollapse: () => void;
  onClearSession: (id: string) => void;
  onRenameSession: (id: string, newTitle: string) => void;
  onClearAll: () => void;
  isOnline: boolean;
  engineVersion: string;
  model: string;
  setModel: (m: string) => void;
  // New Props for Cognition
  provider: "qwen" | "gemini";
  setProvider: (p: "qwen" | "gemini") => void;
  geminiKey: string;
  setGeminiKey: (k: string) => void;
}

export function Sidebar({ 
  t, 
  locale, 
  onLocaleChange, 
  sessions, 
  currentSessionId, 
  onSelectSession, 
  onNewChat,
  onCollapse,
  onClearSession,
  onRenameSession,
  onClearAll,
  isOnline,
  engineVersion,
  model,
  setModel,
  provider,
  setProvider,
  geminiKey,
  setGeminiKey
}: SidebarProps) {
  const [trainingSource, setTrainingSource] = React.useState("godot-4.6");
  const [localProjectPath, setLocalProjectPath] = React.useState<string | null>(null);
  const [isEditingPath, setIsEditingPath] = React.useState(false);
  const [editingChatId, setEditingChatId] = React.useState<string | null>(null);
  const [editTitle, setEditTitle] = React.useState("");
  
  // Dropdown UI for modern 3 dots:
  const [openMenuId, setOpenMenuId] = React.useState<string | null>(null);

  const GODOT_VERSIONS = ["4.0", "4.1", "4.2", "4.3", "4.4", "4.5", "4.6"];

  return (
    <div className="w-64 h-full flex flex-col p-5 border-r border-white/5 bg-[#09090b] shadow-2xl z-20 relative overflow-hidden">
      {/* Dynamic Background decor */}
      <div className="absolute inset-0 bg-gradient-to-b from-indigo-500/[0.015] to-transparent pointer-events-none" />
      
      {/* Brand Header & Collapse */}
      <div className="flex items-center justify-between mb-8 relative z-10">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-600 to-indigo-500 flex items-center justify-center shadow-lg shadow-indigo-500/20 group">
            <Zap className="w-4 h-4 text-white animate-pulse" />
          </div>
          <div className="min-w-0">
            <h1 className="text-sm font-bold text-white tracking-tight leading-none mb-1 truncate">
              {t("app_title")}
            </h1>
            <p className="text-[7px] text-indigo-400 font-bold uppercase tracking-widest opacity-60 truncate">
              {t("app_subtitle")}
            </p>
          </div>
        </div>
        <button 
          onClick={onCollapse}
          className="p-1 rounded-md hover:bg-white/5 text-white/10 hover:text-white/30 transition-all flex-shrink-0"
        >
          <ChevronLeft className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* 1. TOP: Neuro System (Cognitive Selector) */}
      <div className="mb-8 relative z-10 space-y-2.5">
        {/* Provider Selector */}
        <Select.Root value={provider} onValueChange={(p) => setProvider(p as any)}>
          <Select.Trigger className="w-full h-8 flex items-center justify-between px-2.5 rounded-lg bg-indigo-500/5 border border-indigo-500/20 text-[9px] font-black text-indigo-400 hover:text-white transition-all group hover:bg-indigo-500/10">
            <div className="flex items-center gap-2">
              <Zap className="w-2.5 h-2.5 text-indigo-500" />
              <span className="uppercase tracking-widest">{provider === 'qwen' ? 'Qwen Local' : 'Gemini Cloud'}</span>
            </div>
            <ChevronDown className="w-3 h-3 text-indigo-500" />
          </Select.Trigger>
          <Select.Portal>
            <Select.Content className="overflow-hidden bg-zinc-950 border border-white/10 rounded-xl p-1 shadow-2xl z-[200] backdrop-blur-3xl min-w-[140px]">
              <Select.Viewport className="p-1">
                <Select.Item value="qwen" className="px-3 py-2 text-[10px] font-bold text-white/30 hover:text-white hover:bg-white/10 rounded-lg outline-none cursor-pointer">
                  <Select.ItemText>Qwen (Local Native)</Select.ItemText>
                </Select.Item>
                <Select.Item value="gemini" className="px-3 py-2 text-[10px] font-bold text-white/30 hover:text-white hover:bg-white/10 rounded-lg outline-none cursor-pointer">
                  <Select.ItemText>Gemini (Google Cloud)</Select.ItemText>
                </Select.Item>
              </Select.Viewport>
            </Select.Content>
          </Select.Portal>
        </Select.Root>

        {/* Gemini API Key (Visible only when cloud) */}
        {provider === 'gemini' && (
          <div className="relative group">
            <div className="absolute inset-y-0 left-2.5 flex items-center pointer-events-none">
              <Check className={`w-2.5 h-2.5 ${geminiKey?.length > 10 ? 'text-emerald-500' : 'text-white/10'}`} />
            </div>
            <input 
              type="password"
              placeholder="Gemini API Key..."
              value={geminiKey}
              onChange={(e) => setGeminiKey(e.target.value)}
              className="w-full h-8 pl-8 pr-2.5 rounded-lg bg-white/[0.02] border border-white/5 text-[9px] text-white/60 placeholder:text-white/10 outline-none focus:border-indigo-500/40 transition-all font-mono"
            />
          </div>
        )}

        {/* Model Selector */}
        <Select.Root value={model} onValueChange={setModel}>
          <Select.Trigger className="w-full h-8 flex items-center justify-between px-2.5 rounded-lg bg-white/[0.02] border border-white/5 text-[9px] font-bold text-white/30 hover:text-white/60 transition-all group hover:bg-white/[0.04]">
            <div className="flex items-center gap-2">
              <Database className="w-2.5 h-2.5 text-indigo-500/40" />
              <Select.Value />
            </div>
            <ChevronDown className="w-3 h-3 text-white/10 group-hover:text-white/30" />
          </Select.Trigger>
          <Select.Portal>
            <Select.Content className="overflow-hidden bg-zinc-950 border border-white/10 rounded-xl p-1 shadow-2xl z-[200] backdrop-blur-3xl min-w-[140px]">
              <Select.Viewport className="p-1">
                {(provider === 'qwen' 
                  ? ["qwen3-0.6b", "qwen3-1.7b", "qwen3-4b", "qwen3-coder-next"]
                  : ["gemini-1.5-flash", "gemini-1.5-pro", "gemini-2.0-flash-exp"]
                ).map((m) => (
                  <Select.Item
                    key={m}
                    value={m}
                    className="px-3 py-2 text-[10px] font-bold text-white/30 hover:text-white hover:bg-white/10 rounded-lg outline-none cursor-pointer transition-colors"
                  >
                    <Select.ItemText>{m}</Select.ItemText>
                  </Select.Item>
                ))}
              </Select.Viewport>
            </Select.Content>
          </Select.Portal>
        </Select.Root>

        <Select.Root value={trainingSource} onValueChange={setTrainingSource}>
          <Select.Trigger className="w-full h-8 flex items-center justify-between px-2.5 rounded-lg bg-white/[0.02] border border-white/5 text-[9px] font-bold text-white/30 hover:text-white/60 transition-all group hover:bg-white/[0.04]">
            <div className="flex items-center gap-2">
              <Layers className="w-2.5 h-2.5 text-indigo-500/40" />
              <div className="truncate"><Select.Value /></div>
            </div>
            <ChevronDown className="w-3 h-3 text-white/10 group-hover:text-white/30" />
          </Select.Trigger>
          <Select.Portal>
            <Select.Content className="overflow-hidden bg-zinc-950 border border-white/10 rounded-xl p-1 shadow-2xl z-[200] backdrop-blur-3xl min-w-[160px]">
              <Select.Viewport className="p-1 max-h-[260px]">
                <Select.Item
                  value="none"
                  className="px-3 py-2 text-[10px] font-bold text-indigo-400 hover:text-white hover:bg-white/5 focus:bg-white/5 focus:text-white rounded-lg outline-none cursor-pointer transition-all uppercase border-b border-white/5 mb-1"
                >
                  <Select.ItemText>{t("config_training_none")}</Select.ItemText>
                </Select.Item>
                {GODOT_VERSIONS.reverse().map(v => (
                  <Select.Item
                    key={`godot-${v}`}
                    value={`godot-${v}`}
                    className="px-3 py-2 text-[10px] font-bold text-white/40 hover:text-white hover:bg-white/10 rounded-lg outline-none cursor-pointer"
                  >
                    <Select.ItemText>Godot Engine v{v}</Select.ItemText>
                  </Select.Item>
                ))}
              </Select.Viewport>
            </Select.Content>
          </Select.Portal>
        </Select.Root>

        {/* Local Project Directory Selector - same style as selectors above */}
        <div>
          {isEditingPath ? (
            <input 
              type="text"
              autoFocus
              className="w-full h-8 flex items-center justify-between px-2.5 rounded-lg bg-white/[0.02] border border-indigo-500/40 text-[9px] text-white/70 outline-none font-mono placeholder:text-white/20 transition-all"
              placeholder="Ex: C:\Users\bruno\Desktop\Vectora"
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  const val = e.currentTarget.value;
                  if (val.trim()) setLocalProjectPath(val.trim());
                  setIsEditingPath(false);
                } else if (e.key === 'Escape') {
                  setIsEditingPath(false);
                }
              }}
              onBlur={(e) => {
                const val = e.currentTarget.value;
                if (val.trim()) setLocalProjectPath(val.trim());
                setIsEditingPath(false);
              }}
            />
          ) : (
            <button
              onClick={() => setIsEditingPath(true)}
              className={`w-full h-8 flex items-center justify-between px-2.5 rounded-lg bg-white/[0.02] border border-white/5 text-[9px] font-bold hover:bg-white/[0.04] transition-all group ${localProjectPath ? 'text-emerald-400' : 'text-white/30 hover:text-white/60'}`}
            >
              <div className="flex items-center gap-2 min-w-0">
                <FolderOpen className={`w-2.5 h-2.5 shrink-0 ${localProjectPath ? 'text-emerald-400' : 'text-indigo-500/40'}`} />
                <span className="truncate">
                  {localProjectPath
                    ? (localProjectPath.split(/[/\\]/).pop() || localProjectPath)
                    : t("config_local_project_btn")}
                </span>
              </div>
              <ChevronDown className="w-3 h-3 text-white/10 group-hover:text-white/30 shrink-0" />
            </button>
          )}
        </div>
      </div>

      {/* 2. MIDDLE: History (flex-1) */}
      <div className="flex-1 overflow-y-auto pr-1.5 custom-scrollbar relative z-10 mb-4">
        <label className="text-[8px] font-black text-white/10 uppercase tracking-widest mb-4 block px-1">
          {t("history_title")}
        </label>
        <div className="flex flex-col gap-1">
          {sessions.map((chat) => (
            <div
              key={chat.id}
              role="button"
              tabIndex={0}
              onClick={() => onSelectSession(chat.id)}
              onKeyDown={(e) => e.key === 'Enter' && onSelectSession(chat.id)}
              className={`w-full flex items-center gap-2.5 p-2.5 rounded-xl transition-all group text-left border relative overflow-hidden cursor-pointer ${
                currentSessionId === chat.id 
                  ? "bg-indigo-600/10 border-indigo-500/20 shadow-xl" 
                  : "hover:bg-white/[0.03] border-transparent hover:border-white/5 hover:scale-[1.01]"
              }`}
            >
              {currentSessionId === chat.id && (
                <div className="absolute left-0 top-0 bottom-0 w-0.5 bg-indigo-500" />
              )}
              <div className={`w-6 h-6 rounded-md flex items-center justify-center transition-colors flex-shrink-0 ${
                currentSessionId === chat.id ? "bg-indigo-500/20" : "bg-white/[0.02] border border-white/5 group-hover:bg-white/10"
              }`}>
                <MessageSquare className={`w-2.5 h-2.5 transition-colors ${
                  currentSessionId === chat.id ? "text-indigo-400" : "text-white/10 group-hover:text-white/30"
                }`} />
              </div>
              <div className="flex flex-col min-w-0 flex-1 relative">
                {editingChatId === chat.id ? (
                  <input
                    autoFocus
                    value={editTitle}
                    onChange={(e) => setEditTitle(e.target.value)}
                    onBlur={() => {
                      if (editTitle.trim()) onRenameSession(chat.id, editTitle.trim());
                      setEditingChatId(null);
                      setOpenMenuId(null);
                    }}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        if (editTitle.trim()) onRenameSession(chat.id, editTitle.trim());
                        setEditingChatId(null);
                        setOpenMenuId(null);
                      }
                    }}
                    className={`text-[10px] font-bold bg-black/50 border border-indigo-500/50 rounded outline-none px-1 text-white w-full`}
                  />
                ) : (
                  <span className={`text-[10px] font-bold truncate transition-colors ${
                    currentSessionId === chat.id ? "text-white" : "text-white/20 group-hover:text-white/60"
                  }`}>
                    {t(chat.title)}
                  </span>
                )}
                {editingChatId !== chat.id && (
                  <span className={`text-[7px] font-semibold uppercase tracking-tighter ${
                    currentSessionId === chat.id ? "text-indigo-400/40" : "text-white/10"
                  }`}>
                    {t(chat.last)}
                  </span>
                )}
              </div>
              
              {/* 3 dots context menu replacing the plain trash icon */}
              <div className="relative">
                <button 
                  onClick={(e) => { 
                    e.stopPropagation(); 
                    setOpenMenuId(openMenuId === chat.id ? null : chat.id); 
                  }}
                  className="p-1 rounded-md bg-transparent hover:bg-white/10 text-white/20 hover:text-white opacity-0 group-hover:opacity-100 transition-all shrink-0"
                >
                  <MoreHorizontal className="w-3 h-3" />
                </button>
                {openMenuId === chat.id && (
                  <div className="absolute right-0 top-6 w-32 bg-zinc-950 border border-white/10 rounded-xl shadow-2xl z-50 flex flex-col p-1 animate-in zoom-in-95 duration-200"
                    onMouseLeave={() => setOpenMenuId(null)}
                  >
                    <button 
                      onClick={(e) => {
                        e.stopPropagation();
                        setEditTitle(t(chat.title));
                        setEditingChatId(chat.id);
                        setOpenMenuId(null);
                      }}
                      className="w-full text-left px-2.5 py-2 text-[10px] font-bold text-white/60 hover:text-white hover:bg-white/10 rounded-lg flex items-center gap-2"
                    >
                      <Pencil className="w-3 h-3" />
                      Rename
                    </button>
                    <button 
                      onClick={(e) => {
                        e.stopPropagation();
                        onClearSession(chat.id);
                        setOpenMenuId(null);
                      }}
                      className="w-full text-left px-2.5 py-2 text-[10px] font-bold text-rose-500 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg flex items-center gap-2"
                    >
                      <Trash2 className="w-3 h-3" />
                      Delete
                    </button>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 3. BOTTOM: New Chat */}
      <div className="relative z-10 pb-4 border-b border-white/5 mb-4">
        <button 
          onClick={onNewChat}
          className="w-full h-8 flex items-center justify-center gap-2 rounded-xl bg-indigo-600 shadow-xl shadow-indigo-600/10 hover:bg-indigo-500 hover:scale-[1.01] active:scale-95 text-[8px] font-black text-white uppercase tracking-widest transition-all group"
        >
          <Plus className="w-3 h-3 group-hover:rotate-90 transition-transform duration-500" />
          {t("history_new")}
        </button>
      </div>

      <div className="space-y-3 relative z-10 opacity-80 mt-auto">
        <button 
          onClick={onClearAll}
          className="w-full h-8 flex items-center justify-center gap-2 text-[9px] font-bold text-white/30 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg transition-all focus:outline-none focus:ring-1 focus:ring-rose-500/30"
        >
          <Trash2 className="w-3 h-3" />
          {t("action_clear")}
        </button>
        
        <div className="flex items-center justify-between px-2 pt-1">
          <span className={`text-[8px] font-black uppercase tracking-[0.2em] transition-colors ${isOnline ? "text-emerald-500/80" : "text-rose-500/80"}`}>
            {isOnline ? "ONLINE" : "OFFLINE"}
          </span>
          <div className="flex items-center gap-1.5">
            <div className={`w-1.5 h-1.5 rounded-full ${isOnline ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" : "bg-rose-500 shadow-[0_0_8px_rgba(244,63,94,0.5)] animate-pulse"}`} />
            <span className="text-[7px] font-black text-white/50 uppercase tracking-widest">
              ZYRIS RAG {engineVersion}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
