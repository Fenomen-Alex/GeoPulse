import { Show, For } from 'solid-js';
import type { Component } from 'solid-js';
import { BarChart3 } from 'lucide-solid';
import { Panel } from './Panel';
import { analysisResult } from '../store/analysisStore';
import { ISOCHRONE_COLORS } from '../constants/isochrone';

export const AnalyticsDrawer: Component<{ class?: string }> = (props) => {
  return (
    <div class={props.class ?? ''}>
      <Show
        when={analysisResult()}
        fallback={
          <Panel title="Analytics">
            <div class="flex flex-col items-center justify-center py-8 text-zinc-500">
              <BarChart3 class="h-8 w-8 mb-2 opacity-40" />
              <p class="text-xs">Run an analysis to see results</p>
            </div>
          </Panel>
        }
      >
        {(result) => (
          <Panel title="Analysis Results">
            <div class="space-y-3">
              <Show when={result().totalArea != null}>
                <div>
                  <span class="text-xs text-zinc-400">Total Reachable Area</span>
                  <p class="text-lg font-semibold text-zinc-100 font-mono tabular-nums">
                    {result().totalArea?.toFixed(1)} km²
                  </p>
                </div>
              </Show>

              <Show when={result().poiCount != null}>
                <div>
                  <span class="text-xs text-zinc-400">Points of Interest</span>
                  <p class="text-lg font-semibold text-zinc-100 font-mono tabular-nums">
                    {result().poiCount}
                  </p>
                </div>
              </Show>

              <Show when={result().score != null}>
                <div>
                  <span class="text-xs text-zinc-400">Accessibility Score</span>
                  <p class="text-lg font-semibold text-cyan-400 font-mono tabular-nums">
                    {result().score}/100
                  </p>
                </div>
              </Show>

              {/* Isochrone Legend */}
              <div>
                <span class="text-xs text-zinc-400 block mb-2">Isochrone Bands</span>
                <div class="space-y-1.5">
                  <For each={Object.entries(ISOCHRONE_COLORS)}>
                    {([_minutes, config]) => (
                      <div class="flex items-center gap-2">
                        <span
                          class="inline-block h-3 w-3 rounded-sm border"
                          style={{
                            'background-color': config.fill,
                            opacity: config.fillOpacity,
                            'border-color': config.stroke,
                          }}
                        />
                        <span class="text-xs text-zinc-300">{config.label}</span>
                      </div>
                    )}
                  </For>
                </div>
              </div>
            </div>
          </Panel>
        )}
      </Show>
    </div>
  );
};
