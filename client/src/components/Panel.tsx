import type { Component, JSX } from 'solid-js';

export const Panel: Component<{ title: string; children: JSX.Element; class?: string }> = (props) => {
  return (
    <div class={`w-80 rounded-xl border border-zinc-800/80 bg-zinc-900/80 p-4 shadow-2xl backdrop-blur-md transition-all duration-200 ${props.class ?? ''}`}>
      <div class="flex items-center justify-between pb-3 mb-3 border-b border-zinc-800/60">
        <h3 class="text-sm font-semibold text-zinc-100 tracking-wide">{props.title}</h3>
      </div>
      {props.children}
    </div>
  );
};
