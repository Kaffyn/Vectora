"use client";

import React from "react";
import {
  Plus,
  CornerDownLeft,
  Sparkles,
  Image as ImageIcon,
  FileText,
  Code,
  Send,
} from "lucide-react";

interface ChatInputProps {
  onSend: (message: string) => void;
  disabled?: boolean;
  t: (key: string, count?: number, vars?: any) => string;
  isCompact?: boolean;
}

export function ChatInput({ onSend, disabled, t, isCompact }: ChatInputProps) {
  const [value, setValue] = React.useState("");
  const docInputRef = React.useRef<HTMLInputElement>(null);
  const imgInputRef = React.useRef<HTMLInputElement>(null);
  const allInputRef = React.useRef<HTMLInputElement>(null);
  const codeInputRef = React.useRef<HTMLInputElement>(null);

  // Clean hydration trigger for Turbopack
  React.useEffect(() => {
    console.log("[Zyris RAG] UI Components Hydrated and Active.");
  }, []);


  const handleSend = () => {
    if (value.trim() && !disabled) {
      onSend(value.trim());
      setValue("");
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>, type: string) => {
    const file = e.target.files?.[0];
    if (file) {
      console.log(`Analyzing ${type}:`, file.name);
      // Future: Upload file and inject into RAG context
    }
  };

  return (
    <div className="w-full max-w-5xl mx-auto px-6 pb-8 pt-4 lg:pb-12">
      {/* Hidden File Inputs */}
      <input 
        type="file" 
        multiple
        ref={allInputRef} 
        className="hidden" 
        onChange={(e) => handleFileChange(e, "any")} 
      />
      <input 
        type="file" 
        ref={docInputRef} 
        className="hidden" 
        accept=".pdf,.doc,.docx,.txt,.csv"
        onChange={(e) => handleFileChange(e, "document")} 
      />
      <input 
        type="file" 
        ref={imgInputRef} 
        className="hidden" 
        accept="image/*"
        onChange={(e) => handleFileChange(e, "image")} 
      />
      <input 
        type="file" 
        multiple
        ref={codeInputRef} 
        className="hidden" 
        accept=".ts,.tsx,.js,.jsx,.py,.go,.rs,.c,.cpp,.h,.cs,.java,.json,.xml,.yaml,.yml,.md"
        onChange={(e) => handleFileChange(e, "code")} 
      />

      <div className="relative group">
        <div className={`absolute -inset-1 bg-gradient-to-r from-indigo-500/20 via-cyan-500/10 to-indigo-500/20 blur-2xl group-focus-within:opacity-100 opacity-60 transition-all duration-1000 animate-pulse ${isCompact ? 'rounded-2xl' : 'rounded-3xl'}`} />

        <div className={`relative flex flex-col bg-zinc-900/40 border border-white/5 shadow-3xl backdrop-blur-3xl focus-within:ring-2 focus-within:ring-indigo-500/20 focus-within:border-white/10 transition-all duration-500 ease-in-out overflow-hidden group/input ${isCompact ? 'rounded-2xl' : 'rounded-3xl'}`}>
          {/* Header toolbar - Collapsible */}
          <div className={`grid transition-all duration-500 ease-in-out ${isCompact ? 'grid-rows-[0fr] opacity-0' : 'grid-rows-[1fr] opacity-100'}`}>
            <div className="overflow-hidden">
              <div className="flex items-center justify-between px-4 py-3 border-b border-white/5 bg-white/[0.02]">
                <div className="flex items-center gap-2">
                  <button 
                    onClick={() => allInputRef.current?.click()}
                    className="p-2 rounded-xl hover:bg-white/5 text-white/30 hover:text-indigo-400 transition-all group/tool"
                  >
                    <Plus className="w-4 h-4 group-hover/tool:scale-110" />
                  </button>
                  <div className="h-4 w-px bg-white/5 mx-1" />
                  {/* Trigger Doc Picker */}
                  <button 
                    onClick={() => docInputRef.current?.click()}
                    className="p-2 rounded-lg hover:bg-white/5 text-white/20 hover:text-white/40 transition-all flex items-center gap-1.5 px-3"
                  >
                    <FileText className="w-3.5 h-3.5" />
                    <span className="text-[10px] font-bold uppercase tracking-wider">
                      {t("input_doc")}
                    </span>
                  </button>
                  {/* Trigger Img Picker */}
                  <button 
                    onClick={() => imgInputRef.current?.click()}
                    className="p-2 rounded-lg hover:bg-white/5 text-white/20 hover:text-white/40 transition-all flex items-center gap-1.5 px-3"
                  >
                    <ImageIcon className="w-3.5 h-3.5" />
                    <span className="text-[10px] font-bold uppercase tracking-wider">
                      {t("input_img")}
                    </span>
                  </button>
                  <button 
                    onClick={() => codeInputRef.current?.click()}
                    className="p-2 rounded-lg hover:bg-white/5 text-white/20 hover:text-white/40 transition-all flex items-center gap-1.5 px-3"
                  >
                    <Code className="w-3.5 h-3.5" />
                    <span className="text-[10px] font-bold uppercase tracking-wider">
                      {t("input_code")}
                    </span>
                  </button>
                </div>

                <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-indigo-500/5 border border-indigo-500/20 shadow-inner">
                  <Sparkles className="w-3 h-3 text-indigo-400 animate-pulse" />
                  <span className="text-[9px] font-bold text-indigo-400/80 uppercase tracking-[0.1em]">
                    {t("input_deep_memory")}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <div className="relative flex">
            {isCompact && (
              <button 
                onClick={() => allInputRef.current?.click()}
                className="absolute left-3 top-3 p-2 rounded-xl text-white/30 hover:text-indigo-400 hover:bg-white/5 transition-all z-10"
              >
                <Plus className="w-5 h-5" />
              </button>
            )}
            <textarea
              className={`w-full bg-transparent text-[15px] text-white/90 placeholder:text-white/20 focus:outline-none resize-none scrollbar-thin scrollbar-thumb-white/10 scrollbar-track-transparent transition-all duration-500 ${isCompact ? 'pl-12 pr-14 py-[14px] min-h-[52px] max-h-[160px] overflow-y-auto leading-tight' : 'p-6 min-h-[120px] max-h-[240px] overflow-y-auto leading-relaxed'}`}
              placeholder={t("input_placeholder")}
              value={value}
              onChange={(e) => {
                setValue(e.target.value);
                // Auto-expand up to max height
                const base = isCompact ? 52 : 120;
                const max = isCompact ? 160 : 240;
                e.target.style.height = `${base}px`;
                e.target.style.height = `${Math.min(e.target.scrollHeight, max)}px`;
              }}
              onKeyDown={handleKeyDown}
              disabled={disabled}
            />
            {isCompact && (
              <div className="absolute right-2 top-3 z-10 flex items-center">
                 <button
                   onClick={handleSend}
                   disabled={!value.trim() || disabled}
                   className="p-2 bg-indigo-600 rounded-lg text-white shadow-xl hover:bg-indigo-500 disabled:opacity-50 disabled:grayscale transition-all flex items-center justify-center shrink-0"
                 >
                   <Send className="w-4 h-4" />
                 </button>
              </div>
            )}
          </div>

          <div className={`grid transition-all duration-500 ease-in-out ${isCompact ? 'grid-rows-[0fr] opacity-0' : 'grid-rows-[1fr] opacity-100'}`}>
            <div className="overflow-hidden">
              <div className="flex items-center justify-between px-6 py-4 bg-black/40 border-t border-white/5">
                <div className="flex items-center gap-4">
                  <span className="text-[10px] text-white/10 font-bold uppercase tracking-widest flex items-center gap-2">
                    <kbd className="px-1.5 py-0.5 rounded border border-white/5 bg-white/5">
                      Enter
                    </kbd>{" "}
                    {t("input_send")}
                  </span>
                  <span className="text-[10px] text-white/10 font-bold uppercase tracking-widest flex items-center gap-2">
                    <kbd className="px-1.5 py-0.5 rounded border border-white/5 bg-white/5">
                      Shift + Enter
                    </kbd>{" "}
                    {t("input_line")}
                  </span>
                </div>

                <button
                  onClick={handleSend}
                  disabled={!value.trim() || disabled}
                  className="group/btn relative flex items-center gap-2.5 px-6 py-3 bg-gradient-to-r from-indigo-600 via-indigo-500 to-cyan-500 rounded-2xl text-white font-bold text-xs shadow-xl shadow-indigo-600/10 hover:shadow-indigo-600/20 hover:scale-[1.02] active:scale-95 transition-all disabled:opacity-50 disabled:grayscale disabled:scale-100 overflow-hidden"
                >
                  <div className="absolute inset-x-0 bottom-0 h-[100%] bg-gradient-to-t from-white/10 to-transparent group-hover/btn:translate-y-0 translate-y-full transition-transform" />
                  <span className="relative z-10 flex items-center gap-2">
                    {t("input_process")}{" "}
                    <CornerDownLeft className="w-3.5 h-3.5 text-white/60" />
                  </span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
