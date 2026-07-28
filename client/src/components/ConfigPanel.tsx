import { For } from 'solid-js';
import type { Component } from 'solid-js';
import { Footprints, Bike, Car } from 'lucide-solid';
import { Panel } from './Panel';
import { AnalysisButton } from './AnalysisButton';
import { origin, travelMode, setTravelMode, maxMinutes, setMaxMinutes } from '../store/analysisStore';

const TRAVEL_MODES = [
  { value: 'walk' as const, label: 'Walk', icon: Footprints },
  { value: 'bike' as const, label: 'Bike', icon: Bike },
  { value: 'drive' as const, label: 'Drive', icon: Car },
];

const MINUTE_OPTIONS = [5, 10, 15, 30, 45, 60];

export const ConfigPanel: Component<{ class?: string }> = (props) => {
  return (
    <div class={props.class ?? ''}>
      <Panel title="Analysis Configuration">
        <div class="space-y-4">
          {/* Origin Display */}
          <div>
            <label class="text-xs text-zinc-400 mb-1.5 block">Origin</label>
            {origin() ? (
              <p class="text-xs text-zinc-100 font-mono tabular-nums">
                {origin()!.lat.toFixed(4)}° N, {origin()!.lng.toFixed(4)}° W
              </p>
            ) : (
              <p class="text-xs text-zinc-400 italic">Click map or search an address to set an origin point.</p>
            )}
          </div>

          {/* Travel Mode */}
          <div>
            <label class="text-xs text-zinc-400 mb-1.5 block">Travel Mode</label>
            <div class="flex gap-2">
              <For each={TRAVEL_MODES}>
                {(mode) => (
                  <button
                    onClick={() => setTravelMode(mode.value)}
                    class={`flex-1 inline-flex items-center justify-center gap-1.5 rounded-lg px-3 py-2 text-xs font-medium transition-all ${
                      travelMode() === mode.value
                        ? 'bg-cyan-500 text-zinc-950'
                        : 'bg-zinc-800/60 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200'
                    }`}
                  >
                    <mode.icon class="h-3.5 w-3.5" />
                    {mode.label}
                  </button>
                )}
              </For>
            </div>
          </div>

          {/* Max Minutes */}
          <div>
            <label class="text-xs text-zinc-400 mb-1.5 block">
              Max Travel Time: <span class="text-zinc-100 font-mono tabular-nums">{maxMinutes()} min</span>
            </label>
            <input
              type="range"
              min="5"
              max="60"
              step="5"
              value={maxMinutes()}
              onInput={(e) => setMaxMinutes(parseInt(e.currentTarget.value))}
              class="w-full accent-cyan-500"
            />
            <div class="flex justify-between mt-1">
              <For each={MINUTE_OPTIONS}>
                {(m) => (
                  <span class="text-[10px] text-zinc-500 font-mono tabular-nums">{m}</span>
                )}
              </For>
            </div>
          </div>

          {/* CTA */}
          <AnalysisButton />
        </div>
      </Panel>
    </div>
  );
};
