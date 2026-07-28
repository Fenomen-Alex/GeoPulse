import type { Component } from 'solid-js';

export const Hero: Component = () => {
  const handleLogin = () => {
    window.location.href = '/api/v1/auth/login';
  };

  const scrollToContact = () => {
    document.getElementById('contact')?.scrollIntoView({ behavior: 'smooth' });
  };

  return (
    <section class="relative w-full min-h-[80vh] flex items-center justify-center bg-zinc-950 overflow-hidden">
      <div class="absolute inset-0 bg-grid-pattern opacity-5"></div>

      <div class="relative z-10 max-w-6xl mx-auto px-6 py-20 md:py-32 flex flex-col md:flex-row items-center gap-12">
        <div class="md:w-1/2 flex flex-col gap-6 items-start">
          <div class="px-3 py-1.5 rounded-full bg-cyan-500/10 text-cyan-400 text-sm font-medium">
            [NEW] Multi-Modal Isochrones & POI Spatial Engine
          </div>

          <h1 class="text-4xl md:text-5xl lg:text-6xl font-bold text-zinc-100 leading-tight">
            Spatial Intelligence
            <span class="block text-cyan-400">Without the GIS Overhead</span>
          </h1>

          <p class="text-lg md:text-xl text-zinc-400 max-w-2xl">
            Calculate true walk, cycle, and drive travel-time boundaries, index amenity density, and make data-backed location decisions in seconds.
          </p>

          <div class="flex items-center gap-4 mt-4">
            <button
              type="button"
              onClick={handleLogin}
              class="py-3 px-6 bg-cyan-500 hover:bg-cyan-400 text-zinc-950 font-medium rounded-lg transition-colors"
            >
              Get Started Free
            </button>
            <button
              type="button"
              onClick={scrollToContact}
              class="py-3 px-6 bg-transparent border border-zinc-700/50 text-zinc-300 hover:border-cyan-400/30 hover:text-cyan-400 rounded-lg transition-colors"
            >
              Request Enterprise Access
            </button>
          </div>
        </div>

        <div class="md:w-1/2 relative">
          <div class="relative rounded-2xl overflow-hidden border border-zinc-800/50 bg-zinc-900/50 shadow-2xl shadow-cyan-500/10">
            <div class="h-12 px-6 flex items-center gap-3 border-b border-zinc-800/50 bg-zinc-900/50">
              <div class="w-3 h-3 rounded-full bg-zinc-500/40"></div>
              <div class="w-3 h-3 rounded-full bg-zinc-500/40"></div>
              <div class="w-3 h-3 rounded-full bg-zinc-500/40"></div>
              <div class="ml-4 text-sm text-zinc-400">
                map.tsx
              </div>
            </div>
            <div class="p-6 md:p-8">
              <div class="relative w-full h-64 md:h-80 rounded-lg overflow-hidden bg-zinc-900/80">
                <div class="absolute inset-0 bg-gradient-to-br from-cyan-400/20 to-cyan-600/20 rounded-lg"></div>
                <div class="absolute inset-0 flex items-center justify-center">
                  <div class="w-40 h-40 rounded-full border-2 border-cyan-400/30 border-dashed flex items-center justify-center">
                    <div class="w-24 h-24 rounded-full bg-cyan-400/10 flex items-center justify-center">
                      <svg class="w-10 h-10 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
                      </svg>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};