import type { Component } from 'solid-js';

export const QuotaPricing: Component = () => {
  return (
    <section class="max-w-6xl mx-auto px-6 py-16 md:py-24">
      <h2 class="text-3xl md:text-4xl font-bold text-center text-zinc-100 mb-12">
        Quota & Pricing Tiers
      </h2>

      <div class="overflow-x-auto">
        <table class="w-full table-auto border-separate">
          <colgroup>
            <col class="w-1/3" />
            <col class="w-1/3" />
            <col class="w-1/3" />
          </colgroup>
          <thead>
            <tr class="border-b border-zinc-800/80">
              <th class="py-3 px-4 text-left font-semibold text-zinc-300">
                Feature
              </th>
              <th class="py-3 px-4 text-left font-semibold text-zinc-300">
                Standard Access (Free)
              </th>
              <th class="py-3 px-4 text-left font-semibold text-zinc-300">
                Pro / Enterprise Extension
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-800/80">
            <tr>
              <td class="py-4 px-4 text-zinc-400">
                <strong>Auth Requirement</strong>
              </td>
              <td class="py-4 px-4">
                Google Account Sign-In
              </td>
              <td class="py-4 px-4">
                Google Account Sign-In
              </td>
            </tr>

            <tr>
              <td class="py-4 px-4 text-zinc-400">
                <strong>Daily Isochrone Runs</strong>
              </td>
              <td class="py-4 px-4">
                15 runs / day
              </td>
              <td class="py-4 px-4">
                Custom (500+ runs / day)
              </td>
            </tr>

            <tr>
              <td class="py-4 px-4 text-zinc-400">
                <strong>Travel Modes</strong>
              </td>
              <td class="py-4 px-4">
                Foot Walking, Cycling, Driving
              </td>
              <td class="py-4 px-4">
                All Modes + Custom Speeds
              </td>
            </tr>

            <tr>
              <td class="py-4 px-4 text-zinc-400">
                <strong>Export Formats</strong>
              </td>
              <td class="py-4 px-4">
                GeoJSON, Spatial Metrics
              </td>
              <td class="py-4 px-4">
                GeoJSON, CSV, PDF Reports
              </td>
            </tr>

            <tr>
              <td class="py-4 px-4 text-zinc-400">
                <strong>Support</strong>
              </td>
              <td class="py-4 px-4">
                Community
              </td>
              <td class="py-4 px-4">
                Dedicated Engineering Support
              </td>
            </tr>

            <tr>
              <td class="py-4 px-4"></td>
              <td class="py-4 px-4">
              <button
                class="mt-4 w-full py-2 px-4 rounded-lg bg-cyan-500/20 text-cyan-400 hover:bg-cyan-500/30 transition-colors"
                onClick={() => { window.location.href = '/api/v1/auth/login'; }}
              >
                Sign Up Free
              </button>
              </td>
              <td class="py-4 px-4">
              <button
                class="mt-4 w-full py-2 px-4 rounded-lg bg-cyan-500/20 text-cyan-400 hover:bg-cyan-500/30 transition-colors"
                onClick={() => { document.getElementById('contact')?.scrollIntoView({ behavior: 'smooth' }); }}
              >
                Request Extension
              </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  );
};