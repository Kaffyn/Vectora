import React from "react";
import { FileDiff, Check, X } from "lucide-react";

export const BatchDiffApproval = ({ files, ts }: any) => {
  return (
    <div className="flex flex-col gap-2 p-3 bg-vscode-editor-background border border-vscode-widget-border rounded-md mt-2">
      <div className="flex items-center gap-2 text-vscode-foreground font-bold">
        <FileDiff className="w-4 h-4 text-vscode-charts-blue" />
        <span>Alterações Pendentes</span>
      </div>
      <div className="flex flex-col gap-1">
        {files.map((file: any) => (
          <div key={file.path} className="flex items-center justify-between text-xs py-1 px-2 hover:bg-vscode-list-hoverBackground rounded">
            <span className="truncate">{file.path}</span>
            <div className="flex gap-2">
              <span className="text-vscode-charts-green">+{file.additions}</span>
              <span className="text-vscode-charts-red">-{file.deletions}</span>
            </div>
          </div>
        ))}
      </div>
      <div className="flex gap-2 mt-2">
        <button className="flex-1 flex items-center justify-center gap-1 py-1 bg-vscode-button-background hover:bg-vscode-button-hoverBackground text-vscode-button-foreground rounded text-sm">
          <Check className="w-4 h-4" /> Aprovar Tudo
        </button>
        <button className="flex-1 flex items-center justify-center gap-1 py-1 bg-vscode-button-secondaryBackground hover:bg-vscode-button-secondaryHoverBackground text-vscode-button-secondaryForeground rounded text-sm">
          <X className="w-4 h-4" /> Rejeitar
        </button>
      </div>
    </div>
  );
};
