import type { Component } from 'solid-js';
import { origin, isAnalyzing, runAnalysis } from '../store/analysisStore';

export const AnalysisButton: Component = () => {
  return (
    <button
      onClick={runAnalysis}
      disabled={!origin() || isAnalyzing()}
      class="w-full relative inline-flex items-center justify-center gap-2 px-4 py-2.5 text-xs rounded-lg font-semibold bg-cyan-500 text-zinc-950 shadow-lg shadow-cyan-500/20 hover:bg-cyan-400 transition-all focus:outline-none focus:ring-2 focus:ring-cyan-500/50 disabled:opacity-40 disabled:cursor-not-allowed"
    >
      {isAnalyzing() ? (
        <span class="inline-flex items-center gap-2">
          <svg class="animate-spin h-3.5 w-3.5 text-zinc-950" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          Calculating Isochrones...
        </span>
      ) : !origin() ? (
        'Click Map to Set Origin'
      ) : (
        'Run Analysis'
      )}
    </button>
  );
};
