import React, { useState, useRef, useEffect } from "react";
import { cn } from "@utils/cn";

interface DropdownProps {
  trigger: React.ReactNode;
  children: React.ReactNode;
  className?: string;
  contentClassName?: string;
  align?: "left" | "right";
}

export const Dropdown: React.FC<DropdownProps> = ({
  align = "left",
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen]);

  return (
    <div className={cn("relative inline-block", className)} ref={containerRef}>
      <div onClick={() => setIsOpen(!isOpen)} className="cursor-pointer">
        {trigger}
      </div>

      {isOpen && (
        <div
          className={cn(
            "absolute bottom-full mb-2 z-50 min-w-[180px] rounded-lg border border-vscode-border/20 bg-vscode-editor-background shadow-xl animate-in fade-in slide-in-from-bottom-2 duration-200 overflow-hidden backdrop-blur-md",
            align === "left" ? "left-0" : "right-0",
          )}
        >
          {children}
        </div>
      )}
    </div>
  );
};

export const DropdownItem: React.FC<{
  onClick: () => void;
  children: React.ReactNode;
  className?: string;
  active?: boolean;
}> = ({ onClick, children, className, active }) => (
  <div
    onClick={(e) => {
      e.stopPropagation();
      onClick();
    }}
    className={cn(
      "px-3 py-2 text-[12px] cursor-pointer transition-colors flex flex-col gap-0.5",
      active
        ? "bg-vscode-button-background text-vscode-button-foreground"
        : "hover:bg-vscode-toolbar-hoverBackground text-vscode-descriptionForeground hover:text-vscode-editor-foreground",
    )}
  >
    {children}
  </div>
);
