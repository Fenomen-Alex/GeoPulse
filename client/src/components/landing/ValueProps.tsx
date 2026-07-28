import type { Component } from 'solid-js';

export const ValueProps: Component = () => {
  return (
    <section class="max-w-6xl mx-auto px-6 py-16 md:py-24">
      <h2 class="text-3xl md:text-4xl font-bold text-center text-zinc-100 mb-12">
        Core Value Propositions
      </h2>

      <div class="grid md:grid-cols-3 gap-6 md:gap-8">
        <div class="p-6 md:p-8 rounded-xl bg-zinc-900/60 border border-zinc-800/80 backdrop-blur-sm">
          <h3 class="text-xl font-bold text-cyan-400 mb-3">
            True Travel-Time Isochrones
          </h3>
          <p class="text-zinc-400">
            Ditch naive radial circles. Compute realistic reachability polygons based on real road networks and walking paths.
          </p>
        </div>

        <div class="p-6 md:p-8 rounded-xl bg-zinc-900/60 border border-zinc-800/80 backdrop-blur-sm">
          <h3 class="text-xl font-bold text-cyan-400 mb-3">
            Instant Amenity Density Indexing
          </h3>
          <p class="text-zinc-400">
            Filter supermarkets, transit stops, and cafes inside your travel contours with client-side spatial intersection (Turf.js).
          </p>
        </div>

        <div class="p-6 md:p-8 rounded-xl bg-zinc-900/60 border border-zinc-800/80 backdrop-blur-sm">
          <h3 class="text-xl font-bold text-cyan-400 mb-3">
            Zero-Trust Engineering
          </h3>
          <p class="text-zinc-400">
            Sealed API keys, edge caching, and token-bucket protection ensure rapid response times without third-party key leaks.
          </p>
        </div>
      </div>
    </section>
  );
};