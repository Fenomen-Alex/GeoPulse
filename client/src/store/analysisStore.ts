import { createSignal } from 'solid-js';

export interface Location {
  lat: number;
  lng: number;
  address?: string;
}

export const [origin, setOrigin] = createSignal<Location | null>(null);
export const [travelMode, setTravelMode] = createSignal<'walk' | 'bike' | 'drive'>('walk');
export const [maxMinutes, setMaxMinutes] = createSignal<number>(15);
export const [isAnalyzing, setIsAnalyzing] = createSignal<boolean>(false);
export const [analysisResult, setAnalysisResult] = createSignal<any | null>(null);

export function handleMapClick(coords: { lat: number; lng: number }) {
  setOrigin(coords);
  setAnalysisResult(null);
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
    const data = await res.json();
    setAnalysisResult(data);
  } catch (err) {
    console.error('Analysis failed', err);
  } finally {
    setIsAnalyzing(false);
  }
}
