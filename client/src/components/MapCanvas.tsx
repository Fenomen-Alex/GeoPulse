import { onMount, onCleanup } from 'solid-js';
import type { Component } from 'solid-js';
import * as maplibregl from 'maplibre-gl';
import { handleMapClick, setOrigin, setAnalysisResult } from '../store/analysisStore';

export const MapCanvas: Component = () => {
  let container!: HTMLDivElement;
  let map: maplibregl.Map | undefined;
  let marker: maplibregl.Marker | undefined;

  onMount(() => {
    map = new maplibregl.Map({
      container,
      style: 'https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json',
      center: [-122.4194, 37.7749],
      zoom: 12,
    });

    map.addControl(new maplibregl.NavigationControl(), 'bottom-right');

    map.on('click', (e: maplibregl.MapMouseEvent) => {
      const { lng, lat } = e.lngLat;
      handleMapClick({ lat, lng });
      updateMarker(lat, lng);
    });
  });

  function updateMarker(lat: number, lng: number) {
    if (!map) return;

    if (marker) {
      marker.setLngLat([lng, lat]);
    } else {
      marker = new maplibregl.Marker({ color: '#06b6d4', draggable: true })
        .setLngLat([lng, lat])
        .addTo(map);

      marker.on('dragend', () => {
        const lngLat = marker!.getLngLat();
        setOrigin({ lat: lngLat.lat, lng: lngLat.lng });
        setAnalysisResult(null);
      });
    }
  }

  onCleanup(() => {
    marker?.remove();
    map?.remove();
  });

  return <div ref={container} class="h-full w-full" />;
};
