import type { Component } from 'solid-js';

const handleLogin = () => {
  window.location.href = '/api/v1/auth/login';
};

export const LandingHeader: Component = () => {
  return (
    <header class="w-full py-4 px-6 flex items-center justify-between bg-zinc-900/80 border-b border-zinc-800/30">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 rounded-full bg-gradient-to-r from-cyan-400 to-cyan-600 flex items-center justify-center">
          <svg class="w-5 h-5 text-zinc-950" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
          </svg>
        </div>
        <h1 class="text-xl font-bold text-cyan-400">
          GeoPulse
        </h1>
      </div>

      <nav class="hidden md:flex items-center gap-8">
        <a href="#features" class="text-sm text-zinc-300 hover:text-cyan-400 transition-colors">
          Features
        </a>
        <a href="#tech-stack" class="text-sm text-zinc-300 hover:text-cyan-400 transition-colors">
          Tech Stack
        </a>
        <a href="#quotas" class="text-sm text-zinc-300 hover:text-cyan-400 transition-colors">
          Quotas
        </a>
        <a href="#contact" class="text-sm text-zinc-300 hover:text-cyan-400 transition-colors">
          Contact
        </a>
      </nav>

      <div class="hidden md:flex items-center gap-4">
        <button
          type="button"
          onClick={handleLogin}
          class="py-2 px-4 text-sm text-zinc-300 hover:text-cyan-400 border border-zinc-700/50 hover:border-cyan-400/30 rounded-lg transition-all"
        >
          Sign In
        </button>
        <button
          type="button"
          onClick={handleLogin}
          class="py-2 px-4 text-sm bg-cyan-500/10 text-cyan-400 hover:bg-cyan-500/20 rounded-lg transition-colors"
        >
          Get Started
        </button>
      </div>
    </header>
  );
};