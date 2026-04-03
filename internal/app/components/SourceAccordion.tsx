import React from "react";
import { ChevronRight, FileText, ExternalLink } from "lucide-react";
import * as Accordion from "@radix-ui/react-accordion";

interface Source {
  id: string;
  name: string;
  path: string;
  relevance: number;
}

interface SourceAccordionProps {
  sources: Source[];
}

export function SourceAccordion({ sources }: SourceAccordionProps) {
  if (sources.length === 0) return null;

  return (
    <Accordion.Root type="single" collapsible className="mt-6 w-full max-w-3xl">
      <Accordion.Item
        value="sources"
        className="border border-white/5 rounded-xl overflow-hidden bg-white/5"
      >
        <Accordion.Header>
          <Accordion.Trigger className="flex items-center gap-3 w-full px-4 py-3 text-sm font-medium text-white/50 hover:bg-white/5 transition-all group">
            <ChevronRight className="w-4 h-4 text-white/40 group-data-[state=open]:rotate-90 transition-transform" />
            <span className="flex items-center gap-2">
              <span className="text-lg">⪧</span> {sources.length} sources
              consulted
            </span>
          </Accordion.Trigger>
        </Accordion.Header>
        <Accordion.Content className="px-4 pb-4 pt-1 animate-in slide-in-from-top-2 duration-300">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mt-2">
            {sources.map((source) => (
              <div
                key={source.id}
                className="flex items-center justify-between p-3 rounded-lg bg-white/5 border border-white/5 hover:border-white/10 transition-colors cursor-pointer group"
              >
                <div className="flex items-center gap-3 overflow-hidden">
                  <FileText className="w-4 h-4 text-white/30" />
                  <div className="flex flex-col overflow-hidden">
                    <span className="text-xs font-medium text-white/80 truncate">
                      {source.name}
                    </span>
                    <span className="text-[10px] text-white/40 truncate">
                      {source.path}
                    </span>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-[10px] font-bold text-orange-400 bg-orange-400/10 px-1.5 py-0.5 rounded">
                    {Math.round(source.relevance * 100)}%
                  </span>
                  <ExternalLink className="w-3 h-3 text-white/20 group-hover:text-white/60 transition-colors" />
                </div>
              </div>
            ))}
          </div>
        </Accordion.Content>
      </Accordion.Item>
    </Accordion.Root>
  );
}
