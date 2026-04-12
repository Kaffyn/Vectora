import React from "react";
import { ChevronDown, Zap, Terminal, Edit3 } from "lucide-react";
import { Dropdown, DropdownItem } from "./Dropdown";

export type PolicyItem = { id: string; name: string; description: string; icon: React.ReactNode };

export const POLICIES: PolicyItem[] = [
    id: "ask",
    name: "Ask before edits",
    description: "Based on diffs, where you must accept changes",
    icon: <Edit3 size={14} />,
  },
    id: "automatic",
    name: "Edit automatically",
    description: "Can edit files, but terminal/search/embed need auth",
    icon: <Terminal size={14} />,
  },
    id: "yolo",
    name: "YOLO",
    description: "Full autonomous mode, everything allowed",
    icon: <Zap size={14} className="text-yellow-500" />,
  },
];

interface PolicySelectorProps {
  selectedPolicy: PolicyItem;
  onSelect: (policy: PolicyItem) => void;
}

export const PolicySelector: React.FC<PolicySelectorProps> = ({ selectedPolicy, onSelect }) => {
  return (
    <Dropdown
      trigger={
        <div
          className="flex items-center gap-1 px-2 py-1 rounded-md hover:bg-vscode-toolbar-hoverBackground cursor-pointer text-vscode-descriptionForeground group"
          title="Safety Policy"
        >
          {selectedPolicy.icon}
          <ChevronDown size={12} className="opacity-50" />
        </div>
      }
      contentClassName="min-w-[280px]"
      align="right"
    >
      <div className="px-3 py-1.5 text-[10px] font-bold text-vscode-descriptionForeground/60 uppercase">
        Safety Policy
      </div>
      {POLICIES.map((policy) => (
        <DropdownItem
          key={policy.id}
          onClick={() => onSelect(policy)}
          active={selectedPolicy.id === policy.id}
          className="py-2.5"
        >
          <div className="flex items-center gap-2">
            {policy.icon}
            <span className="font-bold">{policy.name}</span>
          </div>
          <span className="opacity-60 text-[10px] leading-tight mt-0.5">{policy.description}</span>
        </DropdownItem>
      ))}
    </Dropdown>
  );
};
