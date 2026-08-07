import { createSignal } from 'solid-js';

export interface Location {
  lat: number;
  lng: number;
  address?: string;
}

export type Tool = 'isochrone' | 'route';

export const [activeTool, setActiveTool] = createSignal<Tool>('isochrone');

export const [origin, setOrigin] = createSignal<Location | null>(null);
export const [travelMode, setTravelMode] = createSignal<'walk' | 'bike' | 'drive'>('walk');
export const [maxMinutes, setMaxMinutes] = createSignal<number>(15);
export const [isAnalyzing, setIsAnalyzing] = createSignal<boolean>(false);
export const [analysisResult, setAnalysisResult] = createSignal<any | null>(null);
export const [remainingQuota, setRemainingQuota] = createSignal<number>(15);
export const [showQuotaModal, setShowQuotaModal] = createSignal<boolean>(false);

export const [routeOrigin, setRouteOrigin] = createSignal<Location | null>(null);
export const [routeDestination, setRouteDestination] = createSignal<Location | null>(null);
export const [routeResult, setRouteResult] = createSignal<any | null>(null);
export const [isRouting, setIsRouting] = createSignal<boolean>(false);
export const [routeError, setRouteError] = createSignal<string | null>(null);

export function handleMapClick(coords: { lat: number; lng: number }) {
  if (activeTool() === 'route') {
    // First click sets the start point; any subsequent click repositions the destination.
    if (!routeOrigin()) {
      setRouteOrigin(coords);
      setRouteDestination(null);
    } else {
      setRouteDestination(coords);
    }
    setRouteResult(null);
    setRouteError(null);
    return;
  }

  setOrigin(coords);
  setAnalysisResult(null);
}

export function resetRoute() {
  setRouteOrigin(null);
  setRouteDestination(null);
  setRouteResult(null);
  setRouteError(null);
}

export async function runAnalysis() {
  const currentOrigin = origin();
  if (!currentOrigin) return;

  setIsAnalyzing(true);
  try {
    const res = await fetch('/api/v1/analysis', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        lat: currentOrigin.lat,
        lng: currentOrigin.lng,
        mode: travelMode(),
        minutes: maxMinutes(),
      }),
    });

    if (res.status === 429) {
      setShowQuotaModal(true);
      return;
    }

    if (!res.ok) throw new Error('Analysis failed');

    const data = await res.json();
    setAnalysisResult(data);
    if (typeof data.remaining_quota === 'number') {
      setRemainingQuota(data.remaining_quota);
    }
  } catch (err) {
    console.error('Analysis failed', err);
  } finally {
    setIsAnalyzing(false);
  }
}

export async function runRoute() {
  const start = routeOrigin();
  const end = routeDestination();
  if (!start || !end) return;

  setIsRouting(true);
  setRouteError(null);
  try {
    const res = await fetch('/api/v1/routes', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        start,
        end,
        mode: travelMode(),
      }),
    });

    if (res.status === 429) {
      setShowQuotaModal(true);
      return;
    }

    if (!res.ok) {
      let message = 'Routing failed';
      try {
        const data = await res.json();
        if (data.error) message = data.error;
      } catch {
        // fall back to generic message
      }
      setRouteError(message);
      return;
    }

    const data = await res.json();
    setRouteResult(data);
    if (typeof data.remaining_quota === 'number') {
      setRemainingQuota(data.remaining_quota);
    }
  } catch (err) {
    console.error('Routing failed', err);
    setRouteError('Network error. Please try again.');
  } finally {
    setIsRouting(false);
  }
}
