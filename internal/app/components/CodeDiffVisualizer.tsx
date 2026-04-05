'use client';

import React, { useState } from 'react';
import { ChevronDown } from 'lucide-react';

export function CodeDiffVisualizer() {
  const [selectedFile, setSelectedFile] = useState<string | null>(null);

  return (
    <div className="w-full h-full flex flex-col bg-zinc-950">
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-zinc-800">
        <h2 className="text-sm font-medium text-zinc-300">Código</h2>
        <span className="text-xs text-zinc-500 ml-auto">Visualizador de Arquivos</span>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-hidden flex">
        {/* File Tree */}
        <div className="w-64 border-r border-zinc-800 overflow-y-auto bg-zinc-900">
          <div className="p-3 space-y-2">
            <div className="text-xs text-zinc-400 font-medium">Arquivos do Workspace</div>
            <div className="text-xs text-zinc-500 italic">
              Nenhum arquivo selecionado
            </div>
          </div>
        </div>

        {/* Code Editor Area */}
        <div className="flex-1 flex flex-col">
          {/* Editor Placeholder */}
          <div className="flex-1 flex items-center justify-center bg-zinc-950">
            <div className="text-center">
              <div className="text-zinc-400 text-sm">
                Selecione um arquivo para visualizá-lo
              </div>
              <div className="text-zinc-600 text-xs mt-2">
                Os arquivos do workspace aparecerão na árvore à esquerda
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
