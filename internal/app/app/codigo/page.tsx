'use client';

import React from 'react';
import { CodeDiffVisualizer } from '@/components/CodeDiffVisualizer';

export default function CodigoPage() {
  return (
    <div className="w-full h-full flex flex-col bg-zinc-950">
      <div className="flex-1 overflow-hidden">
        <CodeDiffVisualizer />
      </div>
    </div>
  );
}
