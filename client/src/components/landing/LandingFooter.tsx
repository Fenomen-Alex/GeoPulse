import type { Component } from 'solid-js';

export const LandingFooter: Component = () => {
  return (
    <footer class="w-full py-8 px-6 border-t border-zinc-800/30">
      <div class="max-w-6xl mx-auto flex flex-col md:flex-row justify-between items-center gap-4">
        <div class="text-sm text-zinc-500">
          &copy; {new Date().getFullYear()} GeoPulse. All rights reserved.
        </div>

        <div class="flex items-center gap-6">
          <a href="#" class="text-sm text-zinc-400 hover:text-cyan-400 transition-colors">
            Privacy Policy
          </a>
          <a href="https://github.com/yourusername/GeoPulse" class="text-sm text-zinc-400 hover:text-cyan-400 transition-colors">
            GitHub
          </a>
        </div>

        <div class="flex items-center gap-2">
          <div class="w-2 h-2 rounded-full bg-green-500"></div>
          <span class="text-sm text-zinc-400">Systems Operational</span>
        </div>
      </div>
    </footer>
  );
};