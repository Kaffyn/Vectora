import React from "react";
import { Plus, Image as ImageIcon, AtSign, Edit3 } from "lucide-react";
import { Dropdown, DropdownItem } from "./Dropdown";

export const ContextSelector: React.FC = () => {
  return (
    <Dropdown
      trigger={
        <div className="p-1.5 rounded-md hover:bg-vscode-toolbar-hoverBackground text-vscode-descriptionForeground transition-colors">
          <Plus size={14} />
        </div>
      }
    >
      <div className="px-3 py-1.5 text-[10px] font-bold text-vscode-descriptionForeground/60 uppercase">
        Add context
      </div>
      <DropdownItem onClick={() => {}} className="flex-row items-center gap-2">
        <ImageIcon size={14} /> <span>Media</span>
      </DropdownItem>
      <DropdownItem onClick={() => {}} className="flex-row items-center gap-2">
        <AtSign size={14} /> <span>Mentions</span>
      </DropdownItem>
      <DropdownItem onClick={() => {}} className="flex-row items-center gap-2 pb-2">
        <Edit3 size={14} /> <span>Workflows</span>
      </DropdownItem>
    </Dropdown>
  );
};
