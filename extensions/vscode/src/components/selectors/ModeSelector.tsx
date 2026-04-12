import React from "react";
import { ChevronDown } from "lucide-react";
import { Dropdown, DropdownItem } from "./Dropdown";

export type ModeItem = { id: string; name: string; description: string };

export const MODES: ModeItem[] = [
    id: "planning",
    name: "Planning",
    description: "Agent can plan before executing tasks. Use for deep research, complex tasks, or collaborative work",
  },
    id: "fast",
    name: "Fast",
    description: "Agent will execute tasks directly. Use for simple tasks that can be completed faster",
  },
];

interface ModeSelectorProps {
  selectedMode: ModeItem;
  onSelect: (mode: ModeItem) => void;
}

export const ModeSelector: React.FC<ModeSelectorProps> = ({ selectedMode, onSelect }) => {
  return (
    <Dropdown
      trigger={
        <div className="flex items-center gap-1 px-2 py-1 rounded-md hover:bg-vscode-toolbar-hoverBackground cursor-pointer text-vscode-descriptionForeground group">
          <span className="text-[11px] font-bold group-hover:text-vscode-editor-foreground transition-colors">
            {selectedMode.name}
          </span>
          <ChevronDown size={12} className="opacity-50" />
        </div>
      }
      contentClassName="min-w-[280px]"
    >
      {MODES.map((mode) => (
        <DropdownItem
          key={mode.id}
          onClick={() => onSelect(mode)}
          active={selectedMode.id === mode.id}
          className="py-2.5"
        >
          <span className="font-bold">{mode.name}</span>
          <span className="opacity-60 text-[10px] leading-tight mt-0.5">{mode.description}</span>
        </DropdownItem>
      ))}
    </Dropdown>
  );
};
