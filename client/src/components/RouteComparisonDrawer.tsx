import type { Component } from 'solid-js';
import { Panel } from './Panel';
import { compareResult, clearCompare } from '../store/analysisStore';

function formatCoords(loc: { lat: number; lng: number }) {
  const ns = loc.lat >= 0 ? 'N' : 'S';
  const ew = loc.lng >= 0 ? 'E' : 'W';
  return `${Math.abs(loc.lat).toFixed(4)}° ${ns}, ${Math.abs(loc.lng).toFixed(4)}° ${ew}`;
}

const RouteComparisonDrawer: Component = () => {
  const cmp = compareResult();
  if (!cmp) return null;

  return (
    <Panel title="Reachability Comparison">
      <div class="space-y-4">
        {/* Left Analysis */}
        <div class="border border-zinc-700/50 rounded-lg p-4">
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Origin A</span>
            <p class="text-xs font-mono text-zinc-100">{formatCoords({ lat: cmp.Left.Lat, lng: cmp.Left.Lng })}</p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Mode A</span>
            <p class="text-xs font-mono text-zinc-100">{cmp.Left.Mode}</p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Travel Time</span>
            <p class="text-xs font-mono text-zinc-100">{cmp.Left.Minutes} min</p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Total Area</span>
            <p class="text-lg font-semibold text-blue-400 font-mono tabular-nums">
              {cmp.Left.TotalArea.toFixed(1)} km²
            </p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">POI Count</span>
            <p class="text-xs font-mono text-zinc-100">{cmp.Left.PoiCount}</p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Accessibility Score</span>
            <p class="text-xs font-mono text-zinc-100">{cmp.Left.Score}/100</p>
          </div>
        </div>

        {/* Right Analysis */}
        <div class="border border-zinc-700/50 rounded-lg p-4">
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Origin B</span>
            <p class="text-xs font-mono text-zinc-100">{formatCoords({ lat: cmp.Right.Lat, lng: cmp.Right.Lng })}</p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Mode B</span>
            <p class="text-xs font-mono text-zinc-100">{cmp.Right.Mode}</p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Travel Time</span>
            <p class="text-xs font-mono text-zinc-100">{cmp.Right.Minutes} min</p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Total Area</span>
            <p class="text-lg font-semibold text-purple-400 font-mono tabular-nums">
              {cmp.Right.TotalArea.toFixed(1)} km²
            </p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">POI Count</span>
            <p class="text-xs font-mono text-zinc-100">{cmp.Right.PoiCount}</p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Accessibility Score</span>
            <p class="text-xs font-mono text-zinc-100">{cmp.Right.Score}/100</p>
          </div>
        </div>

        {/* Delta Summary */}
        <div class="border border-zinc-700/50 rounded-lg p-4">
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Area Difference</span>
            <p class="text-sm font-mono text-zinc-100">
              {Math.abs(cmp.Delta.AreaDiff).toFixed(1)} km² ({cmp.Delta.AreaDiffPct > 0 ? '+' : ''}{cmp.Delta.AreaDiffPct.toFixed(1)}%)
              {' '}— {cmp.Delta.Winner === 'left' ? 'A is larger' : cmp.Delta.Winner === 'right' ? 'B is larger' : 'tie'}
            </p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">POI Difference</span>
            <p class="text-sm font-mono text-zinc-100">
              {Math.abs(cmp.Delta.PoiDiff)} ({cmp.Delta.PoiDiff > 0 ? 'B more' : cmp.Delta.PoiDiff < 0 ? 'A more' : 'tie'})
            </p>
          </div>
          <div class="mb-2">
            <span class="text-xs font-medium text-zinc-300">Score Difference</span>
            <p class="text-sm font-mono text-zinc-100">
              {Math.abs(cmp.Delta.ScoreDiff)} ({cmp.Delta.ScoreDiff > 0 ? 'B higher' : cmp.Delta.ScoreDiff < 0 ? 'A higher' : 'tie'})
            </p>
          </div>
        </div>

        <div class="flex gap-3">
          <button
            type="button"
            onClick={clearCompare}
            class="flex-1 py-2 px-4 rounded-lg bg-zinc-800/60 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200 transition-all text-xs"
          >
            Clear Comparison
          </button>
        </div>
      </div>
    </Panel>
  );
};

export default RouteComparisonDrawer;
