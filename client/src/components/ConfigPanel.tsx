import { For, Show } from 'solid-js';
import type { Component } from 'solid-js';
import { Footprints, Bike, Car, CircleDot, Navigation, X } from 'lucide-solid';
import { Panel } from './Panel';
import { RunButton } from './RunButton';
import {
  activeTool,
  setActiveTool,
  origin,
  isAnalyzing,
  travelMode,
  setTravelMode,
  maxMinutes,
  setMaxMinutes,
  routeOrigin,
  routeDestination,
  isRouting,
  routeError,
  resetRoute,
  runAnalysis,
  runRoute,
} from '../store/analysisStore';

const TRAVEL_MODES = [
  { value: 'walk' as const, label: 'Walk', icon: Footprints },
  { value: 'bike' as const, label: 'Bike', icon: Bike },
  { value: 'drive' as const, label: 'Drive', icon: Car },
];

const MINUTE_OPTIONS = [5, 10, 15, 30, 45, 60];

function formatCoords(loc: { lat: number; lng: number }) {
  const ns = loc.lat >= 0 ? 'N' : 'S';
  const ew = loc.lng >= 0 ? 'E' : 'W';
  return `${Math.abs(loc.lat).toFixed(4)}° ${ns}, ${Math.abs(loc.lng).toFixed(4)}° ${ew}`;
}

const ToolButton: Component<{
  active: boolean;
  label: string;
  icon: Component<{ class?: string }>;
  onClick: () => void;
}> = (props) => (
  <button
    type="button"
    onClick={props.onClick}
    class={`inline-flex items-center justify-center gap-2 rounded-lg px-3 py-2.5 text-xs font-semibold transition-all ${
      props.active
        ? 'bg-cyan-500 text-zinc-950 shadow-lg shadow-cyan-500/20'
        : 'bg-zinc-800/60 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200'
    }`}
  >
    <props.icon class="h-4 w-4" />
    {props.label}
  </button>
);

export const ConfigPanel: Component<{ class?: string }> = (props) => {
  return (
    <div class={props.class ?? ''}>
      <Panel title={activeTool() === 'route' ? 'Route Navigation' : 'Analysis Configuration'}>
        <div class="space-y-4">
          {/* Tool switcher — dedicated route CTA */}
          <div class="grid grid-cols-2 gap-2">
            <ToolButton
              active={activeTool() === 'isochrone'}
              label="Isochrone"
              icon={CircleDot}
              onClick={() => setActiveTool('isochrone')}
            />
            <ToolButton
              active={activeTool() === 'route'}
              label="Route"
              icon={Navigation}
              onClick={() => setActiveTool('route')}
            />
          </div>

          <Show
            when={activeTool() === 'route'}
            fallback={
              <>
                {/* Isochrone Origin */}
                <div>
                  <label class="text-xs text-zinc-400 mb-1.5 block">Origin</label>
                  {origin() ? (
                    <p class="text-xs text-zinc-100 font-mono tabular-nums">{formatCoords(origin()!)}</p>
                  ) : (
                    <p class="text-xs text-zinc-400 italic">Click map or search an address to set an origin point.</p>
                  )}
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
              </>
            }
          >
            {/* Route Start / Destination */}
            <div>
              <label class="text-xs text-zinc-400 mb-1.5 block">Start Point</label>
              {routeOrigin() ? (
                <p class="text-xs text-zinc-100 font-mono tabular-nums">{formatCoords(routeOrigin()!)}</p>
              ) : (
                <p class="text-xs text-zinc-400 italic">Click the map to set your start point.</p>
              )}
            </div>

            <div>
              <label class="text-xs text-zinc-400 mb-1.5 block">Destination</label>
              {routeDestination() ? (
                <p class="text-xs text-zinc-100 font-mono tabular-nums">{formatCoords(routeDestination()!)}</p>
              ) : (
                <p class="text-xs text-zinc-400 italic">
                  {routeOrigin() ? 'Click the map to set your destination.' : 'Set a start point first.'}
                </p>
              )}
            </div>

            <Show when={routeOrigin() || routeDestination()}>
              <button
                type="button"
                onClick={resetRoute}
                class="w-full inline-flex items-center justify-center gap-1.5 rounded-lg px-3 py-2 text-xs font-medium text-zinc-400 bg-zinc-800/60 hover:bg-zinc-800 hover:text-zinc-200 transition-all"
              >
                <X class="h-3.5 w-3.5" />
                Clear Route Points
              </button>
            </Show>
          </Show>

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

          {/* CTA */}
          <Show
            when={activeTool() === 'route'}
            fallback={
              <RunButton
                onClick={runAnalysis}
                disabled={!origin() || isAnalyzing()}
                loading={isAnalyzing()}
                loadingLabel="Calculating Isochrones..."
                hint="Click Map to Set Origin"
                label="Run Analysis"
              />
            }
          >
            <RunButton
              onClick={runRoute}
              disabled={!routeOrigin() || !routeDestination() || isRouting()}
              loading={isRouting()}
              loadingLabel="Calculating Route..."
              hint={!routeOrigin() ? 'Click Map to Set Start Point' : 'Click Map to Set Destination'}
              label="Get Directions"
            />
            <Show when={routeError()}>
              <p class="text-xs text-red-400">{routeError()}</p>
            </Show>
          </Show>
        </div>
      </Panel>
    </div>
  );
};
