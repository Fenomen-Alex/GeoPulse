import { onMount, onCleanup } from 'solid-js';
import type { Component } from 'solid-js';
import * as maplibregl from 'maplibre-gl';

// 💡 1. Native Vite worker isolation pipeline bundle
import workerUrl from 'maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url';
import { handleMapClick, setOrigin, setAnalysisResult } from '../store/analysisStore';

maplibregl.setWorkerUrl(workerUrl);

export const MapCanvas: Component = () => {
  let container!: HTMLDivElement;
  let map: maplibregl.Map | undefined;
  let marker: maplibregl.Marker | undefined;

  onMount(async () => {
    if (!container) return;

    try {
      const baseUrl = 'https://basemaps.cartocdn.com';
      const styleUrl = `${baseUrl}/gl/dark-matter-gl-style/style.json`;

      const response = await fetch(styleUrl);
      const styleData = await response.json();

      // 💡 2. Fix the Glyph (Fonts) fallback tracking route
      if (styleData.glyphs && styleData.glyphs.startsWith('/')) {
        styleData.glyphs = `${baseUrl}${styleData.glyphs}`;
      }

      // 💡 3. Fix the Sprite layout tracking route
      if (styleData.sprite && styleData.sprite.startsWith('/')) {
        styleData.sprite = `${baseUrl}${styleData.sprite}`;
      }

      // 💡 4. Fix every vector source layout tracking route
      if (styleData.sources) {
        Object.keys(styleData.sources).forEach((sourceKey) => {
          const source = styleData.sources[sourceKey];

          if (source.url && source.url.startsWith('/')) {
            source.url = `${baseUrl}${source.url}`;
          }

          if (source.tiles) {
            source.tiles = source.tiles.map((tilePath: string) => {
              if (tilePath.startsWith('/')) {
                return `${baseUrl}${tilePath}`;
              }
              return tilePath;
            });
          }
        });
      }

      // 💡 5. Launch MapLibre using the absolute data configuration
      map = new maplibregl.Map({
        container,
        style: styleData,
        center: [-122.4194, 37.7749],
        zoom: 12,
      });

      map.addControl(new maplibregl.NavigationControl(), 'bottom-right');

      const resizeObserver = new ResizeObserver(() => {
        map?.resize();
      });
      resizeObserver.observe(container);

      map.on('click', (e: maplibregl.MapMouseEvent) => {
        const { lng, lat } = e.lngLat;
        handleMapClick({ lat, lng });
        updateMarker(lat, lng);
      });

      onCleanup(() => {
        resizeObserver.disconnect();
        marker?.remove();
        map?.remove();
      });

    } catch (error) {
      console.error('Failed to securely parse map structure layers:', error);
    }
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

  return <div ref={container} class="absolute inset-0 w-full h-full" />;
};
