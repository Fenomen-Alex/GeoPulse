import { createSignal } from 'solid-js';

export interface Location {
  lat: number;
  lng: number;
  address?: string;
}

export type Tool = 'isochrone' | 'route';

export interface IsochroneBand {
  minutes: number;
  area: number;
  fillColor: string;
  strokeColor: string;
  fillOpacity: number;
  geojson: {
    type: 'Feature';
    geometry: {
      type: 'Polygon';
      coordinates: number[][][];
    };
    properties: Record<string, unknown> | null;
  };
}

export interface AnalysisResult {
  totalArea: number;
  poiCount: number;
  score: number;
  bands: IsochroneBand[];
  remaining_quota: number;
}

export const [activeTool, setActiveTool] = createSignal<Tool>('isochrone');

export const [origin, setOrigin] = createSignal<Location | null>(null);
export const [travelMode, setTravelMode] = createSignal<'walk' | 'bike' | 'drive'>('walk');
export const [maxMinutes, setMaxMinutes] = createSignal<number>(15);
export const [isAnalyzing, setIsAnalyzing] = createSignal<boolean>(false);
export const [analysisResult, setAnalysisResult] = createSignal<AnalysisResult | null>(null);
export const [analysisError, setAnalysisError] = createSignal<string | null>(null);
export const [remainingQuota, setRemainingQuota] = createSignal<number>(15);
export const [showQuotaModal, setShowQuotaModal] = createSignal<boolean>(false);

export const [routeOrigin, setRouteOrigin] = createSignal<Location | null>(null);
export const [routeDestination, setRouteDestination] = createSignal<Location | null>(null);
export const [routeResult, setRouteResult] = createSignal<any | null>(null);
export const [isRouting, setIsRouting] = createSignal<boolean>(false);
export const [routeError, setRouteError] = createSignal<string | null>(null);

// Latest map focus request from the search bar; MapCanvas reacts by flying to it.
export const [mapFocus, setMapFocus] = createSignal<Location | null>(null);

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
  setAnalysisError(null);
}

export function resetRoute() {
  setRouteOrigin(null);
  setRouteDestination(null);
  setRouteResult(null);
  setRouteError(null);
}

// Applies a search result: flies the map to the location and wires it into the
// active tool — origin for isochrones, start/destination for routes.
export function applySearchResult(loc: Location) {
  setMapFocus(loc);

  if (activeTool() === 'route') {
    if (!routeOrigin()) {
      setRouteOrigin(loc);
      setRouteDestination(null);
    } else {
      setRouteDestination(loc);
    }
    setRouteResult(null);
    setRouteError(null);
  } else {
    setOrigin(loc);
    setAnalysisResult(null);
    setAnalysisError(null);
  }
}

export async function loadRemainingQuota() {
  try {
    const res = await fetch('/api/v1/quota');
    if (!res.ok) return;
    const data = await res.json();
    if (typeof data.remaining_quota === 'number') {
      setRemainingQuota(data.remaining_quota);
    }
  } catch {
    // Keep the default until the next successful spatial response.
  }
}

export async function runAnalysis() {
  const currentOrigin = origin();
  if (!currentOrigin) return;

  setIsAnalyzing(true);
  setAnalysisError(null);
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
      setRemainingQuota(0);
      setShowQuotaModal(true);
      return;
    }

    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      setAnalysisError(data.error ?? 'Analysis failed. Please try again.');
      return;
    }

    const data = await res.json();
    setAnalysisResult(data);
    if (typeof data.remaining_quota === 'number') {
      setRemainingQuota(data.remaining_quota);
    }
  } catch (err) {
    console.error('Analysis failed', err);
    setAnalysisError('Network error. Please try again.');
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
      setRemainingQuota(0);
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
