import { For, Show } from 'solid-js';
import type { Component } from 'solid-js';
import { Route } from 'lucide-solid';
import { Panel } from './Panel';
import { routeResult } from '../store/analysisStore';

function formatDistance(m: number) {
  if (m >= 1000) return `${(m / 1000).toFixed(2)} km`;
  return `${Math.round(m)} m`;
}

function formatDuration(s: number) {
  const mins = Math.round(s / 60);
  if (mins < 60) return `${mins} min`;
  return `${Math.floor(mins / 60)} h ${mins % 60} min`;
}

export const RouteDrawer: Component<{ class?: string }> = (props) => {
  return (
    <div class={props.class ?? ''}>
      <Show
        when={routeResult()}
        fallback={
          <Panel title="Route Navigation">
            <div class="flex flex-col items-center justify-center py-8 text-zinc-500">
              <Route class="h-8 w-8 mb-2 opacity-40" />
              <p class="text-xs text-center px-2">
                Set a start point and destination, then get street-level directions.
              </p>
            </div>
          </Panel>
        }
      >
        {(r) => (
          <Panel title="Route Summary">
            <div class="space-y-4">
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <span class="text-xs text-zinc-400">Distance</span>
                  <p class="text-lg font-semibold text-zinc-100 font-mono tabular-nums">
                    {formatDistance(r().distance_m)}
                  </p>
                </div>
                <div>
                  <span class="text-xs text-zinc-400">Duration</span>
                  <p class="text-lg font-semibold text-cyan-400 font-mono tabular-nums">
                    {formatDuration(r().duration_s)}
                  </p>
                </div>
              </div>

              <div>
                <span class="text-xs text-zinc-400 block mb-2">Turn-by-Turn Directions</span>
                <ol class="space-y-3 max-h-80 overflow-y-auto pr-1">
                  <For each={r().steps}>
                    {(step, i) => (
                      <li class="flex gap-2.5 text-xs text-zinc-300">
                        <span class="w-5 h-5 shrink-0 rounded-full bg-zinc-800 text-cyan-400 inline-flex items-center justify-center font-mono tabular-nums">
                          {i() + 1}
                        </span>
                        <div class="min-w-0">
                          <p>{step.instruction}</p>
                          <p class="text-zinc-500">
                            {formatDistance(step.distance)} · {formatDuration(step.duration)}
                          </p>
                        </div>
                      </li>
                    )}
                  </For>
                </ol>
              </div>
            </div>
          </Panel>
        )}
      </Show>
    </div>
  );
};
