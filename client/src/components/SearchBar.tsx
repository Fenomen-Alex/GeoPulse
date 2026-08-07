import { createEffect, createSignal, For, onCleanup, Show } from 'solid-js';
import type { Component } from 'solid-js';
import { Loader2, MapPin, Search } from 'lucide-solid';
import { applySearchResult } from '../store/analysisStore';

interface Suggestion {
  lat: number;
  lng: number;
  label: string;
}

export const SearchBar: Component<{ class?: string }> = (props) => {
  const [query, setQuery] = createSignal('');
  const [suggestions, setSuggestions] = createSignal<Suggestion[]>([]);
  const [loading, setLoading] = createSignal(false);
  const [open, setOpen] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);

  let timer: ReturnType<typeof setTimeout> | undefined;
  let controller: AbortController | undefined;

  const runSearch = async (q: string) => {
    controller?.abort();
    controller = new AbortController();
    try {
      const res = await fetch(`/api/v1/geocode?q=${encodeURIComponent(q)}&limit=8`, {
        signal: controller?.signal,
      });
      if (!res.ok) {
        setError('Search unavailable. Please try again.');
        return;
      }
      const data = await res.json();
      setSuggestions(data.results ?? []);
      setOpen(true);
    } catch (err) {
      if ((err as DOMException)?.name === 'AbortError') return;
      setError('Search unavailable. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const search = (value: string) => {
    clearTimeout(timer);
    const q = value.trim();
    if (q.length < 3) {
      controller?.abort();
      setSuggestions([]);
      setOpen(false);
      setError(null);
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);
    timer = setTimeout(() => runSearch(q), 350);
  };

  const select = (s: Suggestion) => {
    applySearchResult({ lat: s.lat, lng: s.lng, address: s.label });
    setQuery(s.label);
    setSuggestions([]);
    setOpen(false);
    setError(null);
  };

  createEffect(() => {
    search(query());
  });

  onCleanup(() => {
    clearTimeout(timer);
    controller?.abort();
  });

  return (
    <div class={`w-96 ${props.class ?? ''}`}>
      <div class="flex items-center gap-2 rounded-xl border border-zinc-800/80 bg-zinc-900/90 px-3 py-2.5 shadow-2xl backdrop-blur-md focus-within:border-cyan-500/50 transition-colors">
        <Search class="h-4 w-4 shrink-0 text-zinc-400" />
        <input
          type="text"
          value={query()}
          onInput={(e) => setQuery(e.currentTarget.value)}
          onFocus={() => suggestions().length > 0 && setOpen(true)}
          onBlur={() => setTimeout(() => setOpen(false), 150)}
          placeholder="Search city, street or address..."
          class="w-full bg-transparent text-sm text-zinc-100 placeholder-zinc-500 outline-none"
        />
        <Show when={loading()}>
          <Loader2 class="h-4 w-4 shrink-0 animate-spin text-cyan-400" />
        </Show>
      </div>

      <Show when={open()}>
        <div class="absolute left-0 right-0 top-full mt-2 max-h-72 overflow-y-auto rounded-xl border border-zinc-800/80 bg-zinc-900/95 shadow-2xl backdrop-blur-md z-50">
          <Show
            when={suggestions().length > 0}
            fallback={<p class="px-4 py-3 text-xs text-zinc-500">No results found.</p>}
          >
            <For each={suggestions()}>
              {(s) => (
                <button
                  type="button"
                  onMouseDown={(e) => {
                    e.preventDefault();
                    select(s);
                  }}
                  class="flex w-full items-start gap-2 px-4 py-2.5 text-left text-xs text-zinc-200 hover:bg-zinc-800/70 transition-colors"
                >
                  <MapPin class="mt-0.5 h-3.5 w-3.5 shrink-0 text-cyan-400" />
                  <span class="leading-snug">{s.label}</span>
                </button>
              )}
            </For>
          </Show>
        </div>
      </Show>

      <Show when={error()}>
        <p class="mt-1 text-xs text-red-400">{error()}</p>
      </Show>
    </div>
  );
};
