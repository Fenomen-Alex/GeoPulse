import { createEffect, createSignal, onMount, onCleanup } from 'solid-js';
import type { Component } from 'solid-js';
import type * as maplibregl from 'maplibre-gl';

// Native Vite worker isolation pipeline bundle
import workerUrl from 'maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url';
import {
  activeTool,
  origin,
  mapFocus,
  routeOrigin,
  routeDestination,
  routeResult,
  setRouteOrigin,
  setRouteDestination,
  setRouteResult,
  setOrigin,
  setAnalysisResult,
  setAnalysisError,
  analysisResult,
  handleMapClick,
} from '../store/analysisStore';

// Lazy-loaded on first render (inside onMount) so the ~1MB maplibre runtime is
// split out of the landing page bundle. All value references below are guarded
// by `map` being set, which only happens after this module resolves.
let maplibre: typeof maplibregl | undefined;

export const MapCanvas: Component = () => {
  let container!: HTMLDivElement;
  let map: maplibregl.Map | undefined;
  const [mapReady, setMapReady] = createSignal(false);
  let marker: maplibregl.Marker | undefined;
  let startMarker: maplibregl.Marker | undefined;
  let endMarker: maplibregl.Marker | undefined;

  function addIsochroneLayers() {
    if (!map || map.getSource('isochrones')) return;

    map.addSource('isochrones', {
      type: 'geojson',
      data: { type: 'FeatureCollection', features: [] },
    });

    map.addLayer({
      id: 'isochrone-fills',
      type: 'fill',
      source: 'isochrones',
      paint: {
        'fill-color': ['get', 'fillColor'],
        'fill-opacity': ['get', 'fillOpacity'],
      },
    });

    map.addLayer({
      id: 'isochrone-outlines',
      type: 'line',
      source: 'isochrones',
      paint: {
        'line-color': ['get', 'strokeColor'],
        'line-width': 2,
        'line-opacity': 0.9,
      },
    });
  }

  function drawIsochrones() {
    if (!map) return;
    const result = analysisResult();
    if (!result?.bands?.length) return;

    addIsochroneLayers();
    const features = [...result.bands].reverse().map((band) => ({
      ...band.geojson,
      properties: {
        ...(band.geojson.properties ?? {}),
        minutes: band.minutes,
        fillColor: band.fillColor,
        strokeColor: band.strokeColor,
        fillOpacity: band.fillOpacity,
      },
    }));
    const source = map.getSource('isochrones') as maplibregl.GeoJSONSource | undefined;
    source?.setData({ type: 'FeatureCollection', features });

    const coordinates = result.bands.flatMap((band) => band.geojson.geometry.coordinates[0]);
    if (coordinates.length) {
      const bounds = coordinates.reduce(
        (current, coordinate) => current.extend(coordinate as [number, number]),
        new maplibre!.LngLatBounds(coordinates[0] as [number, number], coordinates[0] as [number, number]),
      );
      map.fitBounds(bounds, { padding: 60, maxZoom: 15, duration: 800 });
    }
  }

  function clearIsochroneLayers() {
    if (!map) return;
    if (map.getLayer('isochrone-outlines')) map.removeLayer('isochrone-outlines');
    if (map.getLayer('isochrone-fills')) map.removeLayer('isochrone-fills');
    if (map.getSource('isochrones')) map.removeSource('isochrones');
  }

  function addRouteLayers() {
    if (!map || map.getSource('route-line')) return;

    map.addSource('route-line', {
      type: 'geojson',
      data: {
        type: 'Feature',
        geometry: { type: 'LineString', coordinates: [] },
        properties: {},
      },
    });

    map.addLayer({
      id: 'route-casing',
      type: 'line',
      source: 'route-line',
      layout: { 'line-cap': 'round', 'line-join': 'round' },
      paint: { 'line-color': '#0e7490', 'line-width': 10, 'line-opacity': 0.65 },
    });

    map.addLayer({
      id: 'route-line',
      type: 'line',
      source: 'route-line',
      layout: { 'line-cap': 'round', 'line-join': 'round' },
      paint: { 'line-color': '#06b6d4', 'line-width': 5 },
    });
  }

  function drawRoute() {
    if (!map) return;

    const result = routeResult();
    if (!result?.geojson) return;

    addRouteLayers();
    const source = map.getSource('route-line') as maplibregl.GeoJSONSource | undefined;
    if (source) source.setData(result.geojson);

    const coords: [number, number][] | undefined = result.geojson?.geometry?.coordinates;
    if (coords?.length) {
      const bounds = coords.reduce(
        (b, c) => b.extend([c[0], c[1]]),
        new maplibre!.LngLatBounds(coords[0], coords[0]),
      );
      map.fitBounds(bounds, { padding: 60, maxZoom: 15, duration: 800 });
    }
  }

  function clearRouteLayer() {
    if (!map) return;
    if (map.getLayer('route-line')) map.removeLayer('route-line');
    if (map.getLayer('route-casing')) map.removeLayer('route-casing');
    if (map.getSource('route-line')) map.removeSource('route-line');
  }

  function clearIsochroneMarker() {
    marker?.remove();
    marker = undefined;
  }

  function clearRouteMarkers() {
    startMarker?.remove();
    startMarker = undefined;
    endMarker?.remove();
    endMarker = undefined;
  }

  function updateIsochroneMarker(lat: number, lng: number) {
    if (!map) return;

    if (marker) {
      marker.setLngLat([lng, lat]);
    } else {
      marker = new maplibre!.Marker({ color: '#06b6d4', draggable: true })
        .setLngLat([lng, lat])
        .addTo(map);

      marker.on('dragend', () => {
        const lngLat = marker!.getLngLat();
        setOrigin({ lat: lngLat.lat, lng: lngLat.lng });
        setAnalysisResult(null);
        setAnalysisError(null);
      });
    }
  }

  function updateRouteStartMarker(lat: number, lng: number) {
    if (!map) return;

    if (startMarker) {
      startMarker.setLngLat([lng, lat]);
    } else {
      startMarker = new maplibre!.Marker({ color: '#10b981', draggable: true })
        .setLngLat([lng, lat])
        .addTo(map);

      startMarker.on('dragend', () => {
        const lngLat = startMarker!.getLngLat();
        setRouteOrigin({ lat: lngLat.lat, lng: lngLat.lng });
        setRouteResult(null);
      });
    }
  }

  function updateRouteEndMarker(lat: number, lng: number) {
    if (!map) return;

    if (endMarker) {
      endMarker.setLngLat([lng, lat]);
    } else {
      endMarker = new maplibre!.Marker({ color: '#ef4444', draggable: true })
        .setLngLat([lng, lat])
        .addTo(map);

      endMarker.on('dragend', () => {
        const lngLat = endMarker!.getLngLat();
        setRouteDestination({ lat: lngLat.lat, lng: lngLat.lng });
        setRouteResult(null);
      });
    }
  }

  function syncRouteMarkers() {
    const origin = routeOrigin();
    const destination = routeDestination();

    if (origin) updateRouteStartMarker(origin.lat, origin.lng);
    else {
      startMarker?.remove();
      startMarker = undefined;
    }

    if (destination) updateRouteEndMarker(destination.lat, destination.lng);
    else {
      endMarker?.remove();
      endMarker = undefined;
    }
  }

  function flyToFocus(focus: { lat: number; lng: number } | null | undefined) {
    if (!map || !focus) return;
    map.flyTo({ center: [focus.lng, focus.lat], zoom: 13, duration: 1200 });
  }

  onMount(async () => {
    if (!container) return;

    try {
      // Load the maplibre runtime + its stylesheet only when the workbench
      // actually mounts, so the public landing page ships without either.
      const loaded = await import('maplibre-gl');
      await import('maplibre-gl/dist/maplibre-gl.css');
      loaded.setWorkerUrl(workerUrl);
      maplibre = loaded;

      const baseUrl = 'https://basemaps.cartocdn.com';
      const styleUrl = `${baseUrl}/gl/dark-matter-gl-style/style.json`;

      const response = await fetch(styleUrl);
      const styleData = await response.json();

      if (styleData.glyphs && styleData.glyphs.startsWith('/')) {
        styleData.glyphs = `${baseUrl}${styleData.glyphs}`;
      }

      if (styleData.sprite && styleData.sprite.startsWith('/')) {
        styleData.sprite = `${baseUrl}${styleData.sprite}`;
      }

      if (styleData.sources) {
        Object.keys(styleData.sources).forEach((sourceKey: string) => {
          const source = styleData.sources[sourceKey];

          if (source.url && source.url.startsWith('/')) {
            source.url = `${baseUrl}${source.url}`;
          }

          if (source.tiles) {
            source.tiles = source.tiles.map((tilePath: string) =>
              tilePath.startsWith('/') ? `${baseUrl}${tilePath}` : tilePath,
            );
          }
        });
      }

      map = new maplibre.Map({
        container,
        style: styleData,
        center: [-122.4194, 37.7749],
        zoom: 12,
      });

      map.addControl(new maplibre.NavigationControl(), 'bottom-right');

      const resizeObserver = new ResizeObserver(() => {
        map?.resize();
      });
      resizeObserver.observe(container);

      map.on('load', () => {
        setMapReady(true);
        addRouteLayers();
        drawRoute();
        addIsochroneLayers();
        drawIsochrones();
        syncRouteMarkers();
        flyToFocus(mapFocus());
      });

      map.on('click', (e: maplibregl.MapMouseEvent) => {
        const { lng, lat } = e.lngLat;
        handleMapClick({ lat, lng });

        if (activeTool() === 'route') {
          clearIsochroneMarker();
          syncRouteMarkers();
        } else {
          clearRouteMarkers();
          clearRouteLayer();
          updateIsochroneMarker(lat, lng);
        }
      });

      onCleanup(() => {
        resizeObserver.disconnect();
        marker?.remove();
        startMarker?.remove();
        endMarker?.remove();
        map?.remove();
      });
    } catch (error) {
      console.error('Failed to securely parse map structure layers:', error);
    }
  });

  createEffect(() => {
    if (activeTool() === 'route') {
      clearIsochroneMarker();
      clearIsochroneLayers();
      syncRouteMarkers();
      if (routeResult()?.geojson) {
        drawRoute();
      } else {
        clearRouteLayer();
      }
    } else {
      clearRouteLayer();
      if (analysisResult()?.bands?.length) {
        drawIsochrones();
      } else {
        clearIsochroneLayers();
      }
    }
  });

  // Fly to search results; also keep the isochrone marker in sync when the
  // origin is set programmatically (e.g. via the search bar).
  createEffect(() => {
    if (!mapReady()) return;
    const focus = mapFocus();
    if (focus) flyToFocus(focus);
  });

  createEffect(() => {
    if (!mapReady() || activeTool() !== 'isochrone') return;
    const o = origin();
    if (o) updateIsochroneMarker(o.lat, o.lng);
  });

  return <div ref={container} class="absolute inset-0 w-full h-full" />;
};
