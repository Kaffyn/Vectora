import React from "react";
import type { LucideIcon } from "lucide-react";

interface SuggestionCardProps {
  icon: LucideIcon;
  label: string;
  query: string;
  onClick: (query: string) => void;
}

export function SuggestionCard({
  icon: Icon,
  label,
  query,
  onClick,
}: SuggestionCardProps) {
  return (
    <button
      onClick={() => onClick(query)}
      className="flex flex-col items-start gap-4 p-5 rounded-2xl bg-zinc-900/50 border border-white/5 hover:bg-zinc-800/80 hover:border-indigo-500/30 transition-all text-left group w-full shadow-lg"
    >
      <div className="w-10 h-10 rounded-xl bg-indigo-500/10 flex items-center justify-center group-hover:scale-110 transition-transform duration-500">
        <Icon className="w-5 h-5 text-indigo-400" />
      </div>
      <span className="text-xs font-bold text-white/50 group-hover:text-white/80 transition-colors uppercase tracking-widest leading-relaxed">
        {label}
      </span>
    </button>
  );
}
