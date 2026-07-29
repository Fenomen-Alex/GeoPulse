import type { Component } from 'solid-js';
import { MapCanvas } from '../components/MapCanvas';
import { ConfigPanel } from '../components/ConfigPanel';
import { AnalyticsDrawer } from '../components/AnalyticsDrawer';
import { QuotaBar } from '../components/QuotaBar';
import { TopBar } from '../components/TopBar';

export const AppView: Component<{ user: any }> = (props) => {
  return (
    <div class="h-screen flex flex-col bg-zinc-950 text-zinc-100 overflow-hidden">
      <TopBar />

      <div class="flex items-center gap-3 px-4 py-2 border-b border-zinc-800/60 bg-zinc-900/50">
        <QuotaBar />
        <div class="ml-auto flex items-center gap-3">
          <span class="text-xs text-zinc-400 font-medium">{props.user.name}</span>
          <img
            src={props.user.avatar_url}
            alt={props.user.name}
            class="w-7 h-7 rounded-full border border-zinc-700"
          />
        </div>
      </div>

      <div class="flex-1 flex overflow-hidden">
        <aside class="w-80 flex-shrink-0 border-r border-zinc-800/60 bg-zinc-900/30 overflow-y-auto">
          <ConfigPanel />
        </aside>

          <main class="flex-1 relative h-full overflow-hidden">
              <MapCanvas />
          </main>

        <aside class="w-80 flex-shrink-0 border-l border-zinc-800/60 bg-zinc-900/30 overflow-y-auto">
          <AnalyticsDrawer />
        </aside>
      </div>
    </div>
  );
};