'use client';

import React, { useState } from 'react';
import { Settings, Package, Cog } from 'lucide-react';

export default function ManagerPage() {
  const [activeTab, setActiveTab] = useState<'models' | 'config'>('models');

  return (
    <div className="w-full h-full flex flex-col bg-zinc-950">
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-zinc-800">
        <Settings className="w-5 h-5 text-emerald-500" />
        <h2 className="text-sm font-medium text-zinc-300">Gerenciador</h2>
        <span className="text-xs text-zinc-500 ml-auto">Modelos & Configuração</span>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-zinc-800 px-4">
        <button
          onClick={() => setActiveTab('models')}
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'models'
              ? 'border-emerald-500 text-emerald-400'
              : 'border-transparent text-zinc-400 hover:text-zinc-300'
          }`}
        >
          <div className="flex items-center gap-2">
            <Package className="w-4 h-4" />
            Modelos
          </div>
        </button>
        <button
          onClick={() => setActiveTab('config')}
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
            activeTab === 'config'
              ? 'border-emerald-500 text-emerald-400'
              : 'border-transparent text-zinc-400 hover:text-zinc-300'
          }`}
        >
          <div className="flex items-center gap-2">
            <Cog className="w-4 h-4" />
            Configuração
          </div>
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-6">
        {activeTab === 'models' && (
          <div>
            <h3 className="text-lg font-medium text-zinc-300 mb-4">Gerenciador de Modelos</h3>
            <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-6 text-center">
              <Package className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
              <p className="text-sm text-zinc-400">
                Nenhum modelo instalado ainda.
              </p>
              <p className="text-xs text-zinc-600 mt-2">
                Use o Setup Installer para baixar modelos Qwen.
              </p>
            </div>
          </div>
        )}

        {activeTab === 'config' && (
          <div>
            <h3 className="text-lg font-medium text-zinc-300 mb-4">Configuração Global</h3>
            <div className="space-y-4 max-w-2xl">
              {/* API Key Input */}
              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Chave Gemini API (opcional)
                </label>
                <input
                  type="password"
                  placeholder="sk-..."
                  className="w-full px-3 py-2 bg-zinc-900 border border-zinc-700 rounded text-sm text-zinc-300 placeholder-zinc-600 focus:outline-none focus:border-emerald-500"
                />
              </div>

              {/* Provider Selection */}
              <div>
                <label className="block text-sm font-medium text-zinc-300 mb-2">
                  Provider LLM Preferido
                </label>
                <div className="flex gap-2">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input type="radio" name="provider" value="qwen_local" defaultChecked />
                    <span className="text-sm text-zinc-400">Qwen Local</span>
                  </label>
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input type="radio" name="provider" value="gemini" />
                    <span className="text-sm text-zinc-400">Gemini API</span>
                  </label>
                </div>
              </div>

              {/* Save Button */}
              <div className="pt-4">
                <button className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white text-sm rounded transition-colors">
                  Salvar Configurações
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
