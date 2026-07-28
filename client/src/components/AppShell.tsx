import type { Component } from 'solid-js';
import { MapCanvas } from './MapCanvas';
import { TopBar } from './TopBar';
import { ConfigPanel } from './ConfigPanel';
import { AnalyticsDrawer } from './AnalyticsDrawer';
import { QuotaBar } from './QuotaBar';

export const AppShell: Component = () => {
  return (
    <div class="relative h-screen w-screen overflow-hidden bg-zinc-950 font-sans antialiased select-none">
      {/* Layer 0: Map Canvas (Interactions Enabled) */}
      <div class="absolute inset-0 z-0 pointer-events-auto">
        <MapCanvas />
      </div>

      {/* Layer 1: Floating UI Overlay Container (Click-Through) */}
      <div class="absolute inset-0 z-10 pointer-events-none flex flex-col justify-between p-4">
        <TopBar class="pointer-events-auto" />
        <div class="flex flex-1 items-start justify-between gap-4 my-4">
          <ConfigPanel class="pointer-events-auto" />
          <AnalyticsDrawer class="pointer-events-auto" />
        </div>
        <QuotaBar class="pointer-events-auto" />
      </div>
    </div>
  );
};
