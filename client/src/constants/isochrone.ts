export const ISOCHRONE_COLORS: Record<number, { fill: string; fillOpacity: number; stroke: string; label: string }> = {
  5:  { fill: '#10b981', fillOpacity: 0.40, stroke: '#34d399', label: '0–5 min (Walk)' },
  15: { fill: '#f59e0b', fillOpacity: 0.35, stroke: '#fbbf24', label: '5–15 min' },
  30: { fill: '#ef4444', fillOpacity: 0.25, stroke: '#f87171', label: '15–30 min' },
  45: { fill: '#8b5cf6', fillOpacity: 0.20, stroke: '#a78bfa', label: '30–45 min' },
  60: { fill: '#3b82f6', fillOpacity: 0.15, stroke: '#60a5fa', label: '45–60 min (Drive)' },
};
