import React from "react";
import { 
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue 
} from "./ui/select";
import { Globe } from "lucide-react";
import { Locale } from "../lib/i18n-types";

interface LanguageSelectorProps {
  currentLocale: Locale;
  onLocaleChange: (locale: Locale) => void;
  t: (key: string, count?: number, vars?: any) => string;
}

const LANGUAGES = [
  { code: 'pt', name: 'Português', flag: 'BR' },
  { code: 'en', name: 'English', flag: 'US' },
  { code: 'es', name: 'Español', flag: 'ES' },
  { code: 'it', name: 'Italiano', flag: 'IT' },
];

export function LanguageSelector({ currentLocale, onLocaleChange, t }: LanguageSelectorProps) {
  return (
    <div className="flex items-center gap-2">
      <Select 
        value={currentLocale} 
        onValueChange={(val) => onLocaleChange(val as Locale)}
      >
        <SelectTrigger className="w-[174px] bg-white/[0.03] border-white/10 text-white/70 text-[10px] font-bold uppercase tracking-widest h-9 rounded-xl hover:bg-white/10 transition-all focus:ring-1 focus:ring-indigo-500/30 glass px-3.5 flex items-center justify-between group">
          <div className="flex items-center gap-2.5 leading-none pointer-events-none">
            <Globe className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
            <div className="pt-[0.5px] truncate max-w-[120px]">
              <SelectValue placeholder={t("config_locale_placeholder")} />
            </div>
          </div>
        </SelectTrigger>
        <SelectContent className="bg-zinc-950 border-white/5 rounded-xl p-1 shadow-2xl z-[200] animate-in fade-in zoom-in-95 backdrop-blur-3xl min-w-[174px]">
          {LANGUAGES.map((lang) => (
            <SelectItem 
              key={lang.code} 
              value={lang.code}
              className="text-[10px] font-bold text-white/40 uppercase tracking-widest hover:text-white hover:bg-white/5 focus:bg-white/5 focus:text-white rounded-lg outline-none cursor-pointer mb-0.5 last:mb-0 transition-colors"
            >
              <div className="flex items-center gap-3 py-1">
                <span className="text-[8px] px-1.5 py-0.5 rounded bg-white/5 border border-white/10 text-white/30 font-black leading-none flex items-center justify-center min-w-[22px]">
                  {lang.flag}
                </span> 
                <span className="leading-none pt-[0.5px]">{lang.name}</span>
              </div>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
