import { onMount, Show } from 'solid-js';
import type { Component } from 'solid-js';
import { LogOut } from 'lucide-solid';
import { MapCanvas } from '../components/MapCanvas';
import { SearchBar } from '../components/SearchBar';
import { ConfigPanel } from '../components/ConfigPanel';
import { AnalyticsDrawer } from '../components/AnalyticsDrawer';
import { RouteDrawer } from '../components/RouteDrawer';
import { QuotaBar } from '../components/QuotaBar';
import { TopBar } from '../components/TopBar';
import { QuotaGateModal } from '../components/QuotaGateModal';
import { activeTool, loadRemainingQuota, showQuotaModal } from '../store/analysisStore';
import { loadWorkspace, startWorkspacePersistence } from '../store/workspaceStore';

export const AppView: Component<{ user: any; onLogout: () => void }> = (props) => {
  onMount(async () => {
    void loadRemainingQuota();
    // Restore any saved snapshot first, then subscribe to changes so the
    // hydration writes don't immediately overwrite the saved workspace.
    await loadWorkspace();
    startWorkspacePersistence();
  });

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
          <button
            type="button"
            onClick={props.onLogout}
            class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-700/70 px-2.5 py-1.5 text-xs text-zinc-400 transition-colors hover:border-zinc-600 hover:text-zinc-100"
            aria-label="Sign out"
          >
            <LogOut class="h-3.5 w-3.5" />
            <span class="hidden sm:inline">Sign out</span>
          </button>
        </div>
      </div>

      <div class="flex-1 flex flex-col overflow-y-auto lg:flex-row lg:overflow-hidden">
        <aside class="w-full shrink-0 border-b border-zinc-800/60 bg-zinc-900/30 overflow-y-auto lg:w-80 lg:border-b-0 lg:border-r">
          <ConfigPanel />
        </aside>

        <main class="relative min-h-[55vh] flex-1 overflow-hidden lg:min-h-0">
          <MapCanvas />
          <SearchBar class="absolute top-4 left-4 right-4 z-10000 mx-auto" />
        </main>

        <aside class="w-full shrink-0 border-t border-zinc-800/60 bg-zinc-900/30 overflow-y-auto lg:w-80 lg:border-l lg:border-t-0">
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
