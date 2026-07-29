import type { Component } from 'solid-js';

const STACK = [
	{ name: 'Go 1.22 + Chi', desc: 'Single-binary API execution & embedded SPA assets', icon: (
		<svg class="w-10 h-10 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
		</svg>
	) },
	{ name: 'SolidJS + Vite', desc: 'Fine-grained reactive updates at 60 FPS', icon: (
		<svg class="w-10 h-10 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M22 12H2M12 22V12L9 9"/>
		</svg>
	) },
	{ name: 'MapLibre GL JS', desc: 'High-performance WebGL vector canvas styling', icon: (
		<svg class="w-10 h-10 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10a4 4 0 100-8 4 4 0 000 8z"/>
		</svg>
	) },
	{ name: 'Turf.js', desc: 'In-browser spatial computation & polygon geometry math', icon: (
		<svg class="w-10 h-10 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064"/>
		</svg>
	) },
	{ name: 'OpenRouteService', desc: 'High-precision routing and travel-contour engine', icon: (
		<svg class="w-10 h-10 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"/>
		</svg>
	) },
	{ name: 'Turso (LibSQL)', desc: 'Managed SQLite at the edge for session & quota persistence', icon: (
		<svg class="w-10 h-10 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
		</svg>
	) },
	{ name: 'Resend API', desc: 'High-reliability transactional notification engine for quota requests', icon: (
		<svg class="w-10 h-10 text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
		</svg>
	) },
];

export const TechStack: Component = () => {
	return (
		<section class="max-w-6xl mx-auto px-6 py-16 md:py-24">
			<h2 class="text-2xl md:text-3xl font-bold text-center text-zinc-100 mb-12">
				Technology & Infrastructure
			</h2>

			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 md:gap-8">
				{STACK.map((item) => (
					<div class="p-6 md:p-8 rounded-xl bg-zinc-900/60 border border-zinc-800/80 backdrop-blur-sm">
						<div class="flex items-center gap-2">
							{item.icon}
							<h3 class="text-xl font-bold text-cyan-400 mt-0 mb-1">
								{item.name}
							</h3>
						</div>

						<p class="text-zinc-400 mt-3">
							{item.desc}
						</p>
					</div>
				))}
			</div>
		</section>
	);
};