import React from "react";
import { Check, X } from "lucide-react";

export const BatchFilePermission = ({ files, onPermissionResponse, ts }: any) => {
  return (
    <div className="flex flex-col gap-1 p-2 bg-vscode-input-background border border-vscode-widget-border rounded-md mt-2">
      <span className="text-xs font-bold opacity-80 uppercase">Permissões de Arquivo</span>
      <div className="flex flex-col gap-1">
        {files.map((file: string) => (
          <div key={file} className="flex items-center justify-between text-sm py-1 border-b border-vscode-widget-border last:border-0">
            <span className="truncate flex-1">{file}</span>
            <div className="flex gap-2">
              <Check className="w-4 h-4 text-vscode-charts-green cursor-pointer" />
              <X className="w-4 h-4 text-vscode-errorForeground cursor-pointer" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
