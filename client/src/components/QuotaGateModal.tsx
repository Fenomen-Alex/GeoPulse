import type { Component } from 'solid-js';
import { setShowQuotaModal } from '../store/analysisStore';

export const QuotaGateModal: Component = () => {
  const handleRequestExtension = async () => {
    await fetch('/api/v1/auth/logout', { method: 'POST' });
    window.location.href = '/#contact';
  };

  return (
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="max-w-md w-full rounded-2xl border border-zinc-800/80 bg-zinc-900 p-6 shadow-2xl">
        <h2 class="text-lg font-semibold text-zinc-100 mb-2">Daily Quota Exceeded</h2>
        <p class="text-sm text-zinc-400 mb-6">
          You have reached your limit of 15 queries today. Request an extension to get additional runs.
        </p>
        <div class="flex gap-3">
          <button
            type="button"
            onClick={handleRequestExtension}
            class="flex-1 py-2 px-4 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-zinc-950 font-medium text-sm transition-colors"
          >
            Request Extension
          </button>
          <button
            type="button"
            onClick={() => setShowQuotaModal(false)}
            class="py-2 px-4 rounded-lg border border-zinc-700 text-zinc-300 hover:border-zinc-500 text-sm transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};
