import type { Component } from 'solid-js';
import { Gauge } from 'lucide-solid';

export const QuotaBar: Component<{ class?: string }> = (props) => {
  return (
    <div class={`flex items-center gap-3 rounded-xl border border-zinc-800/80 bg-zinc-900/80 px-4 py-2 shadow-2xl backdrop-blur-md ${props.class ?? ''}`}>
      <Gauge class="h-4 w-4 text-zinc-400" />
      <span class="text-xs text-zinc-400">API Quota:</span>
      <span class="text-xs text-zinc-100 font-mono tabular-nums">—/—</span>
    </div>
  );
};
