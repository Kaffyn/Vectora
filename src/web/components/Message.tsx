"use client";

import React from "react";
import { Copy, Pencil, Loader2 } from "lucide-react";
import { SourceAccordion } from "./SourceAccordion";

interface MessageProps {
  role: "user" | "ai";
  content: string;
  sources?: Array<{ id: string; name: string; path: string; relevance: number }>;
  isThinking?: boolean;
  t: (key: string, count?: number, vars?: any) => string;
}

export function Message({ role, content, sources, isThinking, t }: MessageProps) {
  // ── USER BUBBLE ──────────────────────────────────────────────────────────────
  if (role === "user") {
    return (
      <div className="flex justify-end mb-2 group/user">
        <div className="flex flex-col items-end gap-1 max-w-[78%]">
          <div className="bg-indigo-600/25 border border-indigo-500/20 px-4 py-2.5 rounded-2xl rounded-tr-sm text-left">
            <p className="text-[14px] font-medium text-white/90 leading-relaxed whitespace-pre-wrap">
              {content}
            </p>
          </div>
          {/* Inline micro-actions visible on hover */}
          <div className="flex items-center gap-1.5 opacity-0 group-hover/user:opacity-100 transition-opacity duration-200">
            <button
              className="p-1 hover:bg-white/10 rounded-lg text-white/25 hover:text-white/70 transition-colors"
              title="Editar"
            >
              <Pencil className="w-3 h-3" />
            </button>
            <button
              className="p-1 hover:bg-white/10 rounded-lg text-white/25 hover:text-white/70 transition-colors"
              title="Copiar"
              onClick={() => navigator.clipboard.writeText(content)}
            >
              <Copy className="w-3 h-3" />
            </button>
          </div>
        </div>
      </div>
    );
  }

  // ── AI BUBBLE ────────────────────────────────────────────────────────────────
  return (
    <div className="flex justify-start mb-2">
      <div className="flex flex-col items-start gap-1 max-w-[78%]">
        {/* Thinking indicator */}
        {isThinking && (
          <div className="flex items-center gap-2 px-4 py-2.5 rounded-2xl rounded-tl-sm bg-white/[0.04] border border-white/[0.06]">
            <Loader2 className="w-3.5 h-3.5 text-indigo-400 animate-spin shrink-0" />
            <span className="text-[13px] text-white/40 font-medium">
              {t("message_thinking")}
            </span>
          </div>
        )}

        {!isThinking && (
          <div className="group/ai flex flex-col gap-1">
            {/* Same bubble shape as user but a neutral bg */}
            <div className="bg-white/[0.04] border border-white/[0.06] px-4 py-2.5 rounded-2xl rounded-tl-sm">
              <p className="text-[14px] font-medium text-white/88 leading-relaxed whitespace-pre-wrap">
                {content}
              </p>
              {sources && sources.length > 0 && (
                <div className="mt-3 pt-3 border-t border-white/5">
                  <SourceAccordion sources={sources} />
                </div>
              )}
            </div>
            {/* Copy action, mirrors hover micro-action of user */}
            <div className="flex items-center gap-1.5 opacity-0 group-hover/ai:opacity-100 transition-opacity duration-200">
              <button
                className="p-1 hover:bg-white/10 rounded-lg text-white/25 hover:text-white/70 transition-colors"
                title="Copiar resposta"
                onClick={() => navigator.clipboard.writeText(content)}
              >
                <Copy className="w-3 h-3" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
