import type { Component } from 'solid-js';
import { MapCanvas } from '../components/MapCanvas';

interface AppViewProps {
  user: any;
}

export const AppView: Component<AppViewProps> = (props) => {
  return (
    <div class="min-h-screen bg-zinc-950">
      <div class="p-4 flex items-center justify-between border-b border-zinc-800">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-full bg-gradient-to-r from-cyan-400 to-cyan-600"></div>
          <h1 class="text-xl font-bold text-cyan-400">GeoPulse</h1>
        </div>
        <div class="flex items-center gap-4">
          <div class="px-3 py-1.5 rounded-full bg-cyan-500/10 text-cyan-400 text-sm">
            {props.user.daily_quota} / 15 remaining
          </div>
          <img
            src={props.user.avatar_url}
            alt={props.user.name}
            class="w-8 h-8 rounded-full"
          />
        </div>
      </div>

      <div class="flex-1 p-6">
        <div class="max-w-2xl mx-auto">
          <div class="mb-8">
            <h2 class="text-2xl font-bold text-zinc-100 mb-2">Welcome, {props.user.name}</h2>
            <p class="text-zinc-400">
              You're all set to start using GeoPulse. Click the map to set your origin point and start analyzing.
            </p>
          </div>
          <MapCanvas />
        </div>
      </div>
    </div>
  );
};