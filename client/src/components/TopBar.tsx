import type { Component } from 'solid-js';
import { Navigation } from 'lucide-solid';

export const TopBar: Component<{ class?: string }> = (props) => {
  return (
    <div class={`flex items-center gap-3 rounded-xl border border-zinc-800/80 bg-zinc-900/80 px-4 py-2.5 shadow-2xl backdrop-blur-md ${props.class ?? ''}`}>
      <Navigation class="h-5 w-5 text-cyan-400" />
      <h1 class="text-sm font-semibold text-zinc-100 tracking-wide">GeoPulse</h1>
      <span class="text-xs text-zinc-400">Spatial Analysis Workbench</span>
    </div>
  );
};
