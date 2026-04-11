import React from "react";
import { ChevronDown, Sparkles, AlertTriangle } from "lucide-react";
import { Dropdown, DropdownItem } from "./Dropdown";

export type ModelItem = {
  id: string;
  name: string;
  info?: string;
  isNew?: boolean;
  isThinking?: boolean;
  group: string;
};

export const MODELS: ModelItem[] = [
  { id: "gemini-3.1-pro-high", name: "Gemini 3.1 Pro (High)", info: "High", isNew: true, group: "Google" },
  { id: "gemini-3.1-pro-low", name: "Gemini 3.1 Pro (Low)", info: "Low", isNew: true, group: "Google" },
  { id: "gemini-3-flash", name: "Gemini 3 Flash", group: "Google" },
  { id: "claude-sonnet-4.6", name: "Claude Sonnet 4.6", info: "Thinking", isThinking: true, group: "Anthropic" },
  { id: "claude-opus-4.6", name: "Claude Opus 4.6", info: "Thinking", isThinking: true, group: "Anthropic" },
  { id: "gpt-oss-120b", name: "GPT-OSS 120B", info: "Medium", group: "Open Source" },
];

interface ModelSelectorProps {
  selectedModel: ModelItem;
  onSelect: (model: ModelItem) => void;
}

export const ModelSelector: React.FC<ModelSelectorProps> = ({ selectedModel, onSelect }) => {
  return (
    <Dropdown
      trigger={
        <div className="flex items-center gap-1 px-2 py-1 rounded-md hover:bg-vscode-toolbar-hoverBackground cursor-pointer text-vscode-descriptionForeground group">
          <Sparkles size={12} className="text-vscode-button-background" />
          <span className="text-[11px] font-bold group-hover:text-vscode-editor-foreground transition-colors">
            {selectedModel.name}
          </span>
          <ChevronDown size={12} className="opacity-50" />
        </div>
      }
      contentClassName="min-w-[240px]"
    >
      <div className="px-3 py-1.5 text-[10px] font-bold text-vscode-descriptionForeground/60 uppercase">Model</div>
      {MODELS.map((model) => (
        <DropdownItem
          key={model.id}
          onClick={() => onSelect(model)}
          active={selectedModel.id === model.id}
          className="flex-row items-center justify-between"
        >
          <div className="flex items-center gap-2">
            <span className="font-medium">{model.name}</span>
            {model.isThinking && <AlertTriangle size={10} className="opacity-60" />}
          </div>
          <div className="flex items-center gap-1">
            {model.isNew && (
              <span className="text-[8px] bg-vscode-badge-background px-1 rounded opacity-80 uppercase font-black">
                New
              </span>
            )}
          </div>
        </DropdownItem>
      ))}
    </Dropdown>
  );
};
