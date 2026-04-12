/**
 * WelcomeScreen - Vectora Welcome/Initial State
 *
 * Displays when no session is active, with features overview
 * and quick action buttons.
 */

import React, { memo } from "react";
import { Sparkles, Code, FileText, Zap } from "lucide-react";
import { Button } from "@/components/ui";

interface WelcomeScreenProps {
  provider?: string;
  onStartSession?: () => void;
}

/**
 * Welcome screen shown when chat is idle
 */
export const WelcomeScreen = memo(({ provider, onStartSession }: WelcomeScreenProps) => {
  return (
    <div className="flex flex-col h-full justify-center items-center gap-6 p-8 text-center">
      {/* Logo & Title */}
      <div className="flex flex-col items-center gap-3">
        <div className="text-4xl">🤖</div>
        <h1 className="text-2xl font-bold text-vscode-editor-foreground">Vectora AI</h1>
        <p className="text-sm text-vscode-descriptionForeground max-w-xs">
          Your local AI coding assistant with RAG and agentic tools
        </p>
      </div>

      {/* Provider Info */}
      {provider && (
        <div className="text-xs text-vscode-descriptionForeground bg-vscode-list-hoverBackground px-3 py-2 rounded">
          Connected to <span className="font-mono">{provider}</span>
        </div>
      )}

      {/* Features Grid */}
      <div className="grid grid-cols-1 gap-3 w-full max-w-xs">
        <FeatureItem icon={<Sparkles size={16} />} title="AI Code Assistant" desc="Ask questions about your code" />
        <FeatureItem icon={<FileText size={16} />} title="RAG & Context" desc="Powered by workspace search" />
        <FeatureItem icon={<Code size={16} />} title="Tool Execution" desc="File editing and commands" />
        <FeatureItem icon={<Zap size={16} />} title="Real-time Streaming" desc="See responses as they generate" />
      </div>

      {/* Quick Start */}
      <div className="flex flex-col gap-2 w-full max-w-xs">
        <p className="text-xs text-vscode-descriptionForeground mb-2">Get started by typing a question below</p>
        <Button variant="primary" className="w-full text-sm" onClick={onStartSession}>
          Start Conversation
        </Button>
      </div>

      {/* Footer Info */}
      <div className="text-[10px] text-vscode-descriptionForeground opacity-50">
        Vectora v0.1.0 • Phase 4d
      </div>
    </div>
  );
});

WelcomeScreen.displayName = "WelcomeScreen";

/**
 * Individual feature item
 */
const FeatureItem = memo(
  ({ icon, title, desc }: { icon: React.ReactNode; title: string; desc: string }) => (
    <div className="flex items-start gap-3 p-3 rounded border border-vscode-widget-border hover:bg-vscode-list-hoverBackground transition-colors">
      <div className="text-vscode-button-background flex-shrink-0 mt-0.5">{icon}</div>
      <div className="text-left">
        <div className="text-sm font-medium text-vscode-editor-foreground">{title}</div>
        <div className="text-xs text-vscode-descriptionForeground">{desc}</div>
      </div>
    </div>
  )
);

FeatureItem.displayName = "FeatureItem";
