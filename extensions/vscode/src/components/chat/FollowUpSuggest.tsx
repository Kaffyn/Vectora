import React from "react";
import { MessageCircleQuestion } from "lucide-react";

export const FollowUpSuggest = ({ message, onSuggestionClick }: any) => {
  return (
    <div className="flex flex-col gap-2 p-3 bg-vscode-notifications-background border border-vscode-widget-border rounded-md mt-2">
      <div className="flex items-center gap-2 text-vscode-foreground font-bold">
        <MessageCircleQuestion className="w-4 h-4" />
        <span>Sugestão de Ação</span>
      </div>
      <div className="flex flex-wrap gap-2 mt-1">
        {/* Renderiza sugestões vindas da nossa API */}
      </div>
    </div>
  );
};
