import { Show } from 'solid-js';
import type { Component } from 'solid-js';
import { MapCanvas } from '../components/MapCanvas';
import { SearchBar } from '../components/SearchBar';
import { ConfigPanel } from '../components/ConfigPanel';
import { AnalyticsDrawer } from '../components/AnalyticsDrawer';
import { RouteDrawer } from '../components/RouteDrawer';
import { QuotaBar } from '../components/QuotaBar';
import { TopBar } from '../components/TopBar';
import { QuotaGateModal } from '../components/QuotaGateModal';
import { activeTool, showQuotaModal } from '../store/analysisStore';

export const AppView: Component<{ user: any }> = (props) => {
  return (
    <div class="h-screen flex flex-col bg-zinc-950 text-zinc-100 overflow-hidden">
      <TopBar />

      <div class="flex items-center gap-3 px-4 py-2 border-b border-zinc-800/60 bg-zinc-900/50">
        <QuotaBar />
        <div class="ml-auto flex items-center gap-3">
          <span class="text-xs text-zinc-400 font-medium">{props.user.name}</span>
          <Show when={props.user.avatar_url}>
            <img
              src={props.user.avatar_url}
              alt={props.user.name}
              class="w-7 h-7 rounded-full border border-zinc-700"
            />
          </Show>
        </div>
      </div>

      <div class="flex-1 flex overflow-hidden">
        <aside class="w-80 shrink-0 border-r border-zinc-800/60 bg-zinc-900/30 overflow-y-auto">
          <ConfigPanel />
        </aside>

        <main class="flex-1 relative h-full overflow-hidden">
          <MapCanvas />
          <SearchBar class="absolute top-4 left-1/2 -translate-x-1/2 z-10000" />
        </main>

        <aside class="w-80 shrink-0 border-l border-zinc-800/60 bg-zinc-900/30 overflow-y-auto">
          <Show when={activeTool() === 'route'} fallback={<AnalyticsDrawer />}>
            <RouteDrawer />
          </Show>
        </aside>
      </div>

      <Show when={showQuotaModal()}>
        <QuotaGateModal />
      </Show>
    </div>
  );
};
